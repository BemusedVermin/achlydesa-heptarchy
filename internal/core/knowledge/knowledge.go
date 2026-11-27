package knowledge

import (
	"cmp"
	"fmt"
	"slices"
	"strconv"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// compareKey orders facts by (Subject, Claim, Source).
func compareKey(a, b api.Fact) int {
	return cmp.Or(cmp.Compare(a.Subject, b.Subject), cmp.Compare(a.Claim, b.Claim), cmp.Compare(a.Origin, b.Origin))
}

// Learn records f for the knower, adding comp.Knowledge if it has none, and stamps it with the
// current tick. A fact with the same (Subject, Claim, Source) is replaced, so each source keeps one
// fact per claim. A knower outside the controlled party keeps at most the content's knowledge cap:
// past it, the fact with the lowest confidence, then the oldest tick, is forgotten, which may be f
// itself. A fact it keeps publishes [api.FactLearned], and [api.FactNoticed] when the knower is the
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
	party := anatomy.InControlledParty(w, knower)
	if !party && w.Content != nil && len(facts.Facts) > w.Content.Knowledge().Cap {
		weakest := slices.MinFunc(facts.Facts, func(a, b api.Fact) int {
			return cmp.Or(cmp.Compare(a.Confidence, b.Confidence), cmp.Compare(a.Tick, b.Tick), compareKey(a, b))
		})
		drop, _ := slices.BinarySearchFunc(facts.Facts, weakest, compareKey)
		facts.Facts = slices.Delete(facts.Facts, drop, drop+1)
		if compareKey(weakest, f) == 0 {
			return
		}
	}
	w.Bus.Publish(api.FactLearned{Knower: w.UIDOf(knower), Fact: f})
	if party {
		noticed := api.FactNoticed{Fact: f}
		if w.Content != nil {
			noticed.Shown = w.Content.RenderValue(f.Claim, f.Value)
		}
		w.Bus.Publish(noticed)
	}
}

// Forget drops the knower's fact with the key (subject, claim, origin), and reports whether it held one.
func Forget(w *sim.World, knower ecs.Entity, subject api.UID, claim, origin api.ContentID) bool {
	known := ecs.NewMap[comp.Knowledge](w.ECS)
	if !w.ECS.Alive(knower) || !known.Has(knower) {
		return false
	}
	facts := known.Get(knower)
	at, found := slices.BinarySearchFunc(
		facts.Facts,
		api.Fact{Subject: subject, Claim: claim, Origin: origin},
		compareKey,
	)
	if found {
		facts.Facts = slices.Delete(facts.Facts, at, at+1)
	}
	return found
}

// CellValue writes a cell as a fact's value: "q,r,depth".
func CellValue(c api.Cell) string {
	value := make([]byte, 0, 16)
	value = strconv.AppendInt(value, int64(c.Hex.Q), 10)
	value = strconv.AppendInt(append(value, ','), int64(c.Hex.R), 10)
	value = strconv.AppendInt(append(value, ','), int64(c.Depth), 10)
	return string(value)
}

// ParseCell reads a value [CellValue] wrote.
func ParseCell(value string) (api.Cell, error) {
	var c api.Cell
	if _, err := fmt.Sscanf(value, "%d,%d,%d", &c.Hex.Q, &c.Hex.R, &c.Depth); err != nil {
		return api.Cell{}, fmt.Errorf("knowledge: cell value %q: %w", value, err)
	}
	return c, nil
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
		cmp.Compare(a.Origin, b.Origin),
	) < 0
}

func facts(w *sim.World, knower ecs.Entity) []api.Fact {
	known := ecs.NewMap[comp.Knowledge](w.ECS)
	if !w.ECS.Alive(knower) || !known.Has(knower) {
		return nil
	}
	return known.Get(knower).Facts
}
