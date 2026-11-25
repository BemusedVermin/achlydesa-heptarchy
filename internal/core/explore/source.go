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
}

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
	err = layOut(reg, def, content.PlaceAt{}, open, src.authored, func(name string, cell api.Cell) {
		src.marks[name] = cell
	}, func(spawn Spawn) { src.spawns = append(src.spawns, spawn) })
	if err != nil {
		return nil, err
	}
	return src, nil
}

// layOut adds the authored cells of def's layers, offset by at, to authored, and reports each mark
// and spawn in row order.
func layOut(
	reg *content.Registry,
	def content.MapDef,
	at content.PlaceAt,
	open hexmap.Terrain,
	authored map[hexmap.ChunkKey][]placed,
	mark func(name string, cell api.Cell),
	spawn func(Spawn),
) error {
	for _, layer := range def.Layers {
		for row, text := range layer.Rows {
			for col, char := range []rune(text) {
				if char == ' ' {
					continue
				}
				cell := api.Cell{Hex: hexmap.FromOffsetOddR(at.Col+col, at.Row+row), Depth: at.Depth + layer.Depth}
				t := open
				if id, ok := def.Legend[string(char)]; ok {
					var err error
					if t, err = Terrain(reg, id); err != nil {
						return fmt.Errorf("legend %q: %w", string(char), err)
					}
				} else if name, ok := def.Marks[string(char)]; ok {
					if template, spawns := def.Spawns[string(char)]; spawns {
						spawn(Spawn{Cell: cell, Template: template})
					} else {
						mark(name, cell)
					}
				} else {
					return fmt.Errorf("character %q is in neither the legend nor the marks", string(char))
				}
				key := hexmap.KeyOf(cell)
				authored[key] = append(authored[key], placed{cell: cell, terrain: t})
			}
		}
	}
	return nil
}

// Chunk implements [hexmap.Source].
func (s *Source) Chunk(key hexmap.ChunkKey) (*hexmap.Chunk, error) {
	return filledChunk(key, s.fill, s.authored[key]), nil
}

func filledChunk(key hexmap.ChunkKey, fill hexmap.Terrain, authored []placed) *hexmap.Chunk {
	chunk := hexmap.NewChunk(fill)
	origin := key.Origin()
	for _, p := range authored {
		chunk.Set(p.cell.Hex.Q-origin.Hex.Q, p.cell.Hex.R-origin.Hex.R, p.cell.Depth-origin.Depth, p.terrain)
	}
	return chunk
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
