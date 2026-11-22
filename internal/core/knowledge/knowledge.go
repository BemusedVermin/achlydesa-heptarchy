package knowledge

import (
	"cmp"
	"slices"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// compareKey orders facts by (Subject, Claim, Source).
func compareKey(a, b api.Fact) int {
	return cmp.Or(cmp.Compare(a.Subject, b.Subject), cmp.Compare(a.Claim, b.Claim), cmp.Compare(a.Source, b.Source))
}

// Learn records f for the knower, adding comp.Knowledge if it has none, and stamps it with the
// current tick. A fact with the same (Subject, Claim, Source) is replaced, so each source keeps one
// fact per claim. It publishes [api.FactLearned], and [api.FactNoticed] when the knower is the
// controlled party or one of its members.
func Learn(w *sim.World, knower ecs.Entity, f api.Fact) {
	f.Tick = w.Now
	known := ecs.NewMap[comp.Knowledge](w.ECS)
	if !known.Has(knower) {
		known.Add(knower, &comp.Knowledge{})
	}
	facts := known.Get(knower)
	at, found := slices.BinarySearchFunc(facts.Facts, f, compareKey)
	if found {
		facts.Facts[at] = f
	} else {
		facts.Facts = slices.Insert(facts.Facts, at, f)
	}
	w.Bus.Publish(api.FactLearned{Knower: w.UIDOf(knower), Fact: f})
	if controlled(w, knower) {
		noticed := api.FactNoticed{Fact: f}
		if w.Content != nil {
			noticed.Shown = w.Content.RenderValue(f.Claim, f.Value)
		}
		w.Bus.Publish(noticed)
	}
}

// About returns what the knower holds about the subject, sorted by (Claim, Source). It is a copy.
func About(w *sim.World, knower ecs.Entity, subject api.UID) []api.Fact {
	var out []api.Fact
	for _, f := range facts(w, knower) {
		if f.Subject == subject {
			out = append(out, f)
		}
	}
	return out
}

// Best returns the knower's fact about the subject's claim with the highest confidence, then the
// newest tick, then the smallest Source ID.
func Best(w *sim.World, knower ecs.Entity, subject api.UID, claim api.ContentID) (api.Fact, bool) {
	var best api.Fact
	found := false
	for _, f := range facts(w, knower) {
		if f.Subject != subject || f.Claim != claim {
			continue
		}
		if !found || better(f, best) {
			best, found = f, true
		}
	}
	return best, found
}

// better says whether a beats b: higher confidence, then newer, then the smaller source ID.
func better(a, b api.Fact) bool {
	return cmp.Or(
		cmp.Compare(b.Confidence, a.Confidence),
		cmp.Compare(b.Tick, a.Tick),
		cmp.Compare(a.Source, b.Source),
	) < 0
}

func facts(w *sim.World, knower ecs.Entity) []api.Fact {
	known := ecs.NewMap[comp.Knowledge](w.ECS)
	if !w.ECS.Alive(knower) || !known.Has(knower) {
		return nil
	}
	return known.Get(knower).Facts
}

// controlled says whether the knower is the controlled party or one of its members.
func controlled(w *sim.World, knower ecs.Entity) bool {
	if !w.ECS.Alive(knower) {
		return false
	}
	marks := ecs.NewMap[comp.Controlled](w.ECS)
	if marks.Has(knower) {
		return true
	}
	if members := ecs.NewMap[comp.MemberOf](w.ECS); members.Has(knower) {
		return marks.Has(members.GetRelation(knower))
	}
	return false
}
