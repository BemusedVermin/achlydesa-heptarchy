package timeline_test

import (
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

// Two idle bodies must not wake each other a tick at a time while the player is busy far ahead.
func TestIdleBodiesDoNotWakeEachOther(t *testing.T) {
	fx := newFixture(t, 1)
	fx.spawn(t, timeline.IdleController, 2)
	fx.spawn(t, timeline.IdleController, 2)
	player := fx.spawn(t, timeline.PlayerController, 2)
	if current, err := timeline.Run(fx.w, fx.clk); err != nil || current != player {
		t.Fatalf("Run = %v, %v; want the player", current, err)
	}
	if err := timeline.Occupy(fx.w, fx.bodyChannel(player), 5000); err != nil {
		t.Fatal(err)
	}
	if current, err := timeline.Run(fx.w, fx.clk); err != nil || current != player || fx.w.Now != 5000 {
		t.Fatalf("Run = %v, %v at tick %d; want the player at tick 5000", current, err, fx.w.Now)
	}
}
