package save_test

import (
	"reflect"
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/harness"
)

// landsOf lists the "lands" entries of the timeline, which show the actions in flight.
func landsOf(t testing.TB, g *game.Game) []api.TimelineEntry {
	t.Helper()
	view, err := g.View()
	if err != nil {
		t.Fatal(err)
	}
	var lands []api.TimelineEntry
	for _, entry := range view.Timeline.Entries {
		if entry.Kind == "lands" {
			lands = append(lands, entry)
		}
	}
	return lands
}

// TestSavingMidFightWithActionsInFlightContinuesIdentically saves in the middle of the hall's
// fight, with swings winding up (each a pending entity with its kind-specific landing component),
// loads, runs on, and wants the same actions in flight and the same hash as the game that never saved.
func TestSavingMidFightWithActionsInFlightContinuesIdentically(t *testing.T) {
	// The hall as shipped: the crawler hunts, and the party fights back.
	reg, err := content.Load(sampleDir)
	if err != nil {
		t.Fatal(err)
	}
	cfg := game.Config{Seed: 3, Content: reg}
	g, err := game.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	policy := harness.NewInformedPolicy(reg)
	policy.Bind(g)
	// step has the policy choose from the first game's view, and sends the same command to every game.
	step := func(games ...*game.Game) {
		t.Helper()
		view, err := g.View()
		if err != nil {
			t.Fatal(err)
		}
		option, ok := policy.Choose(view)
		if !ok {
			t.Fatal("the policy has nothing to do")
		}
		for _, each := range games {
			if _, err := each.Apply(option.Command); err != nil {
				t.Fatal(err)
			}
		}
	}
	for range 200 {
		if len(landsOf(t, g)) >= 2 {
			break
		}
		step(g)
	}
	if len(landsOf(t, g)) < 2 {
		t.Fatal("the fight never had two actions in flight")
	}

	loaded, warnings := reload(t, cfg, g)
	if len(warnings) != 0 {
		t.Errorf("warnings = %v, want none", warnings)
	}
	if hash(t, loaded.g) != hash(t, g) {
		t.Fatal("hash changed across save and load")
	}
	if want, got := landsOf(t, g), landsOf(t, loaded.g); !reflect.DeepEqual(want, got) {
		t.Fatalf("the actions in flight differ after loading:\nwant %+v\n got %+v", want, got)
	}
	for i := range 40 {
		step(g, loaded.g)
		if hash(t, loaded.g) != hash(t, g) {
			t.Fatalf("hashes differ %d commands after loading", i+1)
		}
		if want, got := landsOf(t, g), landsOf(t, loaded.g); !reflect.DeepEqual(want, got) {
			t.Fatalf("the actions in flight differ %d commands after loading", i+1)
		}
	}
}
