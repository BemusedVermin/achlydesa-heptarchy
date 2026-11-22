package content

import (
	"fmt"
	"sort"
	"unicode/utf8"

	"github.com/expr-lang/expr"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// Map returns the map with the given ID.
func (r *Registry) Map(id api.ContentID) (MapDef, bool) {
	def, ok := r.bundle.Maps[string(id)]
	return def, ok
}

// MapIDs lists every map ID in sorted order.
func (r *Registry) MapIDs() []api.ContentID {
	ids := make([]api.ContentID, 0, len(r.bundle.Maps))
	for key := range r.bundle.Maps {
		ids = append(ids, api.ContentID(key))
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

// Start returns the start scenario. Its Map is empty when the content has none.
func (r *Registry) Start() Start { return r.bundle.Start }

// ClimbCost is the extra ticks a step up costs a body of the given speed.
func (r *Registry) ClimbCost(env SpeedEnv) (int64, error) {
	if r.climbCost == nil {
		return 0, nil
	}
	out, err := expr.Run(r.climbCost, env)
	if err != nil {
		return 0, fmt.Errorf("content: party.climb_cost: %w", err)
	}
	cost, ok := out.(int64)
	if !ok {
		return 0, fmt.Errorf("content: party.climb_cost returned %T, not an integer", out)
	}
	return cost, nil
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
	var problems []error
	path := fmt.Sprintf("maps[%q]", key)
	checkTerrain := func(field string, id api.ContentID) {
		if _, ok := r.bundle.Terrains[string(id)]; !ok {
			problems = append(problems, fmt.Errorf("%s.%s: unknown terrain %q", path, field, id))
		}
	}
	if _, ok := r.bundle.Text[string(def.Name)]; !ok {
		problems = append(problems, fmt.Errorf("%s.name: text reference %q does not exist", path, def.Name))
	}
	checkTerrain("fill", def.Fill)
	checkTerrain("open", def.Open)
	for _, char := range sortedKeys(def.Legend) {
		checkTerrain(fmt.Sprintf("legend[%q]", char), def.Legend[char])
		if utf8.RuneCountInString(char) != 1 {
			problems = append(problems, fmt.Errorf("%s.legend[%q]: a key is exactly one character", path, char))
		}
		if _, clash := def.Marks[char]; clash {
			problems = append(problems, fmt.Errorf("%s: character %q is both a legend entry and a mark", path, char))
		}
	}
	for _, char := range sortedKeys(def.Marks) {
		if utf8.RuneCountInString(char) != 1 {
			problems = append(problems, fmt.Errorf("%s.marks[%q]: a key is exactly one character", path, char))
		}
	}
	found := map[string]int{}
	depths := map[int]bool{}
	for i, layer := range def.Layers {
		at := fmt.Sprintf("%s.layers[%d]", path, i)
		if depths[layer.Depth] {
			problems = append(problems, fmt.Errorf("%s.depth: depth %d is authored twice", at, layer.Depth))
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
					problems = append(problems, fmt.Errorf(
						"%s.rows[%d]: character %q at column %d is not in the legend or the marks", at, row, char, col))
				}
			}
		}
	}
	for _, char := range sortedKeys(def.Marks) {
		if n := found[char]; n != 1 {
			problems = append(problems, fmt.Errorf(
				"%s.marks[%q]: mark %q must appear exactly once, found %d", path, char, def.Marks[char], n))
		}
	}
	return problems
}

// checkStart validates the start scenario when the content has one.
func (r *Registry) checkStart() []error {
	start := r.bundle.Start
	if start == (Start{}) {
		return nil
	}
	var problems []error
	def, ok := r.bundle.Maps[string(start.Map)]
	if !ok {
		problems = append(problems, fmt.Errorf("start.map: unknown map %q", start.Map))
	}
	if template, found := r.bundle.Templates[string(start.Party)]; !found || template.Party == nil {
		problems = append(problems, fmt.Errorf("start.party: %q is not a party template", start.Party))
	}
	if ok {
		if _, marked := def.MarkChars()[start.PartyAtMark]; !marked {
			problems = append(
				problems,
				fmt.Errorf("start.party_at_mark: map %q has no mark %q", start.Map, start.PartyAtMark),
			)
		}
	}
	return problems
}
