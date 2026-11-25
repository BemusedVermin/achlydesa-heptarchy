package agents

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// genericsSystem runs after every timeline entry so a replacement arrives at its tick.
const genericsSystem = "agents.generics"

// tickGenerics fires the replacements that are due and queues one for each generic a town is
// missing. A town with no residents replaces nobody and forgets its queue.
func tickGenerics(w *sim.World) error {
	type townEntry struct {
		entity ecs.Entity
		region comp.Region
	}
	var towns []townEntry
	query := ecs.NewFilter2[comp.Region, comp.Town](w.ECS).Query()
	for query.Next() {
		region, _ := query.Get()
		towns = append(towns, townEntry{entity: query.Entity(), region: *region})
	}
	slices.SortFunc(towns, func(a, b townEntry) int { return cmp.Compare(a.region.Map, b.region.Map) })
	for _, town := range towns {
		def, ok := w.Content.Map(town.region.Def)
		if !ok || def.Region == nil || def.Region.Town == nil || len(def.Region.Town.Generics) == 0 {
			continue
		}
		if err := tickTownGenerics(w, town.entity, town.region.Map, def.Region.Town); err != nil {
			return fmt.Errorf("agents: generics of map %d: %w", town.region.Map, err)
		}
	}
	return nil
}

func tickTownGenerics(w *sim.World, entity ecs.Entity, id api.MapID, def *content.TownDef) error {
	state := ecs.NewMap[comp.Town](w.ECS).Get(entity)
	if residentsOf(w, id, def) == 0 {
		state.Replacements = nil
		return nil
	}
	var waiting []comp.Replacement
	due := slices.Clone(state.Replacements)
	state.Replacements = nil
	for _, replacement := range due {
		if replacement.At > w.Now {
			waiting = append(waiting, replacement)
			continue
		}
		arrived, err := arrive(w, id, replacement.Template)
		if err != nil {
			return err
		}
		if !arrived {
			waiting = append(waiting, replacement)
		}
	}
	state.Replacements = waiting
	for _, generic := range def.Generics {
		have := genericsOn(w, id, generic.Template)
		for _, replacement := range state.Replacements {
			if replacement.Template == generic.Template {
				have++
			}
		}
		for range generic.Count - have {
			state.Replacements = append(state.Replacements,
				comp.Replacement{Template: generic.Template, At: w.Now + generic.Ticks})
		}
	}
	return nil
}

// residentsOf counts the living bodies of the town's residents template on the map.
func residentsOf(w *sim.World, id api.MapID, def *content.TownDef) int {
	total := 0
	query := ecs.NewFilter1[comp.Pos](w.ECS).Without(ecs.C[comp.Dead]()).Query()
	for query.Next() {
		if query.Get().Map == id && w.TemplateOf(query.Entity()) == def.Residents {
			total++
		}
	}
	return total
}

// genericsOn counts the living generics of the template that belong to the town and stand on its map.
func genericsOn(w *sim.World, id api.MapID, template api.ContentID) int {
	n := 0
	query := ecs.NewFilter2[comp.Generic, comp.Pos](w.ECS).Without(ecs.C[comp.Dead]()).Query()
	for query.Next() {
		generic, pos := query.Get()
		if generic.Town == id && pos.Map == id && w.TemplateOf(query.Entity()) == template {
			n++
		}
	}
	return n
}

// arrive sets a generic down beside the map's first gate. It reports false, leaving the town
// unchanged, when no cell there is free, so the caller can retry later.
func arrive(w *sim.World, id api.MapID, template api.ContentID) (bool, error) {
	m := w.Maps.Get(id)
	if m == nil {
		return false, fmt.Errorf("unknown map %d", id)
	}
	gates := m.Gates()
	if len(gates) == 0 {
		return false, fmt.Errorf("map %d has no gate for arrivals", id)
	}
	cell, free, err := freeCellNear(w, id, gates[0].At)
	if err != nil || !free {
		return false, err
	}
	body, err := w.Spawn(template, &comp.Pos{Map: id, At: cell})
	if err != nil {
		return false, err
	}
	sim.Put(w, body, comp.Generic{Town: id})
	return true, nil
}

// markGenerics makes each living body on a town's map a generic when the town keeps its template.
func markGenerics(w *sim.World, id api.MapID, def *content.TownDef) {
	var bodies []ecs.Entity
	query := ecs.NewFilter1[comp.Pos](w.ECS).Without(ecs.C[comp.Dead]()).Query()
	for query.Next() {
		if query.Get().Map != id {
			continue
		}
		template := w.TemplateOf(query.Entity())
		if slices.ContainsFunc(def.Generics, func(g content.GenericDef) bool { return g.Template == template }) {
			bodies = append(bodies, query.Entity())
		}
	}
	for _, body := range bodies {
		sim.Put(w, body, comp.Generic{Town: id})
	}
}
