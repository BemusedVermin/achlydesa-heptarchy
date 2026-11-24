package actions_test

import (
	"strings"
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/actions"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// firstChooser takes the first intent it is offered.
type firstChooser struct{}

func (firstChooser) Choose(_ *sim.World, _ ecs.Entity, intents []actions.Intent) (actions.Intent, bool) {
	if len(intents) == 0 {
		return actions.Intent{}, false
	}
	return intents[0], true
}

func TestChooserForIsTheTemplatesChooserElseTheDefault(t *testing.T) {
	var defaultChooser actions.Chooser
	fx := newFixture(t, 1, opts{
		install: func(w *sim.World) error {
			if err := actions.RegisterChooser(w, "test.first", firstChooser{}); err != nil {
				return err
			}
			var ok bool
			if defaultChooser, ok = actions.ChooserFor(w, ecs.Entity{}); !ok {
				t.Error("no default chooser is registered")
			}
			return nil
		},
		mutate: func(b *content.Bundle) {
			fighter := b.Templates["template.test.fighter"]
			fighter.Actor = &content.ActorDef{Controller: fighter.Actor.Controller, Chooser: "test.first"}
			b.Templates["template.test.fighter"] = fighter
		},
	})
	if got, ok := actions.ChooserFor(fx.w, fx.fighter); !ok || got != (firstChooser{}) {
		t.Errorf("fighter's chooser = %v, %v; want the test.first chooser", got, ok)
	}
	if got, ok := actions.ChooserFor(fx.w, fx.dummy); !ok || got != defaultChooser {
		t.Errorf("dummy's chooser = %v, %v; want the %q chooser", got, ok, actions.DefaultChooser)
	}
}

func TestRegisterChooserRefusesDuplicatesAndBlanks(t *testing.T) {
	fx := newFixture(t, 1, opts{})
	tests := []struct {
		name    string
		chooser string
		impl    actions.Chooser
	}{
		{"a duplicate name", actions.DefaultChooser, firstChooser{}},
		{"no name", "", firstChooser{}},
		{"no implementation", "test.none", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := actions.RegisterChooser(fx.w, tt.chooser, tt.impl); err == nil {
				t.Error("RegisterChooser accepted it")
			}
		})
	}
}

func TestSpawningATemplateWithAnUnknownChooserFails(t *testing.T) {
	fx := newFixture(t, 1, opts{mutate: func(b *content.Bundle) {
		b.Templates["template.test.lost"] = content.Template{
			Actor: &content.ActorDef{Controller: "idle", Chooser: "no.such.chooser"},
		}
	}})
	_, err := fx.w.Spawn("template.test.lost", nil)
	if err == nil || !strings.Contains(err.Error(), `unknown chooser "no.such.chooser"`) {
		t.Errorf("Spawn error = %v, want an unknown chooser", err)
	}
}
