package timeline

import (
	"cmp"
	"errors"
	"fmt"
	"slices"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/clock"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// canceledGone is the reason when the actor or its part is gone by the time an action lands.
const canceledGone api.ContentID = "text.timeline.canceled.gone"

// Resolver applies one kind of pending action when it lands.
type Resolver struct {
	// Resolve applies the action. The pending entity, with its kind-specific components, still
	// exists while it runs and is removed afterwards. Whether the action hits is its business.
	Resolve func(w *sim.World, pending ecs.Entity) error
	// Label describes the action for the timeline view, e.g. "Clay-thing's Strike lands on the
	// Reaper (left arm)". Without one, the view uses text.timeline.lands.
	Label func(w *sim.World, pending ecs.Entity) string
}

// RegisterResolver makes r available to Schedule as kind. Kinds are unique.
func RegisterResolver(w *sim.World, kind string, r Resolver) error {
	if kind == "" || r.Resolve == nil {
		return errors.New("timeline: a resolver needs a kind and a Resolve function")
	}
	st := stateOf(w)
	if _, dup := st.resolvers[kind]; dup {
		return fmt.Errorf("timeline: resolver %q already registered", kind)
	}
	st.resolvers[kind] = r
	return nil
}

// Schedule creates a pending entity for p, setting its Started and Seq. The caller adds its own
// kind-specific component to the returned entity. Lands must be after now and Kind must have a
// resolver.
func Schedule(w *sim.World, p comp.Pending) (ecs.Entity, error) {
	st := stateOf(w)
	if _, ok := st.resolvers[p.Kind]; !ok {
		return ecs.Entity{}, fmt.Errorf("timeline: no resolver for kind %q", p.Kind)
	}
	if p.Lands <= w.Now {
		return ecs.Entity{}, fmt.Errorf("timeline: %q lands at tick %d, which is not after now (%d)",
			p.Kind, p.Lands, w.Now)
	}
	// Resume above the highest Seq in the world, so a load never reuses one.
	for _, e := range st.allPending(w) {
		st.seq = max(st.seq, st.pendings.Get(e).Seq)
	}
	st.seq++
	p.Seq = st.seq
	p.Started = w.Now
	return st.pendings.NewEntity(&p), nil
}

// Cancel removes a pending action without resolving it and publishes ActionCanceled.
func Cancel(w *sim.World, pending ecs.Entity, reason api.ContentID) {
	st := stateOf(w)
	if !w.ECS.Alive(pending) || !st.pendings.Has(pending) {
		return
	}
	p := *st.pendings.Get(pending)
	w.ECS.RemoveEntity(pending)
	w.Bus.Publish(api.ActionCanceled{
		Kind: p.Kind, Source: p.Actor, Target: p.Target, Tick: w.Now, Reason: api.Text{ID: reason},
	})
}

// allPending lists every pending entity, sorted by landing tick then Seq.
func (st *state) allPending(w *sim.World) []ecs.Entity {
	var out []ecs.Entity
	query := ecs.NewFilter1[comp.Pending](w.ECS).Query()
	for query.Next() {
		out = append(out, query.Entity())
	}
	slices.SortFunc(out, func(a, b ecs.Entity) int {
		pa, pb := st.pendings.Get(a), st.pendings.Get(b)
		return cmp.Or(cmp.Compare(pa.Lands, pb.Lands), cmp.Compare(pa.Seq, pb.Seq))
	})
	return out
}

// PendingOf lists the actions the actor has in flight, sorted by Lands then Seq.
func PendingOf(w *sim.World, actor api.UID) []ecs.Entity {
	st := stateOf(w)
	return slices.DeleteFunc(st.allPending(w), func(e ecs.Entity) bool { return st.pendings.Get(e).Actor != actor })
}

// AimedAt lists the actions aimed at the target, sorted by Lands then Seq.
func AimedAt(w *sim.World, target api.UID) []ecs.Entity {
	st := stateOf(w)
	return slices.DeleteFunc(st.allPending(w), func(e ecs.Entity) bool { return st.pendings.Get(e).Target != target })
}

// landingDue reports whether an action is due to land at or before the current tick.
func (st *state) landingDue(w *sim.World) bool {
	all := st.allPending(w)
	return len(all) > 0 && st.pendings.Get(all[0]).Lands <= w.Now
}

// landNext lands the earliest pending action if it is due no later than the next decision
// (landings go first at the same tick). It reports whether it did.
func (st *state) landNext(w *sim.World, clk *clock.Clock) (bool, error) {
	all := st.allPending(w)
	if len(all) == 0 {
		return false, nil
	}
	first := all[0]
	p := *st.pendings.Get(first)
	if decision, ok := st.peekTick(w); ok && decision < p.Lands {
		return false, nil
	}
	if p.Lands > w.Now {
		if err := clk.Advance(w, p.Lands-w.Now); err != nil {
			return false, err
		}
	}
	st.systemsDue = true
	if !exists(w, p.Actor) || (p.Channel != 0 && !exists(w, p.Channel)) {
		Cancel(w, first, canceledGone)
	} else {
		if err := st.resolvers[p.Kind].Resolve(w, first); err != nil {
			return false, fmt.Errorf("timeline: resolving %q: %w", p.Kind, err)
		}
		w.ECS.RemoveEntity(first)
		w.Bus.Publish(api.ActionLanded{Kind: p.Kind, Source: p.Actor, Target: p.Target, Tick: w.Now})
	}
	if _, err := w.Bus.Drain(w); err != nil {
		return false, fmt.Errorf("timeline: landing %q: %w", p.Kind, err)
	}
	return true, nil
}

func exists(w *sim.World, uid api.UID) bool {
	_, ok := w.ByUID(uid)
	return ok
}

// landsEntries lists the pending actions whose actor or target is visible.
func (st *state) landsEntries(w *sim.World) []api.TimelineEntry {
	appearances := ecs.NewMap[comp.Appearance](w.ECS)
	var entries []api.TimelineEntry
	for _, e := range st.allPending(w) {
		p := st.pendings.Get(e)
		if p.Lands > w.Now+st.cfg.WindowTicks {
			continue
		}
		actor, actorOK := w.ByUID(p.Actor)
		target, targetOK := w.ByUID(p.Target)
		if (!actorOK || !st.visible(w, actor)) && (!targetOK || !st.visible(w, target)) {
			continue
		}
		entry := api.TimelineEntry{Tick: p.Lands, Kind: "lands", Source: p.Actor, Target: p.Target}
		if actorOK && appearances.Has(actor) {
			look := appearances.Get(actor)
			entry.Glyph, entry.Color = look.Glyph, look.Color
		}
		if label := st.resolvers[p.Kind].Label; label != nil {
			entry.Label = label(w, e)
		} else {
			entry.Label = w.Content.Render(api.Text{
				ID: "text.timeline.lands", Args: map[string]string{"kind": p.Kind},
			})
		}
		entries = append(entries, entry)
	}
	return entries
}
