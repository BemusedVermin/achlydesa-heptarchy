package worldgen_test

import (
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/explore"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/worldgen"
)

func decayAt(tb testing.TB, exports uint64) *worldgen.Decay {
	tb.Helper()
	reg := contenttest.Sample(tb)
	fixed, err := explore.NewWorldSource(reg)
	if err != nil {
		tb.Fatal(err)
	}
	decay, err := worldgen.NewDecay(reg, "geography.deep", exports, fixed.Bounds())
	if err != nil {
		tb.Fatal(err)
	}
	return decay
}

func decayRank(tb testing.TB, decay *worldgen.Decay, depth int) int {
	tb.Helper()
	terrain, ok := decay.Terrain(depth)
	switch {
	case !ok:
		return 0
	case terrain.ID == "terrain.decayed_clay":
		return 1
	default:
		return 2
	}
}

func TestDecayGrowsWithExportsAndReachesShallowFirst(t *testing.T) {
	const deepest = 63
	for exports := range uint64(8) {
		now := decayAt(t, exports)
		later := decayAt(t, exports+1)
		for depth := range deepest {
			if decayRank(t, later, depth) < decayRank(t, now, depth) {
				t.Errorf("depth %d: export %d is less decayed than export %d", depth, exports+1, exports)
			}
			if decayRank(t, now, depth+1) > decayRank(t, now, depth) {
				t.Errorf("export %d: depth %d is more decayed than depth %d", exports, depth+1, depth)
			}
		}
	}
	if decayRank(t, decayAt(t, 0), 0) != 0 {
		t.Error("the deep starts decayed before any export")
	}
	if decayRank(t, decayAt(t, 6), 0) != 2 || decayRank(t, decayAt(t, 6), deepest) != 0 {
		t.Error("six exports should rot the top and leave the bottom whole")
	}
}

func cellOf(origin api.Cell, q, r, depth int) api.Cell {
	return api.Cell{Hex: api.Hex{Q: origin.Hex.Q + q, R: origin.Hex.R + r}, Depth: origin.Depth + depth}
}

func TestSourceDecaysOnlySolidGround(t *testing.T) {
	reg := contenttest.Sample(t)
	fixed, err := explore.NewWorldSource(reg)
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
	geography, _ := reg.Geography("geography.deep")
	geography.Stages = geography.Stages[:1]
	const exports = 6
	src, err := worldgen.NewSource(worldgen.Config{
		Seed: 1, ExportCount: exports, Bounds: fixed.Bounds(), Geography: geography, Stages: stages,
		Fixed: fixed, Open: open, Decay: decayAt(t, exports),
	})
	if err != nil {
		t.Fatal(err)
	}
	chunk, err := src.Chunk(hexmap.ChunkKey{Q: 2, R: 1})
	if err != nil {
		t.Fatal(err)
	}
	origin := hexmap.ChunkKey{Q: 2, R: 1}.Origin()
	var rotten, opened int
	for depth := range hexmap.ChunkSize {
		for r := range hexmap.ChunkSize {
			for q := range hexmap.ChunkSize {
				switch got := chunk.Get(q, r, depth); {
				case got == open:
					opened++
				case got.ID == "terrain.rotten_clay":
					rotten++
				case got.ID == "terrain.packed_clay" && depth < 8 && fixed.Bounds().Contains(cellOf(origin, q, r, depth)) &&
					!fixed.Covers(cellOf(origin, q, r, depth)):
					t.Fatalf("solid cell (%d,%d,%d) kept its clay", q, r, depth)
				}
			}
		}
	}
	if rotten == 0 || opened == 0 {
		t.Errorf("got %d rotten and %d open cells, want both", rotten, opened)
	}
}
