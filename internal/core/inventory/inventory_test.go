package inventory_test

import (
	"bytes"
	"slices"
	"strings"
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/clock"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
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

func TestAllElevenTypesLoadAndNameIsNeverRefused(t *testing.T) {
	reg := contenttest.Sample(t)
	if got := len(reg.TypeIDs()); got != 11 {
		t.Errorf("types = %d, want 11", got)
	}
	tests := []struct {
		held map[string]int
		want bool
	}{{map[string]int{"item.name": 1}, true}, {map[string]int{}, true}}
	for _, tc := range tests {
		got, _, err := reg.ModsRequires("type.name", content.ActionEnv{Inventory: tc.held})
		if err != nil || got != tc.want {
			t.Errorf("name with inventory %v = %v, %v; want %v", tc.held, got, err, tc.want)
		}
	}
	if got, _, err := reg.ModsRequires("type.cut", content.ActionEnv{}); err != nil || !got {
		t.Errorf("cut = %v, %v; want always allowed", got, err)
	}
	for _, warning := range reg.Warnings() {
		if strings.Contains(warning, "hostile") {
			t.Errorf("warning %q, want no hostility warnings", warning)
		}
	}
}

func TestOneSidedHostilityWarnsButLoads(t *testing.T) {
	bundle := contenttest.SampleBundle(t)
	deep := bundle.Factions["faction.wild"]
	deep.HostileTo = []api.ContentID{"faction.crawling"} // only the party's side of the feud is left one-sided
	bundle.Factions["faction.wild"] = deep
	reg, err := content.New(bundle)
	if err != nil {
		t.Fatal(err)
	}
	hostility := 0
	for _, warning := range reg.Warnings() {
		if strings.Contains(warning, "hostile") {
			hostility++
		}
	}
	if hostility != 1 {
		t.Errorf("warnings = %v, want one about hostility", reg.Warnings())
	}
}

func TestBadTypeContentIsRejected(t *testing.T) {
	bundle := contenttest.SampleBundle(t)
	bad := bundle.Types["type.press"]
	bad.Requires = "no_such_field == 1"
	bundle.Types["type.press"] = bad
	burn := bundle.Types["type.burn"]
	burn.Lists = map[string][]string{"consumes": {"item.nothing"}}
	bundle.Types["type.burn"] = burn
	if _, err := content.New(bundle); err == nil {
		t.Error("content with a bad condition and an unknown item loaded")
	}
}

func partyOf(t *testing.T, w *sim.World) ecs.Entity {
	t.Helper()
	query := ecs.NewFilter1[comp.Party](w.ECS).Query()
	defer query.Close()
	if !query.Next() {
		t.Fatal("no party")
	}
	return query.Entity()
}

func TestPartyStartsWithConsumablesAndSaveKeepsTheInventory(t *testing.T) {
	reg := contenttest.Sample(t)
	var world *sim.World
	cfg := game.Config{Seed: 1, Content: reg, Install: func(w *sim.World, _ *clock.Clock) error {
		world = w
		return nil
	}}
	g, err := game.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	party := partyOf(t, world)
	for _, item := range []api.ContentID{
		"item.coal", "item.salk_dust", "item.lake_water", "item.daylight_jar", "item.violet_blood", "item.lees",
	} {
		if got := inventory.Count(world, party, item, 0); got < 1 {
			t.Errorf("party has %d of %s, want some", got, item)
		}
	}
	inventory.Take(world, party, "item.coal", 0, 1)
	before := stacks(world, party)

	var buf bytes.Buffer
	if err := g.Save(&buf); err != nil {
		t.Fatal(err)
	}
	if _, _, err := game.Load(cfg, &buf); err != nil {
		t.Fatal(err)
	}
	if got := stacks(world, partyOf(t, world)); !slices.Equal(got, before) {
		t.Errorf("loaded inventory = %+v, want %+v", got, before)
	}
}
