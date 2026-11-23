package anatomy

import (
	"slices"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// ControlledParty returns the party the player commands.
func ControlledParty(w *sim.World) (ecs.Entity, bool) {
	query := ecs.NewFilter1[comp.Controlled](w.ECS).Query()
	for query.Next() {
		party := query.Entity()
		query.Close()
		return party, true
	}
	return ecs.Entity{}, false
}

// PartyOf returns the party a body is a member of.
func PartyOf(w *sim.World, body ecs.Entity) (ecs.Entity, bool) {
	if !w.Alive(body) {
		return ecs.Entity{}, false
	}
	members := ecs.NewMap[comp.MemberOf](w.ECS)
	if !members.Has(body) {
		return ecs.Entity{}, false
	}
	party := members.GetRelation(body)
	return party, !party.IsZero()
}

// InControlledParty reports whether e is the controlled party or one of its members.
func InControlledParty(w *sim.World, e ecs.Entity) bool {
	if !w.Alive(e) {
		return false
	}
	if ecs.NewMap[comp.Controlled](w.ECS).Has(e) {
		return true
	}
	party, ok := PartyOf(w, e)
	return ok && ecs.NewMap[comp.Controlled](w.ECS).Has(party)
}

// OutOfFight reports whether a body is dead or downed, and so cannot act or lead.
func OutOfFight(w *sim.World, body ecs.Entity) bool {
	return ecs.NewMap[comp.Dead](w.ECS).Has(body) || ecs.NewMap[comp.Downed](w.ECS).Has(body)
}

// SetLeader makes the member with the UID the party's leader: it moves to the front of the party
// order, and api.LeaderChanged is published. The caller checks that the member may lead.
func SetLeader(w *sim.World, party ecs.Entity, leader api.UID) {
	state := ecs.NewMap[comp.Party](w.ECS).Get(party)
	order := append([]api.UID{leader}, slices.DeleteFunc(slices.Clone(state.Order), func(uid api.UID) bool {
		return uid == leader
	})...)
	state.Order, state.Leader = order, leader
	w.Bus.Publish(api.LeaderChanged{Party: w.UIDOf(party), Leader: leader, Name: nameOf(w, leader)})
}

// nameOf renders the name of the body with the UID from its template.
func nameOf(w *sim.World, uid api.UID) string {
	e, _ := w.ByUID(uid)
	return w.Content.Render(w.Content.TemplateName(w.TemplateOf(e)))
}
