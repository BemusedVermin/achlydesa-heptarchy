package worldgen_test

import (
	"slices"
	"testing"

	"pgregory.net/rapid"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/explore"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/worldgen"
)

const joinReach = 24

type joinWorld struct {
	fixed *explore.WorldSource
	src   *worldgen.Source
	m     *hexmap.Map
	caves *hexmap.Map
}

func newJoinWorld(tb testing.TB, seed, exportCount uint64) joinWorld {
	tb.Helper()
	reg := contenttest.Sample(tb)
	fixed, err := explore.NewWorldSource(reg)
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
	if err := worldgen.RegisterCaves(stages); err != nil {
		tb.Fatal(err)
	}
	if err := worldgen.RegisterJoin(stages); err != nil {
		tb.Fatal(err)
	}
	geography, _ := reg.Geography("geography.deep")
	geography.Stages = geography.Stages[:2]
	cfg := worldgen.Config{
		Seed: seed, ExportCount: exportCount, Bounds: fixed.Bounds(),
		Geography: geography, Stages: stages, Fixed: fixed, Open: open,
	}
	src, err := worldgen.NewSource(cfg)
	if err != nil {
		tb.Fatal(err)
	}
	cfg.Geography.Stages = geography.Stages[:1]
	cavesOnly, err := worldgen.NewSource(cfg)
	if err != nil {
		tb.Fatal(err)
	}
	m := hexmap.NewMap(src, hexmap.Prisms{})
	m.SetBounds(fixed.Bounds(), edge)
	caves := hexmap.NewMap(cavesOnly, hexmap.Prisms{})
	caves.SetBounds(fixed.Bounds(), edge)
	return joinWorld{fixed: fixed, src: src, m: m, caves: caves}
}

func (w joinWorld) at(tb testing.TB, m *hexmap.Map, cell api.Cell) hexmap.Terrain {
	tb.Helper()
	terrain, err := m.At(cell)
	if err != nil {
		tb.Fatal(err)
	}
	return terrain
}

// walkOut returns the cells reachable from the opening by real steps outside the places, within reach.
func (w joinWorld) walkOut(tb testing.TB, opening worldgen.Opening) []api.Cell {
	tb.Helper()
	seen := map[api.Cell]bool{opening.Cell: true}
	frontier := []api.Cell{opening.Cell}
	var reached []api.Cell
	for range joinReach * 2 {
		var next []api.Cell
		for _, from := range frontier {
			for dir := range api.Dir(api.DirCount) {
				step, ok, err := w.m.Step(from, dir, walkRules)
				if err != nil {
					tb.Fatal(err)
				}
				to := step.To.Cell
				if !ok || seen[to] || w.fixed.Covers(to) {
					continue
				}
				seen[to] = true
				next = append(next, to)
				reached = append(reached, to)
			}
		}
		frontier = next
	}
	return reached
}

func TestJoinLeadsEveryOpeningOut(t *testing.T) {
	for _, seed := range []uint64{1, 2, 3, 4} {
		w := newJoinWorld(t, seed, 0)
		openings := w.src.Openings()
		if len(openings) == 0 {
			t.Fatal("the test places have no openings")
		}
		for _, opening := range openings {
			if w.at(t, w.m, opening.Exit).Solid {
				t.Errorf("seed %d: exit %v of opening %v is solid", seed, opening.Exit, opening.Cell)
				continue
			}
			reached := w.walkOut(t, opening)
			if len(reached) == 0 {
				t.Errorf("seed %d: no step leads out of opening %v", seed, opening.Cell)
				continue
			}
			if !w.cavesWithin(t, opening.Exit, joinReach/3) {
				continue
			}
			if !slices.ContainsFunc(reached, func(cell api.Cell) bool { return w.touchesCave(t, cell) }) {
				t.Errorf("seed %d: opening %v has a cave near but its tunnel does not reach one", seed, opening.Cell)
			}
		}
	}
}

// touchesCave reports whether the cell, the one under it or a neighbor on its level is a cave cell.
func (w joinWorld) touchesCave(tb testing.TB, cell api.Cell) bool {
	tb.Helper()
	isCave := func(c api.Cell) bool { return !w.fixed.Covers(c) && !w.at(tb, w.caves, c).Solid }
	if isCave(cell) || isCave(hexmap.Prisms{}.Below(cell).Cell) {
		return true
	}
	for dir := range api.Dir(api.DirCount) {
		if isCave(hexmap.Prisms{}.Neighbor(cell, dir).Cell) {
			return true
		}
	}
	return false
}

func (w joinWorld) cavesWithin(tb testing.TB, from api.Cell, distance int) bool {
	tb.Helper()
	for _, cell := range w.caves.CellsIn(from.Hex, distance, max(0, from.Depth-distance), from.Depth+distance) {
		if hexmap.CellDistance(cell, from) <= distance && !w.at(tb, w.caves, cell).Solid && !w.fixed.Covers(cell) {
			return true
		}
	}
	return false
}

func TestJoinKeepsPlaces(t *testing.T) {
	w := newJoinWorld(t, 5, 1)
	for r := range 3 {
		for q := range 6 {
			key := hexmap.ChunkKey{Q: q, R: r}
			got, err := w.src.Chunk(key)
			if err != nil {
				t.Fatal(err)
			}
			want, err := w.fixed.Chunk(key)
			if err != nil {
				t.Fatal(err)
			}
			origin := key.Origin()
			for depth := range hexmap.ChunkSize {
				for y := range hexmap.ChunkSize {
					for x := range hexmap.ChunkSize {
						cell := api.Cell{Hex: api.Hex{Q: origin.Hex.Q + x, R: origin.Hex.R + y}, Depth: depth}
						if w.fixed.Covers(cell) && got.Get(x, y, depth) != want.Get(x, y, depth) {
							t.Fatalf("place cell %v changed", cell)
						}
					}
				}
			}
		}
	}
}

func TestJoinIgnoresChunkOrder(t *testing.T) {
	var keys []hexmap.ChunkKey
	for r := range 1 {
		for q := range 4 {
			keys = append(keys, hexmap.ChunkKey{Q: q, R: r, Depth: 0})
		}
	}
	reference := newJoinWorld(t, 3, 0)
	want := map[hexmap.ChunkKey]string{}
	for _, key := range keys {
		want[key] = drawJoined(t, reference, key)
	}
	rapid.Check(t, func(rt *rapid.T) {
		w := newJoinWorld(t, 3, 0)
		order := rapid.Permutation(keys).Draw(rt, "order")
		for _, key := range order {
			if got := drawJoined(t, w, key); got != want[key] {
				rt.Fatalf("chunk %v differs when built in order %v", key, order)
			}
		}
	})
}

func drawJoined(tb testing.TB, w joinWorld, key hexmap.ChunkKey) string {
	tb.Helper()
	chunk, err := w.src.Chunk(key)
	if err != nil {
		tb.Fatal(err)
	}
	out := make([]byte, 0, hexmap.ChunkSize*hexmap.ChunkSize*hexmap.ChunkSize)
	for depth := range hexmap.ChunkSize {
		for r := range hexmap.ChunkSize {
			for q := range hexmap.ChunkSize {
				if chunk.Get(q, r, depth).Solid {
					out = append(out, '#')
				} else {
					out = append(out, '.')
				}
			}
		}
	}
	return string(out)
}
