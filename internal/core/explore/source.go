package explore

import (
	"fmt"
	"slices"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
)

type placed struct {
	cell    api.Cell
	terrain hexmap.Terrain
}

// drawing is a place or a piece laid out in world cells, each list in the order the cells are drawn:
// layers as authored, then rows and columns.
type drawing struct {
	cells     []placed
	marks     map[string]api.Cell
	spawns    []Spawn
	openings  []api.Cell
	doors     []api.Cell
	landmarks []Landmark
}

// draw lays out a drawing, finding the world cell of each authored column, row and layer depth with cellOf.
// Every mark is reported once in its own list: a plain mark by name, the rest by their kind.
func draw(
	reg *content.Registry,
	layout content.Layout,
	open hexmap.Terrain,
	cellOf func(col, row, depth int) api.Cell,
) (drawing, error) {
	out := drawing{marks: map[string]api.Cell{}}
	for _, layer := range layout.Layers {
		for row, text := range layer.Rows {
			for col, char := range []rune(text) {
				if char == ' ' {
					continue
				}
				cell := cellOf(col, row, layer.Depth)
				t := open
				if id, ok := layout.Legend[string(char)]; ok {
					var err error
					if t, err = Terrain(reg, id); err != nil {
						return drawing{}, fmt.Errorf("legend %q: %w", string(char), err)
					}
				} else if name, ok := layout.Marks[string(char)]; ok {
					gives, isLandmark := layout.Affordances[string(char)]
					switch template, spawns := layout.Spawns[string(char)]; {
					case spawns:
						out.spawns = append(out.spawns, Spawn{Cell: cell, Template: template})
					case slices.Contains(layout.Openings, string(char)):
						out.openings = append(out.openings, cell)
					case slices.Contains(layout.Doors, string(char)):
						out.doors = append(out.doors, cell)
					case isLandmark:
						out.landmarks = append(out.landmarks, Landmark{Cell: cell, Gives: gives})
					default:
						out.marks[name] = cell
					}
				} else {
					return drawing{}, fmt.Errorf("character %q is in neither the legend nor the marks", string(char))
				}
				out.cells = append(out.cells, placed{cell: cell, terrain: t})
			}
		}
	}
	return out, nil
}

func filledChunk(key hexmap.ChunkKey, fill hexmap.Terrain, authored []placed) *hexmap.Chunk {
	chunk := hexmap.NewChunk(fill)
	origin := key.Origin()
	for _, p := range authored {
		chunk.Set(p.cell.Hex.Q-origin.Hex.Q, p.cell.Hex.R-origin.Hex.R, p.cell.Depth-origin.Depth, p.terrain)
	}
	return chunk
}

// Spawn is a body a place starts with: a template and the cell it stands in.
type Spawn struct {
	Cell     api.Cell
	Template api.ContentID
}

// Landmark is a marked cell of a place that gives something, with no body of its own, such as a
// graveyard's ground.
type Landmark struct {
	Cell api.Cell
	// Gives lists "affordance." IDs.
	Gives []api.ContentID
	// Place is the place or piece ID the mark is in.
	Place api.ContentID
}

// Terrain resolves a terrain content ID to the cell terrain it stands for.
func Terrain(reg *content.Registry, id api.ContentID) (hexmap.Terrain, error) {
	t, ok := reg.Terrain(id)
	if !ok {
		return hexmap.Terrain{}, fmt.Errorf("unknown terrain %q", id)
	}
	return hexmap.Terrain{ID: id, Solid: t.Solid, Opaque: t.Opaque}, nil
}
