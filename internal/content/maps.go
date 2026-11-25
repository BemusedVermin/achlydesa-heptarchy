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
			p.addf("%s.spawns[%q]: not a mark of the map", path, char)
		}
		ref(p, r.bundle.Templates, fmt.Sprintf("%s.spawns[%q]", path, char), "template", template)
		if found[char] < 1 {
			p.addf("%s.spawns[%q]: the character is in no layer", path, char)
		}
	}
	r.checkGates(p, path, def, found)
	r.checkTown(p, path, def)
	for _, char := range sortedKeys(def.Marks) {
		_, spawns := def.Spawns[char]
		_, gate := def.Gates[char]
		if spawns || gate {
			continue
		}
		if n := found[char]; n != 1 {
			p.addf("%s.marks[%q]: mark %q must appear exactly once, found %d", path, char, def.Marks[char], n)
		}
	}
	return p.errs
}

// checkGates validates a map's gates: each is a mark used in a layer, and leads to a plain mark of a
// map that exists.
func (r *Registry) checkGates(p *problems, path string, def MapDef, found map[string]int) {
	for _, char := range sortedKeys(def.Gates) {
		at := fmt.Sprintf("%s.gates[%q]", path, char)
		gate := def.Gates[char]
		if _, ok := def.Marks[char]; !ok {
			p.addf("%s: not a mark of the map", at)
		}
		if _, spawns := def.Spawns[char]; spawns {
			p.addf("%s: a character is a gate or a spawn, not both", at)
		}
		if found[char] < 1 {
			p.addf("%s: the character is in no layer", at)
		}
		dest, ok := ref(p, r.bundle.Maps, at+".to", "map", gate.To)
		if !ok {
			continue
		}
		destChar, marked := dest.MarkChars()[gate.Mark]
		_, destSpawns := dest.Spawns[destChar]
		_, destGate := dest.Gates[destChar]
		if !marked || destSpawns || destGate {
			p.addf("%s.mark: map %q has no plain mark %q", at, gate.To, gate.Mark)
		}
	}
}

// checkTown validates a town's stock and its stock marks, which must be plain marks of the map.
func (r *Registry) checkTown(p *problems, path string, def MapDef) {
	if def.Region == nil || def.Region.Town == nil {
		return
	}
	town := def.Region.Town
	at := path + ".region.town"
	for i, stack := range town.Stock {
		ref(p, r.bundle.Items, fmt.Sprintf("%s.stock[%d].item", at, i), "item", stack.Item)
		if stack.Count < 1 {
			p.addf("%s.stock[%d].count: %d is not at least 1", at, i, stack.Count)
		}
	}
	r.checkTownStock(p, at, town)
	chars := def.MarkChars()
	for i, mark := range town.StockMarks {
		char, marked := chars[mark]
		_, spawns := def.Spawns[char]
		_, gate := def.Gates[char]
		if !marked || spawns || gate {
			p.addf("%s.stock_marks[%d]: %q is not a plain mark of the map", at, i, mark)
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

// checkTownStock validates what a town's residents use, what refills the stock and how many leave.
func (r *Registry) checkTownStock(p *problems, at string, town *TownDef) {
	if town.Residents != "" {
		ref(p, r.bundle.Families, at+".residents", "family", town.Residents)
	}
	for _, list := range []struct {
		field  string
		stacks []StackDef
	}{{"use", town.Use}, {"refill", town.Refill}} {
		for i, stack := range list.stacks {
			ref(p, r.bundle.Items, fmt.Sprintf("%s.%s[%d].item", at, list.field, i), "item", stack.Item)
			if stack.Count < 1 {
				p.addf("%s.%s[%d].count: %d is not at least 1", at, list.field, i, stack.Count)
			}
		}
	}
	if len(town.Use) > 0 && town.Residents == "" {
		p.addf("%s.use: a town that uses stock needs residents", at)
	}
	if !levelInRange(town.Exodus) {
		p.addf("%s.exodus: must be from 0 to 1, got %v", at, town.Exodus)
	}
	r.checkGenerics(p, at, town)
}

// checkGenerics validates a town's generics: known templates, each the only template of one family.
func (r *Registry) checkGenerics(p *problems, at string, town *TownDef) {
	if len(town.Generics) > 0 && town.Residents == "" {
		p.addf("%s.generics: a town that keeps generics needs residents", at)
	}
	seen := map[api.ContentID]bool{}
	for i, generic := range town.Generics {
		path := fmt.Sprintf("%s.generics[%d]", at, i)
		ref(p, r.bundle.Templates, path+".template", "template", generic.Template)
		if seen[generic.Template] {
			p.addf("%s.template: %q is listed twice", path, generic.Template)
		}
		seen[generic.Template] = true
		if generic.Count < 1 {
			p.addf("%s.count: %d is not at least 1", path, generic.Count)
		}
		if generic.Ticks < 1 {
			p.addf("%s.ticks: %d is not at least 1", path, generic.Ticks)
		}
		owners := 0
		for _, family := range r.bundle.Families {
			if len(family.Templates) == 1 && family.Templates[0].Template == generic.Template {
				owners++
			}
		}
		if owners != 1 {
			p.addf("%s.template: %q must be the only template of exactly one family, found %d",
				path, generic.Template, owners)
		}
	}
}
