package hexmap

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// Edit is one cell changed from what the source supplied.
type Edit struct {
	Cell    api.Cell
	Terrain Terrain
}

// Map is one map: a chunk source, a topology, and an overlay of edits. It is not safe for concurrent use.
type Map struct {
	id     api.MapID
	name   string
	src    Source
	topo   Topology
	chunks map[ChunkKey]*Chunk
	edits  map[api.Cell]Terrain
	// sights remembers Visible results until the next edit. It is derived, never saved or hashed.
	sights map[sightKey][]api.Cell
}

// NewMap returns a map that loads chunks from src on first use.
func NewMap(id api.MapID, name string, src Source, topo Topology) *Map {
	return &Map{
		id: id, name: name, src: src, topo: topo,
		chunks: map[ChunkKey]*Chunk{},
		edits:  map[api.Cell]Terrain{},
	}
}

// ID is the map's ID.
func (m *Map) ID() api.MapID { return m.id }

// Name is the map's display name.
func (m *Map) Name() string { return m.name }

// Marked is a [Source] that names some of its cells, such as where a party starts.
type Marked interface {
	Mark(name string) (api.Cell, bool)
}

// Mark returns the cell of the named mark, when the map's source is [Marked] and has it.
func (m *Map) Mark(name string) (api.Cell, bool) {
	marked, ok := m.src.(Marked)
	if !ok {
		return api.Cell{}, false
	}
	return marked.Mark(name)
}

// At returns the terrain of c, loading its chunk on first use and caching it.
func (m *Map) At(c api.Cell) (Terrain, error) {
	if t, ok := m.edits[c]; ok {
		return t, nil
	}
	key := KeyOf(c)
	chunk, ok := m.chunks[key]
	if !ok {
		var err error
		chunk, err = m.src.Chunk(key)
		if err != nil {
			return Terrain{}, fmt.Errorf("hexmap: load chunk %v of map %d: %w", key, m.id, err)
		}
		if chunk == nil {
			return Terrain{}, fmt.Errorf("hexmap: source returned no chunk for %v of map %d", key, m.id)
		}
		m.chunks[key] = chunk
	}
	return chunk.Get(floorMod(c.Hex.Q), floorMod(c.Hex.R), floorMod(c.Depth)), nil
}

// Set edits c. The edit lives in an overlay, so it can be saved and survives reloading the chunk.
func (m *Map) Set(c api.Cell, t Terrain) error {
	m.edits[c] = t
	clear(m.sights)
	return nil
}

// Edits lists every edit, sorted by Depth, R, Q.
func (m *Map) Edits() []Edit {
	out := make([]Edit, 0, len(m.edits))
	for cell, t := range m.edits {
		out = append(out, Edit{Cell: cell, Terrain: t})
	}
	slices.SortFunc(out, func(a, b Edit) int { return CompareCells(a.Cell, b.Cell) })
	return out
}

// CompareCells orders cells by Depth, then R, then Q.
func CompareCells(a, b api.Cell) int {
	return cmp.Or(cmp.Compare(a.Depth, b.Depth), cmp.Compare(a.Hex.R, b.Hex.R), cmp.Compare(a.Hex.Q, b.Hex.Q))
}

// CellsIn lists the cells within radius hexes of center and between depthFrom and depthTo inclusive, sorted by
// Depth, R, Q.
func (m *Map) CellsIn(center api.Hex, radius, depthFrom, depthTo int) []api.Cell {
	if radius < 0 || depthFrom > depthTo {
		return nil
	}
	var out []api.Cell
	for depth := depthFrom; depth <= depthTo; depth++ {
		for dr := -radius; dr <= radius; dr++ {
			for dq := max(-radius, -dr-radius); dq <= min(radius, -dr+radius); dq++ {
				out = append(out, api.Cell{Hex: api.Hex{Q: center.Q + dq, R: center.R + dr}, Depth: depth})
			}
		}
	}
	return out
}

// local resolves a link to a cell of this map; ok is false when it leads to another map.
func (m *Map) local(l Link) (api.Cell, bool) {
	return l.Cell, l.Map == 0 || l.Map == m.id
}

// solid reports whether a link leads to a solid cell of this map. A link to another map counts as solid, which
// keeps bodies from standing on or passing through cells this map can't see.
func (m *Map) solid(l Link) (bool, error) {
	cell, ok := m.local(l)
	if !ok {
		return true, nil
	}
	t, err := m.At(cell)
	return t.Solid, err
}

// Standable reports whether c is open and the cell below it is solid.
func (m *Map) Standable(c api.Cell) (bool, error) {
	t, err := m.At(c)
	if err != nil || t.Solid {
		return false, err
	}
	return m.solid(m.topo.Below(c))
}

// Topology returns how the map's cells connect, for callers that need neighbors.
func (m *Map) Topology() Topology { return m.topo }
