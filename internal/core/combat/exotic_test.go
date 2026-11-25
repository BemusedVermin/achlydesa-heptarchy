package combat_test

import (
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/effects"
)

func TestALeashedSoulFeederGrowsStrongerNearTheDying(t *testing.T) {
	r := newRig(t)
	feederPos := here
	feeder, err := r.w.Spawn("template.generic.leashed_feeder", &feederPos)
	if err != nil {
		t.Fatal(err)
	}
	strike, ok := r.w.Content.Action("action.strike")
	if !ok {
		t.Fatal("no strike")
	}
	swing := func(part ecs.Entity) {
		t.Helper()
		ctx := &effects.Context{
			World: r.w, Source: feeder, Target: r.reaper, TargetPart: part,
			Choices: map[string]string{"type": "type.press"},
		}
		r.apply(t, ctx, strike.OnLand...)
	}
	feeding := func() bool {
		statuses := ecs.NewMap[comp.Statuses](r.w.ECS)
		if !statuses.Has(feeder) {
			return false
		}
		for _, active := range statuses.Get(feeder).Active {
			if active.Status == "status.feeding" {
				return true
			}
		}
		return false
	}
	arms := r.partsOf(r.reaper, "arm")
	torso := r.partsOf(r.reaper, "torso")[0]

	before := r.integrity(arms[0])
	swing(arms[0])
	if feeding() {
		t.Fatal("the feeder grew stronger from a body that is not dying")
	}
	if got := before - r.integrity(arms[0]); got != 3 {
		t.Fatalf("an ordinary blow took %d, want 3", got)
	}

	ctx := r.ctx(r.reaper)
	ctx.TargetPart = torso
	r.apply(t, ctx, contenttest.Do("damage", "parts", "target_part", "amount", "6"))
	swing(arms[0])
	if !feeding() {
		t.Fatal("the feeder did not grow stronger from a dying body")
	}
	before = r.integrity(arms[1])
	swing(arms[1])
	if got := before - r.integrity(arms[1]); got != 5 {
		t.Errorf("a feeding blow took %d, want 5", got)
	}
}
