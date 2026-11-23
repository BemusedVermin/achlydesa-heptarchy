package combat

import (
	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/effects"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

// installSides registers the alert and death events, the allegiance spawner, the hostility hook,
// the alert system and the death subscribers.
func installSides(w *sim.World) error {
	for _, register := range []func(*api.Codec) error{
		api.RegisterEvent[api.AlertStarted],
		api.RegisterEvent[api.AlertEnded],
		api.RegisterEvent[api.BodyDied],
		api.RegisterEvent[api.BodyDowned],
		api.RegisterEvent[api.PartyLost],
	} {
		if err := register(w.Codec); err != nil {
			return err
		}
	}
	err := w.RegisterSpawner(
		sim.Spawner{Name: "allegiance", Apply: func(w *sim.World, e ecs.Entity, t content.Template) error {
			if t.Faction != nil {
				ecs.NewMap[comp.Allegiance](w.ECS).Add(e, &comp.Allegiance{Faction: *t.Faction})
			}
			return nil
		}},
	)
	if err != nil {
		return err
	}
	if err := timeline.RegisterController(w, restController, rester{}); err != nil {
		return err
	}
	effects.SetHostile(w, Hostile)
	sim.Subscribe(w.Bus, "combat.death", onPartDamaged)
	sim.Subscribe(w.Bus, "combat.death_severed", onPartSevered)
	return w.Systems.Register(sim.System{Name: alertSystem, Cadence: sim.PerTurn, Run: alert})
}
