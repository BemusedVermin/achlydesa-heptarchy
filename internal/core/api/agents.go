package api

import (
	"strconv"
	"strings"
)

// PlanMadeKind is the event kind of [PlanMade].
const PlanMadeKind = "agents.plan_made"

// PlanMade reports that an agent adopted a plan, for reading how it reasons. It is published only
// while the agents.trace_plans config is on.
type PlanMade struct {
	Agent UID         `json:"agent"`
	Goal  ContentID   `json:"goal"`
	Steps []ContentID `json:"steps"`
	Tick  int64       `json:"tick"`
}

// EventKind implements [Event].
func (PlanMade) EventKind() string { return PlanMadeKind }

// LogText implements [Loggable].
func (e PlanMade) LogText() Text {
	steps := make([]string, len(e.Steps))
	for i, step := range e.Steps {
		steps[i] = string(step)
	}
	return Text{ID: "text.agents.plan_made", Args: map[string]string{
		"agent": strconv.FormatUint(uint64(e.Agent), 10), "goal": string(e.Goal), "steps": strings.Join(steps, " > "),
	}}
}
