package timeline

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/clock"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// Reload re-reads the timeline settings from the world's content. game.SwapContent calls it after
// the content changes.
func Reload(w *sim.World) { stateOf(w).cfg = w.Content.Timeline() }

// Install registers the Ready event, the wait command, the Actor spawner and the view builder.
// The world's content must be set. clk gives the phase boundaries that waiting skips to.
func Install(w *sim.World, clk *clock.Clock) error {
	if w.Content == nil {
		return errors.New("timeline: install: world has no content")
	}
	stateOf(w).cfg = w.Content.Timeline()
	if err := api.RegisterEvent[api.Ready](w.Codec); err != nil {
		return err
	}
	if err := api.RegisterEvent[api.ActionLanded](w.Codec); err != nil {
		return err
	}
	if err := api.RegisterEvent[api.ActionCanceled](w.Codec); err != nil {
		return err
	}
	err := sim.Handle(w, func(w *sim.World, cmd api.Wait) (int64, error) {
		return handleWait(w, clk, cmd)
	})
	if err != nil {
		return err
	}
	if err := w.RegisterSpawner(sim.Spawner{Name: "actor", Apply: spawnActor}); err != nil {
		return err
	}
	return w.Views.Register(sim.ViewBuilder{Name: "timeline", Build: func(w *sim.World, v *api.View) error {
		buildView(w, clk, v)
		return nil
	}})
}

// spawnActor puts a body with an Actor template on the timeline.
func spawnActor(w *sim.World, e ecs.Entity, t content.Template) error {
	if t.Actor == nil {
		return nil
	}
	st := stateOf(w)
	if _, ok := st.controllers[t.Actor.Controller]; !ok && t.Actor.Controller != PlayerController {
		return fmt.Errorf("unknown controller %q", t.Actor.Controller)
	}
	st.actors.Add(e, &comp.Actor{Controller: t.Actor.Controller})
	st.schedule(w, e)
	return nil
}

// currentPlayer returns the current body when it is the player-controlled body with the UID.
func currentPlayer(w *sim.World, uid api.UID) (ecs.Entity, bool) {
	e, ok := Current(w)
	if !ok || uid == 0 || w.UIDOf(e) != uid {
		return ecs.Entity{}, false
	}
	return e, stateOf(w).actors.Get(e).Controller == PlayerController
}

func reject(w *sim.World, kind string, reason api.ContentID) {
	w.Bus.Publish(api.Rejected{Command: kind, Reason: api.Text{ID: reason}})
}

// nextPhaseBoundary is the tick at which the phase after the current one starts.
func nextPhaseBoundary(w *sim.World, clk *clock.Clock) int64 {
	return w.Now - clk.TickInPhase(w.Now) + clk.PhaseTicks(clk.Phase(w.Now))
}

func handleWait(w *sim.World, clk *clock.Clock, cmd api.Wait) (int64, error) {
	body, ok := currentPlayer(w, cmd.Source)
	if !ok {
		reject(w, cmd.CommandKind(), "text.timeline.not_current")
		return 0, nil
	}
	st := stateOf(w)
	switch {
	case cmd.Until == api.WaitTicks && cmd.Ticks > 0:
		st.waitUntil(w, body, w.Now+cmd.Ticks)
	case cmd.Until == api.WaitEvent:
		st.waitUntil(w, body, st.nextOtherEntry(w, body))
	case cmd.Until == api.WaitPhase:
		st.waitUntil(w, body, nextPhaseBoundary(w, clk))
	default:
		reject(w, cmd.CommandKind(), "text.timeline.bad_wait")
	}
	return 0, nil
}

// visibleEntries lists the decisions of visible bodies from now to the end of the window.
func (st *state) visibleEntries(w *sim.World) []api.TimelineEntry {
	type decision struct {
		tick int64
		body ecs.Entity
	}
	var decisions []decision
	seen := map[ecs.Entity]bool{}
	if body, ok := Current(w); ok {
		decisions = append(decisions, decision{w.Now, body})
		seen[body] = true
	}
	for _, s := range st.line {
		if !st.live(w, s) || seen[s.e] {
			continue
		}
		seen[s.e] = true
		decisions = append(decisions, decision{s.tick, s.e})
	}
	slices.SortFunc(decisions, func(a, b decision) int {
		return cmp.Or(cmp.Compare(a.tick, b.tick), w.CompareUID(a.body, b.body))
	})
	appearances := ecs.NewMap[comp.Appearance](w.ECS)
	identities := ecs.NewMap[comp.Identity](w.ECS)
	var entries []api.TimelineEntry
	for _, d := range decisions {
		if d.tick > w.Now+st.cfg.WindowTicks || !st.visible(w, d.body) {
			continue
		}
		entry := api.TimelineEntry{Tick: d.tick, Kind: "ready", Source: w.UIDOf(d.body)}
		if appearances.Has(d.body) {
			look := appearances.Get(d.body)
			entry.Glyph, entry.Color = look.Glyph, look.Color
		}
		name := api.Text{ID: "text.template.unknown"}
		if identities.Has(d.body) {
			name.ID = api.ContentID(
				"text.template." + strings.TrimPrefix(string(identities.Get(d.body).Template), "template."),
			)
		}
		entry.Label = w.Content.Render(api.Text{
			ID:   "text.timeline.ready",
			Args: map[string]string{"name": w.Content.Render(name)},
		})
		entries = append(entries, entry)
	}
	return entries
}

// buildView fills the timeline section and, on a player's decision, the wait options.
func buildView(w *sim.World, clk *clock.Clock, v *api.View) {
	st := stateOf(w)
	entries := append(st.visibleEntries(w), st.landsEntries(w)...)
	slices.SortStableFunc(entries, func(a, b api.TimelineEntry) int {
		if a.Tick != b.Tick {
			return cmp.Compare(a.Tick, b.Tick)
		}
		// Landings come before decisions at the same tick.
		switch {
		case a.Kind == "lands" && b.Kind != "lands":
			return -1
		case a.Kind != "lands" && b.Kind == "lands":
			return 1
		}
		return 0
	})
	view := &api.TimelineView{Now: w.Now, Window: st.cfg.WindowTicks, Entries: entries}
	v.Timeline = view
	body, ok := Current(w)
	if !ok {
		return
	}
	view.Current = w.UIDOf(body)
	if _, player := currentPlayer(w, view.Current); !player {
		return
	}
	phases := w.Content.Calendar().Phases
	next := phases[(clk.Phase(w.Now)+1)%len(phases)]
	phaseName := w.Content.Render(api.Text{ID: next.Name})
	v.Options = append(v.Options,
		api.Option{
			ID:      "wait/event",
			Label:   w.Content.Render(api.Text{ID: "text.timeline.wait_event"}),
			Command: api.Wait{Source: view.Current, Until: api.WaitEvent},
			Enabled: true,
		},
		api.Option{
			ID: "wait/phase",
			Label: w.Content.Render(api.Text{
				ID:   "text.timeline.wait_phase",
				Args: map[string]string{"phase": phaseName},
			}),
			Command: api.Wait{Source: view.Current, Until: api.WaitPhase},
			Enabled: true,
		},
	)
}
