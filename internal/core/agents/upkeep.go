package agents

import (
	"cmp"
	"slices"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/inventory"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// Systems that keep recruits in the party.
const (
	upkeepSystem = "agents.upkeep"
	flightSystem = "agents.flight"
)

// recruitsOf lists the living recruits of each party, parties in UID order and recruits in UID
// order within each.
func recruitsOf(w *sim.World) []partyRecruits {
	var parties []ecs.Entity
	query := ecs.NewFilter1[comp.Party](w.ECS).Query()
	for query.Next() {
		parties = append(parties, query.Entity())
	}
	dead := ecs.NewMap[comp.Dead](w.ECS)
	var out []partyRecruits
	for _, party := range w.SortByUID(parties) {
		entry := partyRecruits{party: party}
		for _, member := range anatomy.Members(w, party) {
			def, ok := w.Content.Template(w.TemplateOf(member))
			if ok && def.Recruit != nil && !dead.Has(member) {
				entry.recruits = append(entry.recruits, member)
			}
		}
		slices.SortFunc(entry.recruits, func(a, b ecs.Entity) int { return cmp.Compare(w.UIDOf(a), w.UIDOf(b)) })
		out = append(out, entry)
	}
	return out
}

type partyRecruits struct {
	party    ecs.Entity
	recruits []ecs.Entity
}

// tickUpkeep is the day's upkeep: each recruit takes what its template wants from the party's
// inventory when all of it is there. An unpaid day raises its discontent, a paid one lowers it.
func tickUpkeep(w *sim.World) error {
	cfg := w.Content.Recruits()
	discontent := ecs.NewMap[comp.Discontent](w.ECS)
	for _, group := range recruitsOf(w) {
		for _, recruit := range group.recruits {
			def, _ := w.Content.Template(w.TemplateOf(recruit))
			if !discontent.Has(recruit) {
				discontent.Add(recruit, &comp.Discontent{})
			}
			held := discontent.Get(recruit)
			if canPay(w, group.party, def.Recruit.Upkeep) {
				for _, want := range def.Recruit.Upkeep {
					inventory.Take(w, group.party, want.Item, 0, want.Count)
				}
				held.Days = max(held.Days-1, 0)
				continue
			}
			held.Days++
			partyUID, recruitUID := w.UIDOf(group.party), w.UIDOf(recruit)
			switch {
			case held.Days >= cfg.DesertAt && recruitUID != leaderOf(w, group.party):
				desert(w, group.party, recruit)
				w.Bus.Publish(api.Deserted{Party: partyUID, Recruit: recruitUID})
			case held.Days >= cfg.GrumbleAt:
				w.Bus.Publish(api.Grumbled{Party: partyUID, Recruit: recruitUID})
			}
		}
	}
	return nil
}

func canPay(w *sim.World, party ecs.Entity, upkeep []content.StackDef) bool {
	for _, want := range upkeep {
		if inventory.Count(w, party, want.Item, 0) < want.Count {
			return false
		}
	}
	return true
}

func leaderOf(w *sim.World, party ecs.Entity) api.UID {
	return ecs.NewMap[comp.Party](w.ECS).Get(party).Leader
}

// desert takes the recruit out of the party. It keeps its body, grafts and inventory, and acts on
// its own behaviors from now on.
func desert(w *sim.World, party, recruit ecs.Entity) {
	cfg := w.Content.Recruits()
	state := ecs.NewMap[comp.Party](w.ECS).Get(party)
	uid := w.UIDOf(recruit)
	state.Order = slices.DeleteFunc(state.Order, func(member api.UID) bool { return member == uid })
	ecs.NewMap[comp.MemberOf](w.ECS).Remove(recruit)
	ecs.NewMap[comp.Discontent](w.ECS).Remove(recruit)
	if fleeing := ecs.NewMap[comp.Fleeing](w.ECS); fleeing.Has(recruit) {
		fleeing.Remove(recruit)
	}
	sim.Put(w, recruit, comp.Allegiance{Faction: cfg.DesertersFaction})
	if individual := ecs.NewMap[comp.Individual](w.ECS); !individual.Has(recruit) {
		individual.Add(recruit, &comp.Individual{})
	}
	if actor := ecs.NewMap[comp.Actor](w.ECS); actor.Has(recruit) {
		actor.Get(recruit).Controller = controllerName
	}
}

// tickFlight starts the flight of each recruit whose terror has reached recruits.flee_at, and ends
// it when the terror falls below. The explore package gives a fleeing recruit its controller.
func tickFlight(w *sim.World) error {
	cfg := w.Content.Recruits()
	fleeing := ecs.NewMap[comp.Fleeing](w.ECS)
	needs := ecs.NewMap[comp.Needs](w.ECS)
	for _, group := range recruitsOf(w) {
		for _, recruit := range group.recruits {
			terror := 0.0
			if needs.Has(recruit) {
				if at, ok := find(needs.Get(recruit), cfg.TerrorNeed); ok {
					terror = needs.Get(recruit).Levels[at].Level
				}
			}
			switch {
			case terror >= cfg.FleeAt && !fleeing.Has(recruit):
				fleeing.Add(recruit, &comp.Fleeing{Controller: controllerName})
				w.Bus.Publish(api.Fled{Recruit: w.UIDOf(recruit)})
			case terror < cfg.FleeAt && fleeing.Has(recruit):
				fleeing.Remove(recruit)
			}
		}
	}
	return nil
}
