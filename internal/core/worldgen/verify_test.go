package worldgen_test

import (
	"flag"
	"maps"
	"os"
	"slices"
	"strconv"
	"testing"

	"pgregory.net/rapid"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/explore"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/worldgen"
)

var verifyRules = hexmap.StepRules{MaxStepUp: 1, MaxStepDown: 1, MaxDrop: 1}

// newVerifiedWorld builds the fixture world's deep; a budget above zero replaces verify's max_chunks.
func newVerifiedWorld(tb testing.TB, seed, exportCount uint64, budget int) (*worldgen.Source, *hexmap.Map) {
	tb.Helper()
	return newVerifiedWorldWith(tb, seed, exportCount, budget, 0)
}

// newVerifiedWorldWith is newVerifiedWorld with verify's summary workers fixed; zero leaves the default.
func newVerifiedWorldWith(
	tb testing.TB, seed, exportCount uint64, budget, workers int,
) (*worldgen.Source, *hexmap.Map) {
	tb.Helper()
	return verifiedWorldOf(tb, contenttest.Sample(tb), seed, exportCount, budget, workers)
}

func verifiedWorldOf(
	tb testing.TB, reg *content.Registry, seed, exportCount uint64, budget, workers int,
) (*worldgen.Source, *hexmap.Map) {
	tb.Helper()
	fixed, err := explore.NewWorldSource(reg, 0)
	if err != nil {
		tb.Fatal(err)
	}
	open, err := explore.Terrain(reg, reg.World().Open)
	if err != nil {
		tb.Fatal(err)
	}
	edge, err := explore.Terrain(reg, reg.World().Edge)
	if err != nil {
		tb.Fatal(err)
	}
	stages := worldgen.NewStages()
	for _, register := range []func(*worldgen.Stages) error{
		worldgen.RegisterCaves, worldgen.RegisterJoin, worldgen.RegisterSinks,
		func(stages *worldgen.Stages) error {
			if workers > 0 {
				return worldgen.RegisterVerifyWorkers(stages, workers)
			}
			return worldgen.RegisterVerify(stages)
		},
	} {
		if err := register(stages); err != nil {
			tb.Fatal(err)
		}
	}
	geography, _ := reg.Geography("geography.deep")
	geography.Stages = slices.Clone(geography.Stages)
	for i, stage := range geography.Stages {
		if stage.Stage == worldgen.VerifyStage && budget > 0 {
			geography.Stages[i].Params = maps.Clone(stage.Params)
			geography.Stages[i].Params["max_chunks"] = float64(budget)
		}
	}
	src, err := worldgen.NewSource(worldgen.Config{
		Seed: seed, ExportCount: exportCount, Bounds: fixed.Bounds(),
		Geography: geography, Stages: stages, Fixed: fixed, Open: open,
	})
	if err != nil {
		tb.Fatal(err)
	}
	m := hexmap.NewMap(src, hexmap.Prisms{})
	m.SetBounds(fixed.Bounds(), edge)
	return src, m
}

// withChecks sets rapid's check count for this test alone, unless the run asked for one.
func withChecks(t *testing.T, checks int) {
	t.Helper()
	explicit := os.Getenv("RAPID_CHECKS") != ""
	flag.Visit(func(set *flag.Flag) { explicit = explicit || set.Name == "rapid.checks" })
	if explicit {
		return
	}
	setting := flag.Lookup("rapid.checks").Value
	before := setting.String()
	if err := setting.Set(strconv.Itoa(checks)); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := setting.Set(before); err != nil {
			t.Error(err)
		}
	})
}

func TestVerifyEveryPlaceReachesEveryOther(t *testing.T) {
	withChecks(t, 40)
	rapid.Check(t, func(rt *rapid.T) {
		seed := rapid.Uint64Range(1, 1<<20).Draw(rt, "seed")
		budget := rapid.IntRange(4, 12).Draw(rt, "budget")
		src, m := newVerifiedWorld(t, seed, 0, budget)
		openings := src.Openings()
		if len(openings) == 0 {
			rt.Fatal("the test places have no openings")
		}
		frontier := []api.Cell{openings[0].Cell, openings[0].Exit}
		seen := map[api.Cell]bool{frontier[0]: true, frontier[1]: true}
		for len(frontier) > 0 {
			from := frontier[0]
			frontier = frontier[1:]
			for dir := range api.Dir(api.DirCount) {
				step, ok, err := m.Step(from, dir, verifyRules)
				if err != nil {
					rt.Fatal(err)
				}
				if ok && !seen[step.To.Cell] {
					seen[step.To.Cell] = true
					frontier = append(frontier, step.To.Cell)
				}
			}
		}
		if peak := worldgen.PeakChunks(src); peak < 1 || peak > budget {
			rt.Fatalf("seed %d: verify held %d chunks at once, want 1 to %d", seed, peak, budget)
		}
		for _, opening := range openings {
			if !reaches(rt, m, seen, opening) {
				rt.Fatalf("seed %d: opening %v of %q is not reachable from %v",
					seed, opening.Cell, opening.Place, openings[0].Cell)
			}
		}
	})
}

// reaches reports whether the walk met the opening's cell or ground a step from it or its exit leads to.
func reaches(rt *rapid.T, m *hexmap.Map, seen map[api.Cell]bool, opening worldgen.Opening) bool {
	for _, base := range []api.Cell{opening.Cell, opening.Exit} {
		if seen[base] {
			return true
		}
		for dir := range api.Dir(api.DirCount) {
			step, ok, err := m.Step(base, dir, verifyRules)
			if err != nil {
				rt.Fatal(err)
			}
			if ok && seen[step.To.Cell] {
				return true
			}
		}
	}
	return false
}

func TestVerifyIsWorkerIndependent(t *testing.T) {
	for _, seed := range []uint64{3, 9} {
		one, _ := newVerifiedWorldWith(t, seed, 1, 0, 1)
		many, _ := newVerifiedWorldWith(t, seed, 1, 0, 5)
		for q := range 6 {
			for r := range 3 {
				key := hexmap.ChunkKey{Q: q, R: r}
				first, err := one.Chunk(key)
				if err != nil {
					t.Fatal(err)
				}
				second, err := many.Chunk(key)
				if err != nil {
					t.Fatal(err)
				}
				for slot := range hexmap.ChunkCells {
					cell := key.Cell(slot)
					origin := key.Origin()
					x, y, d := cell.Hex.Q-origin.Hex.Q, cell.Hex.R-origin.Hex.R, cell.Depth-origin.Depth
					if first.Get(x, y, d) != second.Get(x, y, d) {
						t.Fatalf("seed %d: chunk %v differs at %v between 1 and 5 workers", seed, key, cell)
					}
				}
			}
		}
	}
}

func TestVerifyIsOrderIndependent(t *testing.T) {
	for _, seed := range []uint64{3, 9} {
		forward, _ := newVerifiedWorld(t, seed, 1, 0)
		backward, _ := newVerifiedWorld(t, seed, 1, 0)
		var keys []hexmap.ChunkKey
		for q := range 6 {
			for r := range 3 {
				keys = append(keys, hexmap.ChunkKey{Q: q, R: r})
			}
		}
		built := map[hexmap.ChunkKey]*hexmap.Chunk{}
		for _, key := range keys {
			chunk, err := forward.Chunk(key)
			if err != nil {
				t.Fatal(err)
			}
			built[key] = chunk
		}
		for _, key := range slices.Backward(keys) {
			chunk, err := backward.Chunk(key)
			if err != nil {
				t.Fatal(err)
			}
			for slot := range hexmap.ChunkCells {
				cell := key.Cell(slot)
				origin := key.Origin()
				q, r, d := cell.Hex.Q-origin.Hex.Q, cell.Hex.R-origin.Hex.R, cell.Depth-origin.Depth
				if chunk.Get(q, r, d) != built[key].Get(q, r, d) {
					t.Fatalf("seed %d: chunk %v differs at %v", seed, key, cell)
				}
			}
		}
	}
}
