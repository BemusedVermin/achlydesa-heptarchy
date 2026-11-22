package turns

import (
	"errors"
	"fmt"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/clock"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// Install registers the turn events, the end and wait commands, the Turn spawner and the view
// builder. The world's content must be set. clk gives the length of the phase that wait skips.
func Install(w *sim.World, clk *clock.Clock) error {
	if w.Content == nil {
		return errors.New("turns: install: world has no content")
	}
	stateOf(w).cfg = w.Content.Turns()
	if err := api.RegisterEvent[api.TurnStarted](w.Codec); err != nil {
		return err
	}
	if err := api.RegisterEvent[api.TurnEnded](w.Codec); err != nil {
		return err
	}
	if err := sim.Handle(w, handleEnd); err != nil {
		return err
	}
	if err := sim.Handle(w, handleWait); err != nil {
		return err
	}
	if err := w.RegisterSpawner(sim.Spawner{Name: "turn", Apply: spawnTurn}); err != nil {
		return err
	}
	return w.Views.Register(sim.ViewBuilder{Name: "turns", Build: func(w *sim.World, v *api.View) error {
		buildView(w, clk, v)
		return nil
	}})
}

// spawnTurn puts an entity with a Turn template on the timeline.
func spawnTurn(w *sim.World, e ecs.Entity, t content.Template) error {
	if t.Turn == nil {
		return nil
	}
	st := stateOf(w)
	if _, ok := st.controllers[t.Turn.Controller]; !ok && t.Turn.Controller != PlayerController {
		return fmt.Errorf("unknown controller %q", t.Turn.Controller)
	}
	ticks, err := st.turnTicks(w, e)
	if err != nil {
		return err
	}
	st.turns.Add(e, &comp.Turn{Next: w.Now + ticks, Controller: t.Turn.Controller})
	st.schedule(w, e)
	return nil
}

// currentPlayer returns the current entity when it is the player-controlled actor with the UID.
func currentPlayer(w *sim.World, uid api.UID) (ecs.Entity, bool) {
	e, ok := Current(w)
	if !ok || uid == 0 || w.UIDOf(e) != uid {
		return ecs.Entity{}, false
	}
	return e, stateOf(w).turns.Get(e).Controller == PlayerController
}

func reject(w *sim.World, kind string, reason api.ContentID) {
	w.Bus.Publish(api.Rejected{Command: kind, Reason: api.Text{ID: reason}})
}

func handleEnd(w *sim.World, cmd api.End) (int64, error) {
	actor, ok := currentPlayer(w, cmd.Actor)
	if !ok {
		reject(w, cmd.CommandKind(), "text.turns.not_current")
		return 0, nil
	}
	return 0, EndTurn(w, actor)
}

func handleWait(w *sim.World, cmd api.Wait) (int64, error) {
	actor, ok := currentPlayer(w, cmd.Actor)
	switch {
	case !ok:
		reject(w, cmd.CommandKind(), "text.turns.not_current")
		return 0, nil
	case cmd.Ticks <= 0:
		reject(w, cmd.CommandKind(), "text.turns.bad_wait")
		return 0, nil
	}
	AddExtra(w, actor, cmd.Ticks)
	return 0, EndTurn(w, actor)
}

// buildView fills the timeline section and, on a player's turn, the end and wait options.
func buildView(w *sim.World, clk *clock.Clock, v *api.View) {
	view := &api.TurnsView{Upcoming: stateOf(w).upcoming(w, api.UpcomingTurnsMax)}
	v.Turns = view
	e, ok := Current(w)
	if !ok {
		return
	}
	view.Current = w.UIDOf(e)
	if _, player := currentPlayer(w, view.Current); !player {
		return
	}
	untilPhaseEnds := clk.PhaseTicks(clk.Phase(w.Now)) - clk.TickInPhase(w.Now)
	v.Options = append(v.Options,
		api.Option{
			ID:      "end",
			Label:   w.Content.Render(api.Text{ID: "text.turns.end"}),
			Command: api.End{Actor: view.Current},
			Enabled: true,
		},
		api.Option{
			ID:      "wait",
			Label:   w.Content.Render(api.Text{ID: "text.turns.wait"}),
			Command: api.Wait{Actor: view.Current, Ticks: untilPhaseEnds},
			Enabled: true,
		},
	)
}
