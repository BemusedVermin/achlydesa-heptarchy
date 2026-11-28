package worldgen_test

import (
	"reflect"
	"slices"
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/worldgen"
)

var (
	rock = hexmap.Terrain{ID: "rock", Solid: true, Opaque: true}
	air  = hexmap.Terrain{ID: "air"}
)

// tunnelConfig is solid rock with a tunnel at depth 5 along R = 5 through chunks Q 0 to 2, and a sealed chamber
// in chunk (0, 1, 0).
func tunnelConfig(workers int) worldgen.GraphConfig {
	ground := hexmap.FuncSource(func(cell api.Cell) hexmap.Terrain {
		tunnel := cell.Hex.R == 5 && cell.Hex.Q >= 5 && cell.Hex.Q <= 40
		chamber := cell.Hex.R >= 20 && cell.Hex.R <= 22 && cell.Hex.Q >= 5 && cell.Hex.Q <= 8
		if cell.Depth == 5 && (tunnel || chamber) {
			return air
		}
		return rock
	})
	return worldgen.GraphConfig{
		Source:  ground,
		Bounds:  hexmap.Bounds{MaxCol: 70, MaxRow: 40, MaxDepth: 20},
		Edge:    rock,
		Rules:   hexmap.StepRules{MaxStepUp: 1, MaxStepDown: 1, MaxDrop: 3},
		Workers: workers,
	}
}

func TestARouteFollowsTheTunnelBetweenChunks(t *testing.T) {
	graph := worldgen.NewGraph(tunnelConfig(0))
	west, middle, east := hexmap.ChunkKey{}, hexmap.ChunkKey{Q: 1}, hexmap.ChunkKey{Q: 2}
	route, ok, err := graph.Route(west, east, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("no route along the tunnel")
	}
	if want := []hexmap.ChunkKey{west, middle, east}; !slices.Equal(route.Chunks, want) {
		t.Fatalf("route %v, want %v", route.Chunks, want)
	}
	if route.Cost != 50 {
		t.Fatalf("route cost %d, want 50: two links of 10 to the opening, 1 across and 14 on", route.Cost)
	}
	links, err := graph.Neighbors(middle)
	if err != nil {
		t.Fatal(err)
	}
	var linked []hexmap.ChunkKey
	for _, link := range links {
		linked = append(linked, link.To)
	}
	if !slices.Equal(linked, []hexmap.ChunkKey{west, east}) {
		t.Fatalf("middle chunk links %v, want the west and east chunks", links)
	}
	if _, ok, err := graph.Route(west, hexmap.ChunkKey{R: 1}, 0); err != nil || ok {
		t.Fatalf("route into the sealed chamber: ok %v, err %v; want none", ok, err)
	}
}

func TestTheGraphIsTheSameColdOrWarmAndWithAnyWorkers(t *testing.T) {
	var keys []hexmap.ChunkKey
	for r := range 3 {
		for q := range 4 {
			keys = append(keys, hexmap.ChunkKey{Q: q, R: r})
		}
	}
	summaries := func(graph *worldgen.Graph) []worldgen.OpeningSummary {
		var out []worldgen.OpeningSummary
		for _, key := range keys {
			summary, err := graph.Summary(key)
			if err != nil {
				t.Fatal(err)
			}
			out = append(out, summary)
		}
		return out
	}
	cold := worldgen.NewGraph(tunnelConfig(1))
	coldRoute, _, err := cold.Route(hexmap.ChunkKey{}, hexmap.ChunkKey{Q: 2}, 0)
	if err != nil {
		t.Fatal(err)
	}
	warm := worldgen.NewGraph(tunnelConfig(8))
	if err := warm.Prepare(keys); err != nil {
		t.Fatal(err)
	}
	warmRoute, _, err := warm.Route(hexmap.ChunkKey{}, hexmap.ChunkKey{Q: 2}, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(coldRoute, warmRoute) {
		t.Fatalf("cold route %v, warm route %v", coldRoute, warmRoute)
	}
	if !reflect.DeepEqual(summaries(cold), summaries(warm)) {
		t.Fatal("summaries differ between one worker and eight")
	}
}

func TestARouteNeverCrossesAChunkBetweenAreasThatDoNotMeetInIt(t *testing.T) {
	cfg := tunnelConfig(0)
	cfg.Source = hexmap.FuncSource(func(cell api.Cell) hexmap.Terrain {
		if cell.Depth == 5 && cell.Hex.R == 5 && cell.Hex.Q >= 5 && cell.Hex.Q <= 40 && cell.Hex.Q != 24 {
			return air
		}
		return rock
	})
	graph := worldgen.NewGraph(cfg)
	west, middle, east := hexmap.ChunkKey{}, hexmap.ChunkKey{Q: 1}, hexmap.ChunkKey{Q: 2}
	summary, err := graph.Summary(middle)
	if err != nil {
		t.Fatal(err)
	}
	if summary.Areas != 2 {
		t.Fatalf("middle chunk has %d areas, want 2 either side of the wall", summary.Areas)
	}
	if _, ok, err := graph.Route(west, east, 0); err != nil || ok {
		t.Fatalf("route across the walled middle chunk: ok %v, err %v; want none", ok, err)
	}
	if route, ok, err := graph.Route(west, middle, 0); err != nil || !ok || len(route.Chunks) != 2 {
		t.Fatalf("route into the middle chunk's west area: %v, ok %v, err %v", route, ok, err)
	}
}

func TestARouteWithANodeLimitStopsSearchingForAnUnreachableChunk(t *testing.T) {
	cfg := tunnelConfig(0)
	cfg.Source = hexmap.FuncSource(func(cell api.Cell) hexmap.Terrain {
		floor := cell.Depth == 5
		chamber := cell.Depth == 20 && cell.Hex.R == 8 && cell.Hex.Q >= 8 && cell.Hex.Q <= 10
		if floor || chamber {
			return air
		}
		return rock
	})
	cfg.Bounds = hexmap.Bounds{MaxCol: 400, MaxRow: 400, MaxDepth: 40}
	graph := worldgen.NewGraph(cfg)
	if _, ok, err := graph.Route(hexmap.ChunkKey{Q: 5, R: 5}, hexmap.ChunkKey{Depth: 1}, 10); err != nil || ok {
		t.Fatalf("route into the sealed chamber: ok %v, err %v; want none", ok, err)
	}
	if count := worldgen.SummaryCount(graph); count > 100 {
		t.Fatalf("the search summarized %d chunks, want it to stop near its limit of 10 areas", count)
	}
}

func TestARouteFromACellStartsInThatCellsArea(t *testing.T) {
	cfg := tunnelConfig(0)
	cfg.Source = hexmap.FuncSource(func(cell api.Cell) hexmap.Terrain {
		if cell.Depth == 5 && cell.Hex.R == 5 && cell.Hex.Q >= 5 && cell.Hex.Q <= 40 && cell.Hex.Q != 24 {
			return air
		}
		return rock
	})
	graph := worldgen.NewGraph(cfg)
	east := hexmap.ChunkKey{Q: 2}
	west := api.Cell{Hex: api.Hex{Q: hexmap.ChunkSize, R: 5}, Depth: 5}
	if _, ok, err := graph.RouteFrom(west, east, 0); err != nil || ok {
		t.Fatalf("route out of the middle chunk's west area: ok %v, err %v; want none", ok, err)
	}
	inEast := api.Cell{Hex: api.Hex{Q: 2*hexmap.ChunkSize - 1, R: 5}, Depth: 5}
	route, ok, err := graph.RouteFrom(inEast, east, 0)
	if err != nil || !ok {
		t.Fatalf("route out of the middle chunk's east area: ok %v, err %v", ok, err)
	}
	if len(route.Entries) != len(route.Chunks)-1 {
		t.Fatalf("%d entries for %d chunks", len(route.Entries), len(route.Chunks))
	}
}
