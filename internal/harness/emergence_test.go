package harness_test

import (
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/harness"
)

var update = flag.Bool("update", false, "rewrite the golden files")

// sampleRegistry is the sample world's content, starting on the sample start.
func sampleRegistry(tb testing.TB) *content.Registry {
	tb.Helper()
	bundle := contenttest.SampleBundle(tb)
	bundle.Start = bundle.Starts["start.sample"]
	return contenttest.Registry(tb, bundle)
}

// TestNewColumnsCountEvents runs the sample world through Run and expects the new columns to fill.
func TestNewColumnsCountEvents(t *testing.T) {
	reg := sampleRegistry(t)
	var buf bytes.Buffer
	cfg := harness.RunConfig{
		Seeds: 1, SeedStart: 1, Steps: 40, Content: reg,
		NewPolicy: func(uint64) harness.Policy { return harness.NewRestPolicy(reg) },
	}
	if err := harness.Run(cfg, &buf); err != nil {
		t.Fatal(err)
	}
	rows, err := csv.NewReader(&buf).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	last := rows[len(rows)-1]
	for _, name := range []string{"deaths", "migrations"} {
		at := slices.Index(rows[0], name)
		if at < 0 {
			t.Fatalf("no column %q in %v", name, rows[0])
		}
		if n, err := strconv.Atoi(last[at]); err != nil || n == 0 {
			t.Errorf("%s = %q at the end of seed 1, want a count above zero", name, last[at])
		}
	}
	if !slices.Contains(rows[0], "towns_emptied") {
		t.Errorf("no towns_emptied column in %v", rows[0])
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
