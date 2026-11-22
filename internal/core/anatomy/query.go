package anatomy

import (
	"slices"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// byUID sorts entities by UID.
func byUID(w *sim.World, entities []ecs.Entity) []ecs.Entity {
	slices.SortFunc(entities, func(a, b ecs.Entity) int {
		ua, ub := w.UIDOf(a), w.UIDOf(b)
		switch {
		case ua < ub:
			return -1
		case ua > ub:
			return 1
		}
		return 0
	})
	return entities
}

// Parts lists the parts attached to the body, sorted by UID. A part with no PartOf is not listed.
func Parts(w *sim.World, body ecs.Entity) []ecs.Entity {
	if body.IsZero() || !w.ECS.Alive(body) {
		return nil
	}
	var parts []ecs.Entity
	query := ecs.NewFilter1[comp.PartOf](w.ECS).Query(ecs.Rel[comp.PartOf](body))
	for query.Next() {
		parts = append(parts, query.Entity())
	}
	return byUID(w, parts)
}

// Functioning reports whether the part is attached to a living body and has integrity left.
func Functioning(w *sim.World, part ecs.Entity) bool {
	if part.IsZero() || !w.ECS.Alive(part) {
		return false
	}
	partOf := ecs.NewMap[comp.PartOf](w.ECS)
	parts := ecs.NewMap[comp.Part](w.ECS)
	if !partOf.Has(part) || !parts.Has(part) {
		return false
	}
	body := partOf.GetRelation(part)
	return !body.IsZero() && w.ECS.Alive(body) && parts.Get(part).Integrity > 0
}

// Grants totals what the body's functioning parts grant of one kind: for "speed" the sum of the
// values, for "action" and "targeting" the number of parts that grant it.
func Grants(w *sim.World, body ecs.Entity, kind string) int {
	parts := ecs.NewMap[comp.Part](w.ECS)
	total := 0
	for _, part := range Parts(w, body) {
		if !Functioning(w, part) {
			continue
		}
		for _, grant := range parts.Get(part).Grants {
			if grant.Kind != kind {
				continue
			}
			if kind == content.GrantSpeed {
				total += grant.Value
			} else {
				total++
			}
		}
	}
	return total
}

// Members lists the members of a party, sorted by UID. Party order is in comp.Party.Order.
func Members(w *sim.World, party ecs.Entity) []ecs.Entity {
	if party.IsZero() || !w.ECS.Alive(party) {
		return nil
	}
	var members []ecs.Entity
	query := ecs.NewFilter1[comp.MemberOf](w.ECS).Query(ecs.Rel[comp.MemberOf](party))
	for query.Next() {
		members = append(members, query.Entity())
	}
	return byUID(w, members)
}
