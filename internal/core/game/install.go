package game

import (
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/actions"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/clock"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/combat"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/effects"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/explore"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/inventory"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

// installAll calls each package's Install in a fixed order. This is the only place packages are
// named: adding a feature package adds one line here.
func installAll(w *sim.World, clk *clock.Clock) error {
	steps := []func() error{
		func() error { return clock.Install(w, clk) },
		func() error { return timeline.Install(w, clk) },
		func() error { return anatomy.Install(w) },
		func() error { return inventory.Install(w) },
		func() error { return effects.Install(w) },
		func() error { return actions.Install(w) },
		func() error { return combat.Install(w) },
		func() error { return explore.Install(w) },
	}
	for _, step := range steps {
		if err := step(); err != nil {
			return err
		}
	}
	return nil
}
