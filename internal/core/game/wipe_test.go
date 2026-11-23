package game_test

import (
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/clock"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// wipeCommand stands in for a landing that downs the whole party inside one Apply.
type wipeCommand struct{}

func (wipeCommand) CommandKind() string { return "test.wipe" }

func TestWipeInsideApplyEndsTheCallAndLaterCommandsAreRejected(t *testing.T) {
	reg, err := content.Load(sampleDir)
	if err != nil {
		t.Fatal(err)
	}
	g, err := game.New(game.Config{
		Seed:    1,
		Content: reg,
		Install: func(w *sim.World, _ *clock.Clock) error {
			return sim.Handle(w, func(w *sim.World, _ wipeCommand) (int64, error) {
				actors := ecs.NewMap[comp.Actor](w.ECS)
				dead := ecs.NewMap[comp.Dead](w.ECS)
				query := ecs.NewFilter1[comp.Controlled](w.ECS).Query()
				var party ecs.Entity
				for query.Next() {
					party = query.Entity()
				}
				for _, member := range anatomy.Members(w, party) {
					actors.Remove(member)
					dead.Add(member, &comp.Dead{Tick: w.Now})
				}
				return 0, nil
			})
		},
		Spawn: func(w *sim.World, _ *clock.Clock) error {
			_, err := w.Spawn("template.creature.moving_clay", nil)
			return err
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := g.Apply(wipeCommand{}); err != nil {
		t.Fatalf("the wiping Apply failed: %v", err)
	}
	events, err := g.Apply(api.Wait{Source: g.Current(), Until: api.WaitTicks, Ticks: 1})
	if err != nil {
		t.Fatal(err)
	}
	for _, ev := range events {
		if rej, ok := ev.(api.Rejected); ok && rej.Reason.ID == "text.game.lost" {
			return
		}
	}
	t.Errorf("events = %v, want a Rejected with text.game.lost", events)
}
