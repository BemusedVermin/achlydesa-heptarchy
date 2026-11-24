package agents

import (
	"errors"
	"fmt"
	"slices"
	"strconv"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/actions"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/inventory"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

const (
	// pathNodes bounds the search for a way to a behavior's target.
	pathNodes = 200
	// heldCellCost prices a step into a cell a standing body holds so high that a route around is taken
	// whenever the search finds one within pathNodes.
	heldCellCost = 1000
	// huntController is the controller engage hands a decision to unless the behavior names one.
	huntController = "hunt"
)

func primitives() []Primitive {
	return []Primitive{
		{Name: "idle", Step: stepIdle},
		{Name: "wander", Step: stepWander},
		{Name: "move_toward", Step: stepMoveToward},
		{Name: "consume", Params: []string{"action"}, Step: stepConsume},
		{Name: "carry", Params: []string{"item", "count", "to"}, Step: stepCarry},
		{Name: "flee", Step: stepFlee},
		{Name: "guard", Params: []string{"radius"}, Step: stepGuard},
		{Name: "engage", Optional: []string{"controller"}, Step: stepEngage},
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
	return approach(w, body, target.Cell, !target.Body.IsZero())
}

// approach takes the step action along the path to goal and reports whether the body is already
// there: on the cell, or next to it when adjacent.
func approach(w *sim.World, body ecs.Entity, goal api.Cell, adjacent bool) (bool, error) {
	from := *ecs.NewMap[comp.Pos](w.ECS).Get(body)
	if from.At == goal || (adjacent && hexmap.CellDistance(from.At, goal) <= 1) {
		return true, nil
	}
	m := w.Maps.Get(from.Map)
	if m == nil {
		return false, nil
	}
	rules := actions.StepRules(w)
	held := cellsHeldByOthers(w, body, from.Map)
	route, found, err := m.Path(from.At, goal, rules, avoidHeldCells(held, goal), pathNodes)
	if err == nil && (!found || routeCrosses(route, held, goal)) {
		route, found, err = m.Path(from.At, goal, rules, nil, pathNodes)
	}
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

// cellsHeldByOthers lists the cells on mapID where a body other than self stands.
func cellsHeldByOthers(w *sim.World, self ecs.Entity, mapID api.MapID) map[api.Cell]bool {
	held := map[api.Cell]bool{}
	query := ecs.NewFilter2[comp.Pos, comp.Actor](w.ECS).Query()
	for query.Next() {
		if pos, _ := query.Get(); pos.Map == mapID && query.Entity() != self {
			held[pos.At] = true
		}
	}
	return held
}

// avoidHeldCells prices a step into a held cell, except the goal, at heldCellCost. When no route avoids
// them, approach plans through them and waits for the blocker to move.
func avoidHeldCells(held map[api.Cell]bool, goal api.Cell) func(hexmap.Step) float64 {
	return func(step hexmap.Step) float64 {
		if held[step.To.Cell] && step.To.Cell != goal {
			return heldCellCost
		}
		return 1
	}
}

// routeCrosses reports whether route enters a held cell other than the goal.
func routeCrosses(route []api.Cell, held map[api.Cell]bool, goal api.Cell) bool {
	return slices.ContainsFunc(route[1:], func(cell api.Cell) bool { return held[cell] && cell != goal })
}

// stepConsume starts params.action on the target once it can, walking up to it until then. It is
// done when the action starts; what it does lands later.
func stepConsume(w *sim.World, body ecs.Entity, b content.BehaviorDef, target Target) (bool, error) {
	action := api.ContentID(b.Params["action"])
	if _, ok := w.Content.Action(action); !ok {
		return false, fmt.Errorf("unknown action %q", action)
	}
	other, ok := ownerAt(w, body, target)
	if !ok {
		return false, nil
	}
	in := actions.Intent{Actor: body, Action: action, Target: other}
	if valid, _ := actions.Validate(w, in); valid {
		return true, actions.Start(w, in)
	}
	_, err := approach(w, body, ecs.NewMap[comp.Pos](w.ECS).Get(other).At, true)
	return false, err
}

// ownerAt is the target body, or the body standing on the target's cell.
func ownerAt(w *sim.World, body ecs.Entity, target Target) (ecs.Entity, bool) {
	if !target.Body.IsZero() {
		return target.Body, true
	}
	return bodyAt(w, body, target.Cell)
}

// bodyAt is the body other than self standing on the cell, the lowest UID of them.
func bodyAt(w *sim.World, self ecs.Entity, cell api.Cell) (ecs.Entity, bool) {
	mapID := ecs.NewMap[comp.Pos](w.ECS).Get(self).Map
	var found []ecs.Entity
	query := ecs.NewFilter2[comp.Pos, comp.Identity](w.ECS).Query()
	for query.Next() {
		if pos, _ := query.Get(); pos.Map == mapID && pos.At == cell && query.Entity() != self {
			found = append(found, query.Entity())
		}
	}
	if len(found) == 0 {
		return ecs.Entity{}, false
	}
	return slices.MinFunc(found, w.CompareUID), true
}

// stepCarry moves params.count of params.item from the target body's inventory to whatever owns one
// at the params.to mark. It is done when delivered.
func stepCarry(w *sim.World, body ecs.Entity, b content.BehaviorDef, target Target) (bool, error) {
	item := api.ContentID(b.Params["item"])
	count, err := strconv.Atoi(b.Params["count"])
	if err != nil || count <= 0 {
		return false, fmt.Errorf("count %q is not a positive whole number", b.Params["count"])
	}
	if target.Body.IsZero() {
		return false, errors.New("carry needs a body target to take from")
	}
	mine := inventory.HolderOf(w, body)
	if inventory.Count(w, mine, item, 0) < count {
		arrived, err := approach(w, body, target.Cell, true)
		if err != nil || !arrived {
			return false, err
		}
		if !inventory.Take(w, inventory.HolderOf(w, target.Body), item, 0, count) {
			return false, nil
		}
		inventory.Add(w, mine, comp.ItemStack{Item: item, Count: count})
	}
	pos := *ecs.NewMap[comp.Pos](w.ECS).Get(body)
	m := w.Maps.Get(pos.Map)
	if m == nil {
		return false, nil
	}
	dest, ok := m.Mark(b.Params["to"])
	if !ok {
		return false, nil
	}
	owner, owned := bodyAt(w, body, dest)
	arrived, err := approach(w, body, dest, owned)
	if err != nil || !arrived || !owned {
		return false, err
	}
	if !inventory.Take(w, mine, item, 0, count) {
		return false, nil
	}
	inventory.Add(w, inventory.HolderOf(w, owner), comp.ItemStack{Item: item, Count: count})
	return true, nil
}

// hostilesInSight lists the standing hostiles the body sees, by UID.
func hostilesInSight(w *sim.World, body ecs.Entity) ([]ecs.Entity, error) {
	from := *ecs.NewMap[comp.Pos](w.ECS).Get(body)
	m := w.Maps.Get(from.Map)
	if m == nil {
		return nil, nil
	}
	cfg := w.Content.Combat()
	cells, err := m.Visible(from.At, cfg.SightRadius, cfg.SightDepth)
	if err != nil {
		return nil, err
	}
	seen := make(map[api.Cell]bool, len(cells))
	for _, cell := range cells {
		seen[cell] = true
	}
	var out []ecs.Entity
	query := ecs.NewFilter2[comp.Pos, comp.Allegiance](w.ECS).Query()
	for query.Next() {
		other := query.Entity()
		if pos, _ := query.Get(); pos.Map != from.Map || !seen[pos.At] || other == body ||
			anatomy.OutOfFight(w, other) || !anatomy.Hostile(w, body, other) {
			continue
		}
		out = append(out, other)
	}
	return w.SortByUID(out), nil
}

// stepFlee steps to the neighbour farthest from the nearest hostile in sight, when that is farther
// than where the body stands. It is done when no hostile is in sight.
func stepFlee(w *sim.World, body ecs.Entity, _ content.BehaviorDef, _ Target) (bool, error) {
	threats, err := hostilesInSight(w, body)
	if err != nil {
		return false, err
	}
	if len(threats) == 0 {
		return true, nil
	}
	positions := ecs.NewMap[comp.Pos](w.ECS)
	from := *positions.Get(body)
	var threat api.Cell
	nearest := -1
	for _, other := range threats {
		if d := hexmap.CellDistance(from.At, positions.Get(other).At); nearest < 0 || d < nearest {
			threat, nearest = positions.Get(other).At, d
		}
	}
	m := w.Maps.Get(from.Map)
	rules := actions.StepRules(w)
	var best actions.Intent
	bestDistance := nearest
	for _, in := range startableSteps(w, body) {
		step, can, err := m.Step(from.At, *in.Dir, rules)
		if err != nil {
			return false, err
		}
		if !can || (step.To.Map != 0 && step.To.Map != from.Map) {
			continue
		}
		if d := hexmap.CellDistance(step.To.Cell, threat); d > bestDistance {
			best, bestDistance = in, d
		}
	}
	if best.Action == "" {
		return false, nil
	}
	return false, actions.Start(w, best)
}

// stepGuard engages when a hostile in sight is within params.radius of the target, and otherwise
// stays within that radius of the target.
func stepGuard(w *sim.World, body ecs.Entity, b content.BehaviorDef, target Target) (bool, error) {
	radius, err := strconv.Atoi(b.Params["radius"])
	if err != nil || radius < 0 {
		return false, fmt.Errorf("radius %q is not a whole number", b.Params["radius"])
	}
	threats, err := hostilesInSight(w, body)
	if err != nil {
		return false, err
	}
	positions := ecs.NewMap[comp.Pos](w.ECS)
	for _, other := range threats {
		if hexmap.CellDistance(target.Cell, positions.Get(other).At) <= radius {
			return false, engageWith(w, body, huntController)
		}
	}
	if hexmap.CellDistance(positions.Get(body).At, target.Cell) <= radius {
		return false, nil
	}
	_, err = approach(w, body, target.Cell, !target.Body.IsZero())
	return false, err
}

// stepEngage hands the decision to the controller params.controller names, hunt by default.
func stepEngage(w *sim.World, body ecs.Entity, b content.BehaviorDef, _ Target) (bool, error) {
	name := huntController
	if given, ok := b.Params["controller"]; ok {
		name = given
	}
	return false, engageWith(w, body, name)
}

func engageWith(w *sim.World, body ecs.Entity, name string) error {
	if name == controllerName {
		return fmt.Errorf("engage cannot hand the decision back to the %q controller", name)
	}
	controller, ok := timeline.ControllerNamed(w, name)
	if !ok {
		return fmt.Errorf("unknown controller %q", name)
	}
	return controller.Decide(w, body)
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
