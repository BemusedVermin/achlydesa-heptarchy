package agents_test

import (
	"slices"
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/effects"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game/gametest"
)

func TestACorpseGivesTheDeadFromTheTickItDies(t *testing.T) {
	_, w, _ := gametest.New(t, 1, nil)
	body := spawnAt(t, w, 5)
	affordances := ecs.NewMap[comp.Affordance](w.ECS)
	if affordances.Has(body) {
		t.Fatalf("a living beast gives %v", affordances.Get(body).Tags)
	}
	kill := contenttest.Do("damage", "parts", "self:torso", "amount", "1000")
	if err := effects.Apply(&effects.Context{World: w, Source: body}, []content.EffectRef{kill}); err != nil {
		t.Fatal(err)
	}
	if _, err := w.Bus.Drain(w); err != nil {
		t.Fatal(err)
	}
	dead := ecs.NewMap[comp.Dead](w.ECS)
	if !dead.Has(body) || dead.Get(body).Tick != w.Now {
		t.Fatal("the beast did not die this tick")
	}
	if !affordances.Has(body) || !slices.Equal(affordances.Get(body).Tags, []api.ContentID{"affordance.the_dead"}) {
		t.Errorf("the corpse does not give the dead at tick %d", w.Now)
	}
}
