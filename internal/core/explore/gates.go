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

const crossedReason api.ContentID = "text.explore.crossed"

// crossGate moves the whole party through the gate the leader stands on, if it does. The leader
// lands on the destination mark and the other living members on the nearest free cells.
func crossGate(w *sim.World, body ecs.Entity) error {
	positions := ecs.NewMap[comp.Pos](w.ECS)
	if !w.Alive(body) || !positions.Has(body) {
		return nil
	}
	party, ok := leaderParty(w, body)
	if !ok {
		return nil
	}
	here := *positions.Get(body)
	m := w.Maps.Get(here.Map)
	if m == nil {
		return nil
	}
	gate, ok := m.GateAt(here.At)
	if !ok {
		return nil
	}
	dest := w.Maps.Get(gate.To)
	if dest == nil {
		return fmt.Errorf("explore: gate leads to unknown map %d", gate.To)
	}
	dead := ecs.NewMap[comp.Dead](w.ECS)
	taken := map[api.Cell]bool{}
	occupied := ecs.NewFilter2[comp.Pos, comp.Actor](w.ECS).Query()
	for occupied.Next() {
		if pos, _ := occupied.Get(); pos.Map == gate.To {
			taken[pos.At] = true
		}
	}
	rules := stateOf(w).rules
	for _, member := range membersOf(w, party) {
		if dead.Has(member) {
			continue
		}
		at := gate.ToCell
		if member != body {
			var err error
			if at, err = nearestFree(dest, gate.ToCell, rules, taken); err != nil {
				return fmt.Errorf("explore: crossing: %w", err)
			}
		}
		taken[at] = true
		for _, pending := range timeline.PendingOf(w, w.UIDOf(member)) {
			timeline.Cancel(w, pending, crossedReason)
		}
		sim.Put(w, member, comp.Pos{Map: gate.To, At: at})
		sim.Put(w, member, comp.Trail{Last: at})
	}
	if err := markSeenByParty(w, party); err != nil {
		return err
	}
	w.Bus.Publish(api.PartyCrossed{Party: w.UIDOf(party), From: here.Map, To: gate.To})
	return nil
}

// leaderParty returns the party the body leads.
func leaderParty(w *sim.World, body ecs.Entity) (ecs.Entity, bool) {
	party, ok := anatomy.PartyOf(w, body)
	if !ok {
		return ecs.Entity{}, false
	}
	state := ecs.NewMap[comp.Party](w.ECS)
	if !state.Has(party) || state.Get(party).Leader != w.UIDOf(body) {
		return ecs.Entity{}, false
	}
	return party, true
}
