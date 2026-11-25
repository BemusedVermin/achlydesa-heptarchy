package worldgen

import (
	"errors"
	"fmt"
	"math"
	"slices"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
)

// VerifyStage is the name the verify stage registers under.
const VerifyStage = "verify"

var verifyParams = []string{"max_step_up", "max_step_down", "min_pocket"}

// maxRepairRounds bounds the repair loop; a world that needs more is reported as a bug.
const maxRepairRounds = 256

// RegisterVerify adds the verify stage to stages.
//
// The stage checks the ground the earlier stages leave against the guarantees and repairs it: every place's
// openings lead into one walkable whole, and no pocket of walkable ground smaller than min_pocket is left.
// Ground is walkable by steps a body can make both ways, so a drop never joins two areas. A small pocket is
// filled; a larger one that is cut off is joined to the rest by the shortest corridor that can be carved. A
// repair never changes a place's cells, and never opens the ground a walkable cell stands on. A place's
// ground that no opening leads to is left alone. The stage must be the last of its geography, since it reads the
// result of every stage before it. Every param is required:
//
//   - max_step_up, max_step_down: how far a body climbs or descends in one step, as in the party's rules.
//   - min_pocket: walkable cells a pocket needs to be kept; smaller ones are filled.
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

// Run applies the chunk's part of the world's repair. The repair is worked out once from the whole ground
// before this stage, a pure function of the seed, so every chunk agrees on it whatever order chunks load in.
func (verify) Run(ctx *Context, chunk *hexmap.Chunk) error {
	value, err := ctx.Shared(func() (any, error) { return repairWorld(ctx) })
	if err != nil {
		return err
	}
	origin := ctx.Key.Origin()
	for _, edit := range value.(map[hexmap.ChunkKey][]hexmap.Edit)[ctx.Key] {
		chunk.Set(
			edit.Cell.Hex.Q-origin.Hex.Q,
			edit.Cell.Hex.R-origin.Hex.R,
			edit.Cell.Depth-origin.Depth,
			edit.Terrain,
		)
	}
	return nil
}

type repairer struct {
	ctx    *Context
	ground *hexmap.Map
	rules  hexmap.StepRules
	pocket int
	cells  []api.Cell
	fixed  map[hexmap.ChunkKey]*hexmap.Chunk
}

// repairWorld builds the repair as the edits each chunk applies.
func repairWorld(ctx *Context) (map[hexmap.ChunkKey][]hexmap.Edit, error) {
	r := &repairer{
		ctx:    ctx,
		ground: hexmap.NewMap(ctx.Before, hexmap.Prisms{}),
		rules: hexmap.StepRules{
			MaxStepUp:   int(ctx.Params["max_step_up"]),
			MaxStepDown: int(ctx.Params["max_step_down"]),
			MaxDrop:     int(ctx.Params["max_step_down"]),
		},
		pocket: int(ctx.Params["min_pocket"]),
		fixed:  map[hexmap.ChunkKey]*hexmap.Chunk{},
	}
	r.ground.SetBounds(ctx.Bounds, hexmap.Terrain{Solid: true, Opaque: true})
	r.cells = boundsCells(ctx.Bounds)
	if err := r.settle(); err != nil {
		return nil, err
	}
	byChunk := map[hexmap.ChunkKey][]hexmap.Edit{}
	for _, edit := range r.ground.Edits() {
		key := hexmap.KeyOf(edit.Cell)
		byChunk[key] = append(byChunk[key], edit)
	}
	return byChunk, nil
}

// boundsCells lists every cell inside the bounds by depth, then row, then column.
func boundsCells(bounds hexmap.Bounds) []api.Cell {
	var cells []api.Cell
	for depth := bounds.MinDepth; depth <= bounds.MaxDepth; depth++ {
		for row := bounds.MinRow; row <= bounds.MaxRow; row++ {
			for col := bounds.MinCol; col <= bounds.MaxCol; col++ {
				cells = append(cells, api.Cell{Hex: hexmap.FromOffsetOddR(col, row), Depth: depth})
			}
		}
	}
	return cells
}

// walkable is the walkable cells grouped into the areas a body can cross between both ways.
type walkable struct {
	areaOf map[api.Cell]int
	areas  [][]api.Cell
}

// settle repairs the ground until a round finds nothing to change.
func (r *repairer) settle() error {
	for range maxRepairRounds {
		found, err := r.areasOfGround()
		if err != nil {
			return err
		}
		anchored, opened, err := r.anchorAreas(found)
		if err != nil {
			return err
		}
		if opened {
			continue
		}
		keep := make([]bool, len(found.areas))
		changed := false
		for i, members := range found.areas {
			inPlace := slices.ContainsFunc(members, r.ctx.Fixed.Covers)
			anchor := slices.Contains(anchored, i)
			keep[i] = anchor || !inPlace && len(members) >= r.pocket
			if keep[i] || inPlace {
				continue
			}
			for _, cell := range members {
				if err := r.ground.Set(cell, r.fillAt(cell)); err != nil {
					return err
				}
				changed = true
			}
		}
		if changed {
			continue
		}
		targets := r.targets(found, keep, anchored)
		if len(targets) == 0 {
			return nil
		}
		if err := r.join(found, r.mainArea(found, anchored), targets); err != nil {
			return err
		}
	}
	return errors.New("verify: the repair did not settle")
}

// targets lists the kept areas other than the main one.
func (r *repairer) targets(found *walkable, keep []bool, anchored []int) map[int]bool {
	main := r.mainArea(found, anchored)
	targets := map[int]bool{}
	for i := range found.areas {
		if keep[i] && i != main {
			targets[i] = true
		}
	}
	return targets
}

// mainArea is the area of the first place's opening, or the largest area when no opening is walkable.
func (r *repairer) mainArea(found *walkable, anchored []int) int {
	if len(anchored) > 0 {
		return anchored[0]
	}
	largest := 0
	for i, members := range found.areas {
		if len(members) > len(found.areas[largest]) {
			largest = i
		}
	}
	return largest
}

// anchorAreas lists the areas the places' openings lead into, in opening order. An opening leads to the
// standable ground a body reaches by one step from the opening's cell or its exit, or stands on there.
//
// An opening that leads nowhere gets its exit made standable instead, and opened reports it, so the caller
// looks again.
func (r *repairer) anchorAreas(found *walkable) (areas []int, opened bool, err error) {
	add := func(cell api.Cell) {
		if i, ok := found.areaOf[cell]; ok && !slices.Contains(areas, i) {
			areas = append(areas, i)
		}
	}
	for _, opening := range r.ctx.Openings {
		reached := 0
		for _, base := range []api.Cell{opening.Cell, opening.Exit} {
			add(base)
			for dir := range api.Dir(api.DirCount) {
				step, ok, err := r.ground.Step(base, dir, r.rules)
				if err != nil {
					return nil, false, err
				}
				if ok {
					add(step.To.Cell)
					reached++
				}
			}
		}
		_, standsCell := found.areaOf[opening.Cell]
		_, standsExit := found.areaOf[opening.Exit]
		if reached == 0 && !standsCell && !standsExit {
			if err := r.standOn(opening); err != nil {
				return nil, false, err
			}
			return nil, true, nil
		}
	}
	return areas, false, nil
}

// standOn opens the exit of an opening and gives it a floor, so a body can stand there.
func (r *repairer) standOn(opening Opening) error {
	exit, err := r.ground.At(opening.Exit)
	if err != nil {
		return err
	}
	changed := false
	if exit.Solid {
		if err := r.ground.Set(opening.Exit, r.ctx.Open); err != nil {
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
		if err := r.ground.Set(floorCell, r.fillAt(floorCell)); err != nil {
			return err
		}
		changed = true
	}
	if !changed {
		return fmt.Errorf("verify: opening %v of %q leads to no walkable ground", opening.Cell, opening.Place)
	}
	return nil
}

// areasOfGround flood-fills the walkable cells in cell order, so area numbers never depend on map order.
func (r *repairer) areasOfGround() (*walkable, error) {
	found := &walkable{areaOf: map[api.Cell]int{}}
	for _, start := range r.cells {
		if _, seen := found.areaOf[start]; seen {
			continue
		}
		standable, err := r.ground.Standable(start)
		if err != nil {
			return nil, err
		}
		if !standable {
			continue
		}
		index := len(found.areas)
		found.areaOf[start] = index
		members := []api.Cell{start}
		for head := 0; head < len(members); head++ {
			reached, err := r.crossings(members[head])
			if err != nil {
				return nil, err
			}
			for _, to := range reached {
				if _, seen := found.areaOf[to]; !seen {
					found.areaOf[to] = index
					members = append(members, to)
				}
			}
		}
		found.areas = append(found.areas, members)
	}
	return found, nil
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
		back, err := r.steppable(step.To.Cell, from)
		if err != nil {
			return nil, err
		}
		if back {
			reached = append(reached, step.To.Cell)
		}
	}
	return reached, nil
}

func (r *repairer) steppable(from, to api.Cell) (bool, error) {
	for dir := range api.Dir(api.DirCount) {
		step, ok, err := r.ground.Step(from, dir, r.rules)
		if err != nil {
			return false, err
		}
		if ok && step.To.Cell == to {
			return true, nil
		}
	}
	return false, nil
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
		r.fixed[key] = chunk
	}
	origin := key.Origin()
	return chunk.Get(cell.Hex.Q-origin.Hex.Q, cell.Hex.R-origin.Hex.R, cell.Depth-origin.Depth)
}

// bridgeMove is one step of a corridor: the cell stood on after it, the solid cells it opens and the open
// cells it fills to give the landing a floor.
type bridgeMove struct {
	to          api.Cell
	from        api.Cell
	carve, fill []api.Cell
}

// join carves the shortest corridor from the main area to any target area. It searches outward from every
// cell of the main area in a fixed order, over the steps a tunnel can take.
func (r *repairer) join(found *walkable, main int, targets map[int]bool) error {
	for _, fill := range []bool{false, true} {
		joined, err := r.bridge(found, main, targets, fill)
		if err != nil || joined {
			return err
		}
	}
	return errors.New("verify: no corridor can join the walkable areas")
}

// bridge searches for a corridor and carves it, reporting whether it found one. With fill set, a step may
// fill an open cell that is no walkable ground to give its landing a floor.
func (r *repairer) bridge(found *walkable, main int, targets map[int]bool, fill bool) (bool, error) {
	sources := found.areas[main]
	parents := make(map[api.Cell]bridgeMove, len(sources))
	for _, cell := range sources {
		parents[cell] = bridgeMove{from: cell, to: cell}
	}
	queue := slices.Clone(sources)
	for head := 0; head < len(queue); head++ {
		from := queue[head]
		for dir := range api.Dir(api.DirCount) {
			for _, rise := range []int{0, -1, 1} {
				move, ok, err := r.tunnelStep(found, from, dir, rise, fill)
				if err != nil {
					return false, err
				}
				if _, seen := parents[move.to]; !ok || seen {
					continue
				}
				parents[move.to] = move
				if area, walks := found.areaOf[move.to]; walks && targets[area] {
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
			if err := r.ground.Set(carved, r.ctx.Open); err != nil {
				return err
			}
		}
		for _, filled := range move.fill {
			if err := r.ground.Set(filled, r.fillAt(filled)); err != nil {
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
func (r *repairer) tunnelStep(
	found *walkable, from api.Cell, dir api.Dir, rise int, fill bool,
) (bridgeMove, bool, error) {
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
		if _, walkable := found.areaOf[floorCell]; !fill || walkable || r.ctx.Fixed.Covers(floorCell) {
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
			if _, stoodOn := found.areaOf[topology.Above(cell).Cell]; stoodOn {
				return move, false, nil
			}
			move.carve = append(move.carve, cell)
		}
	}
	return move, true, nil
}
