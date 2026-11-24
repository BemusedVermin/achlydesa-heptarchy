package explore

import (
	"cmp"
	"errors"
	"fmt"
	"slices"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/actions"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// state is what explore keeps on the world, as an Ark resource.
type state struct {
	rules hexmap.StepRules
	// names says which content map each map ID was built from.
	names map[api.MapID]api.ContentID
}

func stateOf(w *sim.World) *state {
	res := ecs.NewResource[state](w.ECS)
	if res.Has() {
		return res.Get()
	}
	st := &state{names: map[api.MapID]api.ContentID{}}
	if w.Content != nil {
		st.rules = actions.StepRules(w)
	}
	res.Add(st)
	return st
}

// membersOf lists the members of a party in party order: the leader first, each follower behind the
// member it follows. Anyone the order does not name comes last, by UID.
func membersOf(w *sim.World, party ecs.Entity) []ecs.Entity {
	members := anatomy.Members(w, party)
	rank := map[api.UID]int{}
	if parties := ecs.NewMap[comp.Party](w.ECS); w.Alive(party) && parties.Has(party) {
		for i, uid := range parties.Get(party).Order {
			rank[uid] = i
		}
	}
	rankOf := func(e ecs.Entity) int {
		if r, ok := rank[w.UIDOf(e)]; ok {
			return r
		}
		return len(rank)
	}
	slices.SortStableFunc(members, func(a, b ecs.Entity) int { return cmp.Compare(rankOf(a), rankOf(b)) })
	return members
}

// bodyAt returns the body standing in the cell, other than except. With several, the lowest UID.
func bodyAt(w *sim.World, mapID api.MapID, cell api.Cell, except ecs.Entity) (ecs.Entity, bool) {
	var found ecs.Entity
	query := ecs.NewFilter2[comp.Pos, comp.Actor](w.ECS).Query()
	for query.Next() {
		pos, _ := query.Get()
		e := query.Entity()
		if e == except || pos.Map != mapID || pos.At != cell {
			continue
		}
		if found.IsZero() || w.UIDOf(e) < w.UIDOf(found) {
			found = e
		}
	}
	return found, !found.IsZero()
}

// markSeen records cells a party has seen on a map. cells need not be sorted.
func markSeen(w *sim.World, party ecs.Entity, id api.MapID, cells []api.Cell) {
	seen := ecs.NewMap[comp.Seen](w.ECS)
	if !seen.Has(party) {
		seen.Add(party, &comp.Seen{ByMap: map[api.MapID][]api.Cell{}})
	}
	known := seen.Get(party)
	if known.ByMap == nil {
		known.ByMap = map[api.MapID][]api.Cell{}
	}
	fresh := slices.Clone(cells)
	slices.SortFunc(fresh, hexmap.CompareCells)
	fresh = slices.CompactFunc(fresh, func(a, b api.Cell) bool { return a == b })
	known.ByMap[id] = mergeCells(known.ByMap[id], fresh)
}

// mergeCells merges two sorted, unique slices into one.
func mergeCells(a, b []api.Cell) []api.Cell {
	out := make([]api.Cell, 0, len(a)+len(b))
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		switch c := hexmap.CompareCells(a[i], b[j]); {
		case c < 0:
			out = append(out, a[i])
			i++
		case c > 0:
			out = append(out, b[j])
			j++
		default:
			out = append(out, a[i])
			i++
			j++
		}
	}
	out = append(out, a[i:]...)
	return append(out, b[j:]...)
}

// seenIn returns the cells a party has seen on a map, sorted by depth, then R, then Q.
func seenIn(w *sim.World, party ecs.Entity, id api.MapID) []api.Cell {
	seen := ecs.NewMap[comp.Seen](w.ECS)
	if !w.Alive(party) || !seen.Has(party) {
		return nil
	}
	return seen.Get(party).ByMap[id]
}

func sightOf(w *sim.World, pos comp.Pos) ([]api.Cell, error) {
	m := w.Maps.Get(pos.Map)
	if m == nil {
		return nil, nil
	}
	cfg := w.Content.Party()
	cells, err := m.Visible(pos.At, cfg.SightRadius, cfg.SightDepth)
	if err != nil {
		return nil, fmt.Errorf("explore: sight: %w", err)
	}
	return cells, nil
}

// partySight is every cell any member of the party can see now, by map.
func partySight(w *sim.World, party ecs.Entity) (map[api.MapID]map[api.Cell]bool, error) {
	positions := ecs.NewMap[comp.Pos](w.ECS)
	out := map[api.MapID]map[api.Cell]bool{}
	for _, member := range membersOf(w, party) {
		if !positions.Has(member) {
			continue
		}
		pos := *positions.Get(member)
		cells, err := sightOf(w, pos)
		if err != nil {
			return nil, err
		}
		if out[pos.Map] == nil {
			out[pos.Map] = make(map[api.Cell]bool, len(cells))
		}
		for _, cell := range cells {
			out[pos.Map][cell] = true
		}
	}
	return out, nil
}

// templateName is the text ID that names a body, "text.template.<template without its prefix>".
func templateName(w *sim.World, e ecs.Entity) api.Text {
	return w.Content.TemplateName(w.TemplateOf(e))
}

// LoadMaps builds every map from content and registers it, without starting a game. Loading a save
// calls it instead of Start, then reapplies the saved map edits.
func LoadMaps(w *sim.World) error {
	_, _, err := buildMaps(w)
	return err
}

// buildMaps registers a map per content map, numbered in content ID order from 1.
func buildMaps(w *sim.World) (map[api.ContentID]api.MapID, map[api.ContentID]*Source, error) {
	if w.Content == nil {
		return nil, nil, errors.New("explore: start: world has no content")
	}
	st := stateOf(w)
	ids := map[api.ContentID]api.MapID{}
	sources := map[api.ContentID]*Source{}
	for i, id := range w.Content.MapIDs() {
		def, _ := w.Content.Map(id)
		src, err := NewSource(w.Content, def)
		if err != nil {
			return nil, nil, fmt.Errorf("explore: map %q: %w", id, err)
		}
		mapID := api.MapID(i + 1)
		if err := w.Maps.Add(hexmap.NewMap(mapID, string(id), src, hexmap.Prisms{})); err != nil {
			return nil, nil, fmt.Errorf("explore: map %q: %w", id, err)
		}
		ids[id], sources[id], st.names[mapID] = mapID, src, id
	}
	if err := resolveGates(w, ids, sources); err != nil {
		return nil, nil, err
	}
	return ids, sources, nil
}

// resolveGates fills each map's gate table, once every map has its ID.
func resolveGates(w *sim.World, ids map[api.ContentID]api.MapID, sources map[api.ContentID]*Source) error {
	for _, name := range w.Content.MapIDs() {
		src := sources[name]
		var gates []hexmap.Gate
		for _, site := range src.Gates() {
			dest, ok := sources[site.To]
			if !ok {
				return fmt.Errorf("explore: map %q: gate to unknown map %q", name, site.To)
			}
			cell, ok := dest.Mark(site.Mark)
			if !ok {
				return fmt.Errorf("explore: map %q: gate to map %q has no mark %q", name, site.To, site.Mark)
			}
			gates = append(gates, hexmap.Gate{At: site.Cell, To: ids[site.To], ToCell: cell})
		}
		w.Maps.Get(ids[name]).SetGates(gates)
	}
	return nil
}

// Start builds every map of the content and, when the content has a start scenario, spawns the
// start party on its mark. game.New calls it once, after Install.
func Start(w *sim.World) error {
	ids, sources, err := buildMaps(w)
	if err != nil {
		return err
	}
	start := w.Content.Start()
	if start.Map != "" {
		mark, ok := sources[start.Map].Mark(start.PartyAtMark)
		if !ok {
			return fmt.Errorf("explore: start: map %q has no mark %q", start.Map, start.PartyAtMark)
		}
		if err := spawnParty(w, ids[start.Map], mark, start.Party); err != nil {
			return err
		}
	}
	return spawnMarked(w, ids, sources)
}

// spawnMarked spawns the bodies each map's spawn marks name, maps in content ID order and each
// map's marks in row order.
func spawnMarked(w *sim.World, ids map[api.ContentID]api.MapID, sources map[api.ContentID]*Source) error {
	for _, name := range w.Content.MapIDs() {
		for _, spawn := range sources[name].Spawns() {
			if _, err := w.Spawn(spawn.Template, &comp.Pos{Map: ids[name], At: spawn.Cell}); err != nil {
				return fmt.Errorf("explore: start: spawn %q on %q: %w", spawn.Template, name, err)
			}
		}
	}
	return nil
}

// spawnParty spawns the template's party. The leader stands on the mark; each other member stands
// on the nearest free standable cell, nearest in steps.
func spawnParty(w *sim.World, id api.MapID, mark api.Cell, template api.ContentID) error {
	party, err := w.Spawn(template, nil)
	if err != nil {
		return fmt.Errorf("explore: start party: %w", err)
	}
	m := w.Maps.Get(id)
	positions := ecs.NewMap[comp.Pos](w.ECS)
	trails := ecs.NewMap[comp.Trail](w.ECS)
	taken := map[api.Cell]bool{}
	for i, member := range membersOf(w, party) {
		at := mark
		if i > 0 {
			if at, err = nearestFree(m, mark, stateOf(w).rules, taken); err != nil {
				return err
			}
		}
		taken[at] = true
		positions.Add(member, &comp.Pos{Map: id, At: at})
		trails.Add(member, &comp.Trail{Last: at})
	}
	ecs.NewMap[comp.Controlled](w.ECS).Add(party, &comp.Controlled{})
	SetRestPlace(w, party, id, mark)
	for _, member := range membersOf(w, party) {
		cells, err := sightOf(w, *positions.Get(member))
		if err != nil {
			return err
		}
		markSeen(w, party, id, cells)
	}
	return nil
}

// searchRadius is how many steps from the mark a member is placed before giving up.
const searchRadius = 8

// nearestFree finds the free standable cell fewest steps from the mark, searching breadth-first
// over Step. Ties go to the first cell by depth, then R, then Q.
func nearestFree(
	m *hexmap.Map, mark api.Cell, rules hexmap.StepRules, taken map[api.Cell]bool,
) (api.Cell, error) {
	seen := map[api.Cell]bool{mark: true}
	frontier := []api.Cell{mark}
	for radius := 0; radius <= searchRadius && len(frontier) > 0; radius++ {
		var free []api.Cell
		for _, cell := range frontier {
			if !taken[cell] {
				free = append(free, cell)
			}
		}
		if len(free) > 0 {
			return slices.MinFunc(free, hexmap.CompareCells), nil
		}
		var next []api.Cell
		for _, cell := range frontier {
			for dir := range api.Dir(api.DirCount) {
				step, can, err := m.Step(cell, dir, rules)
				if err != nil {
					return api.Cell{}, err
				}
				if !can || (step.To.Map != 0 && step.To.Map != m.ID()) || seen[step.To.Cell] {
					continue
				}
				seen[step.To.Cell] = true
				next = append(next, step.To.Cell)
			}
		}
		frontier = next
	}
	return api.Cell{}, fmt.Errorf("explore: no free standable cell within %d steps of the start mark", searchRadius)
}

// terrainOf is the content record of a terrain, or the zero record when content lacks it.
func terrainOf(reg *content.Registry, id api.ContentID) content.Terrain {
	t, _ := reg.Terrain(id)
	return t
}
