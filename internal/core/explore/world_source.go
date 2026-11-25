package explore

import (
	"fmt"
	"slices"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
)

// WorldSource is a [hexmap.Source] for the one world: every place's layers at its position, set in the
// world's fill. A [hexmap.Map] over it takes [WorldSource.Bounds] and [WorldSource.Edge].
type WorldSource struct {
	fill, edge hexmap.Terrain
	bounds     hexmap.Bounds
	authored   map[hexmap.ChunkKey][]placed
	marks      map[placeMark]api.Cell
	spawns     []Spawn
}

type placeMark struct {
	place api.ContentID
	name  string
}

// NewWorldSource builds the world of the content's world and places. Marked cells hold the world's
// Open terrain.
func NewWorldSource(reg *content.Registry) (*WorldSource, error) {
	world := reg.World()
	src := &WorldSource{
		bounds: hexmap.Bounds{
			MaxCol: world.Cols - 1, MaxRow: world.Rows - 1,
			MinDepth: world.MinDepth, MaxDepth: world.MaxDepth,
		},
		authored: map[hexmap.ChunkKey][]placed{},
		marks:    map[placeMark]api.Cell{},
	}
	var err error
	if src.fill, err = Terrain(reg, world.Fill); err != nil {
		return nil, fmt.Errorf("world fill: %w", err)
	}
	if src.edge, err = Terrain(reg, world.Edge); err != nil {
		return nil, fmt.Errorf("world edge: %w", err)
	}
	open, err := Terrain(reg, world.Open)
	if err != nil {
		return nil, fmt.Errorf("world open: %w", err)
	}
	for _, id := range reg.PlaceIDs() {
		def, _ := reg.Place(id)
		err := layOut(reg, def, open, src.authored, func(name string, cell api.Cell) {
			src.marks[placeMark{place: id, name: name}] = cell
		}, func(spawn Spawn) { src.spawns = append(src.spawns, spawn) })
		if err != nil {
			return nil, fmt.Errorf("place %q: %w", id, err)
		}
	}
	return src, nil
}

// Chunk implements [hexmap.Source].
func (s *WorldSource) Chunk(key hexmap.ChunkKey) (*hexmap.Chunk, error) {
	return filledChunk(key, s.fill, s.authored[key]), nil
}

// Bounds is the box of cells the world holds.
func (s *WorldSource) Bounds() hexmap.Bounds { return s.bounds }

// Edge is the terrain beyond the bounds.
func (s *WorldSource) Edge() hexmap.Terrain { return s.edge }

// Mark returns the world cell of the named mark of a place.
func (s *WorldSource) Mark(place api.ContentID, name string) (api.Cell, bool) {
	cell, ok := s.marks[placeMark{place: place, name: name}]
	return cell, ok
}

// Spawns lists every place's spawn marks: places in ID order, then layers as authored, rows and columns.
func (s *WorldSource) Spawns() []Spawn { return slices.Clone(s.spawns) }
