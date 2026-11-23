package game_test

import (
	"strings"
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
)

// TestControlReturnsOnlyWhenARealActionIsOffered plays the real sample party and steps repeatedly:
// each time control returns, a step must be offered, with no waits to get through recovery.
func TestControlReturnsOnlyWhenARealActionIsOffered(t *testing.T) {
	reg, err := content.Load(sampleDir)
	if err != nil {
		t.Fatal(err)
	}
	g, err := game.New(game.Config{Seed: 1, Content: reg})
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 6; i++ {
		view, err := g.View()
		if err != nil {
			t.Fatal(err)
		}
		var step *api.Option
		for j, o := range view.Options {
			if o.Enabled && strings.HasPrefix(o.ID, "step/") && o.Command != nil {
				step = &view.Options[j]
				break
			}
		}
		if step == nil {
			t.Fatalf("step %d at tick %d: control returned with no step offered: %+v", i, view.Tick, view.Options)
		}
		if _, err := g.Apply(step.Command); err != nil {
			t.Fatal(err)
		}
	}
}
