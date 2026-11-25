package game

import (
	"fmt"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/explore"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/procgen"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// generator builds the deep around the places from the world's seed and the number of exports that have
// begun. The cells are lazy and depend on nothing else, so a save needs no more than the seed and the count.
func generator(exportCount uint64) explore.Generator {
	return func(w *sim.World, fixed *explore.WorldSource) (hexmap.Source, error) {
		world := w.Content.World()
		if world.Geography == "" {
			return fixed, nil
		}
		stages := procgen.NewStages()
		for _, register := range []func(*procgen.Stages) error{
			procgen.RegisterCaves, procgen.RegisterJoin, procgen.RegisterSinks,
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
		open, err := explore.Terrain(w.Content, world.Open)
		if err != nil {
			return nil, fmt.Errorf("world open: %w", err)
		}
		return procgen.NewSource(procgen.Config{
			Seed:        w.RNG.Seed(),
			ExportCount: exportCount,
			Bounds:      fixed.Bounds(),
			Geography:   geography,
			Stages:      stages,
			Fixed:       fixed,
			Open:        open,
		})
	}
}
