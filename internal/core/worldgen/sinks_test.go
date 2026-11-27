package worldgen_test

import (
	"testing"

	"pgregory.net/rapid"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/explore"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/worldgen"
)

const (
	sinkRim   = 1
	sinkFloor = 7
)

var walkRules = hexmap.StepRules{MaxStepUp: 1, MaxStepDown: 1, MaxDrop: 8}

type sinkWorld struct {
	m      *hexmap.Map
	src    *worldgen.Source
	open   hexmap.Terrain
	bounds hexmap.Bounds
}

func sinkParams(riserChance float64) map[string]float64 {
	return map[string]float64{
		"spacing": 44, "jitter": 0.4, "density": 1, "riser_chance": riserChance,
		"rim_depth": sinkRim, "lap_height": 3, "laps": 2, "headroom": 2,
		"shaft_radius": 1.5, "outer_radius": 8,
	}
}

func newSinkWorld(tb testing.TB, seed, exportCount uint64, riserChance float64) sinkWorld {
	tb.Helper()
	reg := contenttest.Sample(tb)
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
	if err := worldgen.RegisterSinks(stages); err != nil {
		tb.Fatal(err)
	}
	src, err := worldgen.NewSource(worldgen.Config{
		Seed: seed, ExportCount: exportCount, Bounds: fixed.Bounds(),
		Geography: content.GeographyDef{Stages: []content.StageDef{
			{Stage: worldgen.SinksStage, Params: sinkParams(riserChance)},
		}},
		Stages: stages, Fixed: fixed, Open: open,
	})
	if err != nil {
		tb.Fatal(err)
	}
	m := hexmap.NewMap(src, hexmap.Prisms{})
	m.SetBounds(fixed.Bounds(), edge)
	return sinkWorld{m: m, src: src, open: open, bounds: fixed.Bounds()}
}

func (w sinkWorld) standable(tb testing.TB, cell api.Cell) bool {
	tb.Helper()
	ok, err := w.m.Standable(cell)
	if err != nil {
		tb.Fatal(err)
	}
	return ok
}

func (w sinkWorld) findShaft(tb testing.TB) api.Hex {
	tb.Helper()
	for row := w.bounds.MinRow; row <= w.bounds.MaxRow; row++ {
		for col := w.bounds.MinCol; col <= w.bounds.MaxCol; col++ {
			hex := hexmap.FromOffsetOddR(col, row)
			column := true
			for depth := sinkRim; depth <= sinkFloor; depth++ {
				terrain, err := w.m.At(api.Cell{Hex: hex, Depth: depth})
				if err != nil {
					tb.Fatal(err)
				}
				column = column && !terrain.Solid
			}
			if column {
				return hex
			}
		}
	}
	tb.Fatal("no shaft found")
	return api.Hex{}
}

func hexesAround(center api.Hex, radius int) []api.Hex {
	var hexes []api.Hex
	for dq := -radius; dq <= radius; dq++ {
		for dr := -radius; dr <= radius; dr++ {
			if dq+dr >= -radius && dq+dr <= radius {
				hexes = append(hexes, api.Hex{Q: center.Q + dq, R: center.R + dr})
			}
		}
	}
	return hexes
}

func (w sinkWorld) walkable(tb testing.TB, from, to api.Cell) bool {
	tb.Helper()
	seen := map[api.Cell]bool{from: true}
	queue := []api.Cell{from}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		if cur == to {
			return true
		}
		for dir := range api.DirCount {
			step, ok, err := w.m.Step(cur, api.Dir(dir), walkRules)
			if err != nil {
				tb.Fatal(err)
			}
			if !ok || step.Kind == hexmap.StepDrop || seen[step.To.Cell] {
				continue
			}
			seen[step.To.Cell] = true
			queue = append(queue, step.To.Cell)
		}
	}
	return false
}

func TestSinkAndRiserCanBeWalked(t *testing.T) {
	for _, tc := range []struct {
		name        string
		riserChance float64
	}{{"sink", 0}, {"riser", 1}} {
		t.Run(tc.name, func(t *testing.T) {
			w := newSinkWorld(t, 3, 0, tc.riserChance)
			shaft := w.findShaft(t)
			bottom := api.Cell{Hex: shaft, Depth: sinkFloor}
			if !w.standable(t, bottom) {
				t.Fatalf("shaft bottom %v is not standable", bottom)
			}
			var rim api.Cell
			for _, hex := range hexesAround(shaft, 9) {
				if cell := (api.Cell{Hex: hex, Depth: sinkRim}); w.standable(t, cell) {
					rim = cell
					break
				}
			}
			if rim == (api.Cell{}) {
				t.Fatal("no floor at the rim depth")
			}
			if !w.walkable(t, rim, bottom) {
				t.Errorf("no walk without drops from rim %v down to bottom %v", rim, bottom)
			}
			if !w.walkable(t, bottom, rim) {
				t.Errorf("no walk without drops from bottom %v up to rim %v", bottom, rim)
			}
		})
	}
}

func TestSinkShaftIsAFall(t *testing.T) {
	w := newSinkWorld(t, 3, 0, 0)
	shaft := w.findShaft(t)
	var deepest hexmap.Step
	for _, hex := range hexesAround(shaft, 9) {
		from := api.Cell{Hex: hex, Depth: sinkRim}
		if !w.standable(t, from) {
			continue
		}
		for dir := range api.DirCount {
			step, ok, err := w.m.Step(from, api.Dir(dir), walkRules)
			if err != nil {
				t.Fatal(err)
			}
			if ok && step.Kind == hexmap.StepDrop && step.Fall > deepest.Fall {
				deepest = step
			}
		}
	}
	if deepest.Fall < 2 {
		t.Fatalf("deepest drop from the rim level is %d, want several depth steps", deepest.Fall)
	}
	if !w.standable(t, deepest.To.Cell) {
		t.Errorf("fall lands on %v, which is not standable", deepest.To.Cell)
	}
}

func TestSinksIgnoreChunkOrder(t *testing.T) {
	var keys []hexmap.ChunkKey
	for r := range 3 {
		for q := range 4 {
			keys = append(keys, hexmap.ChunkKey{Q: q, R: r})
		}
	}
	reference := newSinkWorld(t, 5, 1, 0.5)
	fixed, err := explore.NewWorldSource(contenttest.Sample(t), 0)
	if err != nil {
		t.Fatal(err)
	}
	want := map[hexmap.ChunkKey]string{}
	changed := false
	for _, key := range keys {
		want[key] = drawChunk(t, reference.src, key, reference.open)
		changed = changed || want[key] != drawChunk(t, fixed, key, reference.open)
	}
	if !changed {
		t.Fatal("no sink reaches the chunks")
	}
	rapid.Check(t, func(rt *rapid.T) {
		w := newSinkWorld(t, 5, 1, 0.5)
		order := rapid.Permutation(keys).Draw(rt, "order")
		for _, key := range order {
			if got := drawChunk(t, w.src, key, w.open); got != want[key] {
				rt.Fatalf("chunk %v differs when built in order %v", key, order)
			}
		}
	})
}
