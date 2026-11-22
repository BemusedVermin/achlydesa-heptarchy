package explore

import (
	"cmp"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
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
		cfg := w.Content.Party()
		st.rules = hexmap.StepRules{MaxStepUp: cfg.MaxStepUp, MaxStepDown: cfg.MaxStepDown, MaxDrop: cfg.MaxDrop}
	}
	res.Add(st)
	return st
}

// compareCells orders cells by depth, then R, then Q.
func compareCells(a, b api.Cell) int {
	return cmp.Or(cmp.Compare(a.Depth, b.Depth), cmp.Compare(a.Hex.R, b.Hex.R), cmp.Compare(a.Hex.Q, b.Hex.Q))
}

// controlledParty returns the party the player commands.
func controlledParty(w *sim.World) (ecs.Entity, bool) {
	query := ecs.NewFilter1[comp.Controlled](w.ECS).Query()
	for query.Next() {
		e := query.Entity()
		query.Close()
		return e, true
	}
	return ecs.Entity{}, false
}

// membersOf lists the members of a party in party order: the leader first, each follower behind the
// member it follows. Anyone the order does not name comes last, by UID.
func membersOf(w *sim.World, party ecs.Entity) []ecs.Entity {
	var members []ecs.Entity
	query := ecs.NewFilter1[comp.MemberOf](w.ECS).Query(ecs.Rel[comp.MemberOf](party))
	for query.Next() {
		members = append(members, query.Entity())
	}
	rank := map[api.UID]int{}
	if parties := ecs.NewMap[comp.Party](w.ECS); alive(w, party) && parties.Has(party) {
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
	slices.SortFunc(members, func(a, b ecs.Entity) int {
		return cmp.Or(cmp.Compare(rankOf(a), rankOf(b)), cmp.Compare(w.UIDOf(a), w.UIDOf(b)))
	})
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

// partyOf returns the party a member belongs to.
func partyOf(w *sim.World, member ecs.Entity) (ecs.Entity, bool) {
	query := ecs.NewFilter1[comp.MemberOf](w.ECS).Query()
	for query.Next() {
		if query.Entity() == member {
			party := query.GetRelation(0)
			query.Close()
			return party, true
		}
	}
	return ecs.Entity{}, false
}

func alive(w *sim.World, e ecs.Entity) bool { return !e.IsZero() && w.ECS.Alive(e) }

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
	slices.SortFunc(fresh, compareCells)
	fresh = slices.CompactFunc(fresh, func(a, b api.Cell) bool { return a == b })
	known.ByMap[id] = mergeCells(known.ByMap[id], fresh)
}

// mergeCells merges two sorted, unique slices into one.
func mergeCells(a, b []api.Cell) []api.Cell {
	out := make([]api.Cell, 0, len(a)+len(b))
	i, j := 0, 0
	for i < len(a) && j < len(b) {
		switch c := compareCells(a[i], b[j]); {
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
	if !alive(w, party) || !seen.Has(party) {
		return nil
	}
	return seen.Get(party).ByMap[id]
}

// sightOf lists what a body standing at pos can see.
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
	identities := ecs.NewMap[comp.Identity](w.ECS)
	if !alive(w, e) || !identities.Has(e) {
		return api.Text{ID: "text.template.unknown"}
	}
	template := strings.TrimPrefix(string(identities.Get(e).Template), "template.")
	return api.Text{ID: api.ContentID("text.template." + template)}
}

// Start builds every map of the content and, when the content has a start scenario, spawns the
// start party on its mark. game.New calls it once, after Install.
func Start(w *sim.World) error {
	if w.Content == nil {
		return errors.New("explore: start: world has no content")
	}
	st := stateOf(w)
	ids := map[api.ContentID]api.MapID{}
	sources := map[api.ContentID]*Source{}
	for i, id := range w.Content.MapIDs() {
		def, _ := w.Content.Map(id)
		src, err := NewSource(w.Content, def)
		if err != nil {
			return fmt.Errorf("explore: map %q: %w", id, err)
		}
		mapID := api.MapID(i + 1)
		if err := w.Maps.Add(hexmap.NewMap(mapID, string(id), src, hexmap.Prisms{})); err != nil {
			return fmt.Errorf("explore: map %q: %w", id, err)
		}
		ids[id], sources[id], st.names[mapID] = mapID, src, id
	}
	start := w.Content.Start()
	if start.Map == "" {
		return nil
	}
	mark, ok := sources[start.Map].Mark(start.PartyAtMark)
	if !ok {
		return fmt.Errorf("explore: start: map %q has no mark %q", start.Map, start.PartyAtMark)
	}
	return spawnParty(w, ids[start.Map], mark, start.Party)
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
			return slices.MinFunc(free, compareCells), nil
		}
		var next []api.Cell
		for _, cell := range frontier {
			for dir := api.Dir(0); dir < api.DirCount; dir++ {
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

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// Content check: the terrain a cell shows in the plan view.
func terrainOf(reg *content.Registry, id api.ContentID) content.Terrain {
	t, _ := reg.Terrain(id)
	return t
}
