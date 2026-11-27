package explore_test

import (
	"errors"
	"maps"
	"slices"
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/explore"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
)

const testPiece = "piece.test.nook"

// pieceWorld builds a small world holding one place at column 2, row 2, with a piece whose rows are
// "+x#" and "gw": a door, a plain mark, a wall, a grave landmark and a well.
func pieceWorld(tb testing.TB) (*content.Registry, *explore.WorldSource) {
	tb.Helper()
	bundle := contenttest.SampleBundle(tb)
	contenttest.NoStart(&bundle)
	bundle.World = content.WorldDef{
		Cols: 32, Rows: 16, MinDepth: 0, MaxDepth: 5,
		Fill: "terrain.packed_clay", Edge: "terrain.packed_clay", Open: "terrain.air",
	}
	legend := map[string]api.ContentID{"#": "terrain.packed_clay", ".": "terrain.air"}
	bundle.Places = map[string]content.PlaceDef{"place.test.a": {
		Name:   "text.place.test.a",
		At:     content.PlaceAt{Col: 2, Row: 2},
		Layers: []content.LayerDef{{Depth: 0, Rows: []string{"###", "#.#"}}},
		Legend: legend,
	}}
	bundle.Pieces[testPiece] = content.PieceDef{
		Name:        "text." + testPiece,
		Layers:      []content.LayerDef{{Depth: 1, Rows: []string{"+x#", "gw"}}},
		Legend:      legend,
		Marks:       map[string]string{"+": "door", "x": "spot", "g": "grave", "w": "well"},
		Spawns:      map[string]api.ContentID{"w": "template.source.well"},
		Affordances: map[string][]api.ContentID{"g": {"affordance.the_dead"}},
		Doors:       []string{"+"},
	}
	bundle.Text["text.place.test.a"] = "a"
	bundle.Text["text."+testPiece] = "a nook"
	reg := contenttest.Registry(tb, bundle)
	src, err := explore.NewWorldSource(reg, 0)
	if err != nil {
		tb.Fatal(err)
	}
	return reg, src
}

func TestAPieceStampsAtAnOffsetAndARotationWithItsMarksMapped(t *testing.T) {
	at := api.Cell{Hex: api.Hex{Q: 10, R: 10}, Depth: 2}
	offset := func(q, r int) api.Cell {
		return api.Cell{Hex: api.Hex{Q: at.Hex.Q + q, R: at.Hex.R + r}, Depth: at.Depth + 1}
	}
	tests := []struct {
		name                          string
		rotation                      api.Dir
		door, spot, wall, grave, well api.Cell
	}{
		{"as drawn", api.DirE, offset(0, 0), offset(1, 0), offset(2, 0), offset(0, 1), offset(1, 1)},
		{"a sixth turn", api.DirNE, offset(0, 0), offset(1, -1), offset(2, -2), offset(1, 0), offset(2, -1)},
		{"a half turn", api.DirW, offset(0, 0), offset(-1, 0), offset(-2, 0), offset(0, -1), offset(-1, -1)},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reg, src := pieceWorld(t)
			stamped, err := src.Stamp(reg, testPiece, at, test.rotation)
			if err != nil {
				t.Fatal(err)
			}
			if want := map[string]api.Cell{"spot": test.spot}; !maps.Equal(stamped.Marks, want) {
				t.Errorf("marks %v, want %v", stamped.Marks, want)
			}
			if want := []api.Cell{test.door}; !slices.Equal(stamped.Doors, want) {
				t.Errorf("doors %v, want %v", stamped.Doors, want)
			}
			if want := []explore.Spawn{
				{Cell: test.well, Template: "template.source.well"},
			}; !slices.Equal(
				src.Spawns(),
				want,
			) {
				t.Errorf("spawns %v, want %v", src.Spawns(), want)
			}
			landmarks := src.Landmarks()
			if len(landmarks) != 1 || landmarks[0].Cell != test.grave || landmarks[0].Place != testPiece {
				t.Errorf("landmarks %v, want the grave at %v from %s", landmarks, test.grave, testPiece)
			}
			world := hexmap.NewMap(src, hexmap.Prisms{})
			world.SetBounds(src.Bounds(), src.Edge())
			for cell, want := range map[api.Cell]api.ContentID{
				test.wall: "terrain.packed_clay", test.spot: "terrain.air", test.door: "terrain.air",
			} {
				if got, err := world.At(cell); err != nil || got.ID != want {
					t.Errorf("At(%v) = %q, %v; want %q", cell, got.ID, err, want)
				}
				if !src.Covers(cell) {
					t.Errorf("Covers(%v) is false for a stamped cell", cell)
				}
			}
		})
	}
}

func TestAStampThatOverlapsOrLeavesTheWorldIsRefused(t *testing.T) {
	place := api.Cell{Hex: hexmap.FromOffsetOddR(3, 3)}
	tests := []struct {
		name string
		at   api.Cell
		want error
	}{
		{"over a place", api.Cell{Hex: place.Hex, Depth: -1}, explore.ErrOverlap},
		{"over an earlier piece", api.Cell{Hex: api.Hex{Q: 10, R: 10}, Depth: 1}, explore.ErrOverlap},
		{"outside the world", api.Cell{Hex: api.Hex{Q: 10, R: 10}, Depth: 5}, explore.ErrOutside},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reg, src := pieceWorld(t)
			if _, err := src.Stamp(
				reg,
				testPiece,
				api.Cell{Hex: api.Hex{Q: 10, R: 10}, Depth: 1},
				api.DirE,
			); err != nil {
				t.Fatal(err)
			}
			covered := src.ChunkKeys()
			_, err := src.Stamp(reg, testPiece, test.at, api.DirSE)
			if !errors.Is(err, test.want) {
				t.Fatalf("Stamp = %v, want %v", err, test.want)
			}
			if len(src.Spawns()) != 1 || len(src.Landmarks()) != 1 || !slices.Equal(src.ChunkKeys(), covered) {
				t.Errorf("a refused stamp left spawns %v, landmarks %v, chunks %v (before %v)",
					src.Spawns(), src.Landmarks(), src.ChunkKeys(), covered)
			}
		})
	}
}
