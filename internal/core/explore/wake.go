package explore

import (
	"fmt"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

// SetRestPlace records where the party last rested, which is where it wakes after a wipe. The
// party starts with the cell it began on, until towns and resting exist to move it.
func SetRestPlace(w *sim.World, party ecs.Entity, id api.MapID, at api.Cell) {
	sim.Put(w, party, comp.RestPlace{Map: id, At: at})
}

// CanWake reports whether the controlled party has a rest place and a downed member to get up. A
// party of the dead only has nothing to wake.
func CanWake(w *sim.World) bool {
	party, ok := controlledParty(w)
	if !ok || !ecs.NewMap[comp.RestPlace](w.ECS).Has(party) {
		return false
	}
	downed := ecs.NewMap[comp.Downed](w.ECS)
	for _, member := range membersOf(w, party) {
		if downed.Has(member) {
			return true
		}
	}
	return false
}

// Wake gets the controlled party's downed members up at its rest place: their parts mended to full
// integrity and back on the timeline under their template's controller. Every member who is not
// dead is set down around the rest cell; the dead stay where they fell. The caller advances the
// clock first, so the members decide from the time they wake. It does nothing unless [CanWake].
func Wake(w *sim.World) error {
	if !CanWake(w) {
		return nil
	}
	party, _ := controlledParty(w)
	place := *ecs.NewMap[comp.RestPlace](w.ECS).Get(party)
	m := w.Maps.Get(place.Map)
	if m == nil {
		return fmt.Errorf("explore: wake: rest place is on unknown map %d", place.Map)
	}
	dead := ecs.NewMap[comp.Dead](w.ECS)
	downed := ecs.NewMap[comp.Downed](w.ECS)
	identities := ecs.NewMap[comp.Identity](w.ECS)
	parts := ecs.NewMap[comp.Part](w.ECS)
	actors := ecs.NewMap[comp.Actor](w.ECS)
	taken := map[api.Cell]bool{}
	occupied := ecs.NewFilter2[comp.Pos, comp.Actor](w.ECS).Query()
	for occupied.Next() {
		if pos, _ := occupied.Get(); pos.Map == place.Map {
			taken[pos.At] = true
		}
	}
	var woke []ecs.Entity
	for _, member := range membersOf(w, party) {
		if dead.Has(member) {
			continue
		}
		at, err := nearestFree(m, place.At, stateOf(w).rules, taken)
		if err != nil {
			return fmt.Errorf("explore: wake: %w", err)
		}
		taken[at] = true
		pos := comp.Pos{Map: place.Map, At: at}
		sim.Put(w, member, pos)
		sim.Put(w, member, comp.Trail{Last: at})
		if !downed.Has(member) {
			continue
		}
		for _, part := range anatomy.Parts(w, member) {
			parts.Get(part).Integrity = parts.Get(part).MaxIntegrity
		}
		downed.Remove(member)
		template, ok := w.Content.Template(identities.Get(member).Template)
		if !ok || template.Actor == nil {
			return fmt.Errorf("explore: wake: member %d has no controller in its template", w.UIDOf(member))
		}
		actors.Add(member, &comp.Actor{Controller: template.Actor.Controller})
		woke = append(woke, member)
	}
	leadFromStanding(w, party)
	applyControllers(w, party)
	for _, member := range woke {
		timeline.Reschedule(w, member)
	}
	if err := markSeenByParty(w, party); err != nil {
		return err
	}
	w.Bus.Publish(api.PartyWoke{Party: w.UIDOf(party), Map: place.Map, At: place.At, Tick: w.Now})
	return nil
}

// leadFromStanding makes the first member in party order who is not dead the leader, when the
// leader is dead, so that the party wakes with somebody deciding for it.
func leadFromStanding(w *sim.World, party ecs.Entity) {
	state := ecs.NewMap[comp.Party](w.ECS).Get(party)
	dead := ecs.NewMap[comp.Dead](w.ECS)
	if leader, ok := w.ByUID(state.Leader); ok && !dead.Has(leader) {
		return
	}
	for _, uid := range state.Order {
		if next, ok := w.ByUID(uid); ok && !dead.Has(next) {
			state.Leader = uid
			return
		}
	}
}

// markSeenByParty records what the woken party sees where it stands.
func markSeenByParty(w *sim.World, party ecs.Entity) error {
	sight, err := partySight(w, party)
	if err != nil {
		return err
	}
	for id, set := range sight {
		cells := make([]api.Cell, 0, len(set))
		for cell := range set {
			cells = append(cells, cell)
		}
		markSeen(w, party, id, cells)
	}
	return nil
}
