package agents_test

import (
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/agents"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
)

const rest = api.ContentID("drive.rest")

func TestChooseWeighsDrivesIntoAGoal(t *testing.T) {
	cases := []struct {
		name                    string
		waterLevel, restLevel   float64
		waterWeight, restWeight float64
		current, want           api.ContentID
	}{
		{"thirstier goes for water", 10, 2, 1, 1, "", "goal.drink"},
		{"more tired goes to sleep", 2, 6, 1, 1, "", "goal.sleep"},
		{"a larger weight on water wins an even level", 6, 6, 2, 1, "", "goal.drink"},
		{"a larger weight on rest wins an even level", 6, 6, 1, 2, "", "goal.sleep"},
		{"a slightly higher goal does not displace the current", 6, 4.2, 1, 1, "goal.drink", "goal.drink"},
		{"a slightly higher goal wins with no current", 6, 4.2, 1, 1, "", "goal.sleep"},
		{"a clearly higher goal displaces the current", 6, 8, 1, 1, "goal.drink", "goal.sleep"},
		{"nothing pressing leaves wandering", 0, 0, 1, 1, "goal.drink", "goal.wander"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fx := newMindFixture(t, 1)
			body := fx.entity(fx.thinkers[0])
			drives := ecs.NewMap[comp.Drives](fx.w.ECS).Get(body)
			drives.Levels[water], drives.Levels[rest] = tc.waterLevel, tc.restLevel
			drives.Weights[water], drives.Weights[rest] = tc.waterWeight, tc.restWeight
			ecs.NewMap[comp.Mind](fx.w.ECS).Get(body).Goal = tc.current
			for range 2 {
				got, err := agents.Choose(fx.w, body)
				if err != nil {
					t.Fatal(err)
				}
				if got != tc.want {
					t.Errorf("Choose = %q, want %q", got, tc.want)
				}
			}
		})
	}
}
