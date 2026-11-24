package agents

import (
	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/actions"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// pathNodes bounds the search for a way to a behavior's target.
const pathNodes = 200

func primitives() []Primitive {
	return []Primitive{
		{Name: "idle", Step: stepIdle},
		{Name: "wander", Step: stepWander},
		{Name: "move_toward", Step: stepMoveToward},
	}
}

// stepIdle does nothing, so the body waits for the next entry on the timeline.
func stepIdle(*sim.World, ecs.Entity, content.BehaviorDef, Target) (bool, error) { return false, nil }

// stepWander takes the step action in a direction drawn from the agents stream, among those it can
// start.
func stepWander(w *sim.World, body ecs.Entity, _ content.BehaviorDef, _ Target) (bool, error) {
	steps := startableSteps(w, body)
	if len(steps) == 0 {
		return false, nil
	}
	return false, actions.Start(w, steps[w.RNG.Get(rngStream).IntN(len(steps))])
}

// stepMoveToward takes the step action along the path to the target. It is done on the target's
// cell, or next to a target body.
func stepMoveToward(w *sim.World, body ecs.Entity, _ content.BehaviorDef, target Target) (bool, error) {
	from := *ecs.NewMap[comp.Pos](w.ECS).Get(body)
	arrived := from.At == target.Cell
	if !target.Body.IsZero() {
		arrived = hexmap.CellDistance(from.At, target.Cell) <= 1
	}
	if arrived {
		return true, nil
	}
	m := w.Maps.Get(from.Map)
	if m == nil {
		return false, nil
	}
	rules := actions.StepRules(w)
	route, found, err := m.Path(from.At, target.Cell, rules, nil, pathNodes)
	if err != nil || !found || len(route) < 2 {
		return false, err
	}
	dir, can, err := m.DirTo(from.At, route[1], rules)
	if err != nil || !can {
		return false, err
	}
	for _, in := range startableSteps(w, body) {
		if *in.Dir == dir {
			return false, actions.Start(w, in)
		}
	}
	return false, nil
}

// startableSteps lists the step action's intents the body can start now, in option order. The step
// action is combat.step_action.
func startableSteps(w *sim.World, body ecs.Entity) []actions.Intent {
	stepID := w.Content.Combat().StepAction
	var steps []actions.Intent
	for _, in := range actions.Startable(w, body) {
		if in.Action == stepID && in.Dir != nil {
			steps = append(steps, in)
		}
	}
	return steps
}
