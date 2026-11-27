package content

import (
	"fmt"
	"maps"
	"slices"
	"unicode/utf8"

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

// Piece returns the piece with the given ID.
func (r *Registry) Piece(id api.ContentID) (PieceDef, bool) {
	def, ok := r.bundle.Pieces[string(id)]
	return def, ok
}

// PieceIDs lists every piece ID in sorted order.
func (r *Registry) PieceIDs() []api.ContentID { return sortedIDs(r.bundle.Pieces) }

// Settlement returns the settlement with the given ID.
func (r *Registry) Settlement(id api.ContentID) (SettlementDef, bool) {
	def, ok := r.bundle.Settlements[string(id)]
	return def, ok
}

// SettlementIDs lists every settlement ID in sorted order.
func (r *Registry) SettlementIDs() []api.ContentID { return sortedIDs(r.bundle.Settlements) }

// Start returns the start scenario. Its Place is empty when the content has none.
func (r *Registry) Start() Start { return r.bundle.Start }

type worldCell struct{ col, row, depth int }

func (c worldCell) String() string {
	return fmt.Sprintf("column %d, row %d, depth %d", c.col, c.row, c.depth)
}

func (w WorldDef) contains(c worldCell) bool {
	return c.col >= 0 && c.col < w.Cols && c.row >= 0 && c.row < w.Rows &&
		c.depth >= w.MinDepth && c.depth <= w.MaxDepth
}

// checkWorld validates the world and its places: each place's layout is sound, it sits on
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
		r.checkLayout(p, path, def.Layout())
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

// checkLayout checks the layers, legend and marks of a place or a piece.
func (r *Registry) checkLayout(p *problems, path string, def Layout) {
	checkTerrain := func(field string, id api.ContentID) {
		ref(p, r.bundle.Terrains, path+"."+field, "terrain", id)
	}
	for _, char := range sortedKeys(def.Legend) {
		checkTerrain(fmt.Sprintf("legend[%q]", char), def.Legend[char])
		if utf8.RuneCountInString(char) != 1 {
			p.addf("%s.legend[%q]: a key is exactly one character", path, char)
		}
		if _, clash := def.Marks[char]; clash {
			p.addf("%s: character %q is both a legend entry and a mark", path, char)
		}
	}
	names := map[string]string{}
	for _, char := range sortedKeys(def.Marks) {
		if utf8.RuneCountInString(char) != 1 {
			p.addf("%s.marks[%q]: a key is exactly one character", path, char)
		}
		if other, taken := names[def.Marks[char]]; taken {
			p.addf("%s.marks[%q]: mark %q is also named by %q", path, char, def.Marks[char], other)
		}
		names[def.Marks[char]] = char
	}
	found := map[string]int{}
	depths := map[int]bool{}
	for i, layer := range def.Layers {
		at := fmt.Sprintf("%s.layers[%d]", path, i)
		if depths[layer.Depth] {
			p.addf("%s.depth: depth %d is authored twice", at, layer.Depth)
		}
		depths[layer.Depth] = true
		for row, text := range layer.Rows {
			for col, char := range []rune(text) {
				_, inLegend := def.Legend[string(char)]
				_, isMark := def.Marks[string(char)]
				switch {
				case char == ' ' || inLegend:
				case isMark:
					found[string(char)]++
				default:
					p.addf(
						"%s.rows[%d]: character %q at column %d is not in the legend or the marks",
						at,
						row,
						char,
						col,
					)
				}
			}
		}
	}
	for _, char := range sortedKeys(def.Spawns) {
		template := def.Spawns[char]
		if _, ok := def.Marks[char]; !ok {
			p.addf("%s.spawns[%q]: not one of the marks", path, char)
		}
		ref(p, r.bundle.Templates, fmt.Sprintf("%s.spawns[%q]", path, char), "template", template)
		if found[char] < 1 {
			p.addf("%s.spawns[%q]: the character is in no layer", path, char)
		}
	}
	exits := func(field string, chars []string) {
		for _, char := range chars {
			if _, ok := def.Marks[char]; !ok {
				p.addf("%s.%s[%q]: not one of the marks", path, field, char)
			}
			if _, spawns := def.Spawns[char]; spawns {
				p.addf("%s.%s[%q]: also a spawn", path, field, char)
			}
			if found[char] < 1 {
				p.addf("%s.%s[%q]: the character is in no layer", path, field, char)
			}
		}
	}
	exits("openings", def.Openings)
	exits("doors", def.Doors)
	repeats := func(char string) bool {
		return slices.Contains(def.Openings, char) || slices.Contains(def.Doors, char)
	}
	for _, char := range sortedKeys(def.Affordances) {
		if _, ok := def.Marks[char]; !ok {
			p.addf("%s.affordances[%q]: not one of the marks", path, char)
		}
		if _, spawns := def.Spawns[char]; spawns || repeats(char) {
			p.addf("%s.affordances[%q]: also a spawn, an opening or a door", path, char)
		}
		if found[char] < 1 {
			p.addf("%s.affordances[%q]: the character is in no layer", path, char)
		}
	}
	for _, char := range sortedKeys(def.Marks) {
		_, spawns := def.Spawns[char]
		_, gives := def.Affordances[char]
		if spawns || gives || repeats(char) {
			continue
		}
		if n := found[char]; n != 1 {
			p.addf("%s.marks[%q]: mark %q must appear exactly once, found %d", path, char, def.Marks[char], n)
		}
	}
}

// checkPieces validates every piece's layout; each has at least one door, so it can be joined to a town.
func (r *Registry) checkPieces() []error {
	p := r.problems()
	for _, key := range sortedKeys(r.bundle.Pieces) {
		def := r.bundle.Pieces[key]
		path := fmt.Sprintf("pieces[%q]", key)
		p.text(path+".name", def.Name)
		if len(def.Doors) == 0 {
			p.addf("%s.doors: a piece needs at least one door", path)
		}
		r.checkLayout(p, path, def.Layout())
	}
	return p.errs
}

// checkStart validates the start scenario when the content has one.
func (r *Registry) checkStart() []error {
	start := r.bundle.Start
	if start == (Start{}) {
		return nil
	}
	p := r.problems()
	def, ok := ref(p, r.bundle.Places, "start.place", "place", start.Place)
	if template, found := r.bundle.Templates[string(start.Party)]; !found || template.Party == nil {
		p.addf("start.party: %q is not a party template", start.Party)
	}
	if ok && !slices.Contains(slices.Collect(maps.Values(def.Marks)), start.Mark) {
		p.addf("start.mark: place %q has no mark %q", start.Place, start.Mark)
	}
	return p.errs
}
