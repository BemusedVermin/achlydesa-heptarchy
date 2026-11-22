package explore_test

import (
	"slices"
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
)

func TestAFollowerStepsWithTheActionPartyFollowActionNames(t *testing.T) {
	for _, known := range []bool{true, false} {
		fx := newFixture(t, "", func(b *content.Bundle) {
			stride := b.Actions["action.step"]
			stride.Name, stride.Key = "text.action.test.stride", "stride"
			b.Actions["action.test.stride"] = stride
			b.Text["text.action.test.stride"] = "stride"
			b.Party.FollowAction = "action.test.stride"
			if !known {
				return
			}
			for id, tmpl := range b.Templates {
				if slices.Contains(tmpl.Actions, "action.step") {
					tmpl.Actions = append(slices.Clone(tmpl.Actions), "action.test.stride")
					b.Templates[id] = tmpl
				}
			}
		})
		members := fx.members()
		got := moves(fx.step("e"), members[1])
		if known && len(got) != 1 {
			t.Errorf("the follower knows the named action but moved %d times, want 1", len(got))
		}
		if !known && len(got) != 0 {
			t.Errorf("the follower does not know the named action but moved %d times, want 0", len(got))
		}
	}
}

func TestTheSpeedFormulaIsContent(t *testing.T) {
	ticks := func(speed string) int64 {
		fx := newFixture(t, "", func(b *content.Bundle) {
			b.Timeline.StepTicks = "speed >= 2 ? 10 : 25"
			b.Timeline.Speed = speed
			legs := b.Parts["part.human_legs"]
			legs.Grants = []comp.Grant{{Kind: "speed", Value: 1}}
			b.Parts["part.human_legs"] = legs
		})
		before := fx.view().Tick
		fx.stepLeader("e")
		fx.stepLeader("e")
		return fx.view().Tick - before
	}
	if plain, boosted := ticks("grant.speed"), ticks("grant.speed + 1"); boosted >= plain {
		t.Errorf("two steps took %d ticks at speed 1 and %d with the formula adding 1; want it faster", plain, boosted)
	}
}
