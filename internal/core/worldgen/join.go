package worldgen

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/fnv"
	"math"
	"slices"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/rng"
)

// JoinStage is the name the join stage registers under.
const JoinStage = "join"

var joinParams = []string{"width", "reach", "bias"}

// RegisterJoin adds the join stage to stages.
//
// From each opening of a fixed place the stage carves a tunnel toward the nearest cave, so a place ends in
// ground that continues as a passage. The tunnel is a seeded random walk that takes level steps and one-cell
// climbs and descents, each a step a body can make. It ends on meeting a cave or after reach steps. A tunnel
// is drawn from the seed, the opening and the caves alone, so every chunk it crosses carves its own part.
// The stage needs the caves stage in the same geography, and should come after it. A tunnel never changes a
// place's cells. When no cave lies within reach, the opening gets only its first cell. Every param is
// required:
//
//   - width: a tunnel's width in hexes; the passage is every cell within width-1 hexes of the walk.
//   - reach: the most steps a tunnel walks, and how far it looks for a cave.
//   - bias: at least 0; how strongly a step toward the cave is preferred; 0 is an unbiased walk.
func RegisterJoin(stages *Stages) error {
	return stages.Register(JoinStage, join{}, checkJoinParams)
}

func checkJoinParams(params map[string]float64) error {
	for _, name := range joinParams {
		if _, ok := params[name]; !ok {
			return fmt.Errorf("missing param %q", name)
		}
	}
	for name := range params {
		if !slices.Contains(joinParams, name) {
			return fmt.Errorf("takes no param %q", name)
		}
	}
	for _, name := range []string{"width", "reach"} {
		if params[name] < 1 || params[name] != math.Trunc(params[name]) {
			return fmt.Errorf("param %q must be a whole number of at least 1", name)
		}
	}
	if params["bias"] < 0 {
		return errors.New(`param "bias" must not be negative`)
	}
	return nil
}

type join struct{}

type joiner struct {
	ctx           *Context
	field         caveField
	radius, reach int
	bias          float64
}

// Run carves the part of every nearby tunnel that lies in the chunk.
func (join) Run(ctx *Context, chunk *hexmap.Chunk) error {
	caveParams, ok := ctx.Siblings[CavesStage]
	if !ok {
		return errors.New("the geography has no caves stage to join to")
	}
	j := joiner{
		ctx:    ctx,
		field:  newCaveField(ctx.Seed, ctx.ExportCount, ctx.Bounds, caveParams),
		radius: int(ctx.Params["width"]) - 1,
		reach:  int(ctx.Params["reach"]),
		bias:   ctx.Params["bias"],
	}
	origin := ctx.Key.Origin()
	middle := api.Hex{Q: origin.Hex.Q + hexmap.ChunkSize/2, R: origin.Hex.R + hexmap.ChunkSize/2}
	span := j.reach + j.radius + hexmap.ChunkSize
	for _, opening := range ctx.Openings {
		exit := opening.Exit
		if hexmap.Distance(exit.Hex, middle) > span ||
			exit.Depth+span < origin.Depth || exit.Depth-span >= origin.Depth+hexmap.ChunkSize {
			continue
		}
		for _, cell := range j.tunnel(opening) {
			if hexmap.KeyOf(cell) == ctx.Key {
				chunk.Set(cell.Hex.Q-origin.Hex.Q, cell.Hex.R-origin.Hex.R, cell.Depth-origin.Depth, ctx.Open)
			}
		}
	}
	return nil
}

// tunnel lists the cells the opening's tunnel carves. The width is added around the walk last and spares the
// ground under each climb's landing, which the climb steps from.
func (j joiner) tunnel(opening Opening) []api.Cell {
	stands := []api.Cell{opening.Exit}
	var path []api.Cell
	spared := map[api.Cell]bool{}
	if !j.joins(opening.Exit) {
		if target, found := j.nearestCave(opening.Exit); found {
			stands, path = j.walk(opening, target, spared)
		}
	}
	cells := path
	for _, stand := range stands {
		for _, cell := range j.disc(stand) {
			if !spared[cell] {
				cells = append(cells, cell)
			}
		}
	}
	return cells
}

// walk takes the biased random walk from the opening's exit, returning the cells stood on and the cells the
// steps need open. It marks in spared the ground under each climb's landing.
func (j joiner) walk(opening Opening, target api.Cell, spared map[api.Cell]bool) (stands, path []api.Cell) {
	gen := rng.New(j.ctx.Seed).Derived(JoinStage, j.ctx.ExportCount^cellHash(opening.Cell))
	visited := map[api.Cell]bool{opening.Exit: true}
	opened := map[api.Cell]bool{opening.Exit: true}
	stand := opening.Exit
	stands = []api.Cell{stand}
	for range j.reach {
		moves := j.moves(stand, target, visited, opened, spared)
		if len(moves) == 0 {
			break
		}
		total := 0.0
		for _, move := range moves {
			total += move.weight
		}
		pick, chosen := gen.Float64()*total, moves[len(moves)-1]
		for _, move := range moves {
			if pick < move.weight {
				chosen = move
				break
			}
			pick -= move.weight
		}
		visited[chosen.stand] = true
		path = append(path, chosen.carve...)
		for _, cell := range chosen.carve {
			opened[cell] = true
		}
		stands = append(stands, chosen.stand)
		if chosen.stand.Depth < stand.Depth {
			spared[hexmap.Prisms{}.Below(chosen.stand).Cell] = true
		}
		stand = chosen.stand
		if j.joins(stand) {
			break
		}
	}
	return stands, path
}

type tunnelMove struct {
	stand  api.Cell
	carve  []api.Cell
	weight float64
}

// moves lists the steps a body could take from stand, each with the cells it needs open: a level step, a
// climb with headroom, or a step down into a lowered cell. A climb needs the ground under its landing solid,
// so it is not offered where the walk already opened that ground, and no step opens ground a climb needs.
func (j joiner) moves(stand, target api.Cell, visited, opened, spared map[api.Cell]bool) []tunnelMove {
	var moves []tunnelMove
	for dir := range api.Dir(api.DirCount) {
		next := hexmap.Prisms{}.Neighbor(stand, dir).Cell
		for _, rise := range []int{0, -1, 1} {
			to := api.Cell{Hex: next.Hex, Depth: stand.Depth + rise}
			carve := []api.Cell{to}
			switch rise {
			case -1:
				carve = append(carve, api.Cell{Hex: stand.Hex, Depth: to.Depth})
			case 1:
				carve = []api.Cell{next, to}
			}
			if visited[to] ||
				slices.ContainsFunc(carve, func(cell api.Cell) bool { return !j.usable(cell) || spared[cell] }) {
				continue
			}
			if rise == -1 && opened[hexmap.Prisms{}.Below(to).Cell] {
				continue
			}
			gain := float64(hexmap.CellDistance(stand, target) - hexmap.CellDistance(to, target))
			moves = append(moves, tunnelMove{stand: to, carve: carve, weight: max(0.1, 1+j.bias*gain)})
		}
	}
	return moves
}

// usable reports whether the tunnel may open the cell.
func (j joiner) usable(cell api.Cell) bool {
	return j.ctx.Bounds.Contains(cell) && !j.ctx.Fixed.Covers(cell)
}

func (j joiner) isCave(cell api.Cell) bool { return j.usable(cell) && j.field.open(cell) }

// joins reports whether a tunnel ending at the cell is open to a cave: the cell, the one under it or a
// neighbor on its level is cave.
func (j joiner) joins(cell api.Cell) bool {
	if j.isCave(cell) || j.isCave(hexmap.Prisms{}.Below(cell).Cell) {
		return true
	}
	for dir := range api.Dir(api.DirCount) {
		if j.isCave(hexmap.Prisms{}.Neighbor(cell, dir).Cell) {
			return true
		}
	}
	return false
}

// disc lists the usable cells within the tunnel's radius of the cell, on its level.
func (j joiner) disc(center api.Cell) []api.Cell {
	var cells []api.Cell
	for dr := -j.radius; dr <= j.radius; dr++ {
		for dq := -j.radius; dq <= j.radius; dq++ {
			if hexmap.Distance(api.Hex{}, api.Hex{Q: dq, R: dr}) > j.radius {
				continue
			}
			cell := api.Cell{Hex: api.Hex{Q: center.Hex.Q + dq, R: center.Hex.R + dr}, Depth: center.Depth}
			if j.usable(cell) {
				cells = append(cells, cell)
			}
		}
	}
	return cells
}

// nearestCave finds the cave cell closest to from within reach, shell by shell of [hexmap.CellDistance].
func (j joiner) nearestCave(from api.Cell) (api.Cell, bool) {
	for distance := 1; distance <= j.reach; distance++ {
		for rise := -distance; rise <= distance; rise++ {
			depth := from.Depth + rise
			if depth < j.ctx.Bounds.MinDepth || depth > j.ctx.Bounds.MaxDepth {
				continue
			}
			for _, hex := range hexRing(from.Hex, distance-max(rise, -rise)) {
				if cell := (api.Cell{Hex: hex, Depth: depth}); j.isCave(cell) {
					return cell, true
				}
			}
		}
	}
	return api.Cell{}, false
}

// hexRing lists the hexes exactly radius steps from center, starting southwest and turning counter-clockwise.
func hexRing(center api.Hex, radius int) []api.Hex {
	if radius == 0 {
		return []api.Hex{center}
	}
	cell := api.Cell{Hex: api.Hex{Q: center.Q - radius, R: center.R + radius}}
	ring := make([]api.Hex, 0, 6*radius)
	for _, side := range []api.Dir{api.DirE, api.DirNE, api.DirNW, api.DirW, api.DirSW, api.DirSE} {
		for range radius {
			ring = append(ring, cell.Hex)
			cell = hexmap.Prisms{}.Neighbor(cell, side).Cell
		}
	}
	return ring
}

func cellHash(cell api.Cell) uint64 {
	hash := fnv.New64a()
	var buf []byte
	for _, n := range []int{cell.Hex.Q, cell.Hex.R, cell.Depth} {
		buf = binary.LittleEndian.AppendUint64(buf, uint64(int64(n)))
	}
	_, _ = hash.Write(buf)
	return hash.Sum64()
}
