package explore

import (
	"cmp"
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/worldgen"
)

// pocketPlace names the pocket's footprint among the fixed places.
const pocketPlace api.ContentID = "pocket"

// pocketFixed is the fixed layer with the pocket added: whole chunks as they were before the export, joined
// to the new deep through a few openings.
type pocketFixed struct {
	worldgen.Fixed
	chunks    map[hexmap.ChunkKey]*hexmap.Chunk
	footprint hexmap.Footprint
}

// Chunk implements [hexmap.Source]. A pocket chunk is shared and must only be read.
func (p *pocketFixed) Chunk(key hexmap.ChunkKey) (*hexmap.Chunk, error) {
	if chunk, ok := p.chunks[key]; ok {
		return chunk, nil
	}
	return p.Fixed.Chunk(key)
}

// Covers implements [worldgen.Fixed].
func (p *pocketFixed) Covers(cell api.Cell) bool {
	_, ok := p.chunks[hexmap.KeyOf(cell)]
	return ok || p.Fixed.Covers(cell)
}

// Places implements [worldgen.Fixed].
func (p *pocketFixed) Places() []hexmap.Footprint {
	return append(p.Fixed.Places(), p.footprint)
}

// regenerate rebuilds the deep from the new export count around the party's pocket, forgets what the
// parties saw of it, and moves the bodies standing in it to open ground.
func regenerate(w *sim.World, _ api.ExportBegan) {
	if err := regenerateDeep(w); err != nil {
		w.Bus.Fail(fmt.Errorf("explore: export: %w", err))
	}
}

func regenerateDeep(w *sim.World) error {
	world := w.Content.World()
	base := stateOf(w).world
	if w.Map == nil || base == nil || world.Geography == "" {
		return nil
	}
	fixed, pocket, err := openPocket(w, base, memberCells(w), world)
	if err != nil {
		return err
	}
	cells, err := cellSource(w, fixed, base.Bounds())
	if err != nil {
		return err
	}
	st := stateOf(w)
	st.pocket = fixed
	st.generated, _ = cells.(*worldgen.Source)
	clear(st.populated)
	for _, key := range pocket {
		st.populated[key] = true
	}
	w.Map.Replace(cells, fixed.Covers)
	forgetSeen(w, slices.Concat(pocket, base.ChunkKeys()))
	return relocate(w, fixed)
}

// openPocket builds the pocket and its openings. A pocket with no opening gets a tunnel cut from the
// party's ground to new ground.
func openPocket(
	w *sim.World, base *WorldSource, members []api.Cell, world content.WorldDef,
) (*pocketFixed, []hexmap.ChunkKey, error) {
	radius := world.PocketChunks
	fixed, pocket, reached, err := buildPocket(w, base, members, world, radius)
	if err != nil {
		return nil, nil, err
	}
	if len(fixed.footprint.Openings) == 0 {
		if err := cutTunnel(w, fixed, reached, base.Bounds()); err != nil {
			return nil, nil, err
		}
		if fixed, pocket, _, err = buildPocket(w, base, members, world, radius); err != nil {
			return nil, nil, err
		}
		if len(fixed.footprint.Openings) == 0 {
			return nil, nil, fmt.Errorf("no way out of the pocket around %v", members)
		}
	}
	return fixed, pocket, nil
}

// buildPocket snapshots the chunks within radius of the members and finds their openings, returning the
// cells the party can walk to.
func buildPocket(
	w *sim.World, base *WorldSource, members []api.Cell, world content.WorldDef, radius int,
) (*pocketFixed, []hexmap.ChunkKey, []api.Cell, error) {
	pocket := pocketKeys(members, radius)
	fixed := &pocketFixed{Fixed: base, chunks: make(map[hexmap.ChunkKey]*hexmap.Chunk, len(pocket))}
	for _, key := range pocket {
		chunk, err := snapshot(w.Map, key)
		if err != nil {
			return nil, nil, nil, err
		}
		fixed.chunks[key] = chunk
	}
	exits, reached, err := pocketExits(w, fixed, members, base.Bounds(), world.PocketExits)
	if err != nil {
		return nil, nil, nil, err
	}
	fixed.footprint = hexmap.Footprint{Place: pocketPlace, Openings: exits}
	return fixed, pocket, reached, nil
}

// cutTunnel opens the shortest line of cells on one level from the reached region to a cell bordering new
// ground, keeping the floor under every cell it opens.
func cutTunnel(w *sim.World, fixed *pocketFixed, reached []api.Cell, bounds hexmap.Bounds) error {
	if len(reached) == 0 {
		return errors.New("no ground to cut a tunnel from")
	}
	topo := hexmap.Prisms{}
	open, err := w.Map.At(reached[0])
	if err != nil {
		return err
	}
	parent := make(map[api.Cell]api.Cell, len(reached))
	for _, at := range reached {
		parent[at] = at
	}
	frontier := slices.Clone(reached)
	for len(frontier) > 0 {
		cell := frontier[0]
		frontier = frontier[1:]
		for dir := range api.Dir(api.DirCount) {
			next := topo.Neighbor(cell, dir).Cell
			if !bounds.Contains(next) {
				continue
			}
			if !fixed.Covers(next) {
				if parent[cell] == cell {
					continue
				}
				return carve(w, parent, cell, open)
			}
			if _, seen := parent[next]; seen {
				continue
			}
			floor, err := w.Map.At(topo.Below(next).Cell)
			if err != nil {
				return err
			}
			if floor.Solid {
				parent[next] = cell
				frontier = append(frontier, next)
			}
		}
	}
	return fmt.Errorf("no tunnel from %v reaches new ground", reached[0])
}

// carve opens every cell on the path from end back to the region.
func carve(w *sim.World, parent map[api.Cell]api.Cell, end api.Cell, open hexmap.Terrain) error {
	for cell := end; parent[cell] != cell; cell = parent[cell] {
		if err := w.Map.Set(cell, open); err != nil {
			return err
		}
	}
	return nil
}

// memberCells lists where every party member stands, by UID.
func memberCells(w *sim.World) []api.Cell {
	type member struct {
		uid api.UID
		at  api.Cell
	}
	var found []member
	query := ecs.NewFilter2[comp.MemberOf, comp.Pos](w.ECS).Query()
	for query.Next() {
		_, pos := query.Get()
		found = append(found, member{uid: w.UIDOf(query.Entity()), at: pos.At})
	}
	slices.SortFunc(found, func(a, b member) int { return cmp.Compare(a.uid, b.uid) })
	cells := make([]api.Cell, len(found))
	for i, m := range found {
		cells[i] = m.at
	}
	return cells
}

// pocketKeys lists the chunks within radius chunks of any member, sorted.
func pocketKeys(members []api.Cell, radius int) []hexmap.ChunkKey {
	set := map[hexmap.ChunkKey]bool{}
	for _, at := range members {
		home := hexmap.KeyOf(at)
		for depth := -radius; depth <= radius; depth++ {
			for r := -radius; r <= radius; r++ {
				for q := -radius; q <= radius; q++ {
					key := hexmap.ChunkKey{Q: home.Q + q, R: home.R + r, Depth: home.Depth + depth}
					if hexmap.ChunkDistance(key, home) <= radius {
						set[key] = true
					}
				}
			}
		}
	}
	keys := make([]hexmap.ChunkKey, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	slices.SortFunc(keys, hexmap.CompareKeys)
	return keys
}

// snapshot copies a chunk as the map shows it now, edits included.
func snapshot(m *hexmap.Map, key hexmap.ChunkKey) (*hexmap.Chunk, error) {
	first, err := m.At(key.Cell(0))
	if err != nil {
		return nil, err
	}
	chunk := hexmap.NewChunk(first)
	origin := key.Origin()
	for slot := range hexmap.ChunkCells {
		cell := key.Cell(slot)
		t, err := m.At(cell)
		if err != nil {
			return nil, err
		}
		chunk.Set(cell.Hex.Q-origin.Hex.Q, cell.Hex.R-origin.Hex.R, cell.Depth-origin.Depth, t)
	}
	return chunk, nil
}

// pocketExits walks from the members over the old ground, inside the pocket, and picks up to most of the
// standable cells that border new ground, spread evenly along the sorted list.
func pocketExits(
	w *sim.World, fixed *pocketFixed, members []api.Cell, bounds hexmap.Bounds, most int,
) (exits, reached []api.Cell, err error) {
	rules := stateOf(w).rules
	topo := hexmap.Prisms{}
	seen := map[api.Cell]bool{}
	var frontier []api.Cell
	for _, at := range members {
		if !seen[at] {
			seen[at] = true
			frontier = append(frontier, at)
		}
	}
	var edge []api.Cell
	for len(frontier) > 0 {
		cell := frontier[0]
		frontier = frontier[1:]
		borders := false
		for dir := range api.Dir(api.DirCount) {
			if next := topo.Neighbor(cell, dir).Cell; bounds.Contains(next) && !fixed.Covers(next) {
				borders = true
			}
			step, can, err := w.Map.Step(cell, dir, rules)
			if err != nil {
				return nil, nil, err
			}
			if to := step.To.Cell; can && fixed.Covers(to) && !seen[to] {
				seen[to] = true
				frontier = append(frontier, to)
			}
		}
		if borders {
			edge = append(edge, cell)
		}
	}
	slices.SortFunc(edge, hexmap.CompareCells)
	reached = slices.Collect(maps.Keys(seen))
	slices.SortFunc(reached, hexmap.CompareCells)
	if len(edge) <= most {
		return edge, reached, nil
	}
	exits = make([]api.Cell, most)
	for i := range exits {
		exits[i] = edge[i*len(edge)/most]
	}
	return exits, reached, nil
}

// forgetSeen drops every party's memory of chunks outside kept.
func forgetSeen(w *sim.World, kept []hexmap.ChunkKey) {
	query := ecs.NewFilter1[comp.Seen](w.ECS).Query()
	for query.Next() {
		seen := query.Get()
		seen.Chunks = slices.DeleteFunc(seen.Chunks, func(chunk comp.SeenChunk) bool {
			return !slices.Contains(kept, seenKey(chunk))
		})
	}
}

// relocate moves every body standing in new ground to the nearest free standable cell outside the pocket,
// bodies in UID order.
func relocate(w *sim.World, fixed *pocketFixed) error {
	type body struct {
		entity ecs.Entity
		uid    api.UID
	}
	var bodies []body
	query := ecs.NewFilter1[comp.Pos](w.ECS).Query()
	for query.Next() {
		if pos := query.Get(); !fixed.Covers(pos.At) {
			bodies = append(bodies, body{entity: query.Entity(), uid: w.UIDOf(query.Entity())})
		}
	}
	slices.SortFunc(bodies, func(a, b body) int { return cmp.Compare(a.uid, b.uid) })
	positions := ecs.NewMap[comp.Pos](w.ECS)
	for _, b := range bodies {
		pos := positions.Get(b.entity)
		to, err := nearestOpen(w, pos.At, b.entity, fixed)
		if err != nil {
			return err
		}
		pos.At = to
	}
	return nil
}

// nearestOpen finds the free standable cell nearest from, in rings of growing radius, ties by cell order.
// A cell in the pocket does not count.
func nearestOpen(w *sim.World, from api.Cell, self ecs.Entity, fixed *pocketFixed) (api.Cell, error) {
	reach := 2 * hexmap.ChunkSize
	for radius := range reach {
		for _, cell := range w.Map.CellsIn(from.Hex, radius, from.Depth-radius, from.Depth+radius) {
			if max(hexmap.Distance(cell.Hex, from.Hex), abs(cell.Depth-from.Depth)) != radius {
				continue
			}
			if _, pocketed := fixed.chunks[hexmap.KeyOf(cell)]; pocketed {
				continue
			}
			ok, err := w.Map.Standable(cell)
			if err != nil {
				return api.Cell{}, err
			}
			if _, taken := bodyAt(w, cell, self); ok && !taken {
				return cell, nil
			}
		}
	}
	return api.Cell{}, fmt.Errorf("no open ground within %d cells of %v", reach, from)
}

func abs(n int) int { return max(n, -n) }
