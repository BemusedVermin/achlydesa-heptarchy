package comp

import "github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"

// Mind is what a body the mind controller decides for is working on. Use the agents package rather
// than changing it.
type Mind struct {
	// Planner names the registered planner that makes the body's plans, from its template.
	Planner string `json:"planner"`
	// Thought is the tick the body last chose a goal, or 0 before its first choice.
	Thought int64 `json:"thought,omitempty"`
	// Goal is the "goal." ID the body works toward, or empty for none.
	Goal api.ContentID `json:"goal,omitempty"`
	// Plan has no steps while the body has no plan.
	Plan Plan `json:"plan"`
	// Step is the index in Plan.Steps of the next step to start.
	Step int `json:"step"`
	// Exports is the world's export count when the plan was made: an export since drops the plan.
	Exports int `json:"exports"`
}

// Plan is the steps a planner found for a goal ("goal." ID), in the order they run, and the claims
// they hold while they run. GOAP and HTN both make one.
type Plan struct {
	Goal   api.ContentID `json:"goal,omitempty"`
	Steps  []PlanStep    `json:"steps,omitempty"`
	Claims []Claim       `json:"claims,omitempty"`
}

// PlanStep is one operator ("operator." ID) of a plan, and what its action is aimed at: nil for an
// action on the agent itself or on nothing.
type PlanStep struct {
	Operator api.ContentID `json:"operator"`
	Target   *PlanTarget   `json:"target,omitempty"`
}

// PlanTarget is the thing a step's action is aimed at, with the cell the agent believes it is in, or
// a cell alone when Subject is 0.
type PlanTarget struct {
	Subject api.UID  `json:"subject,omitempty"`
	At      api.Cell `json:"at"`
}

// Claim reserves a thing for one agent's plan, so no other plan counts on it. With an item ("item."
// ID) it claims N of the thing's stock of that item, and other agents may claim the rest; without
// one it claims the whole thing, such as a bed.
type Claim struct {
	Target api.UID       `json:"target"`
	Item   api.ContentID `json:"item,omitempty"`
	N      int           `json:"n,omitempty"`
}

// Claims are the claims a body holds for its plan; they end when the plan does. Use the agents
// package rather than changing them.
type Claims struct {
	Held []Claim `json:"held"`
}
