package agents_test

import (
	"maps"
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/effects"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game/gametest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

const (
	beast = api.ContentID("template.creature.beast")
	water = api.ContentID("drive.water")
	food  = api.ContentID("drive.food")
)

// thirstyBeasts gives beasts water and food drives, water rising by 1 + weight a phase and hurting the
// torso from level 2.
func thirstyBeasts(bundle *content.Bundle) {
	template := bundle.Templates[string(beast)]
	template.Drives = map[string]float64{string(water): 1, string(food): 2}
	bundle.Templates[string(beast)] = template
	drive := bundle.Drives[string(water)]
	drive.Rise = "1 + weight"
	drive.Thresholds = []content.Threshold{{
		At: 2, Do: []content.EffectRef{{Do: "damage", Params: map[string]string{"parts": "self:torso", "amount": "1"}}},
	}}
	bundle.Drives[string(water)] = drive
}

func spawnAt(tb testing.TB, w *sim.World, q int) ecs.Entity {
	tb.Helper()
	e, err := w.Spawn(beast, &comp.Pos{At: api.Cell{Hex: api.Hex{Q: q, R: 40}, Depth: 3}})
	if err != nil {
		tb.Fatal(err)
	}
	return e
}

func drivesOf(w *sim.World, e ecs.Entity) comp.Drives {
	return *ecs.NewMap[comp.Drives](w.ECS).Get(e)
}

func torso(tb testing.TB, w *sim.World, body ecs.Entity) int {
	tb.Helper()
	parts := ecs.NewMap[comp.Part](w.ECS)
	for _, part := range anatomy.Parts(w, body) {
		if parts.Get(part).Slot == "torso" {
			return parts.Get(part).Integrity
		}
	}
	tb.Fatal("no torso")
	return 0
}

func phase(tb testing.TB, w *sim.World) {
	tb.Helper()
	if err := w.Systems.Run(w, sim.PerPhase); err != nil {
		tb.Fatal(err)
	}
}

func TestLevelsRiseByTheirFormula(t *testing.T) {
	_, w, _ := gametest.New(t, 1, thirstyBeasts)
	body := spawnAt(t, w, 30)
	weight := drivesOf(w, body).Weights[water]
	for range 2 {
		phase(t, w)
	}
	levels := drivesOf(w, body).Levels
	if want := 2 * (1 + weight); levels[water] != want {
		t.Errorf("water level %v after two phases, want %v", levels[water], want)
	}
	if levels[food] != 2 {
		t.Errorf("food level %v after two phases, want 2", levels[food])
	}
}

func TestThresholdHarmsWhileHeldAndSatisfyEndsIt(t *testing.T) {
	_, w, _ := gametest.New(t, 1, func(bundle *content.Bundle) {
		thirstyBeasts(bundle)
		drive := bundle.Drives[string(water)]
		drive.Rise = "1"
		bundle.Drives[string(water)] = drive
	})
	body := spawnAt(t, w, 30)
	whole := torso(t, w, body)
	phase(t, w)
	if got := torso(t, w, body); got != whole {
		t.Fatalf("torso %d below the threshold, want %d", got, whole)
	}
	phase(t, w)
	phase(t, w)
	if got := torso(t, w, body); got != whole-2 {
		t.Fatalf("torso %d after two phases at the threshold, want %d", got, whole-2)
	}
	satisfy := func(amount string) {
		t.Helper()
		ref := content.EffectRef{Do: "satisfy", Params: map[string]string{
			"bodies": "self", "drive": string(water), "amount": amount,
		}}
		if err := effects.Apply(&effects.Context{World: w, Source: body}, []content.EffectRef{ref}); err != nil {
			t.Fatal(err)
		}
	}
	satisfy("10")
	if level := drivesOf(w, body).Levels[water]; level != 0 {
		t.Fatalf("water level %v after drinking more than enough, want 0", level)
	}
	phase(t, w)
	if got := torso(t, w, body); got != whole-2 {
		t.Errorf("torso %d after drinking, want %d", got, whole-2)
	}
	satisfy("-5")
	if level := drivesOf(w, body).Levels[water]; level != 6 {
		t.Errorf("water level %v after a negative satisfy, want 6", level)
	}
}

func TestWeightsAreIndividualAndReproducible(t *testing.T) {
	spawnBoth := func(reverse bool) (game.Config, *game.Game, map[int]comp.Drives, map[int]api.UID) {
		g, w, cfg := gametest.New(t, 7, thirstyBeasts)
		cells := []int{30, 31}
		if reverse {
			cells = []int{31, 30}
		}
		out, uids := map[int]comp.Drives{}, map[int]api.UID{}
		for _, q := range cells {
			e := spawnAt(t, w, q)
			out[q], uids[q] = drivesOf(w, e), w.UIDOf(e)
		}
		return cfg, g, out, uids
	}
	cfg, g, first, uids := spawnBoth(false)
	if maps.Equal(first[30].Weights, first[31].Weights) {
		t.Errorf("two beasts drew the same weights %v", first[30].Weights)
	}
	for _, typical := range []float64{1, 2} {
		id := map[float64]api.ContentID{1: water, 2: food}[typical]
		if weight := first[30].Weights[id]; weight < typical/2 || weight > typical*1.5 {
			t.Errorf("weight %v on %s, want within half of typical %v", weight, id, typical)
		}
	}
	_, _, second, _ := spawnBoth(true)
	_, loaded, _ := gametest.Reload(t, cfg, g)
	for q, drives := range first {
		if !maps.Equal(drives.Weights, second[q].Weights) {
			t.Errorf("beast at q=%d: weights %v, then %v spawning in the other order", q, drives.Weights, second[q].Weights)
		}
		e, ok := loaded.ByUID(uids[q])
		if !ok {
			t.Fatalf("beast at q=%d is gone after a reload", q)
		}
		if got := drivesOf(loaded, e).Weights; !maps.Equal(got, drives.Weights) {
			t.Errorf("beast at q=%d: weights %v after a reload, want %v", q, got, drives.Weights)
		}
	}
}
