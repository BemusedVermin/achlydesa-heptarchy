package game_test

import (
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/clock"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

const sampleDir = "../../../content/data"

// unknownCommand has no handler.
type unknownCommand struct{}

func (unknownCommand) CommandKind() string { return "test.unknown" }

// config spawns one player-controlled actor, standing in for the start scenario of M1-10.
func config(t *testing.T, seed uint64) game.Config {
	t.Helper()
	reg, err := content.Load(sampleDir)
	if err != nil {
		t.Fatal(err)
	}
	return game.Config{Seed: seed, Content: reg, Extra: func(w *sim.World, _ *clock.Clock) error {
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
	events, err := g.Apply(api.Wait{Actor: actor, Until: api.WaitTicks, Ticks: 60})
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
	for i := 0; i < n; i++ {
		cmds = append(cmds, api.Wait{Actor: actor, Until: api.WaitTicks, Ticks: int64(1 + (i*37+11)%70)})
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
	a, errA := original.Hash()
	b, errB := replayed.Hash()
	if errA != nil || errB != nil || a != b {
		t.Errorf("hashes differ: %x vs %x (%v, %v)", a, b, errA, errB)
	}
}

func TestSameSeedAndCommandsGiveTheSameHash(t *testing.T) {
	cmds := waits(t, config(t, 3), 20)
	a, errA := play(t, config(t, 3), cmds).Hash()
	b, errB := play(t, config(t, 3), cmds).Hash()
	if errA != nil || errB != nil || a != b {
		t.Errorf("hashes differ: %x vs %x (%v, %v)", a, b, errA, errB)
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
	events, err := g.Apply(api.Wait{Actor: g.Current() + 99, Until: api.WaitPhase})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 || events[0].EventKind() != api.RejectedKind {
		t.Errorf("events = %v, want one Rejected", events)
	}
}
