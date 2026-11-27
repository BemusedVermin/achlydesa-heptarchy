package agents

import (
	"errors"
	"fmt"
	"slices"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/actions"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/inventory"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

const (
	// TakePicker is the target picker that aims a take at the nearest believed source or container
	// holding an item with the tag the operator's effect adds.
	TakePicker = "take"
	// PutPicker is the target picker that aims a put at the nearest believed container with room for
	// a carried item with the tag the operator's effect removes.
	PutPicker = "put"
	// ConsumePicker is the target picker that aims a consume at the first carried item with the tag
	// the operator's effect removes.
	ConsumePicker = "consume"
)

func installThings(w *sim.World) error {
	if err := RegisterPicker(w, TakePicker, thingPicker{taking: true}); err != nil {
		return err
	}
	if err := RegisterPicker(w, PutPicker, thingPicker{}); err != nil {
		return err
	}
	return RegisterPicker(w, ConsumePicker, consumePicker{})
}

type consumePicker struct{}

func (consumePicker) Pick(w *sim.World, body ecs.Entity, step Step) (actions.Intent, bool, error) {
	op, ok := w.Content.Operator(step.Operator)
	if !ok {
		return actions.Intent{}, false, fmt.Errorf("agents: unknown operator %q", step.Operator)
	}
	action, ok := w.Content.Action(op.Does.Action)
	if !ok || len(action.Choices) != 1 {
		return actions.Intent{}, false, fmt.Errorf("agents: action %q must have one choice, the item", op.Does.Action)
	}
	tag, ok := itemTag(op, false)
	if !ok {
		return actions.Intent{}, false, fmt.Errorf("agents: operator %q consumes no carried item", step.Operator)
	}
	items := stockedWith(w, inventory.HolderOf(w, body), tag)
	if len(items) == 0 {
		return actions.Intent{}, false, nil
	}
	return actions.Intent{Target: body, Choices: map[string]string{action.Choices[0]: string(items[0])}}, true, nil
}

// thingPicker aims a take or a put. A step that names its target uses it; otherwise the picker looks
// through the landmarks the agent remembers for the operator's near fact.
type thingPicker struct{ taking bool }

func (p thingPicker) Pick(w *sim.World, body ecs.Entity, step Step) (actions.Intent, bool, error) {
	op, ok := w.Content.Operator(step.Operator)
	if !ok {
		return actions.Intent{}, false, fmt.Errorf("agents: unknown operator %q", step.Operator)
	}
	action, ok := w.Content.Action(op.Does.Action)
	if !ok || len(action.Choices) != 1 {
		return actions.Intent{}, false, fmt.Errorf("agents: action %q must have one choice, the item", op.Does.Action)
	}
	tag, ok := itemTag(op, p.taking)
	if !ok {
		return actions.Intent{}, false, fmt.Errorf("agents: operator %q changes no carried item", step.Operator)
	}
	subjects, err := p.subjects(w, body, op, step)
	if err != nil {
		return actions.Intent{}, false, err
	}
	positions := ecs.NewMap[comp.Pos](w.ECS)
	for _, subject := range subjects {
		thing, ok := w.ByUID(subject)
		if !ok || !w.Alive(thing) || !positions.Has(thing) || !positions.Has(body) ||
			hexmap.CellDistance(positions.Get(body).At, positions.Get(thing).At) > 1 {
			continue
		}
		for _, item := range p.items(w, body, thing, tag) {
			if p.taking ||
				inventory.CheckTransfer(w, inventory.HolderOf(w, body), thing, body, item, 1) == inventory.None {
				return actions.Intent{
					Target:  thing,
					Choices: map[string]string{action.Choices[0]: string(item)},
				}, true, nil
			}
		}
	}
	return actions.Intent{}, false, nil
}

// subjects lists the things the step may be aimed at, nearest believed first.
func (thingPicker) subjects(w *sim.World, body ecs.Entity, op content.Operator, step Step) ([]api.UID, error) {
	if step.Target != nil && step.Target.Subject != 0 {
		return []api.UID{step.Target.Subject}, nil
	}
	for _, cond := range op.Pre {
		if cond.Fact != content.FactNear || cond.Not {
			continue
		}
		known, err := Known(w, body, cond.Tag)
		if err != nil {
			return nil, err
		}
		out := make([]api.UID, len(known))
		for i, landmark := range known {
			out[i] = landmark.Subject
		}
		return out, nil
	}
	return nil, nil
}

// items lists, in ID order, the items with the tag that a take could draw from the thing or a put
// could draw from the body's holder.
func (p thingPicker) items(w *sim.World, body, thing ecs.Entity, tag string) []api.ContentID {
	from := thing
	if !p.taking {
		from = inventory.HolderOf(w, body)
	}
	return stockedWith(w, from, tag)
}

// stockedWith lists the items with the tag that the entity holds, in its source and its inventory.
func stockedWith(w *sim.World, holder ecs.Entity, tag string) []api.ContentID {
	var items []api.ContentID
	add := func(item api.ContentID, count int) {
		def, ok := w.Content.Item(item)
		if count > 0 && ok && !def.Bound && slices.Contains(def.Tags, tag) && !slices.Contains(items, item) {
			items = append(items, item)
		}
	}
	if sources := ecs.NewMap[comp.Source](w.ECS); sources.Has(holder) {
		source := sources.Get(holder)
		add(source.Item, source.Stock)
	}
	if inventories := ecs.NewMap[comp.Inventory](w.ECS); inventories.Has(holder) {
		for _, stack := range inventories.Get(holder).Stacks {
			if stack.BoundTo == 0 {
				add(stack.Item, stack.Count)
			}
		}
	}
	slices.Sort(items)
	return items
}

// itemTag is the item tag an operator's effect adds (taking) or removes (putting).
func itemTag(op content.Operator, taking bool) (string, bool) {
	for _, eff := range op.Eff {
		if eff.Fact == content.FactHas && (eff.N > 0) == taking {
			return string(eff.Tag), true
		}
	}
	return "", false
}

// TakeClaim is the claim a take step makes on the stock it will empty: as many of the item as the
// operator's effect adds, from the thing the step names. A planner adds it to its plan's claims, so
// that two agents never count on the same last unit. It reports false when the step names no thing,
// or the thing holds no item with the tag.
func TakeClaim(w *sim.World, step Step) (comp.Claim, bool, error) {
	op, ok := w.Content.Operator(step.Operator)
	if !ok {
		return comp.Claim{}, false, fmt.Errorf("agents: unknown operator %q", step.Operator)
	}
	tag, ok := itemTag(op, true)
	if !ok {
		return comp.Claim{}, false, errors.New("agents: " + string(step.Operator) + " takes nothing")
	}
	if step.Target == nil || step.Target.Subject == 0 {
		return comp.Claim{}, false, nil
	}
	thing, ok := w.ByUID(step.Target.Subject)
	if !ok || !w.Alive(thing) {
		return comp.Claim{}, false, nil
	}
	items := stockedWith(w, thing, tag)
	if len(items) == 0 {
		return comp.Claim{}, false, nil
	}
	for _, eff := range op.Eff {
		if eff.Fact == content.FactHas && eff.Tag == api.ContentID(tag) {
			return comp.Claim{Target: step.Target.Subject, Item: items[0], N: eff.N}, true, nil
		}
	}
	return comp.Claim{}, false, nil
}
