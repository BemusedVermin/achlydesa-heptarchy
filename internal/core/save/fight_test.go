package save_test

import (
	"reflect"
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/harness"
)

// landsOf lists the "lands" entries of the timeline, which show the actions in flight.
func landsOf(tb testing.TB, g *game.Game) []api.TimelineEntry {
	tb.Helper()
	view, err := g.View()
	if err != nil {
		tb.Fatal(err)
	}
	var lands []api.TimelineEntry
	for _, entry := range view.Timeline.Entries {
		if entry.Kind == api.EntryLands {
			lands = append(lands, entry)
		}
	}
	return lands
}

// TestSavingMidFightWithActionsInFlightContinuesIdentically saves in the middle of the hall's
// fight, with swings winding up (each a pending entity with its kind-specific landing component),
// loads, runs on, and wants the same actions in flight and the same hash as the game that never saved.
func TestSavingMidFightWithActionsInFlightContinuesIdentically(t *testing.T) {
	reg := contenttest.Sample(t)
	cfg := game.Config{Seed: 3, Content: reg}
	g, err := game.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	policy := harness.NewInformedPolicy(reg)
	policy.Bind(g)
	step := func(each *game.Game, policy harness.Policy) {
		t.Helper()
		if _, ok, err := harness.Step(each, policy); err != nil || !ok {
			t.Fatalf("step: ok %v, err %v", ok, err)
		}
	}
	for range 200 {
		if len(landsOf(t, g)) >= 2 {
			break
		}
		step(g, policy)
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
	loadedPolicy := harness.NewInformedPolicy(reg)
	loadedPolicy.Bind(loaded.g)
	for i := range 40 {
		step(g, policy)
		step(loaded.g, loadedPolicy)
		if hash(t, loaded.g) != hash(t, g) {
			t.Fatalf("hashes differ %d commands after loading", i+1)
		}
		if want, got := landsOf(t, g), landsOf(t, loaded.g); !reflect.DeepEqual(want, got) {
			t.Fatalf("the actions in flight differ %d commands after loading", i+1)
		}
	}
}
