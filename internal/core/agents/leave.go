package agents

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

// leftReason is the text ID given for actions cut short when their body leaves through a gate.
const leftReason api.ContentID = "text.agents.left"

// stepLeave walks to the nearest gate of the body's map and crosses it. It is done once crossed.
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
	nearest := slices.MinFunc(gates, func(a, b hexmap.Gate) int {
		return cmp.Or(
			cmp.Compare(hexmap.CellDistance(from.At, a.At), hexmap.CellDistance(from.At, b.At)),
			hexmap.CompareCells(a.At, b.At),
		)
	})
	if from.At != nearest.At {
		_, err := approach(w, body, nearest.At, false)
		return false, err
	}
	return true, crossGate(w, body, nearest)
}

// crossGate moves the body through the gate. On the party's map it appears at the gate's mark, or
// the nearest free cell; anywhere else it is counted into the destination region's population of
// its family and removed, and [api.Migrated] says so.
func crossGate(w *sim.World, body ecs.Entity, gate hexmap.Gate) error {
	from := ecs.NewMap[comp.Pos](w.ECS).Get(body).Map
	for _, pending := range timeline.PendingOf(w, w.UIDOf(body)) {
		timeline.Cancel(w, pending, leftReason)
	}
	if at, onScreen := partyMap(w); onScreen && at == gate.To {
		placer, err := newPlacer(w, gate.To)
		if err != nil {
			return err
		}
		free, err := placer.free(gate.ToCell)
		if err != nil {
			return err
		}
		if len(free) == 0 {
			return fmt.Errorf("no free cell near gate destination %v", gate.ToCell)
		}
		sim.Put(w, body, comp.Pos{Map: gate.To, At: free[0]})
		return nil
	}
	family, err := familyOf(w, body)
	if err != nil {
		return err
	}
	region, ok := RegionOf(w, gate.To)
	if !ok {
		return fmt.Errorf("map %d has no region", gate.To)
	}
	populations := ecs.NewMap[comp.Populations](w.ECS)
	if !populations.Has(region) {
		sim.Put(w, region, comp.Populations{})
	}
	addToPopulation(populations.Get(region), family, 1)
	w.Despawn(body)
	w.Bus.Publish(api.Migrated{From: from, To: gate.To, Family: family, Count: 1})
	return nil
}

// familyOf is the family the body counts as: the one it was promoted as, or else the first family,
// by ID, that lists its template.
func familyOf(w *sim.World, body ecs.Entity) (api.ContentID, error) {
	if families := ecs.NewMap[comp.Family](w.ECS); families.Has(body) {
		return families.Get(body).ID, nil
	}
	family, err := familyListing(w, w.TemplateOf(body))
	if err != nil {
		return "", fmt.Errorf("body %d: %w", w.UIDOf(body), err)
	}
	return family, nil
}

// familyListing is the first family, by ID, that lists the template.
func familyListing(w *sim.World, template api.ContentID) (api.ContentID, error) {
	for _, id := range w.Content.FamilyIDs() {
		def, _ := w.Content.Family(id)
		if slices.ContainsFunc(def.Templates, func(t content.Weighted) bool { return t.Template == template }) {
			return id, nil
		}
	}
	return "", fmt.Errorf("no family lists template %q, so it cannot be counted", template)
}
