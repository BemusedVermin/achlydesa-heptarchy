package clock_test

import (
	"slices"
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/clock"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// calendar is a 12-tick day of three phases, so the tests don't move when the sample's is tuned.
var calendar = content.Calendar{Phases: []content.Phase{
	{ID: "phase.test.dawn", Name: "text.test.dawn", Ticks: 3},
	{ID: "phase.test.day", Name: "text.test.day", Ticks: 5},
	{ID: "phase.test.night", Name: "text.test.night", Ticks: 4},
}}

const dayTicks = 12

func setup(t *testing.T) (*sim.World, *clock.Clock, *[]string) {
	t.Helper()
	world := sim.NewWorld(1)
	world.Content = contenttest.Sample(t)
	clk := clock.New(calendar)
	if err := clock.Install(world, clk); err != nil {
		t.Fatal(err)
	}
	var calls []string
	record := func(name string, cadence sim.Cadence) {
		err := world.Systems.Register(sim.System{Name: name, Cadence: cadence, Run: func(*sim.World) error {
			calls = append(calls, name)
			return nil
		}})
		if err != nil {
			t.Fatal(err)
		}
	}
	record("test.phase", sim.PerPhase)
	record("test.day", sim.PerDay)
	if err := world.Systems.Freeze(); err != nil {
		t.Fatal(err)
	}
	return world, clk, &calls
}

func TestAdvanceBoundaries(t *testing.T) {
	world, clk, calls := setup(t)
	if err := clk.Advance(world, 2); err != nil || len(*calls) != 0 || world.Now != 2 {
		t.Fatalf("2 ticks: err=%v calls=%v now=%d", err, *calls, world.Now)
	}
	err := clk.Advance(world, 1)
	if err != nil || !slices.Equal(*calls, []string{"test.phase"}) || world.Now != 3 {
		t.Fatalf("3rd tick: err=%v calls=%v now=%d", err, *calls, world.Now)
	}
}

func TestAdvanceDay(t *testing.T) {
	world, clk, calls := setup(t)
	if err := clk.Advance(world, dayTicks); err != nil {
		t.Fatal(err)
	}
	want := []string{"test.phase", "test.phase", "test.phase", "test.day"}
	if !slices.Equal(*calls, want) {
		t.Fatalf("calls = %v, want %v", *calls, want)
	}
}

func TestAdvanceChunkingDoesNotMatter(t *testing.T) {
	const ticks = dayTicks + 1
	bigWorld, bigClock, bigCalls := setup(t)
	if err := bigClock.Advance(bigWorld, ticks); err != nil {
		t.Fatal(err)
	}
	stepWorld, stepClock, stepCalls := setup(t)
	for range ticks {
		if err := stepClock.Advance(stepWorld, 1); err != nil {
			t.Fatal(err)
		}
	}
	if !slices.Equal(*bigCalls, *stepCalls) || len(*bigCalls) == 0 {
		t.Fatalf("batched %v != stepped %v", *bigCalls, *stepCalls)
	}
	if bigWorld.Now != ticks || stepWorld.Now != ticks {
		t.Fatalf("now = %d, %d", bigWorld.Now, stepWorld.Now)
	}
}

func TestAdvanceRejectsNegative(t *testing.T) {
	world, clk, _ := setup(t)
	if err := clk.Advance(world, -1); err == nil {
		t.Fatal("expected an error")
	}
}

func TestView(t *testing.T) {
	world, _, _ := setup(t)
	for _, tc := range []struct {
		now   int64
		day   int
		phase int
		in    int64
	}{
		{0, 1, 0, 0},
		{2, 1, 0, 2},
		{3, 1, 1, 0},
		{6, 1, 1, 3},
		{8, 1, 2, 0},
		{11, 1, 2, 3},
		{12, 2, 0, 0},
	} {
		world.Now = tc.now
		view, err := world.Views.Build(world)
		if err != nil {
			t.Fatal(err)
		}
		phase := calendar.Phases[tc.phase]
		want := api.ClockView{
			Day: tc.day, Phase: world.Content.Render(api.Text{ID: phase.Name}),
			TickInPhase: tc.in, PhaseTicks: phase.Ticks,
		}
		if got := view.Clock; got == nil || *got != want {
			t.Errorf("tick %d: view = %+v, want %+v", tc.now, got, want)
		}
	}
}
