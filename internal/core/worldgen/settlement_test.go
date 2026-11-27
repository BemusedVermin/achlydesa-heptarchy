package worldgen_test

import (
	"slices"
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/explore"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/worldgen"
)

const townID = "settlement.test.town"

func townBundle(tb testing.TB) content.Bundle {
	tb.Helper()
	bundle := contenttest.SampleBundle(tb)
	bundle.World.Cols, bundle.World.Rows = 140, 50
	bundle.Settlements = map[string]content.SettlementDef{townID: {
		Name:   "text." + townID,
		At:     content.SettlementAt{Col: 100, Row: 24},
		Radius: 22,
		Street: "piece.street",
		Pieces: map[string]content.SettlementPiece{
			"piece.street":    {Weight: 4, Count: 6},
			"piece.house":     {Weight: 6, Count: 10},
			"piece.well":      {Weight: 1, Count: 2},
			"piece.shrine":    {Weight: 1, Count: 2},
			"piece.graveyard": {Weight: 1, Count: 2},
		},
		Exits:       2,
		LandmarkMin: map[string]int{"affordance.the_dead": 3, "affordance.worship": 1},
	}}
	bundle.Text["text."+townID] = "a town"
	return bundle
}

func lay(tb testing.TB, bundle content.Bundle, seed uint64) (*content.Registry, worldgen.SettlementLayout) {
	tb.Helper()
	reg := contenttest.Registry(tb, bundle)
	layout, err := worldgen.LaySettlement(reg, townID, seed)
	if err != nil {
		tb.Fatal(err)
	}
	return reg, layout
}

func TestSameSeedGivesTheSameSettlement(t *testing.T) {
	bundle := townBundle(t)
	_, first := lay(t, bundle, 7)
	_, again := lay(t, bundle, 7)
	if !slices.Equal(first.Placements, again.Placements) || !slices.Equal(first.Exits, again.Exits) {
		t.Fatal("the same seed laid out two different settlements")
	}
	differs := false
	for seed := uint64(8); seed < 16 && !differs; seed++ {
		_, other := lay(t, bundle, seed)
		differs = !slices.Equal(first.Placements, other.Placements)
	}
	if !differs {
		t.Error("eight other seeds all laid out the same settlement")
	}
}

func TestSettlementPiecesNeverOverlapAndEveryPieceJoinsAnEarlierOne(t *testing.T) {
	bundle := townBundle(t)
	plain := bundle
	plain.Settlements = nil
	plainReg := contenttest.Registry(t, plain)
	for seed := uint64(1); seed <= 12; seed++ {
		reg, layout := lay(t, bundle, seed)
		src, err := explore.NewWorldSource(plainReg, 0)
		if err != nil {
			t.Fatal(err)
		}
		var earlier []api.Cell
		for i, placement := range layout.Placements {
			stamped, err := src.Stamp(reg, placement.Piece, placement.At, placement.Rotation)
			if err != nil {
				t.Fatalf("seed %d: piece %d: %v", seed, i, err)
			}
			if i > 0 && !slices.ContainsFunc(stamped.Doors, func(door api.Cell) bool {
				return slices.ContainsFunc(earlier, func(other api.Cell) bool {
					return door.Depth == other.Depth && hexmap.Distance(door.Hex, other.Hex) == 1
				})
			}) {
				t.Errorf("seed %d: piece %d (%s) has no door beside an earlier door", seed, i, placement.Piece)
			}
			earlier = append(earlier, stamped.Doors...)
		}
	}
}

func TestAnExportLeavesTheSettlementUnchanged(t *testing.T) {
	reg := contenttest.Registry(t, townBundle(t))
	fixed, err := explore.NewWorldSource(reg, 5)
	if err != nil {
		t.Fatal(err)
	}
	open, err := explore.Terrain(reg, reg.World().Open)
	if err != nil {
		t.Fatal(err)
	}
	stages := worldgen.NewStages()
	if err := worldgen.RegisterCaves(stages); err != nil {
		t.Fatal(err)
	}
	if err := worldgen.RegisterJoin(stages); err != nil {
		t.Fatal(err)
	}
	geography, _ := reg.Geography("geography.deep")
	geography.Stages = geography.Stages[:2]
	_, layout := lay(t, townBundle(t), 5)
	chunks := map[hexmap.ChunkKey]bool{}
	for _, placement := range layout.Placements {
		chunks[hexmap.KeyOf(placement.At)] = true
	}
	var sources []*worldgen.Source
	for exportCount := range uint64(2) {
		src, err := worldgen.NewSource(worldgen.Config{
			Seed: 5, ExportCount: exportCount, Bounds: fixed.Bounds(),
			Geography: geography, Stages: stages, Fixed: fixed, Open: open,
		})
		if err != nil {
			t.Fatal(err)
		}
		sources = append(sources, src)
	}
	for key := range chunks {
		before, err := sources[0].Chunk(key)
		if err != nil {
			t.Fatal(err)
		}
		after, err := sources[1].Chunk(key)
		if err != nil {
			t.Fatal(err)
		}
		origin := key.Origin()
		for _, placement := range layout.Placements {
			cell := placement.At
			if hexmap.KeyOf(cell) != key || !fixed.Covers(cell) {
				continue
			}
			q, r, depth := cell.Hex.Q-origin.Hex.Q, cell.Hex.R-origin.Hex.R, cell.Depth-origin.Depth
			if before.Get(q, r, depth) != after.Get(q, r, depth) {
				t.Fatalf("cell %v changed between export counts", cell)
			}
		}
	}
	if !slices.Equal(sources[0].Openings(), sources[1].Openings()) {
		t.Error("the openings changed between export counts")
	}
}

func TestSettlementLandmarkMinimumsHoldOrLayingFails(t *testing.T) {
	bundle := townBundle(t)
	for seed := uint64(1); seed <= 12; seed++ {
		reg, layout := lay(t, bundle, seed)
		gave := map[api.ContentID]int{}
		for _, placement := range layout.Placements {
			piece, _ := reg.Piece(placement.Piece)
			for char, affordances := range piece.Affordances {
				for _, layer := range piece.Layers {
					for _, row := range layer.Rows {
						for _, c := range row {
							if string(c) == char {
								for _, affordance := range affordances {
									gave[affordance]++
								}
							}
						}
					}
				}
			}
		}
		if gave["affordance.the_dead"] < 3 || gave["affordance.worship"] < 1 {
			t.Errorf("seed %d: landmarks %v miss the minimums", seed, gave)
		}
		if len(layout.Exits) != 2 {
			t.Errorf("seed %d: %d exits, want 2", seed, len(layout.Exits))
		}
	}

	bundle.Settlements = map[string]content.SettlementDef{townID: bundle.Settlements[townID]}
	contenttest.Edit(bundle.Settlements, townID, func(def *content.SettlementDef) {
		def.Pieces = map[string]content.SettlementPiece{
			"piece.street": {Weight: 1, Count: 3},
			"piece.house":  {Weight: 1, Count: 3},
		}
	})
	reg := contenttest.Registry(t, bundle)
	if _, err := worldgen.LaySettlement(reg, townID, 1); err == nil {
		t.Error("laying out a settlement whose pieces give no dead succeeded")
	}
}
