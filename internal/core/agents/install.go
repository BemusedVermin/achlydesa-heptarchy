package agents

import (
	"errors"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/effects"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

// needsSystem is the per-phase system that raises needs and applies their thresholds.
const needsSystem = "agents.needs"

// Install registers the needs spawner, the satisfy primitive, the needs system, the behavior
// primitives and the agents controller. It needs effects and actions installed first.
func Install(w *sim.World) error {
	if w.Content == nil {
		return errors.New("agents: install: world has no content")
	}
	if err := w.RegisterSpawner(sim.Spawner{Name: "needs", Apply: spawnNeeds}); err != nil {
		return err
	}
	if err := effects.Register(w, satisfyPrimitive()); err != nil {
		return err
	}
	if err := w.Systems.Register(sim.System{Name: needsSystem, Cadence: sim.PerPhase, Run: tickNeeds}); err != nil {
		return err
	}
	for _, p := range primitives() {
		if err := RegisterPrimitive(w, p); err != nil {
			return err
		}
	}
	return timeline.RegisterController(w, controllerName, controller{})
}
