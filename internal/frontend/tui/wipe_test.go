package tui_test

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/clock"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/frontend/tui"
)

type downCommand struct{}

func (downCommand) CommandKind() string { return "test.down" }

func TestWipeScreenOpensOnAWipedPartyAndClosesOnTheWakeOption(t *testing.T) {
	g, err := game.New(game.Config{
		Seed:    1,
		Content: contenttest.Sample(t),
		Install: func(w *sim.World, _ *clock.Clock) error {
			return sim.Handle(w, func(w *sim.World, _ downCommand) (int64, error) {
				query := ecs.NewFilter1[comp.Controlled](w.ECS).Query()
				var party ecs.Entity
				for query.Next() {
					party = query.Entity()
				}
				for _, member := range anatomy.Members(w, party) {
					ecs.NewMap[comp.Actor](w.ECS).Remove(member)
					ecs.NewMap[comp.Downed](w.ECS).Add(member, &comp.Downed{})
				}
				return 0, nil
			})
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := g.Apply(downCommand{}); err != nil {
		t.Fatal(err)
	}
	app, err := tui.NewApp(g)
	if err != nil {
		t.Fatal(err)
	}
	if app.State.Wipe == nil || !strings.Contains(app.View().Content, app.State.Wipe.Title) {
		t.Fatalf("the wipe screen is not showing:\n%s", app.View().Content)
	}
	pressCode(app, 'z')
	if app.State.Wipe == nil {
		t.Fatal("a key that is not the wake option woke the party")
	}
	pressCode(app, tea.KeyEnter)
	if app.State.Wipe != nil || strings.Contains(app.View().Content, "has fallen") {
		t.Errorf("the party did not wake:\n%s", app.View().Content)
	}
}
