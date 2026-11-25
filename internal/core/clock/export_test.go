package clock_test

import (
	"slices"
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/clock"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// exportDays advances world one day at a time, days times, and returns the day of each export.
func exportDays(tb testing.TB, world *sim.World, clk *clock.Clock, days int) []int {
	tb.Helper()
	var got []int
	world.Bus.SubscribeAll("test.exports", func(_ *sim.World, ev api.Event) {
		if began, ok := ev.(api.ExportBegan); ok {
			got = append(got, began.Day)
		}
	})
	for range days {
		if err := clk.Advance(world, dayTicks); err != nil {
			tb.Fatal(err)
		}
	}
	return got
}

func TestExportDaysAreIrregularAndRepeatable(t *testing.T) {
	first, firstClock, _ := setup(t)
	got := exportDays(t, first, firstClock, 60)
	second, secondClock, _ := setup(t)
	if again := exportDays(t, second, secondClock, 60); !slices.Equal(got, again) {
		t.Fatalf("export days differ between runs: %v vs %v", got, again)
	}
	schedule := calendar.Exports
	if len(got) < 5 || got[0] != schedule.FirstDay {
		t.Fatalf("export days = %v, want the first on day %d and several more", got, schedule.FirstDay)
	}
	for i := 1; i < len(got); i++ {
		if gap := got[i] - got[i-1]; gap < schedule.MinGap || gap > schedule.MaxGap {
			t.Errorf("gap %d in %v is outside [%d, %d]", gap, got, schedule.MinGap, schedule.MaxGap)
		}
	}
	if first.ExportCount != len(got) {
		t.Errorf("ExportCount = %d, want %d", first.ExportCount, len(got))
	}
}

func TestExportScheduleResumesFromSavedState(t *testing.T) {
	whole, wholeClock, _ := setup(t)
	want := exportDays(t, whole, wholeClock, 60)

	before, beforeClock, _ := setup(t)
	early := exportDays(t, before, beforeClock, 12)
	state, err := before.RNG.State()
	if err != nil {
		t.Fatal(err)
	}
	resumed, resumedClock, _ := setup(t)
	if err := resumed.RNG.Restore(state); err != nil {
		t.Fatal(err)
	}
	resumed.Now, resumed.ExportCount, resumed.NextExportDay = before.Now, before.ExportCount, before.NextExportDay
	late := exportDays(t, resumed, resumedClock, 48)
	if got := append(early, late...); !slices.Equal(got, want) {
		t.Fatalf("export days after resuming = %v, want %v", got, want)
	}
}
