package explore_test

import (
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

func heightGlyph(below int) string {
	switch {
	case below > 0:
		return "v"
	case below < 0:
		return "^"
	}
	return "."
}

func heightWeight(below int) int {
	switch {
	case below <= -4:
		return 1
	case below >= -1 && below < 0:
		return -1
	}
	return 0
}

func drawnRight(c api.CellView, below int) bool {
	return c.Below == below && c.Glyph == heightGlyph(below) && c.Weight == heightWeight(below)
}

func TestADropOptionAsksAndALevelStepDoesNot(t *testing.T) {
	fx := newFixture(t, "place.test.descent", nil)
	for range 7 {
		fx.stepLeader("e")
	}
	fx.toLeader()
	if got := fx.option("step/ne").Confirm; got != "Jump down 3 cells?" {
		t.Errorf("drop confirm = %q, want the question for 3 cells", got)
	}
	for _, id := range []string{"step/e", "step/w"} {
		if got := fx.option(id).Confirm; got != "" {
			t.Errorf("%s confirm = %q, want none", id, got)
		}
	}
}

func TestFloorTheStepRulesCannotReachIsUnreachable(t *testing.T) {
	fx := newFixture(t, "place.test.descent", func(b *content.Bundle) {
		b.Party.MaxStepDown, b.Party.MaxDrop = 0, 0
	})
	for range 5 {
		fx.stepLeader("e")
	}
	fx.toLeader()
	var unreachable, reachable int
	for _, c := range fx.view().Map.Cells {
		if c.Wall {
			if c.Unreachable {
				t.Errorf("wall %v marked unreachable", c.Hex)
			}
			continue
		}
		if c.Unreachable {
			unreachable++
			if c.Below <= 0 {
				t.Errorf("cell %v at Below %d marked unreachable", c.Hex, c.Below)
			}
		} else {
			reachable++
		}
	}
	if unreachable == 0 || reachable == 0 {
		t.Errorf("unreachable %d, reachable %d; want both", unreachable, reachable)
	}
}
