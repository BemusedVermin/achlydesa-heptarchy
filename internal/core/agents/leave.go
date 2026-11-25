package agents

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/effects"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

// leftReason is the text ID given for actions cut short when their body leaves through a gate.
const leftReason api.ContentID = "text.agents.left"

// stepLeave walks to the gate the behavior ranks highest, or else the nearest, and crosses it. It
// is done once crossed.
func stepLeave(w *sim.World, body ecs.Entity, _ content.BehaviorDef, _ Target) (bool, error) {
	from := *ecs.NewMap[comp.Pos](w.ECS).Get(body)
	m := w.Maps.Get(from.Map)
	if m == nil {
		return false, nil
	}
	gates := m.Gates()
	if len(gates) == 0 {
		return false, nil
	}
	gate, err := rankedGate(w, body, from, gates)
	if err != nil {
		return false, err
	}
	if from.At != gate.At {
		_, err := approach(w, body, gate.At, false)
		return false, err
	}
	return true, crossGate(w, body, gate)
}

// rankedGate is the gate the body's current behavior ranks highest, with the nearest winning a tie
// and then the lowest cell, so the choice never depends on the order of gates.
func rankedGate(w *sim.World, body ecs.Entity, from comp.Pos, gates []hexmap.Gate) (hexmap.Gate, error) {
	behavior := ecs.NewMap[comp.Doing](w.ECS).Get(body).Behavior
	source := effects.BodyEnvOf(w, body, ecs.Entity{})
	ranks := make(map[api.Cell]float64, len(gates))
	for _, gate := range gates {
		rank, err := w.Content.BehaviorRank(behavior, content.RankEnv{
			Source: source, Region: regionRecord(w, gate.To), Distance: hexmap.CellDistance(from.At, gate.At),
		})
		if err != nil {
			return hexmap.Gate{}, err
		}
		ranks[gate.At] = rank
	}
	return slices.MinFunc(gates, func(a, b hexmap.Gate) int {
		return cmp.Or(
			cmp.Compare(ranks[b.At], ranks[a.At]),
			cmp.Compare(hexmap.CellDistance(from.At, a.At), hexmap.CellDistance(from.At, b.At)),
			hexmap.CompareCells(a.At, b.At),
		)
	}), nil
}

// crossGate sets the body down at the gate's mark on the destination map, or the nearest free
// cell, and publishes [api.Migrated], followed by [api.TownEmptied] when it was the last resident
// of the party's town.
func crossGate(w *sim.World, body ecs.Entity, gate hexmap.Gate) error {
	from := ecs.NewMap[comp.Pos](w.ECS).Get(body).Map
	for _, pending := range timeline.PendingOf(w, w.UIDOf(body)) {
		timeline.Cancel(w, pending, leftReason)
	}
	cell, free, err := freeCellNear(w, gate.To, gate.ToCell)
	if err != nil {
		return err
	}
	if !free {
		return fmt.Errorf("no free cell near gate destination %v", gate.ToCell)
	}
	sim.Put(w, body, comp.Pos{Map: gate.To, At: cell})
	w.Bus.Publish(api.Migrated{Body: w.UIDOf(body), From: from, To: gate.To})
	publishIfEmptied(w, from, w.TemplateOf(body))
	return nil
}
