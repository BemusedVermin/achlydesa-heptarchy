package explore

import (
	"strconv"
	"strings"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

var dirNames = [api.DirCount]string{"e", "ne", "nw", "w", "sw", "se"}

func bump() api.Text { return api.Text{ID: "text.explore.bump"} }

// refusalText is the reason no step leads from from in dir.
func refusalText(m *hexmap.Map, from api.Cell, dir api.Dir, rules hexmap.StepRules) (api.Text, error) {
	refusal, err := m.Refusal(from, dir, rules)
	if err != nil {
		return api.Text{}, err
	}
	switch refusal.Kind {
	case hexmap.RefusedTooHigh:
		return api.Text{ID: "text.explore.too_high", Args: map[string]string{
			"cells": strconv.Itoa(refusal.Height), "limit": strconv.Itoa(rules.MaxStepUp),
		}}, nil
	case hexmap.RefusedSolid:
		return bump(), nil
	case hexmap.RefusedTooFar:
		return api.Text{ID: "text.explore.too_far", Args: map[string]string{"limit": strconv.Itoa(rules.MaxDrop)}}, nil
	}
	return bump(), nil
}

// movesFrom lists what a step from the body's cell in each direction would do.
func movesFrom(w *sim.World, body ecs.Entity, from api.Cell, m *hexmap.Map) ([]api.MoveView, error) {
	rules := stateOf(w).rules
	moves := make([]api.MoveView, 0, api.DirCount)
	for dir := range api.Dir(api.DirCount) {
		args := map[string]string{
			"dir": w.Content.Render(api.Text{ID: api.ContentID("text.actions.dir." + dirNames[dir])}),
		}
		view := api.MoveView{Dir: dir, Kind: api.MoveBlocked}
		step, can, err := m.Step(from, dir, rules)
		if err != nil {
			return nil, err
		}
		var label api.Text
		if can {
			view.Below = step.To.Cell.Depth - from.Depth
			args["cells"] = strconv.Itoa(max(view.Below, -view.Below))
			if step.Kind == hexmap.StepDrop {
				args["cells"] = strconv.Itoa(step.Fall)
			}
			view.Kind = moveKinds[step.Kind]
			label = api.Text{ID: api.ContentID("text.explore.move." + view.Kind)}
			if _, occupied, swappable := occupant(w, body, step); occupied && !swappable {
				view.Kind = api.MoveBlocked
				label.ID = "text.explore.move.occupied"
			}
		} else {
			refusal, err := m.Refusal(from, dir, rules)
			if err != nil {
				return nil, err
			}
			label = api.Text{ID: "text.explore.move.solid"}
			switch refusal.Kind {
			case hexmap.RefusedTooHigh:
				label.ID = "text.explore.move.too_high"
				args["cells"] = strconv.Itoa(refusal.Height)
			case hexmap.RefusedSolid:
			case hexmap.RefusedTooFar:
				label.ID = "text.explore.move.too_far"
			}
		}
		label.Args = args
		view.Label = strings.TrimSpace(w.Content.Render(label))
		moves = append(moves, view)
	}
	return moves, nil
}
