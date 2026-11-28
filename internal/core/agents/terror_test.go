package agents_test

import (
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game/gametest"
)

func TestSeeingADeathRaisesTerror(t *testing.T) {
	_, w, _ := gametest.New(t, 1, func(bundle *content.Bundle) {
		template := bundle.Templates[string(beast)]
		template.Drives = map[string]float64{safety: 1}
		template.Triggers = append([]content.Trigger{{
			On: api.BodyDiedKind, Role: content.RoleTarget,
			Do: []content.EffectRef{{
				Do: "satisfy", Params: map[string]string{"bodies": "witnesses", "drive": safety, "amount": "-3"},
			}},
		}}, template.Triggers...)
		bundle.Templates[string(beast)] = template
	})
	here, there := openPair(t, w)
	victim, err := w.Spawn(beast, &comp.Pos{At: here})
	if err != nil {
		t.Fatal(err)
	}
	witness, err := w.Spawn(beast, &comp.Pos{At: there})
	if err != nil {
		t.Fatal(err)
	}
	w.Bus.Publish(api.BodyDied{Body: w.UIDOf(victim)})
	if _, err := w.Bus.Drain(w); err != nil {
		t.Fatal(err)
	}
	if level := drivesOf(w, witness).Levels[safety]; level != 3 {
		t.Errorf("witness safety %v after the death, want 3", level)
	}
	if level := drivesOf(w, victim).Levels[safety]; level != 0 {
		t.Errorf("victim safety %v after its own death, want 0", level)
	}
}
