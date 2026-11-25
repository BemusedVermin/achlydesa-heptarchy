package agents_test

import (
	"slices"
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/agents"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game/gametest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/inventory"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// mapNamed is the ID of the map with the content ID.
func mapNamed(tb testing.TB, w *sim.World, name string) api.MapID {
	tb.Helper()
	for _, id := range w.Maps.IDs() {
		if w.Maps.Get(id).Name() == name {
			return id
		}
	}
	tb.Fatalf("no map %q", name)
	return 0
}

// regionMap is the map of the region entity.
func regionMap(w *sim.World, region ecs.Entity) api.MapID {
	return ecs.NewMap[comp.Region](w.ECS).Get(region).Map
}

// hallTown makes the test hall a town with three lake water in stock, filled at the given marks.
func hallTown(marks ...string) func(*content.Bundle) {
	return func(bundle *content.Bundle) {
		contenttest.Edit(bundle.Maps, hallMap, func(hall *content.MapDef) {
			hall.Region = &content.RegionDef{Town: &content.TownDef{
				Stock:      []content.StackDef{{Item: "item.lake_water", Count: 3}},
				StockMarks: marks,
			}}
		})
	}
}

func TestThePlaygroundHasARegionPerMapJoinedByItsGates(t *testing.T) {
	_, w, _ := gametest.New(t, 1, func(bundle *content.Bundle) { bundle.Start.Map = "map.play.hub" })
	count := 0
	query := ecs.NewFilter1[comp.Region](w.ECS).Query()
	for query.Next() {
		count++
	}
	if want := len(w.Maps.IDs()); count != want {
		t.Errorf("%d regions, want one per map (%d)", count, want)
	}
	hub := mapNamed(t, w, "map.play.hub")
	want := []api.MapID{
		mapNamed(t, w, "map.play.den"), mapNamed(t, w, "map.play.descent"), mapNamed(t, w, "map.play.hall"),
		mapNamed(t, w, "map.test.town"), mapNamed(t, w, "map.play.feeders"), mapNamed(t, w, "map.play.shrine"),
	}
	slices.Sort(want)
	tests := []struct {
		name string
		from api.MapID
		want []api.MapID
	}{
		{"the hub", hub, want},
		{"the hall", mapNamed(t, w, "map.play.hall"), []api.MapID{hub}},
		{"a test map with no gates", mapNamed(t, w, "map.test.hall"), []api.MapID{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			region, ok := agents.RegionOf(w, tt.from)
			if !ok {
				t.Fatalf("map %d has no region", tt.from)
			}
			got := []api.MapID{}
			for _, next := range agents.Neighbours(w, region) {
				got = append(got, regionMap(w, next))
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("neighbours %v, want %v", got, tt.want)
			}
		})
	}
	town, ok := agents.TownOf(w, hub)
	if !ok {
		t.Fatal("the hub is not a town")
	}
	if got := inventory.Count(w, town, "item.lake_water", 0); got != 10 {
		t.Errorf("the hub's stock holds %d lake water, want 10", got)
	}
	if _, ok := agents.TownOf(w, mapNamed(t, w, "map.play.hall")); ok {
		t.Error("the hall is a town")
	}
}

func TestAHaulerDeliversIntoTheTownStockAtAStockMark(t *testing.T) {
	tests := []struct {
		name        string
		cisternBody bool
	}{
		{"an empty stock mark", false},
		{"a body standing on the stock mark", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newPrimField(t, 1, []hallMark{{"q", "cistern", 6, 6}}, hallTown("cistern"))
			town, ok := agents.TownOf(f.w, ecs.NewMap[comp.Pos](f.w.ECS).Get(f.leader()).Map)
			if !ok {
				t.Fatal("the hall is not a town")
			}
			if got := f.countWater(town); got != 3 {
				t.Fatalf("the stock starts with %d water, want 3", got)
			}
			var cistern ecs.Entity
			if tt.cisternBody {
				cistern = f.spawn("template.test.cistern", cellAt(6, 6))
			}
			well := f.spawn("template.test.well", cellAt(14, 4))
			hauler := f.spawn("template.test.hauler", cellAt(14, 8))
			if !f.until(600, func() bool { return f.countWater(town) > 3 }) {
				t.Fatalf("nothing reached the stock by tick %d; the hauler is at %v", f.w.Now, f.at(hauler))
			}
			if got := f.countWater(town); got != 5 {
				t.Errorf("the stock holds %d water, want 5", got)
			}
			if got := f.countWater(well); got != 3 {
				t.Errorf("the well holds %d water, want 3", got)
			}
			if tt.cisternBody && f.countWater(cistern) != 0 {
				t.Errorf("the body on the stock mark got %d water, want none", f.countWater(cistern))
			}
		})
	}
}

func TestTheRegionViewNamesThePartysRegionAndShowsATownsStock(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*content.Bundle)
		want   api.RegionView
	}{
		{"a region that is not a town", nil, api.RegionView{Name: "The test hall"}},
		{"a town", hallTown(), api.RegionView{Name: "The test hall", Town: true, Stock: []api.StockLine{
			{Item: "item.lake_water", Name: "lake water", Count: 3},
		}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, w, _ := gametest.New(t, 1, tt.mutate)
			view, err := g.View()
			if err != nil {
				t.Fatal(err)
			}
			if view.Region == nil {
				t.Fatal("the view has no region section")
			}
			tt.want.Map = mapNamed(t, w, hallMap)
			got := *view.Region
			if got.Name != tt.want.Name || got.Map != tt.want.Map || got.Town != tt.want.Town ||
				!slices.Equal(got.Stock, tt.want.Stock) {
				t.Errorf("region view %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestSavingKeepsRegionsAndStocks(t *testing.T) {
	g, w, cfg := gametest.New(t, 1, hallTown())
	hall := mapNamed(t, w, hallMap)
	town, _ := agents.TownOf(w, hall)
	inventory.Add(w, town, comp.ItemStack{Item: "item.ration", Count: 4})
	loaded, world, _ := gametest.Reload(t, cfg, g)
	before, err := g.Hash()
	if err != nil {
		t.Fatal(err)
	}
	after, err := loaded.Hash()
	if err != nil {
		t.Fatal(err)
	}
	if before != after {
		t.Error("the loaded game hashes differently")
	}
	again, ok := agents.TownOf(world, hall)
	if !ok {
		t.Fatal("the loaded hall is not a town")
	}
	if got := inventory.Count(world, again, "item.ration", 0); got != 4 {
		t.Errorf("the loaded stock holds %d rations, want 4", got)
	}
	if got := inventory.Count(world, again, "item.lake_water", 0); got != 3 {
		t.Errorf("the loaded stock holds %d lake water, want 3", got)
	}
	count := 0
	query := ecs.NewFilter1[comp.Region](world.ECS).Query()
	for query.Next() {
		count++
	}
	if want := len(world.Maps.IDs()); count != want {
		t.Errorf("%d regions after loading, want %d", count, want)
	}
}
