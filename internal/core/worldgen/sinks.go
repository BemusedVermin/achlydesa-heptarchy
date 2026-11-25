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

// SinksStage is the name the sinks stage registers under.
const SinksStage = "sinks"

var sinkParams = []string{
	"spacing", "jitter", "density", "riser_chance",
	"rim_depth", "lap_height", "laps", "headroom",
	"shaft_radius", "outer_radius",
}

// RegisterSinks adds the sinks stage to stages.
//
// A sink is a helicoid ramp around a vertical shaft. Floors under a hex are lap_height apart, and walking once
// around the axis lowers the floor by lap_height, so every lap resembles the last and the ramp seam has no
// fault. Far from the axis a step changes the floor by little; near it by more, until the shaft cuts straight
// down. A riser is the same shape with the opposite handedness. The stage sets every cell of its footprint,
// so it should come after stages that carve, such as caves. Every param is required:
//
//   - spacing: hexes between the cells of the coarse grid; each cell may hold one axis.
//   - jitter: in [0, 1]; how far an axis wanders from its grid cell's center, as a fraction of spacing.
//   - density: in [0, 1]; the chance a grid cell holds an axis.
//   - riser_chance: in [0, 1]; the chance an axis is a riser.
//   - rim_depth: depth of the shallowest floor; the depth above it stays as it was.
//   - lap_height: depth steps the floor falls in one lap.
//   - laps: how many laps the ramp makes; the shaft's floor is laps * lap_height below the rim.
//   - headroom: open cells above each floor; at least 2 so a body can climb, and below lap_height.
//   - shaft_radius: hexes from the axis within which the shaft is open from the rim to its floor.
//   - outer_radius: hexes from the axis at which the ramp ends.
func RegisterSinks(stages *Stages) error {
	return stages.Register(SinksStage, sinks{}, checkSinkParams)
}

func checkSinkParams(params map[string]float64) error {
	for _, name := range sinkParams {
		if _, ok := params[name]; !ok {
			return fmt.Errorf("missing param %q", name)
		}
	}
	for name := range params {
		if !slices.Contains(sinkParams, name) {
			return fmt.Errorf("takes no param %q", name)
		}
	}
	for _, name := range []string{"jitter", "density", "riser_chance"} {
		if params[name] < 0 || params[name] > 1 {
			return fmt.Errorf("param %q must be between 0 and 1", name)
		}
	}
	for _, name := range []string{"spacing", "rim_depth", "lap_height", "laps", "headroom"} {
		if params[name] < 1 || params[name] != math.Trunc(params[name]) {
			return fmt.Errorf("param %q must be a whole number of at least 1", name)
		}
	}
	if params["headroom"] < 2 || params["headroom"] >= params["lap_height"] {
		return errors.New(`param "headroom" must be at least 2 and below "lap_height"`)
	}
	shaft, outer := params["shaft_radius"], params["outer_radius"]
	if shaft < 0.5 || outer <= shaft {
		return errors.New(`"shaft_radius" must be at least 0.5 and below "outer_radius"`)
	}
	if params["spacing"]*(1-params["jitter"]) <= 2*outer {
		return errors.New(`"spacing" must exceed twice "outer_radius" even at full "jitter", so sinks never overlap`)
	}
	// A neighbor of a ramp cell is one hex away, so at the shaft wall the floor changes by this much at most.
	steepest := params["lap_height"] * 2 * math.Asin(math.Min(1, 1/(2*shaft))) / (2 * math.Pi)
	if steepest > 1 {
		return errors.New(`"shaft_radius" is too small for "lap_height": the ramp would be steeper than a step`)
	}
	return nil
}

type sinks struct{}

type sinkAxis struct {
	x, y  float64
	riser bool
}

type sinkShape struct {
	params                                         map[string]float64
	rim, lapHeight, laps, headroom, floor, spacing int
	shaft, outer                                   float64
}

func newSinkShape(params map[string]float64) sinkShape {
	shape := sinkShape{
		params:    params,
		rim:       int(params["rim_depth"]),
		lapHeight: int(params["lap_height"]),
		laps:      int(params["laps"]),
		headroom:  int(params["headroom"]),
		spacing:   int(params["spacing"]),
		shaft:     params["shaft_radius"],
		outer:     params["outer_radius"],
	}
	shape.floor = shape.rim + shape.laps*shape.lapHeight
	return shape
}

// Run opens the ramp and shaft of every sink that reaches the chunk and closes the rest of their footprints.
func (sinks) Run(ctx *Context, chunk *hexmap.Chunk) error {
	shape := newSinkShape(ctx.Params)
	if shape.rim <= ctx.Bounds.MinDepth || shape.floor > ctx.Bounds.MaxDepth {
		return fmt.Errorf("a sink from depth %d to %d does not fit between depths %d and %d",
			shape.rim, shape.floor, ctx.Bounds.MinDepth, ctx.Bounds.MaxDepth)
	}
	origin := ctx.Key.Origin()
	if origin.Depth > shape.floor+1 || origin.Depth+hexmap.ChunkSize <= shape.rim {
		return nil
	}
	axes := shape.axesNear(ctx, origin)
	if len(axes) == 0 {
		return nil
	}
	fill, err := ctx.Fixed.Chunk(ctx.Key)
	if err != nil {
		return err
	}
	for r := range hexmap.ChunkSize {
		for q := range hexmap.ChunkSize {
			hex := api.Hex{Q: origin.Hex.Q + q, R: origin.Hex.R + r}
			x, y := planePosition(hex)
			for _, axis := range axes {
				dx, dy := x-axis.x, y-axis.y
				distance := math.Hypot(dx, dy)
				if distance > shape.outer {
					continue
				}
				offset := shape.floorOffset(dx, dy, axis.riser)
				for depth := range hexmap.ChunkSize {
					cell := api.Cell{Hex: hex, Depth: origin.Depth + depth}
					if !ctx.Bounds.Contains(cell) || cell.Depth < shape.rim || cell.Depth > shape.floor+1 {
						continue
					}
					if shape.open(cell.Depth, distance, offset) {
						chunk.Set(q, r, depth, ctx.Open)
					} else {
						chunk.Set(q, r, depth, fill.Get(q, r, depth))
					}
				}
			}
		}
	}
	return nil
}

// floorOffset is how far into its lap the ramp's floor lies at the offset from the axis, in [0, lapHeight).
func (s sinkShape) floorOffset(dx, dy float64, riser bool) int {
	turn := (math.Atan2(dy, dx) + math.Pi) / (2 * math.Pi)
	if riser {
		turn = 1 - turn
	}
	return min(s.lapHeight-1, int(turn*float64(s.lapHeight)))
}

// open reports whether the cell at depth is open, a distance from the axis, for a ramp with the lap offset.
func (s sinkShape) open(depth int, distance float64, offset int) bool {
	if distance < s.shaft {
		return depth <= s.floor
	}
	for lap := range s.laps {
		floor := s.rim + offset + lap*s.lapHeight
		if depth <= floor && depth > floor-s.headroom {
			return true
		}
	}
	return false
}

// axesNear lists the axes whose footprint may reach the chunk's hexes, drawn from the coarse grid alone.
func (s sinkShape) axesNear(ctx *Context, origin api.Cell) []sinkAxis {
	lowX, lowY := math.Inf(1), math.Inf(1)
	highX, highY := math.Inf(-1), math.Inf(-1)
	for _, corner := range [][2]int{
		{0, 0}, {hexmap.ChunkSize - 1, 0}, {0, hexmap.ChunkSize - 1}, {hexmap.ChunkSize - 1, hexmap.ChunkSize - 1},
	} {
		x, y := planePosition(api.Hex{Q: origin.Hex.Q + corner[0], R: origin.Hex.R + corner[1]})
		lowX, highX = min(lowX, x), max(highX, x)
		lowY, highY = min(lowY, y), max(highY, y)
	}
	spacing := float64(s.spacing)
	var axes []sinkAxis
	for gridY := int(math.Floor((lowY - s.outer) / spacing)); gridY <= int(math.Floor((highY+s.outer)/spacing)); gridY++ {
		for gridX := int(math.Floor((lowX - s.outer) / spacing)); gridX <= int(math.Floor((highX+s.outer)/spacing)); gridX++ {
			if axis, ok := s.axisIn(ctx, gridX, gridY); ok {
				axes = append(axes, axis)
			}
		}
	}
	return axes
}

// axisIn draws the axis of one grid cell from the seed alone.
func (s sinkShape) axisIn(ctx *Context, gridX, gridY int) (sinkAxis, bool) {
	hash := fnv.New64a()
	var buf []byte
	for _, n := range []int{gridX, gridY} {
		buf = binary.LittleEndian.AppendUint64(buf, uint64(int64(n)))
	}
	_, _ = hash.Write(buf)
	gen := rng.New(ctx.Seed).Derived(SinksStage, ctx.ExportCount^hash.Sum64())
	present, jitterX, jitterY, handed := gen.Float64(), gen.Float64(), gen.Float64(), gen.Float64()
	if present >= s.params["density"] {
		return sinkAxis{}, false
	}
	jitter, spacing := s.params["jitter"], float64(s.spacing)
	return sinkAxis{
		x:     (float64(gridX) + 0.5 + (jitterX-0.5)*jitter) * spacing,
		y:     (float64(gridY) + 0.5 + (jitterY-0.5)*jitter) * spacing,
		riser: handed < s.params["riser_chance"],
	}, true
}

// planePosition is the hex's center on the plane, where neighbors are one unit apart.
func planePosition(hex api.Hex) (x, y float64) {
	return float64(hex.Q) + float64(hex.R)/2, float64(hex.R) * math.Sqrt(3) / 2
}
