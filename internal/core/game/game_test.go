package game_test

import (
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/clock"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// unknownCommand has no handler.
type unknownCommand struct{}

func (unknownCommand) CommandKind() string { return "test.unknown" }

// config spawns one player-controlled actor, in place of the start scenario.
func config(t *testing.T, seed uint64) game.Config {
	t.Helper()
	bundle := contenttest.SampleBundle(t)
	contenttest.NoStart(&bundle)
	reg := contenttest.Registry(t, bundle)
	return game.Config{Seed: seed, Content: reg, Spawn: func(w *sim.World, _ *clock.Clock) error {
		_, err := w.Spawn("template.person.reaper", nil)
		return err
	}}
}

func TestNewOpensOnTheFirstPlayerDecision(t *testing.T) {
	g, err := game.New(config(t, 1))
	if err != nil {
		t.Fatal(err)
	}
	view, err := g.View()
	if err != nil {
		t.Fatal(err)
	}
	if view.Clock == nil || view.Clock.Day != 1 || view.Clock.Phase != "dawn" {
		t.Errorf("clock = %+v, want day 1, dawn", view.Clock)
	}
	if g.Current() == 0 || view.Timeline == nil || view.Timeline.Current != g.Current() {
		t.Errorf("current = %d, timeline = %+v; want the test actor", g.Current(), view.Timeline)
	}
}

func TestWaitLogsThePhaseAndReturnsToTheActor(t *testing.T) {
	g, err := game.New(config(t, 1))
	if err != nil {
		t.Fatal(err)
	}
	actor := g.Current()
	events, err := g.Apply(api.Wait{Source: actor, Until: api.WaitTicks, Ticks: 60})
	if err != nil {
		t.Fatal(err)
	}
	sawPhase := false
	for _, e := range events {
		if _, ok := e.(api.PhaseStarted); ok {
			sawPhase = true
		}
	}
	if !sawPhase {
		t.Fatalf("events = %v, want a PhaseStarted", events)
	}
	view, err := g.View()
	if err != nil {
		t.Fatal(err)
	}
	const want = "Full day on day 1."
	found := false
	for _, line := range view.Log {
		found = found || line.Text == want
	}
	if !found {
		t.Errorf("log = %+v, want a line %q", view.Log, want)
	}
	if g.Current() != actor || view.Tick != 60 {
		t.Errorf("current = %d at tick %d; want the actor at tick 60", g.Current(), view.Tick)
	}
}

// waits builds n deterministic waits of varied length for the actor of a fresh game.
func waits(t *testing.T, cfg game.Config, n int) []api.Command {
	t.Helper()
	g, err := game.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	actor := g.Current()
	cmds := make([]api.Command, 0, n)
	for i := range n {
		cmds = append(cmds, api.Wait{Source: actor, Until: api.WaitTicks, Ticks: int64(1 + (i*37+11)%70)})
	}
	return cmds
}

func play(t *testing.T, cfg game.Config, cmds []api.Command) *game.Game {
	t.Helper()
	g, err := game.New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	for _, cmd := range cmds {
		if _, err := g.Apply(cmd); err != nil {
			t.Fatal(err)
		}
	}
	return g
}

func TestReplayGivesTheSameHash(t *testing.T) {
	cfg := config(t, 7)
	original := play(t, cfg, waits(t, cfg, 50))
	if len(original.History()) != 50 {
		t.Fatalf("history has %d commands, want 50", len(original.History()))
	}
	replayed, err := game.Replay(cfg, original.History())
	if err != nil {
		t.Fatal(err)
	}
	again := play(t, config(t, 7), original.History())
	a, errA := original.Hash()
	b, errB := replayed.Hash()
	c, errC := again.Hash()
	if errA != nil || errB != nil || errC != nil || a != b || a != c {
		t.Errorf("hashes differ: original %x, replayed %x, played again on fresh content %x (%v, %v, %v)",
			a, b, c, errA, errB, errC)
	}
}

func TestUnknownCommandKindIsAnError(t *testing.T) {
	g, err := game.New(config(t, 1))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := g.Apply(unknownCommand{}); err == nil {
		t.Fatal("want an error for an unknown command kind")
	}
	if _, err := g.View(); err != nil {
		t.Errorf("an unknown kind should leave the game usable: %v", err)
	}
}

func TestRejectedCommandIsAnEventNotAnError(t *testing.T) {
	g, err := game.New(config(t, 1))
	if err != nil {
		t.Fatal(err)
	}
	events, err := g.Apply(api.Wait{Source: g.Current() + 99, Until: api.WaitPhase})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].EventKind() != api.RejectedKind {
		t.Errorf("events = %v, want one Rejected", events)
	}
}

func TestSwapContentRefreshesTheCalendarAndTimeline(t *testing.T) {
	g, err := game.New(config(t, 1))
	if err != nil {
		t.Fatal(err)
	}
	bundle := contenttest.SampleBundle(t)
	contenttest.NoStart(&bundle)
	first := bundle.Calendar.Phases[0]
	bundle.Calendar.Phases[0].ID = "phase.renamed"
	bundle.Calendar.Phases[0].Name = "text.phase.renamed"
	bundle.Text["text.phase.renamed"] = "dusk"
	bundle.Calendar.Phases[0].Ticks = first.Ticks + 7
	bundle.Timeline.WindowTicks++
	reg := contenttest.Registry(t, bundle)
	if err := g.SwapContent(reg); err != nil {
		t.Fatal(err)
	}
	view, err := g.View()
	if err != nil {
		t.Fatal(err)
	}
	if view.Clock == nil || view.Clock.Phase != "dusk" {
		t.Errorf("clock = %+v, want the swapped calendar's first phase (dusk)", view.Clock)
	}
	if view.Timeline == nil || view.Timeline.Window != bundle.Timeline.WindowTicks {
		t.Errorf("timeline = %+v, want window %d", view.Timeline, bundle.Timeline.WindowTicks)
	}
}
