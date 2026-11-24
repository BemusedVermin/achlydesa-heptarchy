package harness_test

import (
	"bytes"
	"encoding/csv"
	"slices"
	"strconv"
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/harness"
)

func TestRunFillsTheCombatColumns(t *testing.T) {
	reg := contenttest.Sample(t)
	var buf bytes.Buffer
	cfg := harness.RunConfig{
		Seeds: 2, SeedStart: 1, Steps: 60, Content: reg,
		NewPolicy: func(uint64) harness.Policy { return harness.NewInformedPolicy(reg) },
	}
	if err := harness.Run(cfg, &buf); err != nil {
		t.Fatal(err)
	}
	rows, err := csv.NewReader(&buf).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	column := func(name string) int {
		i := slices.Index(rows[0], name)
		if i < 0 {
			t.Fatalf("no column %q in %v", name, rows[0])
		}
		return i
	}
	last := rows[len(rows)-1]
	for _, name := range []string{"alert_ticks", "landings", "severed"} {
		if n, err := strconv.Atoi(last[column(name)]); err != nil || n == 0 {
			t.Errorf("%s = %q at the end of an informed fight, want a count above zero", name, last[column(name)])
		}
	}
	if got := last[column("outcome")]; got != harness.OutcomeWon {
		t.Errorf("outcome = %q, want %q", got, harness.OutcomeWon)
	}
	// Totals only grow.
	landings := column("landings")
	previous := 0
	for _, row := range rows[1:] {
		if row[0] != "1" {
			continue
		}
		n, err := strconv.Atoi(row[landings])
		if err != nil || n < previous {
			t.Fatalf("landings went from %d to %q", previous, row[landings])
		}
		previous = n
	}
}
