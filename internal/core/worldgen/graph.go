package worldgen

import (
	"cmp"
	"errors"
	"fmt"
	"iter"
	"runtime"
	"slices"
	"sync"

	"github.com/fzipp/astar"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
)

// GraphConfig is the ground a [Graph] reads: the world's source as it builds chunks, with the bounds and edge a
// world map sets on it, walked by Rules.
type GraphConfig struct {
	Source      hexmap.Source
	Bounds      hexmap.Bounds
	Edge        hexmap.Terrain
	Rules       hexmap.StepRules
	ExportCount uint64
	// Workers is how many goroutines summarize chunks, or zero for one per CPU. The graph is the same whatever
	// the number.
	Workers int
}

// ChunkOpening is one step a body can take out of a chunk: From lies in the chunk, To in a neighbor.
type ChunkOpening struct {
	From, To api.Cell
	// Cost is the cells the step moves, across and up or down.
	Cost int
}

// OpeningSummary is every opening out of one chunk, ordered by the chunk To lies in ([hexmap.CompareKeys]),
// then by From and To ([hexmap.CompareCells]).
type OpeningSummary struct {
	Key      hexmap.ChunkKey
	Openings []ChunkOpening
}

// ChunkLink is an edge of the graph: at least one opening leads from a chunk into To.
type ChunkLink struct {
	To hexmap.ChunkKey
	// Cost is the shortest walk from the chunk's center through one of the openings to To's center, counted
	// in cells as [hexmap.CellDistance] counts them.
	Cost int
}

// Route is a way between two chunks: every chunk on it from first to last, and the sum of its links' costs.
type Route struct {
	Chunks []hexmap.ChunkKey
	Cost   int
}

// Summarize lists the openings out of a chunk. It is a pure function of cfg and key, reading the ground only
// through its own map, so it may run on any goroutine. A chunk with no cell inside the bounds has none.
func Summarize(cfg GraphConfig, key hexmap.ChunkKey) (OpeningSummary, error) {
	summary := OpeningSummary{Key: key}
	ground := hexmap.NewMap(cfg.Source, hexmap.Prisms{})
	ground.SetBounds(cfg.Bounds, cfg.Edge)
	reach := max(cfg.Rules.MaxStepUp, cfg.Rules.MaxStepDown, cfg.Rules.MaxDrop)
	for slot := range hexmap.ChunkCells {
		from := key.Cell(slot)
		if !nearFace(key, from, reach) || !cfg.Bounds.Contains(from) {
			continue
		}
		standable, err := ground.Standable(from)
		if err != nil {
			return OpeningSummary{}, fmt.Errorf("worldgen: graph: chunk %v: %w", key, err)
		}
		if !standable {
			continue
		}
		for dir := range api.Dir(api.DirCount) {
			step, ok, err := ground.Step(from, dir, cfg.Rules)
			if err != nil {
				return OpeningSummary{}, fmt.Errorf("worldgen: graph: chunk %v: %w", key, err)
			}
			to := step.To.Cell
			if !ok || hexmap.KeyOf(to) == key || !cfg.Bounds.Contains(to) {
				continue
			}
			summary.Openings = append(
				summary.Openings,
				ChunkOpening{From: from, To: to, Cost: hexmap.CellDistance(from, to)},
			)
		}
	}
	slices.SortStableFunc(summary.Openings, func(a, b ChunkOpening) int {
		return cmp.Or(
			hexmap.CompareKeys(hexmap.KeyOf(a.To), hexmap.KeyOf(b.To)),
			hexmap.CompareCells(a.From, b.From),
			hexmap.CompareCells(a.To, b.To),
		)
	})
	return summary, nil
}

// nearFace reports whether a step from cell could leave its chunk: one hex across from a side, or reach
// cells up or down from the top or bottom.
func nearFace(key hexmap.ChunkKey, cell api.Cell, reach int) bool {
	origin := key.Origin()
	q, r, depth := cell.Hex.Q-origin.Hex.Q, cell.Hex.R-origin.Hex.R, cell.Depth-origin.Depth
	last := hexmap.ChunkSize - 1
	return q == 0 || q == last || r == 0 || r == last || depth < reach || depth > last-reach
}

type graphKey struct {
	exportCount uint64
	key         hexmap.ChunkKey
}

// Graph is the world seen as chunks joined by the openings between them, for travel that never loads the
// ground. Each chunk's summary is worked out on first use and kept, keyed by export count and chunk key. Only
// the world's goroutine may call its methods; they spread the summarizing over other goroutines themselves.
type Graph struct {
	cfg       GraphConfig
	summaries map[graphKey]OpeningSummary
}

// NewGraph returns a graph over cfg with nothing summarized yet.
func NewGraph(cfg GraphConfig) *Graph {
	return &Graph{cfg: cfg, summaries: map[graphKey]OpeningSummary{}}
}

// Reset points the graph at a new source and export count, dropping every summary it kept.
func (g *Graph) Reset(source hexmap.Source, exportCount uint64) {
	g.cfg.Source, g.cfg.ExportCount = source, exportCount
	clear(g.summaries)
}

// Prepare summarizes every listed chunk not yet summarized. The summaries are computed on several goroutines
// and kept in [hexmap.CompareKeys] order.
func (g *Graph) Prepare(keys []hexmap.ChunkKey) error {
	var missing []hexmap.ChunkKey
	listed := map[hexmap.ChunkKey]bool{}
	for _, key := range keys {
		if _, ok := g.summaries[g.keyOf(key)]; !ok && !listed[key] {
			listed[key] = true
			missing = append(missing, key)
		}
	}
	slices.SortFunc(missing, hexmap.CompareKeys)
	if len(missing) == 0 {
		return nil
	}
	count := g.cfg.Workers
	if count <= 0 {
		count = runtime.GOMAXPROCS(0)
	}
	count = min(count, len(missing))
	results := make([]OpeningSummary, len(missing))
	errs := make([]error, count)
	var wg sync.WaitGroup
	for i := range count {
		from, to := i*len(missing)/count, (i+1)*len(missing)/count
		wg.Go(func() {
			for j, key := range missing[from:to] {
				summary, err := Summarize(g.cfg, key)
				if err != nil {
					errs[i] = err
					return
				}
				results[from+j] = summary
			}
		})
	}
	wg.Wait()
	if err := errors.Join(errs...); err != nil {
		return err
	}
	for i, key := range missing {
		g.summaries[g.keyOf(key)] = results[i]
	}
	return nil
}

func (g *Graph) keyOf(key hexmap.ChunkKey) graphKey {
	return graphKey{exportCount: g.cfg.ExportCount, key: key}
}

// Summary returns a chunk's openings. On a miss it summarizes the chunk's neighbors too, which a route through
// it is likely to ask for next.
func (g *Graph) Summary(key hexmap.ChunkKey) (OpeningSummary, error) {
	if summary, ok := g.summaries[g.keyOf(key)]; ok {
		return summary, nil
	}
	if err := g.Prepare(append(chunkNeighbors(key), key)); err != nil {
		return OpeningSummary{}, err
	}
	return g.summaries[g.keyOf(key)], nil
}

// Neighbors lists the chunks an opening leads into from key, in [hexmap.CompareKeys] order.
func (g *Graph) Neighbors(key hexmap.ChunkKey) ([]ChunkLink, error) {
	summary, err := g.Summary(key)
	if err != nil {
		return nil, err
	}
	var links []ChunkLink
	from := center(key)
	for _, opening := range summary.Openings {
		into := hexmap.KeyOf(opening.To)
		cost := hexmap.CellDistance(from, opening.From) + opening.Cost + hexmap.CellDistance(opening.To, center(into))
		if last := len(links) - 1; last >= 0 && links[last].To == into {
			links[last].Cost = min(links[last].Cost, cost)
			continue
		}
		links = append(links, ChunkLink{To: into, Cost: cost})
	}
	return links, nil
}

// center is the cell in the middle of a chunk.
func center(key hexmap.ChunkKey) api.Cell {
	origin := key.Origin()
	half := hexmap.ChunkSize / 2
	return api.Cell{Hex: api.Hex{Q: origin.Hex.Q + half, R: origin.Hex.R + half}, Depth: origin.Depth + half}
}

// Route finds the cheapest way from one chunk to another by A*. ok is false when no way joins them. A chunk is
// taken to join all its openings, so the cost is an estimate of the walk, never less than the straight
// distance between the two centers.
func (g *Graph) Route(from, to hexmap.ChunkKey) (Route, bool, error) {
	// TODO(M6-06.F1): nodes per walkable area of a chunk, and a reach limit.
	search := &chunkSearch{graph: g, costs: map[[2]hexmap.ChunkKey]int{}}
	price := func(a, b hexmap.ChunkKey) float64 { return float64(search.costs[[2]hexmap.ChunkKey{a, b}]) }
	heuristic := func(a, b hexmap.ChunkKey) float64 { return float64(hexmap.CellDistance(center(a), center(b))) }
	path := astar.FindPath[hexmap.ChunkKey](search, from, to, price, heuristic)
	if search.err != nil {
		return Route{}, false, search.err
	}
	if path == nil {
		return Route{}, false, nil
	}
	route := Route{Chunks: path}
	for i := 1; i < len(path); i++ {
		route.Cost += search.costs[[2]hexmap.ChunkKey{path[i-1], path[i]}]
	}
	return route, true, nil
}

// chunkSearch exposes the graph to astar, keeping each link's cost for the price function.
type chunkSearch struct {
	graph *Graph
	costs map[[2]hexmap.ChunkKey]int
	err   error
}

// Neighbours implements astar.Graph.
func (s *chunkSearch) Neighbours(key hexmap.ChunkKey) iter.Seq[hexmap.ChunkKey] {
	return func(yield func(hexmap.ChunkKey) bool) {
		if s.err != nil {
			return
		}
		links, err := s.graph.Neighbors(key)
		if err != nil {
			s.err = err
			return
		}
		for _, link := range links {
			s.costs[[2]hexmap.ChunkKey{key, link.To}] = link.Cost
			if !yield(link.To) {
				return
			}
		}
	}
}
