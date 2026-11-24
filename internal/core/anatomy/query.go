package anatomy

import (
	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// Parts lists the parts attached to the body, sorted by UID. A part with no PartOf is not listed.
func Parts(w *sim.World, body ecs.Entity) []ecs.Entity {
	if !w.Alive(body) {
		return nil
	}
	var parts []ecs.Entity
	query := ecs.NewFilter1[comp.PartOf](w.ECS).Query(ecs.Rel[comp.PartOf](body))
	for query.Next() {
		parts = append(parts, query.Entity())
	}
	return w.SortByUID(parts)
}

// Functioning reports whether the part is attached to a living body and has integrity left.
func Functioning(w *sim.World, part ecs.Entity) bool {
	if !w.Alive(part) {
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

// Grants totals the values of the body's functioning parts' grants of one kind.
func Grants(w *sim.World, body ecs.Entity, kind string) int { return GrantTotals(w, body)[kind] }

// GrantTotals totals the values of the body's functioning parts' grants, by kind.
func GrantTotals(w *sim.World, body ecs.Entity) map[string]int {
	parts := ecs.NewMap[comp.Part](w.ECS)
	totals := map[string]int{}
	for _, part := range Parts(w, body) {
		if !Functioning(w, part) {
			continue
		}
		for _, grant := range parts.Get(part).Grants {
			totals[grant.Kind] += grant.Value
		}
	}
	return totals
}

// Members lists the members of a party, sorted by UID. Party order is in comp.Party.Order.
func Members(w *sim.World, party ecs.Entity) []ecs.Entity {
	if !w.Alive(party) {
		return nil
	}
	var members []ecs.Entity
	query := ecs.NewFilter1[comp.MemberOf](w.ECS).Query(ecs.Rel[comp.MemberOf](party))
	for query.Next() {
		members = append(members, query.Entity())
	}
	return w.SortByUID(members)
}
