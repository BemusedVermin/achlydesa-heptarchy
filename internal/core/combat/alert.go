package combat

import (
	"fmt"
	"slices"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

const alertSystem = "combat.alert"

// livingMembers lists the members of a party that are neither dead nor downed, by UID.
func livingMembers(w *sim.World, party ecs.Entity) []ecs.Entity {
	return slices.DeleteFunc(anatomy.Members(w, party), func(e ecs.Entity) bool { return anatomy.OutOfFight(w, e) })
}

// hostileCandidates lists every standing body with an allegiance, by UID.
func hostileCandidates(w *sim.World) []ecs.Entity {
	var out []ecs.Entity
	query := ecs.NewFilter2[comp.Pos, comp.Allegiance](w.ECS).Query()
	for query.Next() {
		if e := query.Entity(); !anatomy.OutOfFight(w, e) {
			out = append(out, e)
		}
	}
	w.SortByUID(out)
	return out
}

// allegedBodies lists every body with an allegiance, by UID, standing or not: the caller tests
// OutOfFight only on the few it needs.
func allegedBodies(w *sim.World) []ecs.Entity {
	var out []ecs.Entity
	query := ecs.NewFilter2[comp.Pos, comp.Allegiance](w.ECS).Query()
	for query.Next() {
		out = append(out, query.Entity())
	}
	w.SortByUID(out)
	return out
}

// firstHostileSeen finds the hostile body with the lowest UID that any living member can see.
func firstHostileSeen(w *sim.World, party ecs.Entity) (ecs.Entity, bool, error) {
	positions := ecs.NewMap[comp.Pos](w.ECS)
	candidates := allegedBodies(w)
	cfg := w.Content.Combat()
	var first ecs.Entity
	for _, member := range livingMembers(w, party) {
		if !positions.Has(member) {
			continue
		}
		from := positions.Get(member).At
		if w.Map == nil || !w.Map.IsLoaded(hexmap.KeyOf(from)) {
			continue
		}
		for _, other := range candidates {
			at := positions.Get(other).At
			if hexmap.Distance(from.Hex, at.Hex) > cfg.SightRadius ||
				!anatomy.Hostile(w, member, other) || anatomy.OutOfFight(w, other) {
				continue
			}
			visible, err := w.Map.Sees(from, at, cfg.SightRadius, cfg.SightDepth)
			if err != nil {
				return ecs.Entity{}, false, fmt.Errorf("combat: sight: %w", err)
			}
			if !visible {
				continue
			}
			if first.IsZero() || w.UIDOf(other) < w.UIDOf(first) {
				first = other
			}
			// Candidates are in UID order: nothing later can beat this one.
			break
		}
	}
	return first, !first.IsZero(), nil
}

// alert raises the controlled party's alert while a hostile is in sight, and ends it once the
// calm spell has passed.
func alert(w *sim.World) error {
	party, ok := anatomy.ControlledParty(w)
	if !ok {
		return nil
	}
	calms := ecs.NewMap[comp.Calm](w.ECS)
	if !calms.Has(party) {
		calms.Add(party, &comp.Calm{})
	}
	// Fetch after Add: adding a component moves the entity, which leaves earlier pointers stale.
	state := ecs.NewMap[comp.Party](w.ECS).Get(party)
	calm := calms.Get(party)
	hostile, seen, err := firstHostileSeen(w, party)
	if err != nil {
		return err
	}
	switch {
	case seen:
		calm.Since = w.Now
		if !state.Alert {
			state.Alert = true
			w.Bus.Publish(api.AlertStarted{
				Party: w.UIDOf(party), Hostile: w.UIDOf(hostile), Name: nameOf(w, hostile), Tick: w.Now,
			})
		}
	case state.Alert && w.Now-calm.Since >= w.Content.Combat().CalmTicks:
		state.Alert = false
		w.Bus.Publish(api.AlertEnded{Party: w.UIDOf(party), Tick: w.Now})
	}
	return nil
}
