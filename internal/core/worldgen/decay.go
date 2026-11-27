package worldgen

import (
	"fmt"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
)

// Decay is the terrain decayed ground takes at each depth for one export count.
type Decay struct {
	bounds hexmap.Bounds
	// byDepth is indexed by depth minus the minimum depth; a zero ID leaves the ground as it is.
	byDepth []hexmap.Terrain
}

// NewDecay reads the geography's decay rule for every depth in the bounds. A geography without decay gives
// a nil Decay, which changes nothing.
func NewDecay(
	reg *content.Registry,
	geography api.ContentID,
	exportCount uint64,
	bounds hexmap.Bounds,
) (*Decay, error) {
	def, ok := reg.Geography(geography)
	if !ok {
		return nil, fmt.Errorf("worldgen: decay: unknown geography %q", geography)
	}
	if def.Decay == nil {
		return nil, nil
	}
	terrains := make([]hexmap.Terrain, len(def.Decay.Terrains))
	for i, id := range def.Decay.Terrains {
		terrain, ok := reg.Terrain(id)
		if !ok {
			return nil, fmt.Errorf("worldgen: decay: unknown terrain %q", id)
		}
		terrains[i] = hexmap.Terrain{ID: id, Solid: terrain.Solid, Opaque: terrain.Opaque}
	}
	decay := &Decay{bounds: bounds}
	for depth := bounds.MinDepth; depth <= bounds.MaxDepth; depth++ {
		level, err := reg.DecayLevel(geography, content.DecayEnv{Depth: depth, Exports: int(exportCount)})
		if err != nil {
			return nil, fmt.Errorf("worldgen: decay: depth %d: %w", depth, err)
		}
		var terrain hexmap.Terrain
		if level > 0 {
			terrain = terrains[min(int(level), len(terrains))-1]
		}
		decay.byDepth = append(decay.byDepth, terrain)
	}
	return decay, nil
}

// Terrain is what decayed ground at the depth is made of; ok is false where it is not decayed.
func (d *Decay) Terrain(depth int) (terrain hexmap.Terrain, ok bool) {
	if d == nil || depth < d.bounds.MinDepth || depth > d.bounds.MaxDepth {
		return hexmap.Terrain{}, false
	}
	terrain = d.byDepth[depth-d.bounds.MinDepth]
	return terrain, terrain.ID != ""
}
