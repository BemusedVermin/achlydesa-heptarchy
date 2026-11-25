package explore

import (
	"fmt"
	"maps"
	"slices"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// populateLoaded gives every loaded chunk its starting bodies, once. Chunks go in key order and a chunk's
// bodies in cell order, so UIDs do not depend on how the chunks came to load. A cell already holding a body
// gets none.
func populateLoaded(w *sim.World) error {
	st := stateOf(w)
	if w.Map == nil || st.generated == nil {
		return nil
	}
	for _, key := range w.Map.Loaded() {
		if st.populated[key] {
			continue
		}
		spawns, err := st.generated.Populate(key)
		if err != nil {
			return fmt.Errorf("explore: populate chunk %v: %w", key, err)
		}
		st.populated[key] = true
		for _, spawn := range spawns {
			if _, occupied := bodyAt(w, spawn.Cell, ecs.Entity{}); occupied {
				continue
			}
			if _, err := w.Spawn(spawn.Template, &comp.Pos{At: spawn.Cell}); err != nil {
				return fmt.Errorf("explore: populate: spawn %q at %v: %w", spawn.Template, spawn.Cell, err)
			}
		}
	}
	return nil
}

// Populated lists the chunks whose starting bodies have been spawned, in key order.
func Populated(w *sim.World) []hexmap.ChunkKey {
	keys := slices.Collect(maps.Keys(stateOf(w).populated))
	slices.SortFunc(keys, hexmap.CompareKeys)
	return keys
}

// RestorePopulated marks chunks as populated, as a loaded save had them.
func RestorePopulated(w *sim.World, keys []hexmap.ChunkKey) {
	st := stateOf(w)
	for _, key := range keys {
		st.populated[key] = true
	}
}
