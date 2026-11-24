package inventory_test

import (
	"slices"
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game/gametest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/inventory"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

func newHolder(w *sim.World) ecs.Entity {
	return ecs.NewMap1[comp.Identity](w.ECS).NewEntity(&comp.Identity{UID: w.NewUID()})
}

func stacks(w *sim.World, holder ecs.Entity) []comp.ItemStack {
	return slices.Clone(ecs.NewMap[comp.Inventory](w.ECS).Get(holder).Stacks)
}

func TestStacksMergeAndStaySorted(t *testing.T) {
	w := sim.NewWorld(1)
	holder := newHolder(w)
	for _, s := range []comp.ItemStack{
		{Item: "item.lees", Count: 1},
		{Item: "item.coal", Count: 2},
		{Item: "item.keepsake", Count: 1, BoundTo: 9},
		{Item: "item.coal", Count: 3},
		{Item: "item.keepsake", Count: 1, BoundTo: 4},
		{Item: "item.coal", Count: 0},
	} {
		inventory.Add(w, holder, s)
	}
	want := []comp.ItemStack{
		{Item: "item.coal", Count: 5},
		{Item: "item.keepsake", Count: 1, BoundTo: 4},
		{Item: "item.keepsake", Count: 1, BoundTo: 9},
		{Item: "item.lees", Count: 1},
	}
	if got := stacks(w, holder); !slices.Equal(got, want) {
		t.Errorf("stacks = %+v, want %+v", got, want)
	}
}

func TestTake(t *testing.T) {
	w := sim.NewWorld(1)
	holder := newHolder(w)
	inventory.Add(w, holder, comp.ItemStack{Item: "item.coal", Count: 3})
	inventory.Add(w, holder, comp.ItemStack{Item: "item.keepsake", Count: 1, BoundTo: 4})

	tests := []struct {
		name    string
		item    api.ContentID
		boundTo api.UID
		n       int
		want    bool
		left    int
	}{
		{"more than held fails", "item.coal", 0, 4, false, 3},
		{"exactly held succeeds and empties the stack", "item.keepsake", 4, 1, true, 0},
		{"wrong owner fails", "item.keepsake", 5, 1, false, 0},
		{"some succeeds", "item.coal", 0, 2, true, 1},
		{"missing item fails", "item.lees", 0, 1, false, 0},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := inventory.Take(w, holder, tc.item, tc.boundTo, tc.n); got != tc.want {
				t.Errorf("Take = %v, want %v", got, tc.want)
			}
			if got := inventory.Count(w, holder, tc.item, tc.boundTo); got != tc.left {
				t.Errorf("Count = %d, want %d", got, tc.left)
			}
		})
	}
	if got := stacks(w, holder); len(got) != 1 || got[0].Count != 1 {
		t.Errorf("stacks = %+v, want only one coal", got)
	}
}

func TestCountOfAnEntityWithNoInventoryIsZero(t *testing.T) {
	w := sim.NewWorld(1)
	holder := newHolder(w)
	if got := inventory.Count(w, holder, "item.coal", 0); got != 0 {
		t.Errorf("Count = %d, want 0", got)
	}
	if inventory.Take(w, holder, "item.coal", 0, 1) {
		t.Error("Take from no inventory succeeded")
	}
}

func TestSpawningGivesTheTemplatesInventory(t *testing.T) {
	reg := contenttest.Sample(t)
	_, world, _ := gametest.Start(t, game.Config{Seed: 1, Content: reg})
	query := ecs.NewFilter1[comp.Party](world.ECS).Query()
	if !query.Next() {
		t.Fatal("the start scenario has no party")
	}
	party := query.Entity()
	query.Close()
	id := ecs.NewMap[comp.Identity](world.ECS).Get(party).Template
	template, ok := reg.Template(id)
	if !ok {
		t.Fatalf("unknown template %q", id)
	}
	if len(template.Inventory) == 0 {
		t.Fatal("the sample party starts with nothing, so this test checks nothing")
	}
	for _, def := range template.Inventory {
		if got := inventory.Count(world, party, def.Item, 0); got < def.Count {
			t.Errorf("party has %d of %s, want at least %d", got, def.Item, def.Count)
		}
	}
}
