package explore

import (
	"slices"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/actions"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

const (
	// followController is the controller of a member trailing the one ahead of it.
	followController = "follow"
	// followPathNodes bounds the search for a way to a member that has fallen behind.
	followPathNodes = 200
)

// follower trails the member ahead of it in the party order, in single file.
type follower struct{}

// Decide steps into the cell the member ahead just left. A member that has fallen behind paths
// toward the one ahead. It never steps into an occupied cell, and when it cannot move it does
// nothing, which makes the timeline have it wait for the next entry.
func (follower) Decide(w *sim.World, body ecs.Entity) error {
	ahead, ok := memberAhead(w, body)
	if !ok {
		return nil
	}
	positions := ecs.NewMap[comp.Pos](w.ECS)
	if !positions.Has(body) || !positions.Has(ahead) {
		return nil
	}
	me, them := *positions.Get(body), *positions.Get(ahead)
	if me.Map != them.Map {
		return nil
	}
	m := w.Maps.Get(me.Map)
	if m == nil {
		return nil
	}
	rules := stateOf(w).rules
	if trails := ecs.NewMap[comp.Trail](w.ECS); trails.Has(ahead) {
		last := trails.Get(ahead).Last
		if dir, can, err := dirTo(w, me, last); err != nil {
			return err
		} else if can && last != them.At {
			return stepTo(w, body, me, dir)
		}
	}
	// Close enough already: the cell left behind is out of reach, but so is any reason to move.
	if _, adjacent, err := dirTo(w, me, them.At); err != nil || adjacent {
		return err
	}
	route, found, err := m.Path(me.At, them.At, rules, nil, followPathNodes)
	if err != nil || !found || len(route) < 2 {
		return err
	}
	dir, can, err := dirTo(w, me, route[1])
	if err != nil || !can {
		return err
	}
	return stepTo(w, body, me, dir)
}

// memberAhead is the member just before body in the party order.
func memberAhead(w *sim.World, body ecs.Entity) (ecs.Entity, bool) {
	party, ok := anatomy.PartyOf(w, body)
	if !ok {
		return ecs.Entity{}, false
	}
	members := membersOf(w, party)
	i := slices.Index(members, body)
	if i < 1 {
		return ecs.Entity{}, false
	}
	return members[i-1], true
}

// dirTo finds the direction of the one step from a body's position that ends in the cell, on the
// same map.
func dirTo(w *sim.World, from comp.Pos, cell api.Cell) (api.Dir, bool, error) {
	m := w.Maps.Get(from.Map)
	if m == nil {
		return 0, false, nil
	}
	for dir := api.Dir(0); dir < api.DirCount; dir++ {
		step, can, err := m.Step(from.At, dir, stateOf(w).rules)
		if err != nil {
			return 0, false, err
		}
		if can && (step.To.Map == 0 || step.To.Map == from.Map) && step.To.Cell == cell {
			return dir, true, nil
		}
	}
	return 0, false, nil
}

// stepTo steps the body one step in a direction, unless somebody stands there or the step is not
// allowed, in which case it does nothing.
func stepTo(w *sim.World, body ecs.Entity, from comp.Pos, dir api.Dir) error {
	m := w.Maps.Get(from.Map)
	step, can, err := m.Step(from.At, dir, stateOf(w).rules)
	if err != nil || !can {
		return err
	}
	if _, occupied := bodyAt(w, from.Map, step.To.Cell, body); occupied {
		return nil
	}
	in := actions.Intent{Actor: body, Action: w.Content.Party().FollowAction, Dir: &dir}
	if ok, _ := actions.Validate(w, in); !ok {
		return nil
	}
	return actions.Start(w, in)
}
