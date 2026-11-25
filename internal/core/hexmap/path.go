package hexmap

import (
	"iter"

	"github.com/fzipp/astar"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// Path finds a walking route using Step in six directions, so it climbs, descends and drops by the same rules as
// the party. cost prices each step (nil prices every step at 1); the heuristic is hex distance plus depth
// difference. It gives up after maxNodes expansions, so a search in an infinite map always ends. The route
// includes both ends; ok is false when there is none. The first error from loading a
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
	heuristic := func(a, b api.Cell) float64 { return float64(CellDistance(a, b)) }
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
		for step, err := range g.m.legalSteps(n, g.rules) {
			if err != nil {
				g.err = err
				return
			}
			g.steps[[2]api.Cell{n, step.To.Cell}] = step
			if !yield(step.To.Cell) {
				return
			}
		}
	}
}

// Reachable finds every cell a body can walk to from the start, using Step in six directions like Path, within
// maxSteps steps. It returns each cell with the fewest steps that reach it; the start is included at 0. allow, if
// not nil, limits the search to cells it accepts (the start is always kept). The first
// error from loading a chunk is returned.
func (m *Map) Reachable(
	from api.Cell,
	rules StepRules,
	maxSteps int,
	allow func(api.Cell) bool,
) (map[api.Cell]int, error) {
	reached := map[api.Cell]int{from: 0}
	frontier := []api.Cell{from}
	for steps := 1; steps <= maxSteps && len(frontier) > 0; steps++ {
		var next []api.Cell
		for _, cell := range frontier {
			for step, err := range m.legalSteps(cell, rules) {
				if err != nil {
					return nil, err
				}
				to := step.To.Cell
				if _, seen := reached[to]; seen || (allow != nil && !allow(to)) {
					continue
				}
				reached[to] = steps
				next = append(next, to)
			}
		}
		frontier = next
	}
	return reached, nil
}

// legalSteps yields the legal steps from the cell in direction order. A loading error is
// yielded once, with a zero Step, and ends the sequence.
func (m *Map) legalSteps(from api.Cell, rules StepRules) iter.Seq2[Step, error] {
	return func(yield func(Step, error) bool) {
		for dir := range api.Dir(api.DirCount) {
			step, ok, err := m.Step(from, dir, rules)
			if err != nil {
				yield(Step{}, err)
				return
			}
			if !ok {
				continue
			}
			if !yield(step, nil) {
				return
			}
		}
	}
}

// DirTo finds the direction of the one legal step from a cell that ends in the target cell.
func (m *Map) DirTo(from, to api.Cell, rules StepRules) (api.Dir, bool, error) {
	for dir := range api.Dir(api.DirCount) {
		step, ok, err := m.Step(from, dir, rules)
		if err != nil {
			return 0, false, err
		}
		if !ok {
			continue
		}
		if step.To.Cell == to {
			return dir, true, nil
		}
	}
	return 0, false, nil
}
