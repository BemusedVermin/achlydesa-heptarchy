package content

import (
	"fmt"
	"unicode/utf8"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// Map returns the map with the given ID.
func (r *Registry) Map(id api.ContentID) (MapDef, bool) {
	def, ok := r.bundle.Maps[string(id)]
	return def, ok
}

// MapIDs lists every map ID in sorted order.
func (r *Registry) MapIDs() []api.ContentID { return sortedIDs(r.bundle.Maps) }

// Start returns the start scenario. Its Map is empty when the content has none.
func (r *Registry) Start() Start { return r.bundle.Start }

// ClimbCost is the extra ticks a step up costs a body of the given speed.
func (r *Registry) ClimbCost(env SpeedEnv) (int64, error) {
	if r.climbCost == nil {
		return 0, nil
	}
	return run[int64](r.climbCost, env, "party.climb_cost")
}

// MarkChars returns the character of each mark name of the map.
func (d MapDef) MarkChars() map[string]string {
	out := make(map[string]string, len(d.Marks))
	for char, name := range d.Marks {
		out[name] = char
	}
	return out
}

// checkMaps validates every map: terrains exist, every character of every layer is in the legend or
// the marks, and each mark appears exactly once.
func (r *Registry) checkMaps() []error {
	var problems []error
	for _, key := range sortedKeys(r.bundle.Maps) {
		problems = append(problems, r.checkMap(key, r.bundle.Maps[key])...)
	}
	return problems
}

func (r *Registry) checkMap(key string, def MapDef) []error {
	p := r.problems()
	path := fmt.Sprintf("maps[%q]", key)
	checkTerrain := func(field string, id api.ContentID) {
		ref(p, r.bundle.Terrains, path+"."+field, "terrain", id)
	}
	p.text(path+".name", def.Name)
	checkTerrain("fill", def.Fill)
	checkTerrain("open", def.Open)
	r.checkLayout(p, path, def)
	return p.errs
}

// checkLayout checks the layers, legend, marks and spawns of a map or a place; Name, Fill and Open
// are the caller's.
func (r *Registry) checkLayout(p *problems, path string, def MapDef) {
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
	for _, char := range sortedKeys(def.Marks) {
		if utf8.RuneCountInString(char) != 1 {
			p.addf("%s.marks[%q]: a key is exactly one character", path, char)
		}
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
	for _, char := range sortedKeys(def.Marks) {
		if _, spawns := def.Spawns[char]; spawns {
			continue
		}
		if n := found[char]; n != 1 {
			p.addf("%s.marks[%q]: mark %q must appear exactly once, found %d", path, char, def.Marks[char], n)
		}
	}
}

// checkStart validates the start scenario when the content has one.
func (r *Registry) checkStart() []error {
	start := r.bundle.Start
	if start == (Start{}) {
		return nil
	}
	p := r.problems()
	def, ok := ref(p, r.bundle.Maps, "start.map", "map", start.Map)
	if template, found := r.bundle.Templates[string(start.Party)]; !found || template.Party == nil {
		p.addf("start.party: %q is not a party template", start.Party)
	}
	if ok {
		if _, marked := def.MarkChars()[start.PartyAtMark]; !marked {
			p.addf("start.party_at_mark: map %q has no mark %q", start.Map, start.PartyAtMark)
		}
	}
	return p.errs
}
