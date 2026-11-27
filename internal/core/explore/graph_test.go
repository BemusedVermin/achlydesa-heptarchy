package explore_test

import (
	"cmp"
	"reflect"
	"slices"
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/actions"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/worldgen"
)

// loadedOpenings lists every step out of a chunk the world's own map offers, chunk by chunk.
func (ew *exportWorld) loadedOpenings() map[hexmap.ChunkKey][]worldgen.ChunkOpening {
	ew.t.Helper()
	rules := actions.StepRules(ew.w)
	bounds := ew.src.Bounds()
	out := map[hexmap.ChunkKey][]worldgen.ChunkOpening{}
	for _, from := range ew.cells {
		key := hexmap.KeyOf(from)
		if _, ok := out[key]; !ok {
			out[key] = []worldgen.ChunkOpening{}
		}
		standable, err := ew.w.Map.Standable(from)
		if err != nil {
			ew.t.Fatal(err)
		}
		if !standable {
			continue
		}
		for dir := range api.Dir(api.DirCount) {
			step, ok, err := ew.w.Map.Step(from, dir, rules)
			if err != nil {
				ew.t.Fatal(err)
			}
			if to := step.To.Cell; ok && hexmap.KeyOf(to) != key && bounds.Contains(to) {
				out[key] = append(out[key], worldgen.ChunkOpening{From: from, To: to, Cost: hexmap.CellDistance(from, to)})
			}
		}
	}
	for _, openings := range out {
		slices.SortFunc(openings, func(a, b worldgen.ChunkOpening) int {
			return cmp.Or(
				hexmap.CompareKeys(hexmap.KeyOf(a.To), hexmap.KeyOf(b.To)),
				hexmap.CompareCells(a.From, b.From),
				hexmap.CompareCells(a.To, b.To),
			)
		})
	}
	return out
}

func (ew *exportWorld) checkGraphMatchesTheMap(seed uint64, when string) {
	ew.t.Helper()
	res := ecs.NewResource[worldgen.Graph](ew.w.ECS)
	graph := res.Get()
	loaded := ew.loadedOpenings()
	keys := make([]hexmap.ChunkKey, 0, len(loaded))
	for key := range loaded {
		keys = append(keys, key)
	}
	slices.SortFunc(keys, hexmap.CompareKeys)
	total := 0
	for _, key := range keys {
		summary, err := graph.Summary(key)
		if err != nil {
			ew.t.Fatal(err)
		}
		if len(summary.Openings) == 0 && len(loaded[key]) == 0 {
			continue
		}
		if !reflect.DeepEqual(summary.Openings, loaded[key]) {
			ew.t.Fatalf("seed %d %s: chunk %v has %d openings in the graph, %d on the map",
				seed, when, key, len(summary.Openings), len(loaded[key]))
		}
		total += len(summary.Openings)
	}
	if total == 0 {
		ew.t.Fatalf("seed %d %s: no chunk has an opening", seed, when)
	}
}

// TestTheGraphHasTheOpeningsOfTheLoadedChunksBeforeAndAfterAnExport reads every cell of three small worlds
// twice, which takes over a second.
func TestTheGraphHasTheOpeningsOfTheLoadedChunksBeforeAndAfterAnExport(t *testing.T) {
	for _, seed := range []uint64{1, 2, 3} {
		ew := newExportWorld(t, seed)
		ew.checkGraphMatchesTheMap(seed, "at start")
		ew.w.ExportCount = 1
		ew.w.Bus.Publish(api.ExportBegan{Day: 30, Count: 1})
		if _, err := ew.w.Bus.Drain(ew.w); err != nil {
			t.Fatal(err)
		}
		ew.checkGraphMatchesTheMap(seed, "after an export")
	}
}

func TestTheSameSeedAndExportCountGiveTheSameGraph(t *testing.T) {
	graphOf := func() []worldgen.OpeningSummary {
		ew := newExportWorld(t, 1)
		res := ecs.NewResource[worldgen.Graph](ew.w.ECS)
		var keys []hexmap.ChunkKey
		for _, cell := range ew.cells {
			if key := hexmap.KeyOf(cell); !slices.Contains(keys, key) {
				keys = append(keys, key)
			}
		}
		if err := res.Get().Prepare(keys); err != nil {
			t.Fatal(err)
		}
		out := make([]worldgen.OpeningSummary, len(keys))
		for i, key := range keys {
			summary, err := res.Get().Summary(key)
			if err != nil {
				t.Fatal(err)
			}
			out[i] = summary
		}
		return out
	}
	if !reflect.DeepEqual(graphOf(), graphOf()) {
		t.Fatal("two worlds of seed 1 have different graphs")
	}
}
