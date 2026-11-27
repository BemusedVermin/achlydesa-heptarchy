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

func TestWakeWithADeadLeaderHandsTheLeadOverAndReordersTheParty(t *testing.T) {
	reg := contenttest.Sample(t)
	g, world, _ := gametest.Start(t, game.Config{
		Seed:    1,
		Content: reg,
		Install: func(w *sim.World, _ *clock.Clock) error {
			return sim.Handle(w, func(w *sim.World, _ downCommand) (int64, error) {
				party := controlledParty(w)
				leader := ecs.NewMap[comp.Party](w.ECS).Get(party).Leader
				for _, member := range anatomy.Members(w, party) {
					ecs.NewMap[comp.Actor](w.ECS).Remove(member)
					if w.UIDOf(member) == leader {
						ecs.NewMap[comp.Dead](w.ECS).Add(member, &comp.Dead{})
						continue
					}
					ecs.NewMap[comp.Downed](w.ECS).Add(member, &comp.Downed{})
				}
				return 0, nil
			})
		},
	})
	party := controlledParty(world)
	state := ecs.NewMap[comp.Party](world.ECS).Get(party)
	if len(state.Order) < 2 {
		t.Fatalf("party order = %v, want at least two members", state.Order)
	}
	deadLeader := state.Leader
	if _, err := g.Apply(downCommand{}); err != nil {
		t.Fatal(err)
	}
	events, err := g.Apply(wakeCommand(t, g))
	if err != nil {
		t.Fatal(err)
	}
	state = ecs.NewMap[comp.Party](world.ECS).Get(party)
	if state.Leader == deadLeader || state.Leader == 0 {
		t.Fatalf("leader = %d, want a member who is not the dead one (%d)", state.Leader, deadLeader)
	}
	if state.Order[0] != state.Leader {
		t.Errorf("order = %v, want it to start with the leader %d", state.Order, state.Leader)
	}
	var changed *api.LeaderChanged
	for _, ev := range events {
		if lc, ok := ev.(api.LeaderChanged); ok {
			changed = &lc
		}
	}
	if changed == nil || changed.Leader != state.Leader {
		t.Errorf("LeaderChanged = %+v, want one naming leader %d", changed, state.Leader)
	}
}
