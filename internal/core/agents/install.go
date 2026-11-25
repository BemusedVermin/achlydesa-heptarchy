package agents

import (
	"errors"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/actions"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/effects"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

// needsSystem is the per-phase system that raises needs and applies their thresholds.
const needsSystem = "agents.needs"

// Install registers the needs spawner, the satisfy primitive, the needs system, the behavior
// primitives, the agents chooser, the region view and the agents controller. It needs effects and actions installed first.
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
	if err := effects.Register(w, takeStockPrimitive()); err != nil {
		return err
	}
	for _, system := range []sim.System{
		{Name: needsSystem, Cadence: sim.PerPhase, Run: tickNeeds},
		{Name: offscreenSystem, Cadence: sim.PerPhase, Run: tickOffscreen},
		{Name: genericsSystem, Cadence: sim.PerEntry, Run: tickGenerics},
		{Name: upkeepSystem, Cadence: sim.PerDay, Run: tickUpkeep},
		{Name: flightSystem, Cadence: sim.PerEntry, Run: tickFlight},
	} {
		if err := w.Systems.Register(system); err != nil {
			return err
		}
	}
	for _, p := range primitives() {
		if err := RegisterPrimitive(w, p); err != nil {
			return err
		}
	}
	if err := actions.RegisterChooser(w, chooserName, attackChooser{}); err != nil {
		return err
	}
	for _, register := range []func(*api.Codec) error{
		api.RegisterEvent[api.Promoted],
		api.RegisterEvent[api.Demoted],
		api.RegisterEvent[api.Migrated],
		api.RegisterEvent[api.TownEmptied],
		api.RegisterEvent[api.PopulationDied],
	} {
		if err := register(w.Codec); err != nil {
			return err
		}
	}
	sim.Subscribe(w.Bus, "agents.lod", onPartyCrossed)
	sim.Subscribe(w.Bus, "agents.emptied", onBodyDied)
	if err := w.Views.Register(sim.ViewBuilder{Name: regionView, Build: buildRegion}); err != nil {
		return err
	}
	if err := installRecruits(w); err != nil {
		return err
	}
	return timeline.RegisterController(w, controllerName, controller{})
}
