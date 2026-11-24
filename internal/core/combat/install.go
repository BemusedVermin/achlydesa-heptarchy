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

const expireSystem = "combat.expire_statuses"

// Install registers the combat events, the primitives and the system that expires statuses. It
// needs effects installed first.
func Install(w *sim.World) error {
	for _, register := range []func(*api.Codec) error{
		api.RegisterEvent[api.PartDamaged],
		api.RegisterEvent[api.PartSevered],
		api.RegisterEvent[api.PartAttached],
		api.RegisterEvent[api.StatusAdded],
	} {
		if err := register(w.Codec); err != nil {
			return err
		}
	}
	for _, p := range primitives() {
		if err := effects.Register(w, p); err != nil {
			return err
		}
	}
	if err := installSides(w); err != nil {
		return err
	}
	return w.Systems.Register(sim.System{Name: expireSystem, Cadence: sim.PerEntry, Run: expireStatuses})
}

func primitives() []effects.Primitive {
	parts := map[string]effects.Kind{"parts": effects.KindChannels}
	return []effects.Primitive{
		{Name: "damage", Params: []string{"parts", "amount"}, Kinds: parts, Apply: applyDamage},
		{Name: "sever", Params: []string{"parts"}, Optional: []string{"when"}, Kinds: parts, Apply: applySever},
		{
			Name: "add_status", Params: []string{"bodies", "status", "ticks"},
			Kinds:         map[string]effects.Kind{"bodies": effects.KindBodies, "status": effects.KindText},
			VerifyContent: verifyKnownIDs, Apply: applyAddStatus,
		},
		{
			Name: "attach_part", Params: []string{"from"}, Optional: []string{"item"},
			Kinds:  map[string]effects.Kind{"from": effects.KindText, "item": effects.KindText},
			Verify: verifyAttach, VerifyContent: verifyAttachContent, Apply: applyAttach, Check: checkAttach,
		},
		{Name: "raise_max_integrity", Params: []string{"parts", "amount"}, Kinds: parts, Apply: applyRaiseMaxIntegrity},
		{
			Name: "reveal", Params: []string{"claim", "value", "origin"}, Optional: []string{"parts"},
			Kinds: map[string]effects.Kind{
				"claim": effects.KindText, "value": effects.KindText, "origin": effects.KindText,
				"parts": effects.KindChannels,
			},
			VerifyContent: verifyKnownIDs, Apply: applyReveal,
		},
		{
			Name: "consume", Params: []string{"item"}, Kinds: map[string]effects.Kind{"item": effects.KindText},
			VerifyContent: verifyKnownIDs, Apply: applyConsume, Check: checkConsume,
		},
		{Name: "act_alone", Apply: applyActAlone},
		{
			Name: "spawn_from_part", Params: []string{"faction", "controller"},
			Kinds:         map[string]effects.Kind{"faction": effects.KindText, "controller": effects.KindText},
			VerifyContent: verifyKnownIDs, Apply: applySpawnFromPart,
		},
	}
}

func installSides(w *sim.World) error {
	for _, register := range []func(*api.Codec) error{
		api.RegisterEvent[api.AlertStarted],
		api.RegisterEvent[api.AlertEnded],
		api.RegisterEvent[api.BodyDied],
		api.RegisterEvent[api.BodyDowned],
		api.RegisterEvent[api.PartyWiped],
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
	if err := timeline.RegisterController(w, stillController, stiller{}); err != nil {
		return err
	}
	if err := timeline.RegisterController(w, huntController, hunter{}); err != nil {
		return err
	}
	if err := RegisterChooser(w, scoreChooser, ScoreChooser{}); err != nil {
		return err
	}
	if err := w.Views.Register(sim.ViewBuilder{Name: bodiesView, Build: buildBodies}); err != nil {
		return err
	}
	effects.SetHostile(w, Hostile)
	timeline.SetHalt(w, PartyWiped)
	sim.Subscribe(w.Bus, "combat.death", onPartDamaged)
	sim.Subscribe(w.Bus, "combat.death_severed", onPartSevered)
	return w.Systems.Register(sim.System{Name: alertSystem, Cadence: sim.PerEntry, Run: alert})
}
