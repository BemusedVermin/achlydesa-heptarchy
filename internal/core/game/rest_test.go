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
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

const (
	townMap = "map.test.town"
	hallMap = "map.test.hall"
)

// startRest starts a game on the named map with a command that downs the whole party.
func startRest(t *testing.T, start string) (*game.Game, *sim.World) {
	t.Helper()
	bundle := contenttest.SampleBundle(t)
	bundle.Start.Map = api.ContentID(start)
	g, world, _ := gametest.Start(t, game.Config{
		Seed:    1,
		Content: contenttest.Registry(t, bundle),
		Install: func(w *sim.World, _ *clock.Clock) error {
			return sim.Handle(w, func(w *sim.World, _ downCommand) (int64, error) {
				for _, member := range anatomy.Members(w, controlledParty(w)) {
					for _, part := range anatomy.Parts(w, member) {
						ecs.NewMap[comp.Part](w.ECS).Get(part).Integrity = 0
					}
					ecs.NewMap[comp.Actor](w.ECS).Remove(member)
					ecs.NewMap[comp.Downed](w.ECS).Add(member, &comp.Downed{})
				}
				return 0, nil
			})
		},
	})
	return g, world
}

func restOption(t *testing.T, g *game.Game) (api.Option, bool) {
	t.Helper()
	view, err := g.View()
	if err != nil {
		t.Fatal(err)
	}
	for _, option := range view.Options {
		if option.ID == "rest" {
			return option, true
		}
	}
	return api.Option{}, false
}

func armsOf(w *sim.World, body ecs.Entity) []ecs.Entity {
	var arms []ecs.Entity
	for _, part := range anatomy.Parts(w, body) {
		if ecs.NewMap[comp.Part](w.ECS).Get(part).Slot == "arm" {
			arms = append(arms, part)
		}
	}
	return arms
}

func TestRestHealsAnArmAndLeavesASeveredOneGone(t *testing.T) {
	g, world := startRest(t, townMap)
	if _, err := g.Apply(api.Wait{Source: g.Current(), Until: api.WaitTicks, Ticks: 120}); err != nil {
		t.Fatal(err)
	}
	body, ok := world.ByUID(g.Current())
	if !ok {
		t.Fatal("nobody is deciding")
	}
	parts := ecs.NewMap[comp.Part](world.ECS)
	arms := armsOf(world, body)
	if len(arms) != 2 {
		t.Fatalf("body has %d arms, want 2", len(arms))
	}
	hurt, cut := arms[0], arms[1]
	parts.Get(hurt).Integrity = 1
	parts.Get(cut).Integrity = 2
	ecs.NewMap[comp.PartOf](world.ECS).Remove(cut)
	option, offered := restOption(t, g)
	if !offered || !option.Enabled {
		t.Fatalf("rest offered = %v (%+v), want an enabled option in town", offered, option)
	}
	if _, err := g.Apply(option.Command); err != nil {
		t.Fatal(err)
	}
	for range 40 {
		if world.Now >= 240 {
			break
		}
		if _, err := g.Apply(api.Wait{Source: g.Current(), Until: api.WaitEvent}); err != nil {
			t.Fatal(err)
		}
	}
	if world.Now < 240 {
		t.Fatalf("clock = %d, want the rest to run to the next dawn at 240", world.Now)
	}
	if got := parts.Get(hurt).Integrity; got != 2 {
		t.Errorf("hurt arm integrity = %d, want 1 plus one for the 120 ticks rested", got)
	}
	if got := parts.Get(cut).Integrity; got != 2 {
		t.Errorf("severed arm integrity = %d, want it untouched at 2", got)
	}
	if ecs.NewMap[comp.PartOf](world.ECS).Has(cut) {
		t.Error("the severed arm is attached again")
	}
}

func TestAWipeAfterRestingInATownWakesThere(t *testing.T) {
	g, world := startRest(t, townMap)
	body, ok := world.ByUID(g.Current())
	if !ok {
		t.Fatal("nobody is deciding")
	}
	party := controlledParty(world)
	town := ecs.NewMap[comp.Pos](world.ECS).Get(body).Map
	if rest := ecs.NewMap[comp.RestPlace](world.ECS).Get(party); rest.Map != town {
		t.Fatalf("rest place starts on map %d, want the start map %d", rest.Map, town)
	}
	option, offered := restOption(t, g)
	if !offered {
		t.Fatal("rest is not offered in town")
	}
	ecs.NewMap[comp.RestPlace](world.ECS).Get(party).Map = 0
	if _, err := g.Apply(option.Command); err != nil {
		t.Fatal(err)
	}
	for range 40 {
		if world.Now >= 240 {
			break
		}
		if _, err := g.Apply(api.Wait{Source: g.Current(), Until: api.WaitEvent}); err != nil {
			t.Fatal(err)
		}
	}
	if rest := ecs.NewMap[comp.RestPlace](world.ECS).Get(party); rest.Map != town {
		t.Fatalf("rest place is on map %d, want the town %d", rest.Map, town)
	}
	events, err := g.Apply(downCommand{})
	if err != nil {
		t.Fatal(err)
	}
	var woke *api.PartyWoke
	for _, ev := range events {
		if e, ok := ev.(api.PartyWoke); ok {
			woke = &e
		}
	}
	if woke == nil || woke.Map != town {
		t.Fatalf("woke = %+v, want a wake in the town (map %d)", woke, town)
	}
}

func TestRestIsNotOfferedOutsideATown(t *testing.T) {
	g, _ := startRest(t, hallMap)
	if option, offered := restOption(t, g); offered {
		t.Errorf("rest is offered in a hall: %+v", option)
	}
}

func TestRestIsNotOfferedWithAHostileInSight(t *testing.T) {
	g, world := startRest(t, townMap)
	ecs.NewMap[comp.Party](world.ECS).Get(controlledParty(world)).Alert = true
	if option, offered := restOption(t, g); offered {
		t.Errorf("rest is offered on alert: %+v", option)
	}
}
