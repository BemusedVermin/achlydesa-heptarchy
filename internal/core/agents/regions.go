package agents

import (
	"fmt"
	"slices"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/inventory"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// regionView is the view builder of the region section.
const regionView = "agents.region"

// SpawnRegions makes a region entity for every map, in map ID order, fills each town's stock from
// content and marks the bodies a town keeps as its generics. Call it once, after the maps are built and before anything reads a region.
func SpawnRegions(w *sim.World) error {
	regions := ecs.NewMap[comp.Region](w.ECS)
	towns := ecs.NewMap[comp.Town](w.ECS)
	for _, id := range w.Maps.IDs() {
		name := api.ContentID(w.Maps.Get(id).Name())
		def, ok := w.Content.Map(name)
		if !ok {
			return fmt.Errorf("agents: region of map %d: unknown map %q", id, name)
		}
		region := regions.NewEntity(&comp.Region{Map: id, Def: name})
		if def.Region == nil || def.Region.Town == nil {
			continue
		}
		towns.Add(region, &comp.Town{})
		for _, stack := range def.Region.Town.Stock {
			inventory.Add(w, region, comp.ItemStack{Item: stack.Item, Count: stack.Count})
		}
		markGenerics(w, id, def.Region.Town)
	}
	return nil
}

// RegionOf finds the region entity of a map.
func RegionOf(w *sim.World, id api.MapID) (ecs.Entity, bool) {
	query := ecs.NewFilter1[comp.Region](w.ECS).Query()
	for query.Next() {
		if query.Get().Map == id {
			region := query.Entity()
			query.Close()
			return region, true
		}
	}
	return ecs.Entity{}, false
}

// TownOf finds the town entity of a map, when the map's region is a town.
func TownOf(w *sim.World, id api.MapID) (ecs.Entity, bool) {
	region, ok := RegionOf(w, id)
	if !ok || !ecs.NewMap[comp.Town](w.ECS).Has(region) {
		return ecs.Entity{}, false
	}
	return region, true
}

// Neighbours lists the regions the region's gates lead to, each once, in map ID order.
func Neighbours(w *sim.World, region ecs.Entity) []ecs.Entity {
	regions := ecs.NewMap[comp.Region](w.ECS)
	if !w.Alive(region) || !regions.Has(region) {
		return nil
	}
	own := regions.Get(region).Map
	m := w.Maps.Get(own)
	if m == nil {
		return nil
	}
	var ids []api.MapID
	for _, gate := range m.Gates() {
		if gate.To != own && !slices.Contains(ids, gate.To) {
			ids = append(ids, gate.To)
		}
	}
	slices.Sort(ids)
	out := make([]ecs.Entity, 0, len(ids))
	for _, id := range ids {
		if next, ok := RegionOf(w, id); ok {
			out = append(out, next)
		}
	}
	return out
}

// stockAt is the town whose stock the named mark of the map feeds.
func stockAt(w *sim.World, id api.MapID, mark string) (ecs.Entity, bool) {
	town, ok := TownOf(w, id)
	if !ok {
		return ecs.Entity{}, false
	}
	def, _ := w.Content.Map(ecs.NewMap[comp.Region](w.ECS).Get(town).Def)
	if def.Region == nil || def.Region.Town == nil || !slices.Contains(def.Region.Town.StockMarks, mark) {
		return ecs.Entity{}, false
	}
	return town, true
}

// buildRegion fills the view's region section from the map the controlled party's leader is on.
func buildRegion(w *sim.World, v *api.View) error {
	at, ok := partyMap(w)
	if !ok {
		return nil
	}
	region, ok := RegionOf(w, at)
	if !ok {
		return nil
	}
	def, _ := w.Content.Map(ecs.NewMap[comp.Region](w.ECS).Get(region).Def)
	view := &api.RegionView{Map: at, Name: w.Content.Render(api.Text{ID: def.Name})}
	if ecs.NewMap[comp.Town](w.ECS).Has(region) {
		view.Town = true
		if inventories := ecs.NewMap[comp.Inventory](w.ECS); inventories.Has(region) {
			for _, stack := range inventories.Get(region).Stacks {
				item, _ := w.Content.Item(stack.Item)
				view.Stock = append(view.Stock, api.StockLine{
					Item: stack.Item, Name: w.Content.Render(api.Text{ID: item.Name}), Count: stack.Count,
				})
			}
		}
	}
	v.Region = view
	return nil
}

// partyMap is the map of the controlled party's leader, or of its first placed member when the
// leader has no position.
func partyMap(w *sim.World) (api.MapID, bool) {
	party, ok := anatomy.ControlledParty(w)
	if !ok {
		return 0, false
	}
	positions := ecs.NewMap[comp.Pos](w.ECS)
	members := anatomy.Members(w, party)
	leader := ecs.NewMap[comp.Party](w.ECS).Get(party).Leader
	if at, found := w.ByUID(leader); found && positions.Has(at) {
		return positions.Get(at).Map, true
	}
	for _, member := range members {
		if positions.Has(member) {
			return positions.Get(member).Map, true
		}
	}
	return 0, false
}

// regionRecord is the record of the map's region, or the empty record when the map has none.
func regionRecord(w *sim.World, id api.MapID) content.Record {
	region, ok := RegionOf(w, id)
	if !ok {
		return content.EmptyRecord()
	}
	record, _ := w.Content.Record(ecs.NewMap[comp.Region](w.ECS).Get(region).Def)
	return record
}
