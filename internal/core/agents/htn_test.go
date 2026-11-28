package agents

import (
	"slices"
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

var (
	htnWater  = api.ContentID("affordance.water")
	htnWell   = Target{Subject: 7, At: api.Cell{Hex: api.Hex{Q: 3, R: 4}}}
	htnKnown  = Fact{Kind: content.FactKnown, Tag: htnWater}
	htnNear   = Fact{Kind: content.FactNear, Tag: htnWater}
	htnPlanOf = func(reg *content.Registry, goal api.ContentID, beliefs ...Fact) (Plan, bool) {
		snapshot := Snapshot{
			Beliefs: State(beliefs), Nearest: map[api.ContentID]Target{htnWater: htnWell}, Content: reg,
		}
		return htnPlanner{}.Plan(snapshot, goal)
	}
)

func htnOperators(plan Plan) []api.ContentID {
	var out []api.ContentID
	for _, step := range plan.Steps {
		out = append(out, step.Operator)
	}
	return out
}

func TestHTNPlansTheRecipeForTheGoal(t *testing.T) {
	reg := contenttest.Sample(t)
	cases := []struct {
		name    string
		goal    api.ContentID
		beliefs []Fact
		want    []api.ContentID
	}{
		{
			"walk, take, drink",
			"goal.drink",
			[]Fact{htnKnown},
			[]api.ContentID{"operator.go_to", "operator.take", "operator.consume"},
		},
		{"take, drink", "goal.drink", []Fact{htnKnown, htnNear}, []api.ContentID{"operator.take", "operator.consume"}},
		{
			"drink what is carried", "goal.drink", []Fact{{Kind: content.FactHas, Tag: "drinkable", N: 1}},
			[]api.ContentID{"operator.consume"},
		},
		{"look for water", "goal.drink", nil, []api.ContentID{"operator.wander"}},
		{"wander", "goal.wander", nil, []api.ContentID{"operator.roam"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			plan, found := htnPlanOf(reg, tc.goal, tc.beliefs...)
			if !found {
				t.Fatal("no plan")
			}
			if got := htnOperators(plan); !slices.Equal(got, tc.want) {
				t.Errorf("operators %v, want %v", got, tc.want)
			}
		})
	}

	plan, _ := htnPlanOf(reg, "goal.drink", htnKnown)
	if plan.Steps[0].Target == nil || *plan.Steps[0].Target != htnWell {
		t.Errorf("walk target %v, want the believed well %v", plan.Steps[0].Target, htnWell)
	}
}

func TestHTNFallsToTheNextRecipeWhenOneDoesNotHold(t *testing.T) {
	bundle := contenttest.SampleBundle(t)
	bundle.Recipes["recipe.a_first"] = content.Recipe{
		Goal: "goal.drink", Steps: []content.RecipeStep{{Operator: "operator.take"}, {Operator: "operator.consume"}},
	}
	reg := contenttest.Registry(t, bundle)
	plan, found := htnPlanOf(reg, "goal.drink", htnKnown)
	if !found {
		t.Fatal("no plan")
	}
	want := []api.ContentID{"operator.go_to", "operator.take", "operator.consume"}
	if got := htnOperators(plan); !slices.Equal(got, want) {
		t.Errorf("operators %v, want %v", got, want)
	}
}

func TestHTNFindsNoPlanWithoutARecipe(t *testing.T) {
	reg := contenttest.Sample(t)
	if plan, found := htnPlanOf(reg, "goal.chat", htnKnown); found {
		t.Errorf("plan %v for a goal with no recipe", htnOperators(plan))
	}
}
