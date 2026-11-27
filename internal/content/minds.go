package content

import (
	"fmt"

	"github.com/expr-lang/expr"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// FactKind is the form of a symbolic fact a planner reasons over.
type FactKind string

// Kinds of fact.
const (
	// FactHas counts an item the agent carries: its tag is an "item." ID.
	FactHas FactKind = "has"
	// FactNear is a landmark the agent believes is within agents.near_radius: its tag is an
	// "affordance." ID.
	FactNear FactKind = "near"
	// FactKnown is a landmark the agent remembers and knows where to find: its tag is an
	// "affordance." ID.
	FactKnown FactKind = "known"
	// FactDriveMet is a drive the agent believes is met, by agents.met: its tag is a "drive." ID.
	FactDriveMet FactKind = "drive_met"
)

// factKinds lists the kinds a condition may name, for bundles built in Go, which skip the schema.
func factKinds() []FactKind { return []FactKind{FactHas, FactNear, FactKnown, FactDriveMet} }

// Condition is a fact an operator or goal names. As a precondition or a wanted state it holds when
// the fact does: for has, when at least N are carried. Not turns it around. As an effect it makes the
// fact true, or false with Not; for has, it adds N, which may be negative, instead.
type Condition struct {
	Fact FactKind      `json:"fact"`
	Tag  api.ContentID `json:"tag"`
	// N is the count of a has fact, and 0 for any other.
	N   int  `json:"n,omitempty"`
	Not bool `json:"not,omitempty"`
}

// Operator is one step a planner can take: what must hold first, what it changes in the agent's
// beliefs, what it costs, and the action that carries it out. GOAP and HTN share them.
type Operator struct {
	Pre []Condition `json:"pre,omitempty"`
	Eff []Condition `json:"eff"`
	// Cost is above 0, so a search prefers fewer and cheaper steps.
	Cost float64 `json:"cost"`
	Does Does    `json:"does"`
	// Kind is a free word that sorts operators, such as "travel" or "work".
	Kind string `json:"kind"`
}

// Does is the action an operator's step starts, and the target picker that aims it.
type Does struct {
	// Action is an "action." ID.
	Action api.ContentID `json:"action"`
	// Target names the picker that chooses what the action is aimed at.
	Target string `json:"target"`
}

// Goal is a state an agent works toward to meet a drive.
type Goal struct {
	// Drive is the "drive." ID the goal serves.
	Drive api.ContentID `json:"drive"`
	// Want holds once the goal is reached.
	Want []Condition `json:"want"`
}

// Recipe is a written plan for a goal: its steps in order, each an operator or another recipe.
type Recipe struct {
	// Goal is the "goal." ID the recipe reaches.
	Goal  api.ContentID `json:"goal"`
	Steps []RecipeStep  `json:"steps"`
}

// RecipeStep names exactly one of an operator and a recipe.
type RecipeStep struct {
	Operator api.ContentID `json:"operator,omitempty"`
	Recipe   api.ContentID `json:"recipe,omitempty"`
}

// Operator returns the operator with the given ID. Its slices are shared with the registry: don't
// change them.
func (r *Registry) Operator(id api.ContentID) (Operator, bool) {
	op, ok := r.bundle.Operators[string(id)]
	return op, ok
}

// OperatorIDs lists every operator's ID in order.
func (r *Registry) OperatorIDs() []api.ContentID { return sortedIDs(r.bundle.Operators) }

// Goal returns the goal with the given ID. Its Want slice is shared with the registry: don't change it.
func (r *Registry) Goal(id api.ContentID) (Goal, bool) {
	goal, ok := r.bundle.Goals[string(id)]
	return goal, ok
}

// Recipe returns the recipe with the given ID. Its Steps slice is shared with the registry: don't
// change it.
func (r *Registry) Recipe(id api.ContentID) (Recipe, bool) {
	recipe, ok := r.bundle.Recipes[string(id)]
	return recipe, ok
}

// DriveMet says by agents.met whether a drive at env's level counts as met. It is false when
// agents.met is empty, which content allows only while nothing names drive_met.
func (r *Registry) DriveMet(env DriveEnv) (bool, error) {
	if r.met == nil {
		return false, nil
	}
	return run[bool](r.met, env, "agents.met")
}

// checkMinds checks operators, goals and recipes, and compiles agents.met.
func (r *Registry) checkMinds() []error {
	p := r.problems()
	namesMet := false
	conditions := func(path string, conds []Condition, effect bool) {
		for i, cond := range conds {
			r.checkCondition(p, fmt.Sprintf("%s[%d]", path, i), cond, effect)
			namesMet = namesMet || cond.Fact == FactDriveMet
		}
	}
	for _, key := range sortedKeys(r.bundle.Operators) {
		op := r.bundle.Operators[key]
		path := fmt.Sprintf("operators[%q]", key)
		conditions(path+".pre", op.Pre, false)
		if len(op.Eff) == 0 {
			p.addf("%s.eff: at least one effect is required", path)
		}
		conditions(path+".eff", op.Eff, true)
		if op.Cost <= 0 {
			p.addf("%s.cost: must be above 0, got %v", path, op.Cost)
		}
		ref(p, r.bundle.Actions, path+".does.action", "action", op.Does.Action)
		if op.Does.Target == "" {
			p.addf("%s.does.target: required", path)
		}
		if op.Kind == "" {
			p.addf("%s.kind: required", path)
		}
	}
	for _, key := range sortedKeys(r.bundle.Goals) {
		goal := r.bundle.Goals[key]
		path := fmt.Sprintf("goals[%q]", key)
		ref(p, r.bundle.Drives, path+".drive", "drive", goal.Drive)
		if len(goal.Want) == 0 {
			p.addf("%s.want: at least one condition is required", path)
		}
		conditions(path+".want", goal.Want, false)
	}
	for _, key := range sortedKeys(r.bundle.Recipes) {
		recipe := r.bundle.Recipes[key]
		path := fmt.Sprintf("recipes[%q]", key)
		ref(p, r.bundle.Goals, path+".goal", "goal", recipe.Goal)
		if len(recipe.Steps) == 0 {
			p.addf("%s.steps: at least one step is required", path)
		}
		for i, step := range recipe.Steps {
			stepPath := fmt.Sprintf("%s.steps[%d]", path, i)
			switch {
			case (step.Operator == "") == (step.Recipe == ""):
				p.addf("%s: name exactly one of operator and recipe", stepPath)
			case step.Operator != "":
				ref(p, r.bundle.Operators, stepPath+".operator", "operator", step.Operator)
			default:
				ref(p, r.bundle.Recipes, stepPath+".recipe", "recipe", step.Recipe)
			}
		}
	}
	switch met := r.bundle.Agents.Met; {
	case met != "":
		program, err := r.Compile(met, DriveEnv{}, expr.AsBool())
		if err != nil {
			p.addf("agents.met: %w", err)
		}
		r.met = program
	case namesMet:
		p.addf("agents.met: required when an operator or goal names drive_met")
	}
	if r.bundle.Agents.NearRadius < 0 {
		p.addf("agents.near_radius: must not be negative, got %d", r.bundle.Agents.NearRadius)
	}
	return p.errs
}

// checkCondition checks that a condition names a known fact and a tag of the collection that fact
// is over, with a count only where it means something.
func (r *Registry) checkCondition(p *problems, path string, cond Condition, effect bool) {
	switch cond.Fact {
	case FactHas:
		ref(p, r.bundle.Items, path+".tag", "item", cond.Tag)
		switch {
		case effect && cond.Not:
			p.addf("%s.not: a has effect changes the count by n instead", path)
		case effect && cond.N == 0:
			p.addf("%s.n: a has effect needs a count other than 0", path)
		case !effect && cond.N < 1:
			p.addf("%s.n: a has condition needs a count of at least 1, got %d", path, cond.N)
		}
		return
	case FactNear, FactKnown:
		ref(p, r.bundle.Affordances, path+".tag", "affordance", cond.Tag)
	case FactDriveMet:
		ref(p, r.bundle.Drives, path+".tag", "drive", cond.Tag)
	default:
		p.addf("%s.fact: unknown fact %q (want one of %v)", path, cond.Fact, factKinds())
		return
	}
	if cond.N != 0 {
		p.addf("%s.n: only a has fact takes a count", path)
	}
}
