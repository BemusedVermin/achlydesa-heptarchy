package explore

import (
	"fmt"
	"slices"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
)

// Source is a [hexmap.Source] built from a hand-made [content.MapDef]: the authored layers, set in
// endless fill.
type Source struct {
	fill     hexmap.Terrain
	authored map[hexmap.ChunkKey][]placed
	marks    map[string]api.Cell
	spawns   []Spawn
	gates    []GateSite
}

// GateSite is a gate cell and the mark it leads to, before the maps are numbered.
type GateSite struct {
	Cell api.Cell
	To   api.ContentID
	Mark string
}

// Gates lists the map's gate cells in row order.
func (s *Source) Gates() []GateSite { return slices.Clone(s.gates) }

type placed struct {
	cell    api.Cell
	terrain hexmap.Terrain
}

// NewSource builds the source of a map definition. Marked cells hold the map's Open terrain.
func NewSource(reg *content.Registry, def content.MapDef) (*Source, error) {
	terrain := func(id api.ContentID) (hexmap.Terrain, error) { return Terrain(reg, id) }
	src := &Source{authored: map[hexmap.ChunkKey][]placed{}, marks: map[string]api.Cell{}}
	var err error
	if src.fill, err = terrain(def.Fill); err != nil {
		return nil, fmt.Errorf("fill: %w", err)
	}
	open, err := terrain(def.Open)
	if err != nil {
		return nil, fmt.Errorf("open: %w", err)
	}
	for _, layer := range def.Layers {
		for row, text := range layer.Rows {
			for col, char := range []rune(text) {
				if char == ' ' {
					continue
				}
				cell := api.Cell{Hex: hexmap.FromOffsetOddR(col, row), Depth: layer.Depth}
				t := open
				if id, ok := def.Legend[string(char)]; ok {
					if t, err = terrain(id); err != nil {
						return nil, fmt.Errorf("legend %q: %w", string(char), err)
					}
				} else if name, ok := def.Marks[string(char)]; ok {
					if gate, isGate := def.Gates[string(char)]; isGate {
						src.gates = append(src.gates, GateSite{Cell: cell, To: gate.To, Mark: gate.Mark})
					} else if template, spawns := def.Spawns[string(char)]; spawns {
						src.spawns = append(src.spawns, Spawn{Cell: cell, Template: template})
					} else {
						src.marks[name] = cell
					}
				} else {
					return nil, fmt.Errorf("character %q is in neither the legend nor the marks", string(char))
				}
				key := hexmap.KeyOf(cell)
				src.authored[key] = append(src.authored[key], placed{cell: cell, terrain: t})
			}
		}
	}
	return src, nil
}

// Chunk implements [hexmap.Source].
func (s *Source) Chunk(key hexmap.ChunkKey) (*hexmap.Chunk, error) {
	chunk := hexmap.NewChunk(s.fill)
	origin := key.Origin()
	for _, p := range s.authored[key] {
		chunk.Set(p.cell.Hex.Q-origin.Hex.Q, p.cell.Hex.R-origin.Hex.R, p.cell.Depth-origin.Depth, p.terrain)
	}
	return chunk, nil
}

// Mark returns the cell of the named mark.
func (s *Source) Mark(name string) (api.Cell, bool) {
	cell, ok := s.marks[name]
	return cell, ok
}

// Spawn is a body the map starts with: a template and the cell it stands in.
type Spawn struct {
	Cell     api.Cell
	Template api.ContentID
}

// Spawns lists the map's spawn marks in row order: layers as authored, then rows, then columns.
func (s *Source) Spawns() []Spawn { return slices.Clone(s.spawns) }

// Terrain resolves a terrain content ID to the map cell it stands for.
func Terrain(reg *content.Registry, id api.ContentID) (hexmap.Terrain, error) {
	t, ok := reg.Terrain(id)
	if !ok {
		return hexmap.Terrain{}, fmt.Errorf("unknown terrain %q", id)
	}
	return hexmap.Terrain{ID: id, Solid: t.Solid, Opaque: t.Opaque}, nil
}
