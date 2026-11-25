package content

import (
	"fmt"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// World returns the world's size and terrains. It is the zero WorldDef when the content has no world.
func (r *Registry) World() WorldDef { return r.bundle.World }

// Place returns the place with the given ID.
func (r *Registry) Place(id api.ContentID) (PlaceDef, bool) {
	def, ok := r.bundle.Places[string(id)]
	return def, ok
}

// PlaceIDs lists every place ID in sorted order.
func (r *Registry) PlaceIDs() []api.ContentID { return sortedIDs(r.bundle.Places) }

func (d PlaceDef) layout() MapDef {
	return MapDef{Name: d.Name, Layers: d.Layers, Legend: d.Legend, Marks: d.Marks, Spawns: d.Spawns}
}

type worldCell struct{ col, row, depth int }

func (c worldCell) String() string { return fmt.Sprintf("column %d, row %d, depth %d", c.col, c.row, c.depth) }

func (w WorldDef) contains(c worldCell) bool {
	return c.col >= 0 && c.col < w.Cols && c.row >= 0 && c.row < w.Rows &&
		c.depth >= w.MinDepth && c.depth <= w.MaxDepth
}

// checkWorld validates the world and its places: each place's layout follows a map's rules, sits on
// an even row, lies inside the world's bounds and shares no cell with another place.
func (r *Registry) checkWorld() []error {
	world := r.bundle.World
	if world == (WorldDef{}) && len(r.bundle.Places) == 0 {
		return nil
	}
	p := r.problems()
	if world.Cols < 1 || world.Rows < 1 {
		p.addf("world: cols and rows must be at least 1, got %d and %d", world.Cols, world.Rows)
	}
	if world.MinDepth > world.MaxDepth {
		p.addf("world: min_depth %d is deeper than max_depth %d", world.MinDepth, world.MaxDepth)
	}
	ref(p, r.bundle.Terrains, "world.fill", "terrain", world.Fill)
	ref(p, r.bundle.Terrains, "world.open", "terrain", world.Open)
	if edge, ok := ref(p, r.bundle.Terrains, "world.edge", "terrain", world.Edge); ok && (!edge.Solid || !edge.Opaque) {
		p.addf("world.edge: terrain %q must be solid and opaque", world.Edge)
	}
	owners := map[worldCell]string{}
	for _, key := range sortedKeys(r.bundle.Places) {
		def := r.bundle.Places[key]
		path := fmt.Sprintf("places[%q]", key)
		p.text(path+".name", def.Name)
		if def.At.Row%2 != 0 {
			p.addf("%s.at.row: %d is odd; a place must start on an even row", path, def.At.Row)
		}
		r.checkLayout(p, path, def.layout())
		outside := false
		overlaps := map[string]bool{}
		for _, layer := range def.Layers {
			for row, text := range layer.Rows {
				for col, char := range []rune(text) {
					if char == ' ' {
						continue
					}
					cell := worldCell{col: def.At.Col + col, row: def.At.Row + row, depth: def.At.Depth + layer.Depth}
					if !world.contains(cell) {
						if !outside {
							p.addf("%s: cell at %v is outside the world", path, cell)
						}
						outside = true
						continue
					}
					if other, taken := owners[cell]; taken {
						if !overlaps[other] {
							p.addf("%s: cell at %v is also authored by place %q", path, cell, other)
						}
						overlaps[other] = true
						continue
					}
					owners[cell] = key
				}
			}
		}
	}
	return p.errs
}
