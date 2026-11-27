package agents

import "github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"

// Plan is the steps a planner found for a goal ("goal." ID), in the order they run. GOAP and HTN
// both make one.
type Plan struct {
	Goal  api.ContentID `json:"goal"`
	Steps []Step        `json:"steps"`
}

// Step is one operator ("operator." ID) of a plan, and what its action is aimed at: nil for an action
// on the agent itself or on nothing.
type Step struct {
	Operator api.ContentID `json:"operator"`
	Target   *Target       `json:"target,omitempty"`
}

// Target is the thing a step's action is aimed at, with the cell the agent believes it is in, or a
// cell alone when Subject is 0.
type Target struct {
	Subject api.UID  `json:"subject,omitempty"`
	At      api.Cell `json:"at"`
}
