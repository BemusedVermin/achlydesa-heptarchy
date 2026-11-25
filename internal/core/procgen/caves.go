package procgen

import (
	"errors"
	"fmt"
	"math"
	"slices"

	"github.com/ojrac/opensimplex-go"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/rng"
)

// CavesStage is the name the caves stage registers under.
const CavesStage = "caves"

var caveParams = []string{
	"cheese_frequency", "cheese_threshold",
	"spaghetti_frequency", "spaghetti_width",
	"noodle_frequency", "noodle_width",
	"vertical_scale", "solid_layers", "fade_layers", "deep_rarity",
}

// RegisterCaves adds the caves stage to stages.
//
// The stage carves the solid fill with three world-space 3D noises, joined by union: cheese (one noise above
// a threshold) makes chambers, spaghetti (two noises both near zero) makes long tunnels, and noodle is
// spaghetti finer and tighter, for crawl soops. Every param is required:
//
//   - cheese_frequency, spaghetti_frequency, noodle_frequency: noise cycles per hex.
//   - cheese_threshold: in (0, 1); a cell is chamber when its noise is above it.
//   - spaghetti_width, noodle_width: in (0, 1); a cell is tunnel when both noises are within it of zero.
//   - vertical_scale: multiplies depth before sampling; above 1 makes passages lower.
//   - solid_layers: depth steps below the minimum depth that stay solid.
//   - fade_layers: depth steps over which carving grows from none to full after the solid layers.
//   - deep_rarity: in [0, 1); how far carving thins out by the maximum depth.
func RegisterCaves(stages *Stages) error {
	return stages.Register(CavesStage, caves{}, checkCaveParams)
}

func checkCaveParams(params map[string]float64) error {
	for _, name := range caveParams {
		if _, ok := params[name]; !ok {
			return fmt.Errorf("missing param %q", name)
		}
	}
	for name := range params {
		if !slices.Contains(caveParams, name) {
			return fmt.Errorf("takes no param %q", name)
		}
	}
	for _, name := range []string{
		"cheese_frequency", "spaghetti_frequency", "noodle_frequency", "vertical_scale",
	} {
		if params[name] <= 0 {
			return fmt.Errorf("param %q must be above 0", name)
		}
	}
	for _, name := range []string{"cheese_threshold", "spaghetti_width", "noodle_width"} {
		if params[name] <= 0 || params[name] >= 1 {
			return fmt.Errorf("param %q must be between 0 and 1", name)
		}
	}
	if params["solid_layers"] < 0 || params["fade_layers"] < 0 {
		return errors.New("solid_layers and fade_layers must not be negative")
	}
	if params["deep_rarity"] < 0 || params["deep_rarity"] >= 1 {
		return errors.New(`param "deep_rarity" must be in [0, 1)`)
	}
	return nil
}

type caves struct{}

type caveNoise struct {
	cheese, spaghettiA, spaghettiB, noodleA, noodleB opensimplex.Noise
}

func newCaveNoise(seed, exportCount uint64) caveNoise {
	gen := rng.New(seed).Derived(CavesStage, exportCount)
	next := func() opensimplex.Noise { return opensimplex.New(int64(gen.Uint64())) }
	return caveNoise{cheese: next(), spaghettiA: next(), spaghettiB: next(), noodleA: next(), noodleB: next()}
}

// caveField says which cells the caves stage opens, from the seed alone.
type caveField struct {
	noise  caveNoise
	bounds hexmap.Bounds
	params map[string]float64
}

func newCaveField(seed, exportCount uint64, bounds hexmap.Bounds, params map[string]float64) caveField {
	return caveField{noise: newCaveNoise(seed, exportCount), bounds: bounds, params: params}
}

// open reports whether the stage carves the cell, ignoring the fixed places.
func (f caveField) open(cell api.Cell) bool {
	if !f.bounds.Contains(cell) {
		return false
	}
	strength := caveStrength(f.bounds, cell.Depth, f.params)
	return strength > 0 && carved(f.noise, cell, strength, f.params)
}

// Run carves every in-bounds cell of the chunk that the noises open.
func (caves) Run(ctx *Context, chunk *hexmap.Chunk) error {
	field := newCaveField(ctx.Seed, ctx.ExportCount, ctx.Bounds, ctx.Params)
	origin := ctx.Key.Origin()
	for depth := range hexmap.ChunkSize {
		for r := range hexmap.ChunkSize {
			for q := range hexmap.ChunkSize {
				cell := api.Cell{
					Hex:   api.Hex{Q: origin.Hex.Q + q, R: origin.Hex.R + r},
					Depth: origin.Depth + depth,
				}
				if field.open(cell) {
					chunk.Set(q, r, depth, ctx.Open)
				}
			}
		}
	}
	return nil
}

// caveStrength is 0 where the ground stays solid and 1 at full carving, thinning toward the deepest layer.
func caveStrength(bounds hexmap.Bounds, depth int, params map[string]float64) float64 {
	below := float64(depth - bounds.MinDepth)
	solid := params["solid_layers"]
	if below < solid {
		return 0
	}
	strength := 1.0
	if fade := params["fade_layers"]; fade > 0 {
		strength = min(1, (below-solid+1)/(fade+1))
	}
	if span := float64(bounds.MaxDepth - bounds.MinDepth); span > 0 {
		strength *= 1 - params["deep_rarity"]*below/span
	}
	return strength
}

// carved samples each noise at the cell's world-space position: axial hex to plane, depth as the third axis.
func carved(noise caveNoise, cell api.Cell, strength float64, params map[string]float64) bool {
	x := float64(cell.Hex.Q) + float64(cell.Hex.R)/2
	y := float64(cell.Hex.R) * math.Sqrt(3) / 2
	z := float64(cell.Depth) * params["vertical_scale"]
	sample := func(n opensimplex.Noise, frequency float64) float64 {
		return n.Eval3(x*frequency, y*frequency, z*frequency)
	}

	threshold := 1 - (1-params["cheese_threshold"])*strength
	if sample(noise.cheese, params["cheese_frequency"]) > threshold {
		return true
	}
	spaghetti, noodle := params["spaghetti_frequency"], params["noodle_frequency"]
	return nearZero(
		sample(noise.spaghettiA, spaghetti),
		sample(noise.spaghettiB, spaghetti),
		params["spaghetti_width"]*strength,
	) ||
		nearZero(sample(noise.noodleA, noodle), sample(noise.noodleB, noodle), params["noodle_width"]*strength)
}

func nearZero(a, b, width float64) bool { return math.Abs(a) < width && math.Abs(b) < width }
