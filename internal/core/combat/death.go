package combat

import (
	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

// actorGoneReason is the text ID of why an out-of-fight body's pending actions are canceled.
const actorGoneReason api.ContentID = "text.combat.actor_gone"

// onPartDamaged and onPartSevered check the body the part belongs to (or belonged to).
func onPartDamaged(w *sim.World, ev api.PartDamaged) { checkBody(w, ev.Body) }
func onPartSevered(w *sim.World, ev api.PartSevered) { checkBody(w, ev.Body) }

// planHasVital reports whether the body's plan includes a vital part.
func planHasVital(w *sim.World, body ecs.Entity) bool {
	bodies := ecs.NewMap[comp.Body](w.ECS)
	if !bodies.Has(body) || bodies.Get(body).Plan == "" {
		return false
	}
	plan, ok := w.Content.Body(bodies.Get(body).Plan)
	if !ok {
		return false
	}
	for _, entry := range plan.Parts {
		if def, ok := w.Content.Part(entry.Part); ok && def.Vital {
			return true
		}
	}
	return false
}

// shouldBeOut applies the death rule: a body with vital parts is out when none is left, and one with
// none is out when it has no parts left at all.
func shouldBeOut(w *sim.World, body ecs.Entity) bool {
	parts := ecs.NewMap[comp.Part](w.ECS)
	attached := anatomy.Parts(w, body)
	hadVital := planHasVital(w, body)
	for _, part := range attached {
		if parts.Get(part).Vital {
			hadVital = true
		}
	}
	if !hadVital {
		return len(attached) == 0
	}
	for _, part := range attached {
		if parts.Get(part).Vital && anatomy.Functioning(w, part) {
			return false
		}
	}
	return true
}

// checkBody kills or downs the body when the death rule says so.
func checkBody(w *sim.World, uid api.UID) {
	body, ok := w.ByUID(uid)
	if !ok || !ecs.NewMap[comp.Body](w.ECS).Has(body) || outOfFight(w, body) || !shouldBeOut(w, body) {
		return
	}
	leaveTimeline(w, body)
	if ecs.NewMap[comp.Protected](w.ECS).Has(body) {
		ecs.NewMap[comp.Downed](w.ECS).Add(body, &comp.Downed{})
		w.Bus.Publish(api.BodyDowned{Body: uid, Name: nameOf(w, body), Tick: w.Now})
	} else {
		ecs.NewMap[comp.Dead](w.ECS).Add(body, &comp.Dead{Tick: w.Now})
		makeCorpse(w, body)
		w.Bus.Publish(api.BodyDied{Body: uid, Name: nameOf(w, body), Tick: w.Now})
	}
	handOverLead(w, body)
	checkPartyWiped(w, body)
}

// handOverLead gives the lead to the first member in party order still standing when the leader
// is out of the fight, so that somebody is always deciding for the party in follow mode.
func handOverLead(w *sim.World, body ecs.Entity) {
	memberOf := ecs.NewMap[comp.MemberOf](w.ECS)
	if !memberOf.Has(body) {
		return
	}
	party := memberOf.GetRelation(body)
	state := ecs.NewMap[comp.Party](w.ECS).Get(party)
	if state.Leader != w.UIDOf(body) {
		return
	}
	for _, uid := range state.Order {
		next, ok := w.ByUID(uid)
		if !ok || outOfFight(w, next) {
			continue
		}
		order := []api.UID{uid}
		for _, other := range state.Order {
			if other != uid {
				order = append(order, other)
			}
		}
		state.Order, state.Leader = order, uid
		w.Bus.Publish(api.LeaderChanged{Party: w.UIDOf(party), Leader: uid, Name: nameOf(w, next)})
		return
	}
}

// leaveTimeline cancels the body's pending actions and takes it off the timeline.
func leaveTimeline(w *sim.World, body ecs.Entity) {
	for _, pending := range timeline.PendingOf(w, w.UIDOf(body)) {
		timeline.Cancel(w, pending, actorGoneReason)
	}
	actors := ecs.NewMap[comp.Actor](w.ECS)
	if actors.Has(body) {
		actors.Remove(body)
	}
}

// makeCorpse draws the body as a corpse.
func makeCorpse(w *sim.World, body ecs.Entity) {
	cfg := w.Content.Combat()
	sim.Put(w, body, comp.Appearance{Glyph: cfg.CorpseGlyph, Color: cfg.CorpseColor})
}

// checkPartyWiped publishes PartyWiped when the body belonged to the controlled party and every
// member of it is now dead or downed.
func checkPartyWiped(w *sim.World, body ecs.Entity) {
	memberOf := ecs.NewMap[comp.MemberOf](w.ECS)
	if !memberOf.Has(body) {
		return
	}
	party := memberOf.GetRelation(body)
	if !ecs.NewMap[comp.Controlled](w.ECS).Has(party) || len(livingMembers(w, party)) > 0 {
		return
	}
	w.Bus.Publish(api.PartyWiped{Party: w.UIDOf(party), Tick: w.Now})
}

// PartyWiped reports whether the controlled party has members and all of them are dead or downed.
func PartyWiped(w *sim.World) bool {
	party, ok := controlledParty(w)
	return ok && len(anatomy.Members(w, party)) > 0 && len(livingMembers(w, party)) == 0
}
