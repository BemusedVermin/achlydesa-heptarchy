package harness_test

import (
	"sync"
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/harness"
)

const (
	emergenceSeeds = 100
	emergenceDays  = 30
)

// TestNoDegenerateCamping plays the sample world for 30 days on 100 seeds and expects no family to
// settle into one town.
func TestNoDegenerateCamping(t *testing.T) {
	if testing.Short() {
		t.Skip("plays 100 games of 30 days")
	}
	bundle := contenttest.SampleBundle(t)
	bundle.Start = bundle.Starts["start.sample"]
	reg := contenttest.Registry(t, bundle)
	results := make([]harness.EmergenceResult, emergenceSeeds)
	errs := make([]error, emergenceSeeds)
	var wg sync.WaitGroup
	for i := range results {
		wg.Go(func() {
			seed := uint64(i + 1)
			results[i], errs[i] = harness.PlayDays(reg, seed, harness.NewRestPolicy(reg), emergenceDays, 5000, harness.DefaultCamping())
		})
	}
	wg.Wait()
	camped := 0
	for i, result := range results {
		if errs[i] != nil {
			t.Fatalf("seed %d: %v", i+1, errs[i])
		}
		if len(result.Camps) > 0 {
			camped++
			if camped <= 3 {
				t.Errorf("seed %d: camping %+v", result.Seed, result.Camps)
			}
		}
	}
	if camped > 0 {
		t.Errorf("%d of %d seeds showed camping", camped, emergenceSeeds)
	}
}
