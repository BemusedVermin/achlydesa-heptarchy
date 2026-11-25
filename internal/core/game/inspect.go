package game

import (
	"cmp"
	"slices"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
)

// PositionOf returns where the entity with the UID stands, including its depth, which the view
// leaves out. It is a read-only accessor for the harness and tests, not for frontends.
func (g *Game) PositionOf(uid api.UID) (comp.Pos, bool) { return component[comp.Pos](g, uid) }

// PartOfUID returns a copy of the part with the UID, including its true material, which the view
// hides until the party learns it. It is a read-only accessor for the harness and tests, not for
// frontends.
func (g *Game) PartOfUID(uid api.UID) (comp.Part, bool) { return component[comp.Part](g, uid) }

// component returns a copy of the T of the entity with the UID.
func component[T any](g *Game, uid api.UID) (T, bool) {
	var zero T
	e, ok := g.world.ByUID(uid)
	if !ok {
		return zero, false
	}
	values := ecs.NewMap[T](g.world.ECS)
	if !values.Has(e) {
		return zero, false
	}
	return *values.Get(e), true
}

// FamilyCount is how many living members of a family a map holds, counting both the bodies on it
// and its region's off-screen population. Town is true when the map's region is a town.
type FamilyCount struct {
	Map    api.MapID
	Town   bool
	Family api.ContentID
	Count  int
}

// FamilyCensus counts the living members of every family on every map joined to the party's map by
// gates, sorted by map and then family. It is a read-only accessor for the harness and tests, not for frontends.
func (g *Game) FamilyCensus() []FamilyCount {
	type key struct {
		mapID  api.MapID
		family api.ContentID
	}
	counts := map[key]int{}
	towns := map[api.MapID]bool{}
	dead := ecs.NewMap[comp.Dead](g.world.ECS)
	bodies := ecs.NewFilter2[comp.Pos, comp.Family](g.world.ECS).Query()
	for bodies.Next() {
		if dead.Has(bodies.Entity()) {
			continue
		}
		pos, family := bodies.Get()
		counts[key{pos.Map, family.ID}]++
	}
	townMarks := ecs.NewMap[comp.Town](g.world.ECS)
	regions := ecs.NewFilter2[comp.Region, comp.Populations](g.world.ECS).Query()
	for regions.Next() {
		region, populations := regions.Get()
		towns[region.Map] = townMarks.Has(regions.Entity())
		for _, population := range populations.Of {
			counts[key{region.Map, population.Family}] += population.Count
		}
	}
	out := make([]FamilyCount, 0, len(counts))
	linked := g.linkedMaps()
	for k, count := range counts {
		if count > 0 && linked[k.mapID] {
			out = append(out, FamilyCount{Map: k.mapID, Town: towns[k.mapID], Family: k.family, Count: count})
		}
	}
	slices.SortFunc(out, func(a, b FamilyCount) int {
		return cmp.Or(cmp.Compare(a.Map, b.Map), cmp.Compare(a.Family, b.Family))
	})
	return out
}

// linkedMaps is the set of maps reachable from the party's map through gates, the party's own
// included.
func (g *Game) linkedMaps() map[api.MapID]bool {
	linked := map[api.MapID]bool{}
	party, ok := anatomy.ControlledParty(g.world)
	if !ok {
		return linked
	}
	positions := ecs.NewMap[comp.Pos](g.world.ECS)
	var queue []api.MapID
	for _, member := range anatomy.Members(g.world, party) {
		if positions.Has(member) {
			queue = append(queue, positions.Get(member).Map)
			break
		}
	}
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		if linked[id] {
			continue
		}
		linked[id] = true
		if m := g.world.Maps.Get(id); m != nil {
			for _, gate := range m.Gates() {
				queue = append(queue, gate.To)
			}
		}
	}
	return linked
}
