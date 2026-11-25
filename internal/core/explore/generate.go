package explore

import (
	"fmt"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/worldgen"
)

// cellSource generates the deep around the places from the world's seed. With no geography the places sit in
// the world's fill. A loaded game builds it before its saved export count is restored, so it regenerates on
// its next export.
func cellSource(w *sim.World, fixed worldgen.Fixed, bounds hexmap.Bounds) (hexmap.Source, error) {
	world := w.Content.World()
	if world.Geography == "" {
		return fixed, nil
	}
	stages := worldgen.NewStages()
	for _, register := range []func(*worldgen.Stages) error{
		worldgen.RegisterCaves, worldgen.RegisterJoin, worldgen.RegisterSinks, worldgen.RegisterVerify,
	} {
		if err := register(stages); err != nil {
			return nil, err
		}
	}
	if err := stages.Validate(w.Content); err != nil {
		return nil, err
	}
	geography, ok := w.Content.Geography(world.Geography)
	if !ok {
		return nil, fmt.Errorf("unknown geography %q", world.Geography)
	}
	open, err := Terrain(w.Content, world.Open)
	if err != nil {
		return nil, fmt.Errorf("world open: %w", err)
	}
	return worldgen.NewSource(worldgen.Config{
		Seed:        w.RNG.Seed(),
		ExportCount: uint64(max(w.ExportCount, 0)),
		Bounds:      bounds,
		Geography:   geography,
		Stages:      stages,
		Fixed:       fixed,
		Open:        open,
	})
}
