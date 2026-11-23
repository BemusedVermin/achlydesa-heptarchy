package harness_test

import (
	"flag"
	"slices"
	"sync"
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/harness"
)

const lethalitySteps = 400

// lethalitySeeds is how many seeds each policy plays. The acceptance targets are over 100.
var lethalitySeeds = flag.Int("lethality-seeds", 100, "seeds per policy in the lethality test")

// playFights plays the hall's fight over seeds 1..lethalitySeeds with a policy per seed.
func playFights(t *testing.T, reg *content.Registry, newPolicy func() harness.Policy) []harness.FightResult {
	t.Helper()
	results := make([]harness.FightResult, *lethalitySeeds)
	errs := make([]error, *lethalitySeeds)
	var wg sync.WaitGroup
	for i := range results {
		wg.Add(1)
		go func() {
			defer wg.Done()
			results[i], errs[i] = harness.PlayFight(reg, uint64(i+1), newPolicy(), lethalitySteps)
		}()
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Fatalf("seed %d: %v", i+1, err)
		}
	}
	return results
}

func share(results []harness.FightResult, outcome string) float64 {
	n := 0
	for _, r := range results {
		if r.Outcome == outcome {
			n++
		}
	}
	return float64(n) / float64(len(results))
}

func median(results []harness.FightResult) int64 {
	var ticks []int64
	for _, r := range results {
		ticks = append(ticks, r.Ticks)
	}
	slices.Sort(ticks)
	return ticks[len(ticks)/2]
}

// The targets are placeholders for M7's tuning: combat is lethal, and knowing the materials is what
// wins it.
func TestLethalityTargets(t *testing.T) {
	reg, err := content.Load(sampleDir)
	if err != nil {
		t.Fatal(err)
	}
	blind := playFights(t, reg, func() harness.Policy { return harness.NewBlindPolicy(reg) })
	informed := playFights(t, reg, func() harness.Policy { return harness.NewInformedPolicy(reg) })
	t.Logf("blind: lost %.2f won %.2f undecided %.2f, median %d ticks", share(blind, harness.OutcomeLost),
		share(blind, harness.OutcomeWon), share(blind, harness.OutcomeNone), median(blind))
	t.Logf("informed: lost %.2f won %.2f undecided %.2f, median %d ticks", share(informed, harness.OutcomeLost),
		share(informed, harness.OutcomeWon), share(informed, harness.OutcomeNone), median(informed))
	if got := share(blind, harness.OutcomeLost); got < 0.60 {
		t.Errorf("a blind party lost %.0f%% of fights, want at least 60%%", got*100)
	}
	if got := share(informed, harness.OutcomeWon); got < 0.80 {
		t.Errorf("an informed party won %.0f%% of fights, want at least 80%%", got*100)
	}
	if got := median(informed); got >= 60 {
		t.Errorf("the median informed fight lasted %d ticks, want under 60", got)
	}
	if got := median(blind); got >= 60 {
		t.Errorf("the median blind fight lasted %d ticks, want under 60", got)
	}
}
