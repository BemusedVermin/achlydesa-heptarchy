package harness_test

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/harness"
)

const (
	emergenceSeeds = 100
	emergenceDays  = 30
)

var update = flag.Bool("update", false, "rewrite the golden files")

// sampleRegistry is the sample world's content, starting on the sample start.
func sampleRegistry(tb testing.TB) *content.Registry {
	tb.Helper()
	bundle := contenttest.SampleBundle(tb)
	bundle.Start = bundle.Starts["start.sample"]
	return contenttest.Registry(tb, bundle)
}

// TestNoDegenerateCamping plays the sample world for 30 days on 100 seeds and expects no family to
// settle into one town.
func TestNoDegenerateCamping(t *testing.T) {
	if testing.Short() {
		t.Skip("plays 100 games of 30 days")
	}
	reg := sampleRegistry(t)
	results := make([]harness.EmergenceResult, emergenceSeeds)
	errs := make([]error, emergenceSeeds)
	var wg sync.WaitGroup
	limit := make(chan struct{}, runtime.NumCPU())
	for i := range results {
		wg.Go(func() {
			limit <- struct{}{}
			defer func() { <-limit }()
			seed := uint64(i + 1)
			results[i], errs[i] = harness.PlayDays(reg, seed, harness.NewRestPolicy(reg), emergenceDays, 5000, harness.DefaultCamping())
		})
	}
	wg.Wait()
	camped, migrations, emptied := 0, 0, 0
	for i, result := range results {
		if errs[i] != nil {
			t.Fatalf("seed %d: %v", i+1, errs[i])
		}
		migrations += result.Migrations
		emptied += result.TownsEmptied
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
	t.Logf("migrations %d, towns emptied %d over %d seeds", migrations, emptied, emergenceSeeds)
}

// TestNewColumnsCountEvents plays the sample world and expects the new columns to be filled.
func TestNewColumnsCountEvents(t *testing.T) {
	reg := sampleRegistry(t)
	result, err := harness.PlayDays(reg, 1, harness.NewRestPolicy(reg), 12, 2000, harness.DefaultCamping())
	if err != nil {
		t.Fatal(err)
	}
	if result.Deaths == 0 || result.Migrations == 0 {
		t.Errorf("seed 1 over 12 days: deaths %d, migrations %d, want both above zero", result.Deaths, result.Migrations)
	}
}

// TestWaterScenario replays seed 1 on the sample world: residents of the shallow town fetch water
// and move about, the town runs short and its people die off. The golden file lists the deaths and
// migrations and how many moves the town's bodies made, each by day.
func TestWaterScenario(t *testing.T) {
	reg := sampleRegistry(t)
	g, err := game.New(game.Config{Seed: 1, Content: reg})
	if err != nil {
		t.Fatal(err)
	}
	policy := harness.NewRestPolicy(reg)
	var lines []string
	moves := map[int]int{}
	day := 1
	for day <= 12 {
		events, ok, err := harness.Step(g, policy)
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			break
		}
		for _, event := range events {
			switch e := event.(type) {
			case api.Moved:
				moves[day]++
			case api.BodyDied:
				lines = append(lines, fmt.Sprintf("day %02d: died: %s", day, e.Name))
			case api.Migrated:
				lines = append(lines, fmt.Sprintf("day %02d: migrated %d %s from map %d to map %d, short %q",
					day, e.Count, e.Family, e.From, e.To, e.Short))
			}
		}
		view, err := g.View()
		if err != nil {
			t.Fatal(err)
		}
		day = view.Clock.Day
	}
	for d := 1; d <= 12; d++ {
		lines = append(lines, fmt.Sprintf("day %02d: %d moves", d, moves[d]))
	}
	slices.Sort(lines)
	got := strings.Join(lines, "\n") + "\n"
	path := filepath.Join("testdata", "water_scenario.golden")
	if *update {
		if err := os.MkdirAll("testdata", 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != string(want) {
		t.Errorf("scenario differs from %s:\n%s", path, got)
	}
}
