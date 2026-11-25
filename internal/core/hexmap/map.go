package hexmap

import (
	"cmp"
	"errors"
	"fmt"
	"slices"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// Edit is one cell changed from what the source supplied.
type Edit struct {
	Cell    api.Cell
	Terrain Terrain
}

// Map is the world's cells: a chunk source, a topology, and an overlay of edits. It is not safe for concurrent use.
type Map struct {
	src    Source
	topo   Topology
	chunks map[ChunkKey]*Chunk
	edits  map[api.Cell]Terrain
	// bounded is false for an endless map, which ignores bounds and edge.
	bounded bool
	bounds  Bounds
	edge    Terrain
	// sights remembers Visible results until the next edit. It is derived, never saved or hashed.
	sights map[sightKey][]api.Cell
}

// NewMap returns a map that loads chunks from src on first use.
func NewMap(src Source, topo Topology) *Map {
	return &Map{
		src: src, topo: topo,
		chunks: map[ChunkKey]*Chunk{},
		edits:  map[api.Cell]Terrain{},
	}
}

// ErrOutOfBounds is returned by [Map.Set] for a cell outside the map's bounds.
var ErrOutOfBounds = errors.New("hexmap: cell outside the map's bounds")

// SetBounds limits the map to b. Every cell outside it reads as edge, which should be solid and opaque so that
// nothing stands on, walks into or sees through it, and no chunk outside is ever loaded.
func (m *Map) SetBounds(b Bounds, edge Terrain) {
	m.bounded, m.bounds, m.edge = true, b, edge
	clear(m.sights)
}

// At returns the terrain of c, loading its chunk on first use and caching it. Outside the map's bounds it is the
// edge terrain.
func (m *Map) At(c api.Cell) (Terrain, error) {
	if m.bounded && !m.bounds.Contains(c) {
		return m.edge, nil
	}
	if t, ok := m.edits[c]; ok {
		return t, nil
	}
	key := KeyOf(c)
	chunk, ok := m.chunks[key]
	if !ok {
		var err error
		chunk, err = m.src.Chunk(key)
		if err != nil {
			return Terrain{}, fmt.Errorf("hexmap: load chunk %v: %w", key, err)
		}
		if chunk == nil {
			return Terrain{}, fmt.Errorf("hexmap: source returned no chunk for %v", key)
		}
		m.chunks[key] = chunk
	}
	return chunk.Get(floorMod(c.Hex.Q), floorMod(c.Hex.R), floorMod(c.Depth)), nil
}

// Set edits c. The edit lives in an overlay, so it can be saved and survives reloading the chunk.
func (m *Map) Set(c api.Cell, t Terrain) error {
	if m.bounded && !m.bounds.Contains(c) {
		return fmt.Errorf("set %v: %w", c, ErrOutOfBounds)
	}
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

func (m *Map) solid(l Link) (bool, error) {
	t, err := m.At(l.Cell)
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
