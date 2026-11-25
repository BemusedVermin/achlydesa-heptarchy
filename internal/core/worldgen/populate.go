package worldgen

import (
	"fmt"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
)

// Spawn is a body a chunk starts with: a template and the cell it stands in.
type Spawn struct {
	Cell     api.Cell
	Template api.ContentID
}

// ExportCount is the export count the source generates for.
func (s *Source) ExportCount() uint64 { return s.exportCount }

// Populate lists the bodies the chunk starts with, in depth, R, Q order. A cell holds one at most, and only
// an open cell with solid ground below, outside every fixed place.
//
// Each such cell draws once. Of the geography's spawn entries whose depth band holds the cell, an entry with
// weight w and density d takes the share w*d/W of the draw, where W is the sum of those entries' weights;
// what the entries leave of it is no body. It is a pure function of the seed, export count and chunk key.
func (s *Source) Populate(key hexmap.ChunkKey) ([]Spawn, error) {
	if len(s.spawns) == 0 || !s.touchesBounds(key) {
		return nil, nil
	}
	ground := &groundReader{source: s, chunks: map[hexmap.ChunkKey]*hexmap.Chunk{}}
	gen := s.streams.Derived("procgen", s.salt(key, len(s.stages), PopulateName))
	origin := key.Origin()
	var spawns []Spawn
	for depth := range hexmap.ChunkSize {
		for r := range hexmap.ChunkSize {
			for q := range hexmap.ChunkSize {
				cell := api.Cell{Hex: api.Hex{Q: origin.Hex.Q + q, R: origin.Hex.R + r}, Depth: origin.Depth + depth}
				if !s.bounds.Contains(cell) || s.fixed.Covers(cell) {
					continue
				}
				entries, total := s.spawnsAt(cell.Depth)
				if len(entries) == 0 {
					continue
				}
				standable, err := ground.standable(cell)
				if err != nil {
					return nil, err
				}
				if !standable {
					continue
				}
				pick := gen.Float64() * total
				for _, entry := range entries {
					share := entry.Weight * entry.Density
					if pick < share {
						spawns = append(spawns, Spawn{Cell: cell, Template: entry.Template})
						break
					}
					pick -= share
				}
			}
		}
	}
	return spawns, nil
}

// PopulateName names the stream a chunk's spawns are drawn from.
const PopulateName = "populate"

// spawnsAt lists the entries that apply at the depth and the sum of their weights.
func (s *Source) spawnsAt(depth int) ([]content.GeographySpawn, float64) {
	var entries []content.GeographySpawn
	var total float64
	for _, entry := range s.spawns {
		if depth >= entry.MinDepth && depth <= entry.MaxDepth {
			entries = append(entries, entry)
			total += entry.Weight
		}
	}
	return entries, total
}

type groundReader struct {
	source *Source
	chunks map[hexmap.ChunkKey]*hexmap.Chunk
}

func (g *groundReader) terrain(cell api.Cell) (hexmap.Terrain, error) {
	key := hexmap.KeyOf(cell)
	chunk, ok := g.chunks[key]
	if !ok {
		var err error
		if chunk, err = g.source.Chunk(key); err != nil {
			return hexmap.Terrain{}, fmt.Errorf("worldgen: populate: %w", err)
		}
		g.chunks[key] = chunk
	}
	origin := key.Origin()
	return chunk.Get(cell.Hex.Q-origin.Hex.Q, cell.Hex.R-origin.Hex.R, cell.Depth-origin.Depth), nil
}

// standable reports whether the cell is open and the cell below is solid, as the map defines it; ground
// below the bounds counts as solid, like the edge.
func (g *groundReader) standable(cell api.Cell) (bool, error) {
	here, err := g.terrain(cell)
	if err != nil || here.Solid {
		return false, err
	}
	below := hexmap.Prisms{}.Below(cell).Cell
	if !g.source.bounds.Contains(below) {
		return true, nil
	}
	floor, err := g.terrain(below)
	return floor.Solid, err
}
