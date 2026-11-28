package agents

import (
	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/actions"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

const (
	// GoToPicker is the target picker that steps toward the cell a step's target names.
	GoToPicker = "go_to"
	// WanderPicker is the target picker that steps in a random open direction.
	WanderPicker = "wander"
	wanderStream = "agents.wander"
	// walkNodes bounds the search for a route to a believed cell.
	walkNodes = 200
	// heldCellCost prices a step into a cell a standing body holds: a short detour is taken first, but a
	// crowd cannot make the search run out its nodes.
	heldCellCost = 4
)

func installMovement(w *sim.World) error {
	if err := RegisterPicker(w, GoToPicker, goToPicker{}); err != nil {
		return err
	}
	return RegisterPicker(w, WanderPicker, wanderPicker{})
}

// goToPicker aims a step action one cell along the route to the step's target cell. The step is
// finished once the agent is there, or within agents.near_radius of a landmark.
type goToPicker struct{}

func (goToPicker) Pick(w *sim.World, body ecs.Entity, step Step) (actions.Intent, bool, error) {
	positions := ecs.NewMap[comp.Pos](w.ECS)
	if step.Target == nil || w.Map == nil || !positions.Has(body) {
		return actions.Intent{}, false, nil
	}
	from := positions.Get(body)
	rules := actions.StepRules(w)
	held := cellsHeldByOthers(w, body)
	goal := step.Target.At
	route, found, err := w.Map.Path(from.At, goal, rules, avoidHeldCells(held, goal), walkNodes)
	if err == nil && !found {
		route, found, err = w.Map.Path(from.At, goal, rules, nil, walkNodes)
	}
	if err != nil || !found || len(route) < 2 {
		return actions.Intent{}, false, err
	}
	dir, can, err := w.Map.DirTo(from.At, route[1], rules)
	if err != nil || !can {
		return actions.Intent{}, false, err
	}
	return actions.Intent{Dir: &dir}, true, nil
}

func (goToPicker) Finished(w *sim.World, body ecs.Entity, step Step) bool {
	positions := ecs.NewMap[comp.Pos](w.ECS)
	if step.Target == nil || !positions.Has(body) {
		return false
	}
	from := positions.Get(body).At
	if from == step.Target.At {
		return true
	}
	return step.Target.Subject != 0 && hexmap.CellDistance(from, step.Target.At) <= w.Content.Agents().NearRadius
}

// wanderPicker aims a step action at a random open direction, never back into the cell the agent
// just left unless that is the only way.
type wanderPicker struct{}

func (wanderPicker) Pick(w *sim.World, body ecs.Entity, _ Step) (actions.Intent, bool, error) {
	positions := ecs.NewMap[comp.Pos](w.ECS)
	if w.Map == nil || !positions.Has(body) {
		return actions.Intent{}, false, nil
	}
	from := positions.Get(body)
	trails := ecs.NewMap[comp.Trail](w.ECS)
	if !trails.Has(body) {
		trails.Add(body, &comp.Trail{Last: from.At})
	}
	left := trails.Get(body).Last
	held := cellsHeldByOthers(w, body)
	rules := actions.StepRules(w)
	var open, forward []api.Dir
	for dir := range api.Dir(api.DirCount) {
		step, can, err := w.Map.Step(from.At, dir, rules)
		if err != nil {
			return actions.Intent{}, false, err
		}
		if !can || held[step.To.Cell] {
			continue
		}
		open = append(open, dir)
		if step.To.Cell != left {
			forward = append(forward, dir)
		}
	}
	if len(forward) > 0 {
		open = forward
	}
	if len(open) == 0 {
		return actions.Intent{}, false, nil
	}
	dir := open[w.RNG.Derived(wanderStream, planSalt(w.UIDOf(body), w.Now)).IntN(len(open))]
	return actions.Intent{Dir: &dir}, true, nil
}

// cellsHeldByOthers lists the cells where a body other than self stands.
func cellsHeldByOthers(w *sim.World, self ecs.Entity) map[api.Cell]bool {
	held := map[api.Cell]bool{}
	query := ecs.NewFilter2[comp.Pos, comp.Actor](w.ECS).Query()
	for query.Next() {
		if pos, _ := query.Get(); query.Entity() != self {
			held[pos.At] = true
		}
	}
	return held
}

func avoidHeldCells(held map[api.Cell]bool, goal api.Cell) func(hexmap.Step) float64 {
	return func(step hexmap.Step) float64 {
		if held[step.To.Cell] && step.To.Cell != goal {
			return heldCellCost
		}
		return 1
	}
}
