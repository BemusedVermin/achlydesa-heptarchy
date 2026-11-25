package explore_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
)

const (
	roadLength = 80
	walk       = 64
	budget     = 4
)

// roadWorld replaces the places with one long east-west road, three rows of air through the clay, under the
// hall's name.
func roadWorld(bundle *content.Bundle) {
	bundle.World.Cols = roadLength + 10
	bundle.World.ChunkBudget = budget
	bundle.World.NearChunks = 1
	road := strings.Repeat(".", roadLength)
	bundle.Places = map[string]content.PlaceDef{"place.test.hall": {
		Name:   "text.place.test.hall",
		Layers: []content.LayerDef{{Rows: []string{road, "..@" + road[3:], road}}},
		Legend: map[string]api.ContentID{".": "terrain.air"},
		Marks:  map[string]string{"@": "party_start"},
	}}
	bundle.Start.Place = "place.test.hall"
}

func TestFarChunksUnloadUnderTheBudgetAndReloadTheSame(t *testing.T) {
	fx := newFixture(t, "", roadWorld)
	leader := fx.view().Party.Members[0].UID
	home := hexmap.KeyOf(fx.pos(leader).At)
	edited := home.Cell(hexmap.SlotOf(api.Cell{Hex: api.Hex{Q: 5, R: 6}, Depth: 0}))
	open := hexmap.Terrain{ID: "terrain.air"}
	if err := fx.w.Map.Set(edited, open); err != nil {
		t.Fatal(err)
	}
	before := make([]hexmap.Terrain, hexmap.ChunkCells)
	for slot := range before {
		var err error
		if before[slot], err = fx.w.Map.At(home.Cell(slot)); err != nil {
			t.Fatal(err)
		}
	}
	checkBudget := func(when string) {
		t.Helper()
		if loaded := fx.w.Map.Loaded(); len(loaded) > budget {
			t.Fatalf("%s: %d chunks loaded, budget %d: %v", when, len(loaded), budget, loaded)
		}
	}
	for range walk {
		fx.stepLeader("e")
		checkBudget("walking out")
	}
	if slices.Contains(fx.w.Map.Loaded(), home) {
		t.Fatalf("the home chunk %v is still loaded %d hexes away", home, walk)
	}
	for range walk {
		fx.stepLeader("w")
		checkBudget("walking back")
	}
	for slot, want := range before {
		got, err := fx.w.Map.At(home.Cell(slot))
		if err != nil {
			t.Fatal(err)
		}
		if got != want {
			t.Fatalf("%v after reload = %v, want %v", home.Cell(slot), got, want)
		}
	}
	if got, _ := fx.w.Map.At(edited); got != open {
		t.Errorf("edit at %v = %v after the chunk reloaded, want %v", edited, got, open)
	}
}
