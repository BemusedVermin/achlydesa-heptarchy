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
	// FromArea is the local area of From in its chunk, and ToArea that of To in the neighbor. [Summarize]
	// leaves ToArea at -1; [Graph.Summary] fills it in from the neighbor's summary.
	FromArea, ToArea int
}

// AreaJoin is a one-way step inside a chunk from one of its local areas to another, such as a drop.
type AreaJoin struct {
	From, To int
	Cost     int
}

// BorderCell is a standable cell near a face of its chunk, which a step into or out of the chunk may use.
type BorderCell struct {
	Cell api.Cell
	Area int
}

// OpeningSummary is every opening out of one chunk, ordered by the chunk To lies in ([hexmap.CompareKeys]),
// then by From and To ([hexmap.CompareCells]). A local area is the cells of the chunk a body can walk between
// both ways without leaving it.
type OpeningSummary struct {
	Key      hexmap.ChunkKey
	Openings []ChunkOpening
	// Areas is how many local areas the chunk has, numbered from 0 in the order of their first slot.
	Areas int
	// Joins are ordered by From, then To, and hold the cheapest step of each pair.
	Joins []AreaJoin
	// Border is in [hexmap.CompareCells] order. Only these cells keep their area, not every slot.
	Border []BorderCell
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

// Summarize lists the openings out of a chunk and splits its walkable cells into local areas. It is a pure
// function of cfg and key, reading the ground only through its own map, so it may run on any goroutine. A
// chunk with no cell inside the bounds has none.
func Summarize(cfg GraphConfig, key hexmap.ChunkKey) (OpeningSummary, error) {
	ground := hexmap.NewMap(cfg.Source, hexmap.Prisms{})
	ground.SetBounds(cfg.Bounds, cfg.Edge)
	scan := chunkScan{cfg: cfg, ground: ground, key: key, summary: OpeningSummary{Key: key}}
	if err := scan.flood(); err != nil {
		return OpeningSummary{}, fmt.Errorf("worldgen: graph: chunk %v: %w", key, err)
	}
	return scan.finish(), nil
}

type oneWay struct {
	from int
	to   api.Cell
	cost int
}

// chunkScan floods one chunk into local areas, collecting its openings and one-way steps on the way.
type chunkScan struct {
	cfg     GraphConfig
	ground  *hexmap.Map
	key     hexmap.ChunkKey
	table   []int16
	oneWays []oneWay
	summary OpeningSummary
}

func (s *chunkScan) flood() error {
	s.table = make([]int16, hexmap.ChunkCells)
	for slot := range s.table {
		s.table[slot] = -1
	}
	for slot := range hexmap.ChunkCells {
		start := s.key.Cell(slot)
		if s.table[slot] >= 0 || !s.cfg.Bounds.Contains(start) {
			continue
		}
		standable, err := s.ground.Standable(start)
		if err != nil {
			return err
		}
		if !standable {
			continue
		}
		area := s.summary.Areas
		s.summary.Areas++
		s.table[slot] = int16(area)
		members := []api.Cell{start}
		for head := 0; head < len(members); head++ {
			reached, err := s.visit(members[head], area)
			if err != nil {
				return err
			}
			members = append(members, reached...)
		}
	}
	return nil
}

// visit tries the six steps from a cell of the area. A step out of the chunk is an opening; one that can be
// walked back joins its cell to the area; any other is a one-way step, resolved once every area is known.
func (s *chunkScan) visit(from api.Cell, area int) ([]api.Cell, error) {
	var reached []api.Cell
	for dir := range api.Dir(api.DirCount) {
		step, ok, err := s.ground.Step(from, dir, s.cfg.Rules)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		to := step.To.Cell
		cost := hexmap.CellDistance(from, to)
		if hexmap.KeyOf(to) != s.key {
			if s.cfg.Bounds.Contains(to) {
				s.summary.Openings = append(s.summary.Openings, ChunkOpening{
					From: from, To: to, Cost: cost, FromArea: area, ToArea: -1,
				})
			}
			continue
		}
		if s.table[hexmap.SlotOf(to)] >= 0 {
			if int(s.table[hexmap.SlotOf(to)]) != area {
				s.oneWays = append(s.oneWays, oneWay{from: area, to: to, cost: cost})
			}
			continue
		}
		back := false
		if step.Kind != hexmap.StepDrop {
			reverse, ok, err := s.ground.Step(to, (dir+api.DirCount/2)%api.DirCount, s.cfg.Rules)
			if err != nil {
				return nil, err
			}
			back = ok && reverse.To.Cell == from
		}
		if !back {
			s.oneWays = append(s.oneWays, oneWay{from: area, to: to, cost: cost})
			continue
		}
		s.table[hexmap.SlotOf(to)] = int16(area)
		reached = append(reached, to)
	}
	return reached, nil
}

func (s *chunkScan) finish() OpeningSummary {
	summary := s.summary
	slices.SortStableFunc(summary.Openings, func(a, b ChunkOpening) int {
		return cmp.Or(
			hexmap.CompareKeys(hexmap.KeyOf(a.To), hexmap.KeyOf(b.To)),
			hexmap.CompareCells(a.From, b.From),
			hexmap.CompareCells(a.To, b.To),
		)
	})
	for _, way := range s.oneWays {
		if to := int(s.table[hexmap.SlotOf(way.to)]); to >= 0 && to != way.from {
			summary.Joins = append(summary.Joins, AreaJoin{From: way.from, To: to, Cost: way.cost})
		}
	}
	slices.SortFunc(summary.Joins, func(a, b AreaJoin) int {
		return cmp.Or(cmp.Compare(a.From, b.From), cmp.Compare(a.To, b.To), cmp.Compare(a.Cost, b.Cost))
	})
	summary.Joins = slices.CompactFunc(summary.Joins, func(a, b AreaJoin) bool {
		return a.From == b.From && a.To == b.To
	})
	reach := max(s.cfg.Rules.MaxStepUp, s.cfg.Rules.MaxStepDown, s.cfg.Rules.MaxDrop)
	for slot, area := range s.table {
		if cell := s.key.Cell(slot); area >= 0 && nearFace(s.key, cell, reach) {
			summary.Border = append(summary.Border, BorderCell{Cell: cell, Area: int(area)})
		}
	}
	slices.SortFunc(summary.Border, func(a, b BorderCell) int { return hexmap.CompareCells(a.Cell, b.Cell) })
	return summary
}

// areaOf is the local area of a border cell, or -1 when the cell is not one.
func (summary OpeningSummary) areaOf(cell api.Cell) int {
	at, found := slices.BinarySearchFunc(summary.Border, cell, func(b BorderCell, c api.Cell) int {
		return hexmap.CompareCells(b.Cell, c)
	})
	if !found {
		return -1
	}
	return summary.Border[at].Area
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

// graphEntry is a chunk's summary, and whether the areas its openings lead into have been looked up.
type graphEntry struct {
	summary  OpeningSummary
	resolved bool
}

// Graph is the world seen as chunks joined by the openings between them, for travel that never loads the
// ground. Each chunk's summary is worked out on first use and kept, keyed by export count and chunk key. Only
// the world's goroutine may call its methods; they spread the summarizing over other goroutines themselves.
type Graph struct {
	cfg       GraphConfig
	summaries map[graphKey]*graphEntry
}

// NewGraph returns a graph over cfg with nothing summarized yet.
func NewGraph(cfg GraphConfig) *Graph {
	return &Graph{cfg: cfg, summaries: map[graphKey]*graphEntry{}}
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
		g.summaries[g.keyOf(key)] = &graphEntry{summary: results[i]}
	}
	return nil
}

func (g *Graph) keyOf(key hexmap.ChunkKey) graphKey {
	return graphKey{exportCount: g.cfg.ExportCount, key: key}
}

// Summary returns a chunk's openings, each with the area it leads into. On a miss it summarizes the chunk's
// neighbors too, which it needs for those areas and a route through it is likely to ask for next.
func (g *Graph) Summary(key hexmap.ChunkKey) (OpeningSummary, error) {
	entry, ok := g.summaries[g.keyOf(key)]
	if ok && entry.resolved {
		return entry.summary, nil
	}
	if err := g.Prepare(append(chunkNeighbors(key), key)); err != nil {
		return OpeningSummary{}, err
	}
	entry = g.summaries[g.keyOf(key)]
	into := make([]hexmap.ChunkKey, len(entry.summary.Openings))
	for i, opening := range entry.summary.Openings {
		into[i] = hexmap.KeyOf(opening.To)
	}
	if err := g.Prepare(into); err != nil {
		return OpeningSummary{}, err
	}
	for i := range entry.summary.Openings {
		opening := &entry.summary.Openings[i]
		if into, ok := g.summaries[g.keyOf(hexmap.KeyOf(opening.To))]; ok {
			opening.ToArea = into.summary.areaOf(opening.To)
		}
	}
	entry.resolved = true
	return entry.summary, nil
}

// Neighbors lists the chunks an opening leads into from key, in [hexmap.CompareKeys] order, whatever local
// area of key each opening leaves from.
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

// Center is the cell in the middle of a chunk, which a route's costs are counted from. It may be solid.
func Center(key hexmap.ChunkKey) api.Cell { return center(key) }

// areaNode is a local area of a chunk. Area -1 stands for the whole chunk, the start or end of a route.
type areaNode struct {
	key  hexmap.ChunkKey
	area int
}

// Route finds the cheapest way from one chunk to another by A* over the chunks' local areas, so a route
// enters and leaves a chunk only where its areas join. It may start from any area of from. ok is false when
// no way joins them, or when maxNodes areas have been searched without finding one; maxNodes of 0 or less
// searches without limit. The cost is an estimate of the walk, counted from chunk center to chunk center,
// never less than the straight distance between the two centers.
func (g *Graph) Route(from, to hexmap.ChunkKey, maxNodes int) (Route, bool, error) {
	start, dest := areaNode{key: from, area: -1}, areaNode{key: to, area: -1}
	search := &areaSearch{graph: g, start: start, dest: dest, maxNodes: maxNodes, costs: map[[2]areaNode]int{}}
	price := func(a, b areaNode) float64 { return float64(search.costs[[2]areaNode{a, b}]) }
	heuristic := func(a, b areaNode) float64 { return float64(hexmap.CellDistance(center(a.key), center(b.key))) }
	path := astar.FindPath[areaNode](search, start, dest, price, heuristic)
	if search.err != nil {
		return Route{}, false, search.err
	}
	if path == nil {
		return Route{}, false, nil
	}
	route := Route{Chunks: []hexmap.ChunkKey{from}}
	for i := 1; i < len(path); i++ {
		route.Cost += search.costs[[2]areaNode{path[i-1], path[i]}]
		if key := path[i].key; key != route.Chunks[len(route.Chunks)-1] {
			route.Chunks = append(route.Chunks, key)
		}
	}
	return route, true, nil
}

// areaSearch exposes the areas to astar, keeping each link's cost for the price function.
type areaSearch struct {
	graph       *Graph
	start, dest areaNode
	maxNodes    int
	expanded    int
	costs       map[[2]areaNode]int
	err         error
}

// Neighbours implements astar.Graph.
func (s *areaSearch) Neighbours(node areaNode) iter.Seq[areaNode] {
	return func(yield func(areaNode) bool) {
		if s.err != nil || (s.maxNodes > 0 && s.expanded >= s.maxNodes) {
			return
		}
		s.expanded++
		links, err := s.links(node)
		if err != nil {
			s.err = err
			return
		}
		for _, link := range links {
			s.costs[[2]areaNode{node, link.to}] = link.cost
			if !yield(link.to) {
				return
			}
		}
	}
}

type areaLink struct {
	to   areaNode
	cost int
}

// links lists where a step leads from an area: into a neighbor's area through an opening, into another area
// of the chunk by a one-way step, and to the route's end when the area lies in its chunk. The start leads
// into each area of its chunk.
func (s *areaSearch) links(node areaNode) ([]areaLink, error) {
	if node.area < 0 && node != s.start {
		return nil, nil
	}
	summary, err := s.graph.Summary(node.key)
	if err != nil {
		return nil, err
	}
	var links []areaLink
	if node.area < 0 {
		for area := range summary.Areas {
			links = append(links, areaLink{to: areaNode{key: node.key, area: area}})
		}
		return links, nil
	}
	if node.key == s.dest.key {
		links = append(links, areaLink{to: s.dest})
	}
	for _, join := range summary.Joins {
		if join.From == node.area {
			links = append(links, areaLink{to: areaNode{key: node.key, area: join.To}, cost: join.Cost})
		}
	}
	from := center(node.key)
	for _, opening := range summary.Openings {
		if opening.FromArea != node.area || opening.ToArea < 0 {
			continue
		}
		into := hexmap.KeyOf(opening.To)
		cost := hexmap.CellDistance(from, opening.From) + opening.Cost + hexmap.CellDistance(opening.To, center(into))
		next := areaNode{key: into, area: opening.ToArea}
		if at := slices.IndexFunc(links, func(l areaLink) bool { return l.to == next }); at >= 0 {
			links[at].cost = min(links[at].cost, cost)
			continue
		}
		links = append(links, areaLink{to: next, cost: cost})
	}
	return links, nil
}
