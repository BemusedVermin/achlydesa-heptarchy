package explore_test

import (
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

func TestMoveReadoutNamesEachKindOfStep(t *testing.T) {
	fx := newFixture(t, "place.test.descent", nil)
	for range 7 {
		fx.stepLeader("e")
	}
	want := map[api.Dir]struct {
		kind  string
		below int
	}{
		api.DirE:  {api.MoveDown, 1},
		api.DirNE: {api.MoveDrop, 3},
		api.DirNW: {api.MoveBlocked, 0},
		api.DirW:  {api.MoveUp, -1},
	}
	moves := fx.view().Map.Moves
	if len(moves) != api.DirCount {
		t.Fatalf("%d moves, want %d", len(moves), api.DirCount)
	}
	for dir, w := range want {
		if got := moves[dir]; got.Dir != dir || got.Kind != w.kind || got.Below != w.below || got.Label == "" {
			t.Errorf("move %d = %+v, want kind %s below %d", dir, got, w.kind, w.below)
		}
	}
	if got := fx.view().Map.Elevation; got != "Depth 1" {
		t.Errorf("elevation = %q, want Depth 1", got)
	}
}

func TestOnlySolidGroundBumps(t *testing.T) {
	fx := newFixture(t, "place.test.descent", nil)
	for range 7 {
		fx.stepLeader("e")
	}
	if o := fx.option("step/nw"); o.Enabled || o.Reason != "Packed clay. No way through." {
		t.Errorf("step/nw = %+v, want the bump reason", o)
	}
}
