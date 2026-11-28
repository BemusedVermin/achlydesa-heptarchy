package agents

import (
	"slices"
	"strings"
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

const (
	storeGoal   = api.ContentID("goal.test.stock")
	toStoreOp   = "operator.test.go_to_store"
	unusableOp  = "operator.test.unusable"
	stockOp     = "operator.test.stock"
	waterTag    = api.ContentID("affordance.water")
	storeTag    = api.ContentID("affordance.store")
	stockedPlan = "operator.go_to>operator.take>operator.test.go_to_store>operator.test.stock"
)

func goapContent(tb testing.TB, edit func(*content.Bundle)) *content.Registry {
	tb.Helper()
	bundle := contenttest.SampleBundle(tb)
	bundle.Operators[toStoreOp] = content.Operator{
		Pre: []content.Condition{
			{Fact: content.FactKnown, Tag: storeTag},
			{Fact: content.FactNear, Tag: storeTag, Not: true},
		},
		Eff:  []content.Condition{{Fact: content.FactNear, Tag: storeTag}},
		Cost: 2,
		Does: content.Does{Action: "action.step", Target: "go_to"},
		Kind: "travel",
	}
	bundle.Operators[unusableOp] = content.Operator{
		Pre:  []content.Condition{{Fact: content.FactKnown, Tag: "affordance.the_dead"}},
		Eff:  []content.Condition{{Fact: content.FactNear, Tag: storeTag}, {Fact: content.FactNear, Tag: waterTag}},
		Cost: 0.1, Does: content.Does{Action: "action.step", Target: "go_to"}, Kind: "travel",
	}
	bundle.Operators[stockOp] = content.Operator{
		Pre: []content.Condition{
			{Fact: content.FactHas, Tag: "drinkable", N: 1},
			{Fact: content.FactNear, Tag: storeTag},
		},
		Eff: []content.Condition{
			{Fact: content.FactHas, Tag: "drinkable", N: -1},
			{Fact: content.FactDriveMet, Tag: "drive.food"},
		},
		Cost: 1,
		Does: content.Does{Action: "action.put", Target: "put"},
		Kind: "work",
	}
	bundle.Goals[string(storeGoal)] = content.Goal{Weight: "0.5", Want: []content.Condition{
		{Fact: content.FactDriveMet, Tag: "drive.food"},
	}}
	if edit != nil {
		edit(&bundle)
	}
	return contenttest.Registry(tb, bundle)
}

func goapSnapshot(reg *content.Registry) Snapshot {
	return Snapshot{Content: reg, Beliefs: State{
		{Kind: content.FactKnown, Tag: storeTag}, {Kind: content.FactKnown, Tag: waterTag},
	}}
}

func operatorsOf(plan Plan) string {
	ids := make([]string, len(plan.Steps))
	for i, step := range plan.Steps {
		ids[i] = string(step.Operator)
	}
	return strings.Join(ids, ">")
}

func TestGOAPPlansFromBeliefsToTheGoal(t *testing.T) {
	reg := goapContent(t, nil)
	plan, ok := newGOAP().Plan(goapSnapshot(reg), storeGoal)
	if !ok || operatorsOf(plan) != stockedPlan {
		t.Errorf("plan = %q, %v; want %q", operatorsOf(plan), ok, stockedPlan)
	}
}

func TestGOAPNeverUsesAnOperatorWhosePreconditionsFail(t *testing.T) {
	reg := goapContent(t, nil)
	plan, _ := newGOAP().Plan(goapSnapshot(reg), storeGoal)
	if slices.ContainsFunc(plan.Steps, func(s Step) bool { return s.Operator == unusableOp }) {
		t.Errorf("plan %q uses an operator whose precondition does not hold", operatorsOf(plan))
	}
}

func TestGOAPGivesUpAtItsCaps(t *testing.T) {
	cases := []struct {
		name string
		edit func(*content.Bundle)
	}{
		{"node cap", func(b *content.Bundle) { b.Agents.GoapNodes = 3 }},
		{"depth cap", func(b *content.Bundle) { b.Agents.GoapDepth = 3 }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			plan, ok := newGOAP().Plan(goapSnapshot(goapContent(t, tc.edit)), storeGoal)
			if ok {
				t.Errorf("found plan %q within the %s", operatorsOf(plan), tc.name)
			}
		})
	}
}

func TestGOAPGivesTheSamePlanEveryTimeAndFromItsCache(t *testing.T) {
	reg := goapContent(t, nil)
	cached := newGOAP()
	first, _ := cached.Plan(goapSnapshot(reg), storeGoal)
	for range 5 {
		fresh, _ := newGOAP().Plan(goapSnapshot(reg), storeGoal)
		again, _ := cached.Plan(goapSnapshot(reg), storeGoal)
		if operatorsOf(fresh) != operatorsOf(first) || operatorsOf(again) != operatorsOf(first) {
			t.Fatalf("plans differ: %q, %q, %q", operatorsOf(first), operatorsOf(fresh), operatorsOf(again))
		}
	}
}

func BenchmarkGOAPPlan(b *testing.B) {
	snapshot := goapSnapshot(goapContent(b, nil))
	for b.Loop() {
		searchGOAP(snapshot, storeGoal)
	}
}
