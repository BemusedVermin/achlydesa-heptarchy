package hexmap

import (
	"iter"

	"github.com/fzipp/astar"
	hex "github.com/gravitton/hexagon"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// Path finds a walking route using Step in six directions, so it climbs, descends and drops by the same rules as
// the party. cost prices each step (nil prices every step at 1); the heuristic is hex distance plus depth
// difference. It gives up after maxNodes expansions, so a search in an infinite map always ends. The route
// includes both ends; ok is false when there is none. Routes stay on this map. The first error from loading a
// chunk is returned.
func (m *Map) Path(
	from, to api.Cell,
	rules StepRules,
	cost func(Step) float64,
	maxNodes int,
) ([]api.Cell, bool, error) {
	if cost == nil {
		cost = func(Step) float64 { return 1 }
	}
	graph := &walkGraph{m: m, rules: rules, maxNodes: maxNodes, steps: map[[2]api.Cell]Step{}}
	price := func(a, b api.Cell) float64 { return cost(graph.steps[[2]api.Cell{a, b}]) }
	heuristic := func(a, b api.Cell) float64 {
		h := hex.Pt(a.Hex.Q, a.Hex.R).DistanceTo(hex.Pt(b.Hex.Q, b.Hex.R))
		return float64(h + absInt(a.Depth-b.Depth))
	}
	route := astar.FindPath[api.Cell](graph, from, to, price, heuristic)
	if graph.err != nil {
		return nil, false, graph.err
	}
	if route == nil {
		return nil, false, nil
	}
	return route, true, nil
}

// walkGraph exposes the cells Step reaches as an A* graph.
type walkGraph struct {
	m        *Map
	rules    StepRules
	maxNodes int
	expanded int
	steps    map[[2]api.Cell]Step
	err      error
}

// Neighbours implements astar.Graph, trying the six directions in order.
func (g *walkGraph) Neighbours(n api.Cell) iter.Seq[api.Cell] {
	return func(yield func(api.Cell) bool) {
		if g.err != nil || g.expanded >= g.maxNodes {
			return
		}
		g.expanded++
		for dir := api.Dir(0); dir < api.DirCount; dir++ {
			step, ok, err := g.m.Step(n, dir, g.rules)
			if err != nil {
				g.err = err
				return
			}
			if !ok {
				continue
			}
			to, local := g.m.local(step.To)
			if !local {
				continue
			}
			g.steps[[2]api.Cell{n, to}] = step
			if !yield(to) {
				return
			}
		}
	}
}

func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
