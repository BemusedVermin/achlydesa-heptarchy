package explore

import (
	"cmp"
	"errors"
	"fmt"
	"math/bits"
	"slices"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/actions"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/worldgen"
)

// state is what explore keeps on the world, as an Ark resource.
type state struct {
	rules hexmap.StepRules
	// generated is the world's generated ground, nil when the world has none.
	generated *worldgen.Source
	// populated holds the chunks whose starting bodies have been spawned for generated's export count.
	populated map[hexmap.ChunkKey]bool
}

func stateOf(w *sim.World) *state {
	res := ecs.NewResource[state](w.ECS)
	if res.Has() {
		return res.Get()
	}
	st := &state{populated: map[hexmap.ChunkKey]bool{}}
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
func bodyAt(w *sim.World, cell api.Cell, except ecs.Entity) (ecs.Entity, bool) {
	var found ecs.Entity
	query := ecs.NewFilter2[comp.Pos, comp.Actor](w.ECS).Query()
	for query.Next() {
		pos, _ := query.Get()
		e := query.Entity()
		if e == except || pos.At != cell {
			continue
		}
		if found.IsZero() || w.UIDOf(e) < w.UIDOf(found) {
			found = e
		}
	}
	return found, !found.IsZero()
}

// markSeen records cells a party has seen. cells need not be sorted.
func markSeen(w *sim.World, party ecs.Entity, cells []api.Cell) {
	seen := ecs.NewMap[comp.Seen](w.ECS)
	if !seen.Has(party) {
		seen.Add(party, &comp.Seen{})
	}
	known := seen.Get(party)
	for _, cell := range cells {
		key := hexmap.KeyOf(cell)
		i, found := slices.BinarySearchFunc(known.Chunks, key, compareSeenChunk)
		if !found {
			chunk := comp.SeenChunk{Q: key.Q, R: key.R, Depth: key.Depth, Bits: make([]uint64, seenWords)}
			known.Chunks = slices.Insert(known.Chunks, i, chunk)
		}
		slot := hexmap.SlotOf(cell)
		known.Chunks[i].Bits[slot/64] |= 1 << (slot % 64)
	}
}

const seenWords = hexmap.ChunkCells / 64

func seenKey(chunk comp.SeenChunk) hexmap.ChunkKey {
	return hexmap.ChunkKey{Q: chunk.Q, R: chunk.R, Depth: chunk.Depth}
}

func compareSeenChunk(chunk comp.SeenChunk, key hexmap.ChunkKey) int {
	return hexmap.CompareKeys(seenKey(chunk), key)
}

// seenIn returns the cells a party has seen, sorted by depth, then R, then Q.
func seenIn(w *sim.World, party ecs.Entity) []api.Cell {
	seen := ecs.NewMap[comp.Seen](w.ECS)
	if !w.Alive(party) || !seen.Has(party) {
		return nil
	}
	var out []api.Cell
	for _, chunk := range seen.Get(party).Chunks {
		key := seenKey(chunk)
		for word, set := range chunk.Bits {
			for set != 0 {
				bit := bits.TrailingZeros64(set)
				set &^= 1 << bit
				out = append(out, key.Cell(word*64+bit))
			}
		}
	}
	slices.SortFunc(out, hexmap.CompareCells)
	return out
}

func sightOf(w *sim.World, pos comp.Pos) ([]api.Cell, error) {
	if w.Map == nil {
		return nil, nil
	}
	cfg := w.Content.Party()
	cells, err := w.Map.Visible(pos.At, cfg.SightRadius, cfg.SightDepth)
	if err != nil {
		return nil, fmt.Errorf("explore: sight: %w", err)
	}
	return cells, nil
}

// partySight is every cell any member of the party can see now.
func partySight(w *sim.World, party ecs.Entity) (map[api.Cell]bool, error) {
	positions := ecs.NewMap[comp.Pos](w.ECS)
	out := map[api.Cell]bool{}
	for _, member := range membersOf(w, party) {
		if !positions.Has(member) {
			continue
		}
		cells, err := sightOf(w, *positions.Get(member))
		if err != nil {
			return nil, err
		}
		for _, cell := range cells {
			out[cell] = true
		}
	}
	return out, nil
}

// templateName is the text ID that names a body, "text.template.<template without its prefix>".
func templateName(w *sim.World, e ecs.Entity) api.Text {
	return w.Content.TemplateName(w.TemplateOf(e))
}

// LoadWorld builds the world's cells from content, without starting a game. Loading a save calls
// it instead of Start, then reapplies the saved edits.
func LoadWorld(w *sim.World) error {
	_, err := buildWorld(w)
	return err
}

// buildWorld sets w.Map to the cells of the content's world and places, bounded by the world's
// size. Content with no world leaves w.Map nil.
func buildWorld(w *sim.World) (*WorldSource, error) {
	if w.Content == nil {
		return nil, errors.New("explore: start: world has no content")
	}
	if w.Content.World() == (content.WorldDef{}) {
		return nil, nil
	}
	src, err := NewWorldSource(w.Content)
	if err != nil {
		return nil, fmt.Errorf("explore: world: %w", err)
	}
	cells, err := cellSource(w, src)
	if err != nil {
		return nil, fmt.Errorf("explore: world: %w", err)
	}
	st := stateOf(w)
	st.generated, _ = cells.(*worldgen.Source)
	clear(st.populated)
	w.Map = hexmap.NewMap(cells, hexmap.Prisms{})
	w.Map.SetBounds(src.Bounds(), src.Edge())
	return src, nil
}

// Start builds the world and, when the content has a start scenario, spawns the start party on its
// mark, then every place's spawns. game.New calls it once, after Install.
func Start(w *sim.World) error {
	src, err := buildWorld(w)
	if err != nil || src == nil {
		return err
	}
	start := w.Content.Start()
	if start.Place != "" {
		mark, ok := src.Mark(start.Place, start.Mark)
		if !ok {
			return fmt.Errorf("explore: start: place %q has no mark %q", start.Place, start.Mark)
		}
		if err := spawnParty(w, mark, start.Party); err != nil {
			return err
		}
	}
	for _, spawn := range src.Spawns() {
		if _, err := w.Spawn(spawn.Template, &comp.Pos{At: spawn.Cell}); err != nil {
			return fmt.Errorf("explore: start: spawn %q at %v: %w", spawn.Template, spawn.Cell, err)
		}
	}
	return populateLoaded(w)
}

// spawnParty spawns the template's party. The leader stands on the mark; each other member stands
// on the nearest free standable cell, nearest in steps.
func spawnParty(w *sim.World, mark api.Cell, template api.ContentID) error {
	party, err := w.Spawn(template, nil)
	if err != nil {
		return fmt.Errorf("explore: start party: %w", err)
	}
	positions := ecs.NewMap[comp.Pos](w.ECS)
	trails := ecs.NewMap[comp.Trail](w.ECS)
	taken := map[api.Cell]bool{}
	for i, member := range membersOf(w, party) {
		at := mark
		if i > 0 {
			if at, err = nearestFree(w.Map, mark, stateOf(w).rules, taken); err != nil {
				return err
			}
		}
		taken[at] = true
		positions.Add(member, &comp.Pos{At: at})
		trails.Add(member, &comp.Trail{Last: at})
	}
	ecs.NewMap[comp.Controlled](w.ECS).Add(party, &comp.Controlled{})
	SetRestPlace(w, party, mark)
	for _, member := range membersOf(w, party) {
		cells, err := sightOf(w, *positions.Get(member))
		if err != nil {
			return err
		}
		markSeen(w, party, cells)
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
				if !can || seen[step.To.Cell] {
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
