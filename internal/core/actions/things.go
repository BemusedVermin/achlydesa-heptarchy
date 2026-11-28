package actions

import (
	"errors"
	"fmt"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/effects"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/inventory"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

const (
	// TakeEffect moves items from the target source or container into the actor's inventory.
	TakeEffect = "take"
	// PutEffect moves items from the actor's inventory into the target source or container.
	PutEffect = "put"
)

// thingMove is one take or put worked out from the effect's parameters.
type thingMove struct {
	item         api.ContentID
	count        int
	from, to, by ecs.Entity
}

func registerThings(w *sim.World) error {
	for _, p := range []effects.Primitive{
		thingPrimitive(TakeEffect, true),
		thingPrimitive(PutEffect, false),
	} {
		if err := effects.Register(w, p); err != nil {
			return err
		}
	}
	return nil
}

func thingPrimitive(name string, taking bool) effects.Primitive {
	return effects.Primitive{
		Name: name, Params: []string{"choice", "count"},
		Kinds: map[string]effects.Kind{"choice": effects.KindText},
		Apply: func(ctx *effects.Context, p effects.Params) error {
			move, err := readThingMove(ctx, p, taking)
			if err != nil {
				return err
			}
			if reason := inventory.Transfer(
				ctx.World,
				move.from,
				move.to,
				move.by,
				move.item,
				move.count,
			); reason != inventory.None {
				return fmt.Errorf("actions: %s: refused with reason %d after its check passed", name, reason)
			}
			return nil
		},
		Check: func(ctx *effects.Context, p effects.Params) (bool, api.Text) {
			move, reason, err := checkThingMove(ctx, p, taking)
			if err != nil {
				return false, api.Text{ID: why("broken").ID, Args: map[string]string{"error": err.Error()}}
			}
			if reason == inventory.None {
				return true, api.Text{}
			}
			return false, thingRefusal(ctx.World, reason, move.item, taking)
		},
		Impossible: func(ctx *effects.Context, p effects.Params) (bool, api.Text) {
			move, reason, err := checkThingMove(ctx, p, taking)
			if err != nil || !reason.Impossible() {
				return false, api.Text{}
			}
			return true, thingRefusal(ctx.World, reason, move.item, taking)
		},
	}
}

func checkThingMove(ctx *effects.Context, p effects.Params, taking bool) (thingMove, inventory.Refusal, error) {
	move, err := readThingMove(ctx, p, taking)
	if err != nil {
		return thingMove{}, inventory.None, err
	}
	return move, inventory.CheckTransfer(ctx.World, move.from, move.to, move.by, move.item, move.count), nil
}

// readThingMove reads the item the choice names and the count, and sets the ends of the move: the
// target and the actor's holder, in the order the effect moves them.
func readThingMove(ctx *effects.Context, p effects.Params, taking bool) (thingMove, error) {
	choice, err := p.String("choice")
	if err != nil {
		return thingMove{}, err
	}
	count, err := p.Int("count")
	if err != nil {
		return thingMove{}, err
	}
	item := api.ContentID(ctx.Choices[choice])
	if _, ok := ctx.World.Content.Item(item); !ok {
		return thingMove{}, errors.New("choice " + choice + " names no item: " + string(item))
	}
	holder := inventory.HolderOf(ctx.World, ctx.Source)
	move := thingMove{item: item, count: count, from: ctx.Target, to: holder, by: ctx.Source}
	if !taking {
		move.from, move.to = holder, ctx.Target
	}
	return move, nil
}

// thingRefusal is the reason text for a refused take or put.
func thingRefusal(w *sim.World, reason inventory.Refusal, item api.ContentID, taking bool) api.Text {
	args := map[string]string{"item": string(item)}
	if def, ok := w.Content.Item(item); ok {
		args["item"] = w.Content.Render(api.Text{ID: def.Name})
	}
	switch {
	case reason == inventory.OutOfReach:
		return why("out_of_range")
	case reason == inventory.Full || reason == inventory.Unsuited:
		return api.Text{ID: why("no_room").ID, Args: args}
	case taking:
		return api.Text{ID: why("none_there").ID, Args: args}
	default:
		return api.Text{ID: why("none_carried").ID, Args: args}
	}
}
