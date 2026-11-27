package game_test

import (
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/clock"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game/gametest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// downCommand stands in for a landing that downs the whole party, parts broken and moved away.
type downCommand struct{}

func (downCommand) CommandKind() string { return "test.down" }

func TestADownedPartyWakesAtItsRestPlaceAfterTheWakeTime(t *testing.T) {
	reg := contenttest.Sample(t)
	g, world, _ := gametest.Start(t, game.Config{
		Seed:    1,
		Content: reg,
		Install: func(w *sim.World, _ *clock.Clock) error {
			return sim.Handle(w, func(w *sim.World, _ downCommand) (int64, error) {
				for _, member := range anatomy.Members(w, controlledParty(w)) {
					for _, part := range anatomy.Parts(w, member) {
						ecs.NewMap[comp.Part](w.ECS).Get(part).Integrity = 0
					}
					ecs.NewMap[comp.Actor](w.ECS).Remove(member)
					ecs.NewMap[comp.Downed](w.ECS).Add(member, &comp.Downed{})
					ecs.NewMap[comp.Pos](w.ECS).Get(member).At.Hex.Q += 2
				}
				return 0, nil
			})
		},
	})
	party := controlledParty(world)
	place := *ecs.NewMap[comp.RestPlace](world.ECS).Get(party)
	start := world.Now
	if _, err := g.Apply(downCommand{}); err != nil {
		t.Fatal(err)
	}
	if world.Now >= start+reg.Party().WakeTicks {
		t.Errorf("clock = %d: the party woke without being told to", world.Now)
	}
	events, err := g.Apply(wakeCommand(t, g))
	if err != nil {
		t.Fatal(err)
	}
	var woke bool
	for _, ev := range events {
		if _, ok := ev.(api.PartyWoke); ok {
			woke = true
		}
	}
	if !woke {
		t.Fatalf("events = %v, want a PartyWoke", events)
	}
	if want := start + reg.Party().WakeTicks; world.Now < want {
		t.Errorf("clock = %d, want at least %d", world.Now, want)
	}
	for _, member := range anatomy.Members(world, party) {
		if ecs.NewMap[comp.Downed](world.ECS).Has(member) || !ecs.NewMap[comp.Actor](world.ECS).Has(member) {
			t.Error("a member is still downed or off the timeline")
		}
		for _, part := range anatomy.Parts(world, member) {
			if p := ecs.NewMap[comp.Part](world.ECS).Get(part); p.Integrity != p.MaxIntegrity {
				t.Errorf("part %q is not mended", p.Slot)
			}
		}
		if at := ecs.NewMap[comp.Pos](world.ECS).Get(member).At; hexmap.CellDistance(at, place.At) > 8 {
			t.Errorf("member woke at %v, far from the rest place %v", at, place.At)
		}
	}
	if g.Current() == 0 {
		t.Fatal("nobody is deciding after the wake")
	}
	events, err = g.Apply(api.Wait{Source: g.Current(), Until: api.WaitTicks, Ticks: 1})
	if err != nil {
		t.Fatal(err)
	}
	for _, ev := range events {
		if rej, ok := ev.(api.Rejected); ok {
			t.Errorf("play did not continue: %v", rej.Reason)
		}
	}
}

// wakeCommand is the command of the only option a wiped party's view offers.
func wakeCommand(tb testing.TB, g *game.Game) api.Command {
	tb.Helper()
	view, err := g.View()
	if err != nil {
		tb.Fatal(err)
	}
	if view.Wipe == nil || !view.Wipe.CanWake || len(view.Options) != 1 || view.Options[0].ID != api.OptionWake {
		tb.Fatalf("wipe = %+v, options = %v, want the wake option alone", view.Wipe, view.Options)
	}
	return view.Options[0].Command
}

func TestAGameSavedWhileWipedLoadsOntoTheWipeScreenAndWakesTheSame(t *testing.T) {
	g, _, cfg := gametest.Start(t, game.Config{
		Seed:    1,
		Content: contenttest.Sample(t),
		Install: func(w *sim.World, _ *clock.Clock) error {
			return sim.Handle(w, func(w *sim.World, _ downCommand) (int64, error) {
				for _, member := range anatomy.Members(w, controlledParty(w)) {
					ecs.NewMap[comp.Actor](w.ECS).Remove(member)
					ecs.NewMap[comp.Downed](w.ECS).Add(member, &comp.Downed{})
				}
				return 0, nil
			})
		},
	})
	if _, err := g.Apply(downCommand{}); err != nil {
		t.Fatal(err)
	}
	loaded, _, _ := gametest.Reload(t, cfg, g)
	for _, played := range []*game.Game{g, loaded} {
		if _, err := played.Apply(wakeCommand(t, played)); err != nil {
			t.Fatal(err)
		}
	}
	want, err := g.Hash()
	if err != nil {
		t.Fatal(err)
	}
	got, err := loaded.Hash()
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Error("the loaded game woke differently from the original")
	}
}

func controlledParty(w *sim.World) ecs.Entity {
	query := ecs.NewFilter1[comp.Controlled](w.ECS).Query()
	defer query.Close()
	if query.Next() {
		return query.Entity()
	}
	return ecs.Entity{}
}
