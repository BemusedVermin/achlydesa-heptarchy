package agents

import (
	"fmt"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// HTNPlanner is the name templates give as mind.planner to plan by recipe.
const HTNPlanner = "htn"

// htnPlanner turns a goal into the steps of the first recipe for it that applies and can be walked
// from the agent's beliefs. It searches nothing: its cost is the recipes for the goal, each at most as
// long as its expansion.
type htnPlanner struct{}

// Plan tries the recipes for the goal in ID order. A recipe is taken when its when formula holds and
// each operator of its expansion holds in the beliefs as the steps before it have changed them. A
// recipe whose formula fails, or that names a missing recipe, is an error naming it.
func (htnPlanner) Plan(snapshot Snapshot, goal api.ContentID) (Plan, bool, error) {
	env := beliefEnv(snapshot.Beliefs)
	for _, id := range snapshot.Content.RecipeIDs() {
		recipe, _ := snapshot.Content.Recipe(id)
		if recipe.Goal != goal {
			continue
		}
		applies, err := snapshot.Content.RecipeApplies(id, env)
		if err != nil {
			return Plan{}, false, err
		}
		if !applies {
			continue
		}
		operators, err := expandRecipe(snapshot.Content, id)
		if err != nil {
			return Plan{}, false, fmt.Errorf("agents: recipe %q: %w", id, err)
		}
		if plan, ok := walkOperators(snapshot, operators); ok {
			return plan, true, nil
		}
	}
	return Plan{}, false, nil
}

// expandRecipe lists the operators of the recipe in order, sub-recipes flattened in place. Content
// loading rejects cycles, so this ends.
func expandRecipe(reg *content.Registry, id api.ContentID) ([]api.ContentID, error) {
	recipe, ok := reg.Recipe(id)
	if !ok {
		return nil, fmt.Errorf("agents: unknown recipe %q", id)
	}
	var out []api.ContentID
	for _, step := range recipe.Steps {
		if step.Recipe == "" {
			out = append(out, step.Operator)
			continue
		}
		sub, err := expandRecipe(reg, step.Recipe)
		if err != nil {
			return nil, err
		}
		out = append(out, sub...)
	}
	return out, nil
}

// walkOperators makes a plan of the operators if each holds in turn, and each that moves the agent to
// a landmark has one to go to.
func walkOperators(snapshot Snapshot, operators []api.ContentID) (Plan, bool) {
	state := snapshot.Beliefs
	steps := make([]Step, 0, len(operators))
	for _, id := range operators {
		op, ok := snapshot.Content.Operator(id)
		if !ok || !Holds(op, state) {
			return Plan{}, false
		}
		step := Step{Operator: id}
		for _, eff := range op.Eff {
			if eff.Fact != content.FactNear || eff.Not {
				continue
			}
			target, ok := snapshot.Nearest[eff.Tag]
			if !ok {
				return Plan{}, false
			}
			step.Target = &target
			break
		}
		steps = append(steps, step)
		state = Apply(op, state)
	}
	return Plan{Steps: steps}, len(steps) > 0
}
