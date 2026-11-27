package agents

import (
	"slices"
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

func stateOf(tb testing.TB, w *sim.World, body ecs.Entity) State {
	tb.Helper()
	s, err := StateOf(w, body)
	if err != nil {
		tb.Fatal(err)
	}
	return s
}

func TestAStateIgnoresAWellTheAgentHasNotSeen(t *testing.T) {
	w := landmarkWorld(t)
	if _, err := w.Spawn(well, &comp.Pos{At: at(0, 1, 0)}); err != nil {
		t.Fatal(err)
	}
	unseeing := spawnBody(w, at(0, 0, 0))
	looker := spawnBody(w, at(0, 2, 0))
	far := spawnBody(w, at(0, 6, 0))
	ecs.NewMap[comp.Inventory](w.ECS).Add(unseeing, &comp.Inventory{Stacks: []comp.ItemStack{
		{Item: "item.water", Count: 2}, {Item: "item.water", Count: 1, BoundTo: 7},
	}})
	ecs.NewMap[comp.Drives](w.ECS).Add(unseeing, &comp.Drives{
		Levels:  map[api.ContentID]float64{"drive.water": 0, "drive.food": 30},
		Weights: map[api.ContentID]float64{"drive.water": 1, "drive.food": 1},
	})
	for _, body := range []ecs.Entity{looker, far} {
		if err := sense(w, body); err != nil {
			t.Fatal(err)
		}
	}
	cases := []struct {
		name string
		body ecs.Entity
		want State
	}{
		{"one that never saw the well", unseeing, State{
			{Kind: content.FactDriveMet, Tag: "drive.water"},
			{Kind: content.FactHas, Tag: "item.water", N: 3},
		}},
		{"one beside the well it saw", looker, State{
			{Kind: content.FactKnown, Tag: waterGiven},
			{Kind: content.FactNear, Tag: waterGiven},
		}},
		{"one far from the well it saw", far, State{{Kind: content.FactKnown, Tag: waterGiven}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := stateOf(t, w, tc.body); !slices.Equal(got, tc.want) {
				t.Errorf("state = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestApplyingAnOperatorChangesTheStateAsWritten(t *testing.T) {
	start := State{
		{Kind: content.FactHas, Tag: "item.ration", N: 1},
		{Kind: content.FactHas, Tag: "item.water", N: 2},
		{Kind: content.FactKnown, Tag: waterGiven},
	}
	drink := content.Operator{
		Pre: []content.Condition{
			{Fact: content.FactHas, Tag: "item.water", N: 2},
			{Fact: content.FactDriveMet, Tag: "drive.water", Not: true},
		},
		Eff: []content.Condition{
			{Fact: content.FactHas, Tag: "item.water", N: -2},
			{Fact: content.FactHas, Tag: "item.coal", N: 1},
			{Fact: content.FactKnown, Tag: waterGiven, Not: true},
			{Fact: content.FactDriveMet, Tag: "drive.water"},
		},
	}
	if !Holds(drink, start) {
		t.Fatalf("preconditions do not hold in %v", start)
	}
	before := slices.Clone(start)
	got := Apply(drink, start)
	want := State{
		{Kind: content.FactDriveMet, Tag: "drive.water"},
		{Kind: content.FactHas, Tag: "item.coal", N: 1},
		{Kind: content.FactHas, Tag: "item.ration", N: 1},
	}
	if !slices.Equal(got, want) {
		t.Errorf("state after = %v, want %v", got, want)
	}
	if !slices.Equal(start, before) {
		t.Errorf("Apply changed its input to %v, want %v", start, before)
	}
	if start.Key() == got.Key() {
		t.Errorf("states %v and %v share the key %q", start, got, got.Key())
	}
	if Holds(drink, got) {
		t.Errorf("preconditions hold again after the water is drunk, in %v", got)
	}
}
