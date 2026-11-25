package worldgen

import (
	"cmp"
	"errors"
	"fmt"
	"maps"
	"math"
	"slices"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
)

// VerifyStage is the name the verify stage registers under.
const VerifyStage = "verify"

var verifyParams = []string{"max_step_up", "max_step_down", "min_pocket", "max_chunks"}

// maxRepairRounds bounds the repair loop beyond one round per area found at the start; a world that needs
// more is reported as a bug.
const maxRepairRounds = 256

// corridorWidths are how many chunks around the chunk path a corridor search may use, widest last; zero
// stands for the whole ground.
var corridorWidths = []int{1, 2, 4, 0}

// RegisterVerify adds the verify stage to stages.
//
// The stage checks the ground the earlier stages leave against the guarantees and repairs it: every place's
// openings lead into one walkable whole, and no pocket of walkable ground smaller than min_pocket is left.
// Ground is walkable by steps a body can make both ways, so a drop never joins two areas. A small pocket is
// filled; a larger one that is cut off is joined to the nearest part of the rest by a corridor carved along
// the shortest chain of chunks between them. A repair never changes a place's cells, and never opens the
// ground a walkable cell stands on. A place's ground that no opening leads to is left alone. The stage must be
// the last of its geography, since it reads the result of every stage before it. Every param is required:
//
//   - max_step_up, max_step_down: how far a body climbs or descends in one step, as in the party's rules.
//   - min_pocket: walkable cells a pocket needs to be kept; smaller ones are filled.
//   - max_chunks: chunks the check holds in memory at once; it reads the ground one chunk at a time and
//     keeps only a summary of each.
func RegisterVerify(stages *Stages) error {
	return stages.Register(VerifyStage, verify{}, checkVerifyParams)
}

func checkVerifyParams(params map[string]float64) error {
	for _, name := range verifyParams {
		if _, ok := params[name]; !ok {
			return fmt.Errorf("missing param %q", name)
		}
	}
	for name := range params {
		if !slices.Contains(verifyParams, name) {
			return fmt.Errorf("takes no param %q", name)
		}
	}
	for _, name := range verifyParams {
		if params[name] < 1 || params[name] != math.Trunc(params[name]) {
			return fmt.Errorf("param %q must be a whole number of at least 1", name)
		}
	}
	return nil
}

type verify struct{}

// Run applies the chunk's part of the world's repair. The repair is worked out once from the ground before
// this stage, a pure function of the seed, so every chunk agrees on it whatever order chunks load in.
func (verify) Run(ctx *Context, chunk *hexmap.Chunk) error {
	value, err := ctx.Shared(func() (any, error) { return repairWorld(ctx) })
	if err != nil {
		return err
	}
	origin := ctx.Key.Origin()
	for _, edit := range value.(*repair).edits[ctx.Key] {
		chunk.Set(
			edit.Cell.Hex.Q-origin.Hex.Q,
			edit.Cell.Hex.R-origin.Hex.R,
			edit.Cell.Depth-origin.Depth,
			edit.Terrain,
		)
	}
	return nil
}

// repair is the verify stage's result: the edits each chunk applies, and the most chunks it held at once.
type repair struct {
	edits      map[hexmap.ChunkKey][]hexmap.Edit
	peakChunks int
}

// localArea is the cells of one chunk a body can cross between both ways without leaving the chunk.
type localArea struct {
	size    int
	inPlace bool
	// cells is kept only for an area smaller than a pocket, which may need filling.
	cells []api.Cell
}

// crossing is a both-ways step from a local area into another chunk.
type crossing struct {
	area int
	to   api.Cell
}

// chunkSummary is what the check keeps of a chunk once it has read it.
type chunkSummary struct {
	areas     []localArea
	crossings []crossing
	// border holds the local area of every cell a crossing leaves from.
	border map[api.Cell]int
	// node is the graph node of the chunk's first local area; the others follow it.
	node int
}

// node is a local area of one chunk.
type node struct {
	key  hexmap.ChunkKey
	area int
}

// component is the local areas joined by crossings: the walkable areas of the whole ground.
type component struct {
	nodes   []node
	size    int
	inPlace bool
}

type repairer struct {
	ctx    *Context
	ground *hexmap.Map
	rules  hexmap.StepRules
	pocket int
	budget int
	// loaded is the chunks the ground holds, oldest first; peak is the most it held at once.
	loaded     []hexmap.ChunkKey
	peak       int
	keys []hexmap.ChunkKey
	// sweep is keys column by column: a chunk's summary reads its neighbors, and a column's neighbors are
	// read again soon after, while they are still held.
	sweep      []hexmap.ChunkKey
	summaries  map[hexmap.ChunkKey]*chunkSummary
	dirty      map[hexmap.ChunkKey]bool
	tables     map[hexmap.ChunkKey][]int32
	fixed      map[hexmap.ChunkKey]*hexmap.Chunk
	components []component
	// componentOf is indexed by node.
	componentOf []int
}

// boundedSource loads chunks for the repairer's ground, first dropping the chunk loaded longest ago so that
// the ground never holds more than its budget.
type boundedSource struct{ r *repairer }

// Chunk implements [hexmap.Source]. The ground calls it from inside a read, so dropping another chunk here is
// safe: a read holds no chunk past its return.
func (s boundedSource) Chunk(key hexmap.ChunkKey) (*hexmap.Chunk, error) {
	r := s.r
	for len(r.loaded) >= r.budget {
		oldest := r.loaded[0]
		r.loaded = r.loaded[1:]
		r.ground.Unload(func(other hexmap.ChunkKey) bool { return other != oldest })
	}
	chunk, err := r.ctx.Before.Chunk(key)
	if err != nil {
		return nil, err
	}
	dbgLoads++
	r.loaded = append(r.loaded, key)
	r.peak = max(r.peak, len(r.loaded))
	return chunk, nil
}

// repairWorld builds the repair as the edits each chunk applies.
func repairWorld(ctx *Context) (*repair, error) {
	r := &repairer{
		ctx: ctx,
		rules: hexmap.StepRules{
			MaxStepUp:   int(ctx.Params["max_step_up"]),
			MaxStepDown: int(ctx.Params["max_step_down"]),
			MaxDrop:     int(ctx.Params["max_step_down"]),
		},
		pocket:    int(ctx.Params["min_pocket"]),
		budget:    int(ctx.Params["max_chunks"]),
		keys:      boundsChunks(ctx.Bounds),
		summaries: map[hexmap.ChunkKey]*chunkSummary{},
		dirty:     map[hexmap.ChunkKey]bool{},
		tables:    map[hexmap.ChunkKey][]int32{},
		fixed:     map[hexmap.ChunkKey]*hexmap.Chunk{},
	}
	r.ground = hexmap.NewMap(boundedSource{r: r}, hexmap.Prisms{})
	r.ground.SetBounds(ctx.Bounds, hexmap.Terrain{Solid: true, Opaque: true})
	r.sweep = slices.Clone(r.keys)
	slices.SortFunc(r.sweep, func(a, b hexmap.ChunkKey) int {
		return cmp.Or(cmp.Compare(a.R, b.R), cmp.Compare(a.Q, b.Q), cmp.Compare(a.Depth, b.Depth))
	})
	for _, key := range r.keys {
		r.dirty[key] = true
	}
	if err := r.settle(); err != nil {
		return nil, err
	}
	byChunk := map[hexmap.ChunkKey][]hexmap.Edit{}
	for _, edit := range r.ground.Edits() {
		key := hexmap.KeyOf(edit.Cell)
		byChunk[key] = append(byChunk[key], edit)
	}
	return &repair{edits: byChunk, peakChunks: r.peak}, nil
}

// boundsChunks lists every chunk with a cell inside the bounds, in [hexmap.CompareKeys] order.
func boundsChunks(bounds hexmap.Bounds) []hexmap.ChunkKey {
	columns := map[hexmap.ChunkKey]bool{}
	for row := bounds.MinRow; row <= bounds.MaxRow; row++ {
		for col := bounds.MinCol; col <= bounds.MaxCol; col++ {
			key := hexmap.KeyOf(api.Cell{Hex: hexmap.FromOffsetOddR(col, row)})
			columns[hexmap.ChunkKey{Q: key.Q, R: key.R}] = true
		}
	}
	var keys []hexmap.ChunkKey
	first := hexmap.KeyOf(api.Cell{Depth: bounds.MinDepth}).Depth
	last := hexmap.KeyOf(api.Cell{Depth: bounds.MaxDepth}).Depth
	for column := range columns {
		for depth := first; depth <= last; depth++ {
			keys = append(keys, hexmap.ChunkKey{Q: column.Q, R: column.R, Depth: depth})
		}
	}
	slices.SortFunc(keys, hexmap.CompareKeys)
	return keys
}

// settle repairs the ground until a round finds nothing to change.
func (r *repairer) settle() error {
	rounds := maxRepairRounds
	for round := 0; round < rounds; round++ {
		if err := r.summarize(); err != nil {
			return err
		}
		if round == 0 {
			rounds += len(r.components)
		}
		anchored, opened, err := r.anchorComponents()
		if err != nil {
			return err
		}
		if opened {
			continue
		}
		keep := make([]bool, len(r.components))
		var pockets []api.Cell
		for i, found := range r.components {
			anchor := slices.Contains(anchored, i)
			keep[i] = anchor || !found.inPlace && found.size >= r.pocket
			if keep[i] || found.inPlace {
				continue
			}
			for _, n := range found.nodes {
				pockets = append(pockets, r.summaries[n.key].areas[n.area].cells...)
			}
		}
		for _, cell := range pockets {
			if err := r.set(cell, r.fillAt(cell)); err != nil {
				return err
			}
		}
		if len(pockets) > 0 {
			continue
		}
		main := r.mainComponent(anchored)
		targets := map[int]bool{}
		for i := range r.components {
			if keep[i] && i != main {
				targets[i] = true
			}
		}
		if len(targets) == 0 {
			return nil
		}
		dbg("round %d comps %d targets %d loads %d", round, len(r.components), len(targets), dbgLoads)
		if err := r.join(main, targets); err != nil {
			return err
		}
	}
	return errors.New("verify: the repair did not settle")
}

// summarize reads every chunk changed since the last call and joins the summaries into components. Nodes and
// components are numbered in chunk order, so the numbers never depend on map order.
func (r *repairer) summarize() error {
	for _, key := range r.sweep {
		if !r.dirty[key] {
			continue
		}
		summary, table, err := r.scan(key)
		if err != nil {
			return err
		}
		r.summaries[key] = summary
		r.keepTable(key, table)
	}
	dbg("summarized %d dirty, loads %d", len(r.dirty), dbgLoads)
	clear(r.dirty)
	nodes := 0
	for _, key := range r.keys {
		r.summaries[key].node = nodes
		nodes += len(r.summaries[key].areas)
	}
	parent := make([]int, nodes)
	for i := range parent {
		parent[i] = i
	}
	root := func(i int) int {
		for parent[i] != i {
			parent[i] = parent[parent[i]]
			i = parent[i]
		}
		return i
	}
	for _, key := range r.keys {
		summary := r.summaries[key]
		for _, cross := range summary.crossings {
			other, ok, err := r.nodeOf(cross.to)
			if err != nil {
				return err
			}
			if !ok {
				continue
			}
			a, b := root(summary.node+cross.area), root(other)
			parent[max(a, b)] = min(a, b)
		}
	}
	r.components = r.components[:0]
	r.componentOf = make([]int, nodes)
	index := map[int]int{}
	for _, key := range r.keys {
		summary := r.summaries[key]
		for area, local := range summary.areas {
			top := root(summary.node + area)
			i, seen := index[top]
			if !seen {
				i = len(r.components)
				index[top] = i
				r.components = append(r.components, component{})
			}
			r.componentOf[summary.node+area] = i
			found := &r.components[i]
			found.nodes = append(found.nodes, node{key: key, area: area})
			found.size += local.size
			found.inPlace = found.inPlace || local.inPlace
		}
	}
	return nil
}

// scan flood-fills the walkable cells of one chunk in slot order, over crossings inside the chunk. The table
// holds each slot's local area, or -1 for a cell no body stands in.
func (r *repairer) scan(key hexmap.ChunkKey) (*chunkSummary, []int32, error) {
	summary := &chunkSummary{border: map[api.Cell]int{}}
	table := make([]int32, hexmap.ChunkCells)
	for slot := range table {
		table[slot] = -1
	}
	for slot := range hexmap.ChunkCells {
		start := key.Cell(slot)
		if table[slot] >= 0 || !r.ctx.Bounds.Contains(start) {
			continue
		}
		standable, err := r.ground.Standable(start)
		if err != nil {
			return nil, nil, err
		}
		if !standable {
			continue
		}
		index := len(summary.areas)
		table[slot] = int32(index)
		members := []api.Cell{start}
		inPlace := false
		for head := 0; head < len(members); head++ {
			from := members[head]
			inPlace = inPlace || r.ctx.Fixed.Covers(from)
			reached, err := r.crossings(from)
			if err != nil {
				return nil, nil, err
			}
			for _, to := range reached {
				if hexmap.KeyOf(to) != key {
					summary.crossings = append(summary.crossings, crossing{area: index, to: to})
					summary.border[from] = index
					continue
				}
				if table[hexmap.SlotOf(to)] < 0 {
					table[hexmap.SlotOf(to)] = int32(index)
					members = append(members, to)
				}
			}
		}
		local := localArea{size: len(members), inPlace: inPlace}
		if local.size < r.pocket {
			local.cells = members
		}
		summary.areas = append(summary.areas, local)
	}
	return summary, table, nil
}

// nodeOf is the node a walkable cell belongs to.
func (r *repairer) nodeOf(cell api.Cell) (int, bool, error) {
	key := hexmap.KeyOf(cell)
	summary, ok := r.summaries[key]
	if !ok {
		return 0, false, nil
	}
	if area, ok := summary.border[cell]; ok {
		return summary.node + area, true, nil
	}
	if _, ok := r.tables[key]; !ok {
		standable, err := r.ground.Standable(cell)
		if err != nil || !standable {
			return 0, false, err
		}
	}
	table, err := r.table(key)
	if err != nil {
		return 0, false, err
	}
	area := table[hexmap.SlotOf(cell)]
	if area < 0 {
		return 0, false, nil
	}
	return summary.node + int(area), true, nil
}

// componentAt is the component a walkable cell belongs to.
func (r *repairer) componentAt(cell api.Cell) (int, bool, error) {
	n, ok, err := r.nodeOf(cell)
	if err != nil || !ok {
		return 0, false, err
	}
	return r.componentOf[n], true, nil
}

// table is a chunk's slots by local area, read again from the ground when it is not kept.
func (r *repairer) table(key hexmap.ChunkKey) ([]int32, error) {
	if table, ok := r.tables[key]; ok {
		return table, nil
	}
	_, table, err := r.scan(key)
	if err != nil {
		return nil, err
	}
	r.keepTable(key, table)
	return table, nil
}

// keepTable keeps a chunk's table until the chunk's ground changes, and at most as many tables as chunks.
func (r *repairer) keepTable(key hexmap.ChunkKey, table []int32) {
	if len(r.tables) >= r.budget {
		clear(r.tables)
	}
	r.tables[key] = table
}

// cellsOf lists a component's cells within the given chunks, in chunk order then slot order.
func (r *repairer) cellsOf(i int, within []hexmap.ChunkKey) ([]api.Cell, error) {
	var cells []api.Cell
	for _, key := range within {
		summary, ok := r.summaries[key]
		if !ok {
			continue
		}
		table, err := r.table(key)
		if err != nil {
			return nil, err
		}
		for slot, area := range table {
			if area >= 0 && r.componentOf[summary.node+int(area)] == i {
				cells = append(cells, key.Cell(slot))
			}
		}
	}
	return cells, nil
}

// chunksOf lists the chunks a component has cells in, in chunk order.
func (found component) chunksOf() []hexmap.ChunkKey {
	var keys []hexmap.ChunkKey
	for _, n := range found.nodes {
		if len(keys) == 0 || keys[len(keys)-1] != n.key {
			keys = append(keys, n.key)
		}
	}
	return keys
}

// set edits the ground and marks every chunk whose summary the edit may change.
func (r *repairer) set(cell api.Cell, terrain hexmap.Terrain) error {
	if err := r.ground.Set(cell, terrain); err != nil {
		return err
	}
	reach := r.rules.MaxStepUp + max(r.rules.MaxStepDown, r.rules.MaxDrop) + 2
	for _, near := range r.ground.CellsIn(cell.Hex, 2, cell.Depth-reach, cell.Depth+reach) {
		key := hexmap.KeyOf(near)
		if _, ok := r.summaries[key]; ok {
			r.dirty[key] = true
			delete(r.tables, key)
		}
	}
	return nil
}

// mainComponent is the component of the first place's opening, or the largest when no opening is walkable.
func (r *repairer) mainComponent(anchored []int) int {
	if len(anchored) > 0 {
		return anchored[0]
	}
	largest := 0
	for i, found := range r.components {
		if found.size > r.components[largest].size {
			largest = i
		}
	}
	return largest
}

// anchorComponents lists the components the places' openings lead into, in opening order. An opening leads
// to the standable ground a body reaches by one step from the opening's cell or its exit, or stands on there.
//
// An opening that leads nowhere gets its exit made standable instead, and opened reports it, so the caller
// looks again.
func (r *repairer) anchorComponents() (found []int, opened bool, err error) {
	add := func(cell api.Cell) (bool, error) {
		i, ok, err := r.componentAt(cell)
		if ok && !slices.Contains(found, i) {
			found = append(found, i)
		}
		return ok, err
	}
	for _, opening := range r.ctx.Openings {
		reached := 0
		stands := false
		for _, base := range []api.Cell{opening.Cell, opening.Exit} {
			ok, err := add(base)
			if err != nil {
				return nil, false, err
			}
			stands = stands || ok
			for dir := range api.Dir(api.DirCount) {
				step, ok, err := r.ground.Step(base, dir, r.rules)
				if err != nil {
					return nil, false, err
				}
				if ok {
					if _, err := add(step.To.Cell); err != nil {
						return nil, false, err
					}
					reached++
				}
			}
		}
		if reached == 0 && !stands {
			if err := r.standOn(opening); err != nil {
				return nil, false, err
			}
			return nil, true, nil
		}
	}
	return found, false, nil
}

// standOn opens the exit of an opening and gives it a floor, so a body can stand there.
func (r *repairer) standOn(opening Opening) error {
	exit, err := r.ground.At(opening.Exit)
	if err != nil {
		return err
	}
	changed := false
	if exit.Solid {
		if err := r.set(opening.Exit, r.ctx.Open); err != nil {
			return err
		}
		changed = true
	}
	floorCell := hexmap.Prisms{}.Below(opening.Exit).Cell
	floor, err := r.ground.At(floorCell)
	if err != nil {
		return err
	}
	if !floor.Solid && !r.ctx.Fixed.Covers(floorCell) {
		if err := r.set(floorCell, r.fillAt(floorCell)); err != nil {
			return err
		}
		changed = true
	}
	if !changed {
		return fmt.Errorf("verify: opening %v of %q leads to no walkable ground", opening.Cell, opening.Place)
	}
	return nil
}

// crossings lists the cells a body at from can step to and step back from.
func (r *repairer) crossings(from api.Cell) ([]api.Cell, error) {
	var reached []api.Cell
	for dir := range api.Dir(api.DirCount) {
		step, ok, err := r.ground.Step(from, dir, r.rules)
		if err != nil {
			return nil, err
		}
		if !ok || step.Kind == hexmap.StepDrop {
			continue
		}
		back, ok, err := r.ground.Step(step.To.Cell, (dir+api.DirCount/2)%api.DirCount, r.rules)
		if err != nil {
			return nil, err
		}
		if ok && back.To.Cell == from {
			reached = append(reached, step.To.Cell)
		}
	}
	return reached, nil
}

// fillAt is the terrain the fixed layer holds at a cell outside the places.
func (r *repairer) fillAt(cell api.Cell) hexmap.Terrain {
	key := hexmap.KeyOf(cell)
	chunk, ok := r.fixed[key]
	if !ok {
		var err error
		if chunk, err = r.ctx.Fixed.Chunk(key); err != nil {
			chunk = hexmap.NewChunk(hexmap.Terrain{Solid: true, Opaque: true})
		}
		if len(r.fixed) >= r.budget {
			clear(r.fixed)
		}
		r.fixed[key] = chunk
	}
	origin := key.Origin()
	return chunk.Get(cell.Hex.Q-origin.Hex.Q, cell.Hex.R-origin.Hex.R, cell.Depth-origin.Depth)
}

// chunkNeighbors lists the chunks sharing a face with key, in a fixed order.
func chunkNeighbors(key hexmap.ChunkKey) []hexmap.ChunkKey {
	var out []hexmap.ChunkKey
	for dir := range api.Dir(api.DirCount) {
		next := hexmap.Prisms{}.Neighbor(api.Cell{Hex: api.Hex{Q: key.Q, R: key.R}}, dir).Cell
		out = append(out, hexmap.ChunkKey{Q: next.Hex.Q, R: next.Hex.R, Depth: key.Depth})
	}
	return append(out,
		hexmap.ChunkKey{Q: key.Q, R: key.R, Depth: key.Depth - 1},
		hexmap.ChunkKey{Q: key.Q, R: key.R, Depth: key.Depth + 1},
	)
}

// chunkPath is the shortest chain of chunks from one the main component has cells in to one a target has
// cells in, searched breadth first in chunk order.
func (r *repairer) chunkPath(main int, targets map[int]bool) []hexmap.ChunkKey {
	goal := map[hexmap.ChunkKey]bool{}
	for i := range targets {
		for _, key := range r.components[i].chunksOf() {
			goal[key] = true
		}
	}
	sources := r.components[main].chunksOf()
	parents := map[hexmap.ChunkKey]hexmap.ChunkKey{}
	for _, key := range sources {
		parents[key] = key
	}
	queue := slices.Clone(sources)
	for head := 0; head < len(queue); head++ {
		at := queue[head]
		if goal[at] {
			path := []hexmap.ChunkKey{at}
			for parents[at] != at {
				at = parents[at]
				path = append(path, at)
			}
			return path
		}
		for _, next := range chunkNeighbors(at) {
			if _, ok := r.summaries[next]; !ok {
				continue
			}
			if _, seen := parents[next]; !seen {
				parents[next] = at
				queue = append(queue, next)
			}
		}
	}
	return nil
}

// widen adds every chunk within width chunks of the path, or every chunk for width zero, in chunk order.
func (r *repairer) widen(path []hexmap.ChunkKey, width int) []hexmap.ChunkKey {
	if width == 0 {
		return r.keys
	}
	in := map[hexmap.ChunkKey]bool{}
	frontier := slices.Clone(path)
	for _, key := range path {
		in[key] = true
	}
	for range width {
		var next []hexmap.ChunkKey
		for _, key := range frontier {
			for _, near := range chunkNeighbors(key) {
				if _, ok := r.summaries[near]; ok && !in[near] {
					in[near] = true
					next = append(next, near)
				}
			}
		}
		frontier = next
	}
	keys := slices.Collect(maps.Keys(in))
	slices.SortFunc(keys, hexmap.CompareKeys)
	return keys
}

// bridgeMove is one step of a corridor: the cell stood on after it, the solid cells it opens and the open
// cells it fills to give the landing a floor.
type bridgeMove struct {
	to          api.Cell
	from        api.Cell
	carve, fill []api.Cell
}

// join carves a corridor from the main component to the nearest target component. It finds the shortest
// chain of chunks between them, then searches cells only in the chunks near that chain, widening the search
// when no corridor fits.
func (r *repairer) join(main int, targets map[int]bool) error {
	path := r.chunkPath(main, targets)
	if path == nil {
		return errors.New("verify: no chain of chunks joins the walkable areas")
	}
	for _, width := range corridorWidths {
		dbg("  width %d path %d", width, len(path))
		within := r.widen(path, width)
		inside := map[hexmap.ChunkKey]bool{}
		for _, key := range within {
			inside[key] = true
		}
		sources, err := r.cellsOf(main, within)
		if err != nil {
			return err
		}
		for _, fill := range []bool{false, true} {
			joined, err := r.bridge(sources, inside, targets, fill)
			if err != nil || joined {
				return err
			}
		}
	}
	return errors.New("verify: no corridor can join the walkable areas")
}

// bridge searches for a corridor inside the given chunks and carves it, reporting whether it found one. With
// fill set, a step may fill an open cell that is no walkable ground to give its landing a floor.
func (r *repairer) bridge(
	sources []api.Cell, inside map[hexmap.ChunkKey]bool, targets map[int]bool, fill bool,
) (bool, error) {
	parents := make(map[api.Cell]bridgeMove, len(sources))
	for _, cell := range sources {
		parents[cell] = bridgeMove{from: cell, to: cell}
	}
	queue := slices.Clone(sources)
	for head := 0; head < len(queue); head++ {
		from := queue[head]
		for dir := range api.Dir(api.DirCount) {
			for _, rise := range []int{0, -1, 1} {
				move, ok, err := r.tunnelStep(from, dir, rise, fill)
				if err != nil {
					return false, err
				}
				if _, seen := parents[move.to]; !ok || seen || !inside[hexmap.KeyOf(move.to)] {
					continue
				}
				parents[move.to] = move
				area, walks, err := r.componentAt(move.to)
				if err != nil {
					return false, err
				}
				if walks && targets[area] {
					return true, r.carvePath(parents, move.to)
				}
				queue = append(queue, move.to)
			}
		}
	}
	return false, nil
}

func (r *repairer) carvePath(parents map[api.Cell]bridgeMove, end api.Cell) error {
	for cell := end; ; {
		move := parents[cell]
		for _, carved := range move.carve {
			if err := r.set(carved, r.ctx.Open); err != nil {
				return err
			}
		}
		for _, filled := range move.fill {
			if err := r.set(filled, r.fillAt(filled)); err != nil {
				return err
			}
		}
		if move.from == cell {
			return nil
		}
		cell = move.from
	}
}

// tunnelStep is one step of a tunnel: a level step, a climb with headroom, or a step down. It never changes
// a place's cells, and fills the open cell under the landing when there is one.
func (r *repairer) tunnelStep(from api.Cell, dir api.Dir, rise int, fill bool) (bridgeMove, bool, error) {
	topology := hexmap.Prisms{}
	next := topology.Neighbor(from, dir).Cell
	move := bridgeMove{from: from}
	var needs []api.Cell
	switch rise {
	case 0:
		move.to, needs = next, []api.Cell{next}
	case -1:
		move.to = topology.Above(next).Cell
		needs = []api.Cell{move.to, topology.Above(from).Cell}
		ground, err := r.ground.At(next)
		if err != nil || !ground.Solid {
			return move, false, err
		}
	default:
		move.to = topology.Below(next).Cell
		needs = []api.Cell{next, move.to}
	}
	floorCell := topology.Below(move.to).Cell
	floor, err := r.ground.At(floorCell)
	if err != nil {
		return move, false, err
	}
	if !floor.Solid {
		_, walkable, err := r.nodeOf(floorCell)
		if err != nil {
			return move, false, err
		}
		if !fill || walkable || r.ctx.Fixed.Covers(floorCell) {
			return move, false, nil
		}
		move.fill = append(move.fill, floorCell)
	}
	for _, cell := range needs {
		terrain, err := r.ground.At(cell)
		if err != nil {
			return move, false, err
		}
		if !r.ctx.Bounds.Contains(cell) || r.ctx.Fixed.Covers(cell) && terrain.Solid {
			return move, false, nil
		}
		if terrain.Solid {
			_, stoodOn, err := r.nodeOf(topology.Above(cell).Cell)
			if err != nil {
				return move, false, err
			}
			if stoodOn {
				return move, false, nil
			}
			move.carve = append(move.carve, cell)
		}
	}
	return move, true, nil
}
