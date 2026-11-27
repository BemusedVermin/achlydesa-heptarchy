package explore_test

import (
	"slices"
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/explore"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
)

func TestWorldSourcePlacesEachPlaceInTheFill(t *testing.T) {
	bundle := contenttest.SampleBundle(t)
	contenttest.NoStart(&bundle)
	bundle.World = content.WorldDef{
		Cols: 32, Rows: 16, MinDepth: 0, MaxDepth: 5,
		Fill: "terrain.air", Edge: "terrain.packed_clay", Open: "terrain.air",
	}
	legend := map[string]api.ContentID{"#": "terrain.packed_clay"}
	bundle.Places = map[string]content.PlaceDef{
		"place.test.b": {
			Name:   "text.place.test.b",
			At:     content.PlaceAt{Col: 20, Row: 10, Depth: 3},
			Layers: []content.LayerDef{{Depth: 1, Rows: []string{"#c#", " # "}}},
			Legend: legend,
			Marks:  map[string]string{"c": "creature"},
			Spawns: map[string]api.ContentID{"c": "template.creature.moving_clay"},
		},
		"place.test.a": {
			Name:   "text.place.test.a",
			At:     content.PlaceAt{Col: 2, Row: 2},
			Layers: []content.LayerDef{{Depth: 0, Rows: []string{"###", "#@#", "#c#"}}},
			Legend: legend,
			Marks:  map[string]string{"@": "start", "c": "creature"},
			Spawns: map[string]api.ContentID{"c": "template.creature.moving_clay"},
		},
	}
	bundle.Text["text.place.test.a"] = "a"
	bundle.Text["text.place.test.b"] = "b"
	src, err := explore.NewWorldSource(contenttest.Registry(t, bundle), 0)
	if err != nil {
		t.Fatal(err)
	}
	world := hexmap.NewMap(src, hexmap.Prisms{})
	world.SetBounds(src.Bounds(), src.Edge())

	cell := func(col, row, depth int) api.Cell {
		return api.Cell{Hex: hexmap.FromOffsetOddR(col, row), Depth: depth}
	}
	tests := []struct {
		name string
		cell api.Cell
		want api.ContentID
	}{
		{"a wall of place a", cell(2, 3, 0), "terrain.packed_clay"},
		{"the start of place a", cell(3, 3, 0), "terrain.air"},
		{"a wall of place b, deeper", cell(21, 11, 4), "terrain.packed_clay"},
		{"a space in place b leaves the fill", cell(20, 11, 4), "terrain.air"},
		{"the fill between them", cell(10, 6, 2), "terrain.air"},
		{"place b's cell at the wrong depth", cell(21, 11, 3), "terrain.air"},
		{"west of the bounds", cell(-1, 3, 0), "terrain.packed_clay"},
		{"below the bounds", cell(10, 6, 6), "terrain.packed_clay"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := world.At(test.cell)
			if err != nil {
				t.Fatal(err)
			}
			if got.ID != test.want {
				t.Errorf("At(%v) = %q, want %q", test.cell, got.ID, test.want)
			}
		})
	}

	if got, ok := src.Mark("place.test.a", "start"); !ok || got != cell(3, 3, 0) {
		t.Errorf("Mark(place.test.a, start) = %v, %v; want %v", got, ok, cell(3, 3, 0))
	}
	if _, ok := src.Mark("place.test.b", "start"); ok {
		t.Error("place.test.b has a start mark, want marks kept per place")
	}
	var spawned []api.Cell
	for _, spawn := range src.Spawns() {
		spawned = append(spawned, spawn.Cell)
	}
	if want := []api.Cell{cell(3, 4, 0), cell(21, 10, 4)}; !slices.Equal(spawned, want) {
		t.Errorf("spawns at %v, want %v (places in ID order)", spawned, want)
	}
}
