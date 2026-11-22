package clock_test

import (
	"reflect"
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/clock"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

const sampleDir = "../../../content/data"

// setup returns a world with a clock and recording systems.
func setup(t *testing.T) (*sim.World, *clock.Clock, *[]string) {
	t.Helper()
	reg, err := content.Load(sampleDir)
	if err != nil {
		t.Fatal(err)
	}
	world := sim.NewWorld(1)
	clk := clock.New(reg.Calendar())
	if err := clock.Install(world, clk, reg); err != nil {
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
	if err := clk.Advance(world, 59); err != nil || len(*calls) != 0 || world.Now != 59 {
		t.Fatalf("59 ticks: err=%v calls=%v now=%d", err, *calls, world.Now)
	}
	if err := clk.Advance(
		world,
		1,
	); err != nil || !reflect.DeepEqual(*calls, []string{"test.phase"}) ||
		world.Now != 60 {
		t.Fatalf("60th tick: err=%v calls=%v now=%d", err, *calls, world.Now)
	}
}

func TestAdvanceDay(t *testing.T) {
	world, clk, calls := setup(t)
	if err := clk.Advance(world, 240); err != nil {
		t.Fatal(err)
	}
	want := []string{"test.phase", "test.phase", "test.phase", "test.phase", "test.day"}
	if !reflect.DeepEqual(*calls, want) {
		t.Fatalf("calls = %v, want %v", *calls, want)
	}
}

func TestAdvanceChunkingDoesNotMatter(t *testing.T) {
	bigWorld, bigClock, bigCalls := setup(t)
	if err := bigClock.Advance(bigWorld, 1000); err != nil {
		t.Fatal(err)
	}
	stepWorld, stepClock, stepCalls := setup(t)
	for range 1000 {
		if err := stepClock.Advance(stepWorld, 1); err != nil {
			t.Fatal(err)
		}
	}
	if !reflect.DeepEqual(*bigCalls, *stepCalls) || len(*bigCalls) == 0 {
		t.Fatalf("batched %v != stepped %v", *bigCalls, *stepCalls)
	}
	if bigWorld.Now != 1000 || stepWorld.Now != 1000 {
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
		phase string
		in    int64
	}{{0, 1, "dawn", 0}, {75, 1, "day", 15}, {239, 1, "night", 59}, {240, 2, "dawn", 0}} {
		world.Now = tc.now
		view, err := world.Views.Build(world)
		if err != nil {
			t.Fatal(err)
		}
		got := view.Clock
		if got == nil || got.Day != tc.day || got.Phase != tc.phase || got.TickInPhase != tc.in ||
			got.TicksPerPhase != 60 {
			t.Errorf("tick %d: view = %+v", tc.now, got)
		}
	}
}
