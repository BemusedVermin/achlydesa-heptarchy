package game_test

import (
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/actions"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/clock"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game/gametest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

// passiveController never acts, so the loop makes its body wait.
type passiveController struct{}

func (passiveController) Decide(*sim.World, ecs.Entity) error { return nil }

// firstChooser takes the first intent it is offered.
type firstChooser struct{}

func (firstChooser) Choose(_ *sim.World, _ ecs.Entity, intents []actions.Intent) (actions.Intent, bool) {
	if len(intents) == 0 {
		return actions.Intent{}, false
	}
	return intents[0], true
}

func TestAStartMapBodyMayNameAControllerAndChooserThatInstallRegisters(t *testing.T) {
	g, world, cfg := gametest.New(t, 1, func(b *content.Bundle) {
		movingClay := b.Templates["template.creature.moving_clay"]
		movingClay.Actor = &content.ActorDef{Controller: "test.passive", Chooser: "test.first"}
		b.Templates["template.creature.moving_clay"] = movingClay
	}, func(cfg *game.Config) {
		cfg.Install = func(w *sim.World, _ *clock.Clock) error {
			if err := timeline.RegisterController(w, "test.passive", passiveController{}); err != nil {
				return err
			}
			return actions.RegisterChooser(w, "test.first", firstChooser{})
		}
	})
	var named int
	query := ecs.NewFilter1[comp.Actor](world.ECS).Query()
	for query.Next() {
		if query.Get().Controller == "test.passive" {
			named++
		}
	}
	if named == 0 {
		t.Fatal("no body on the start map names the test.passive controller")
	}
	gametest.Reload(t, cfg, g)
}
