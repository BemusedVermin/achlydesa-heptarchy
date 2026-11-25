package agents

import (
	"fmt"
	"slices"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
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
	slices.SortFunc(towns, func(a, b townEntry) int { return int(a.region.Map) - int(b.region.Map) })
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
	if residentsOf(w, entity, id, def) == 0 {
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
		if err := arrive(w, entity, id, replacement.Template); err != nil {
			return err
		}
	}
	state.Replacements = waiting
	for _, generic := range def.Generics {
		family, err := familyListing(w, generic.Template)
		if err != nil {
			return err
		}
		have := genericsOn(w, id, generic.Template) + populationOf(w, entity, family)
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

// residentsOf counts the town's residents: its family's bodies alive on the map, plus its count.
func residentsOf(w *sim.World, town ecs.Entity, id api.MapID, def *content.TownDef) int {
	total := populationOf(w, town, def.Residents)
	query := ecs.NewFilter2[comp.Family, comp.Pos](w.ECS).Without(ecs.C[comp.Dead]()).Query()
	for query.Next() {
		family, pos := query.Get()
		if family.ID == def.Residents && pos.Map == id {
			total++
		}
	}
	return total
}

func populationOf(w *sim.World, region ecs.Entity, family api.ContentID) int {
	populations := ecs.NewMap[comp.Populations](w.ECS)
	if !populations.Has(region) {
		return 0
	}
	return countOf(populations.Get(region), family)
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

// arrive brings a generic into the town: a body at the gate when the party is on the map, else one
// more in the family's count.
func arrive(w *sim.World, town ecs.Entity, id api.MapID, template api.ContentID) error {
	family, err := familyListing(w, template)
	if err != nil {
		return err
	}
	if at, onScreen := partyMap(w); !onScreen || at != id {
		populations := ecs.NewMap[comp.Populations](w.ECS)
		if !populations.Has(town) {
			sim.Put(w, town, comp.Populations{})
		}
		addToPopulation(populations.Get(town), family, 1)
		return nil
	}
	cell, err := arrivalCell(w, id)
	if err != nil {
		return err
	}
	body, err := w.Spawn(template, &comp.Pos{Map: id, At: cell})
	if err != nil {
		return err
	}
	sim.Put(w, body, comp.Family{ID: family})
	sim.Put(w, body, comp.Generic{Town: id})
	return nil
}

// arrivalCell picks, with the lod stream, one of the free cells nearest the map's first gate.
func arrivalCell(w *sim.World, id api.MapID) (api.Cell, error) {
	m := w.Maps.Get(id)
	if m == nil {
		return api.Cell{}, fmt.Errorf("unknown map %d", id)
	}
	gates := m.Gates()
	if len(gates) == 0 {
		return api.Cell{}, fmt.Errorf("map %d has no gate for arrivals", id)
	}
	placer, err := newPlacer(w, id)
	if err != nil {
		return api.Cell{}, err
	}
	free, err := placer.free(gates[0].At)
	if err != nil {
		return api.Cell{}, err
	}
	if len(free) == 0 {
		return api.Cell{}, fmt.Errorf("no free cell near the gate at %v", gates[0].At)
	}
	nearest := hexmap.CellDistance(gates[0].At, free[0])
	ties := 0
	for ties < len(free) && hexmap.CellDistance(gates[0].At, free[ties]) == nearest {
		ties++
	}
	return free[w.RNG.Get(lodStream).IntN(ties)], nil
}

// markGeneric makes a promoted body a generic when the town of its map keeps its template.
func markGeneric(w *sim.World, body ecs.Entity, id api.MapID, template api.ContentID) {
	town, ok := TownOf(w, id)
	if !ok {
		return
	}
	def, _ := w.Content.Map(ecs.NewMap[comp.Region](w.ECS).Get(town).Def)
	if def.Region == nil || def.Region.Town == nil {
		return
	}
	if slices.ContainsFunc(
		def.Region.Town.Generics,
		func(g content.GenericDef) bool { return g.Template == template },
	) {
		sim.Put(w, body, comp.Generic{Town: id})
	}
}
