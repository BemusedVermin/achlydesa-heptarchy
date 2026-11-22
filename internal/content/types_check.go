package content

import (
	"fmt"
	"slices"

	"github.com/expr-lang/expr"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// Type returns the damage type with the given ID.
func (r *Registry) Type(id api.ContentID) (DamageType, bool) {
	def, ok := r.bundle.Types[string(id)]
	return def, ok
}

// TypeIDs lists every damage type ID in sorted order.
func (r *Registry) TypeIDs() []api.ContentID { return sortedIDs(r.bundle.Types) }

// Item returns the item with the given ID.
func (r *Registry) Item(id api.ContentID) (Item, bool) {
	def, ok := r.bundle.Items[string(id)]
	return def, ok
}

// Status returns the status with the given ID.
func (r *Registry) Status(id api.ContentID) (Status, bool) {
	def, ok := r.bundle.Statuses[string(id)]
	return def, ok
}

// Faction returns the faction with the given ID. Its HostileTo slice is shared with the registry:
// don't change it.
func (r *Registry) Faction(id api.ContentID) (Faction, bool) {
	def, ok := r.bundle.Factions[string(id)]
	return def, ok
}

// Warnings lists problems that don't make the content invalid, such as one-sided hostility.
func (r *Registry) Warnings() []string { return append([]string(nil), r.warnings...) }

// TypeAllowed runs the type's condition against env. A type with no condition is always allowed.
func (r *Registry) TypeAllowed(id api.ContentID, env TypeEnv) (bool, error) {
	if _, ok := r.bundle.Types[string(id)]; !ok {
		return false, fmt.Errorf("content: unknown type %q", id)
	}
	program, ok := r.types[id]
	if !ok {
		return true, nil
	}
	out, err := expr.Run(program, env)
	if err != nil {
		return false, fmt.Errorf("content: type %q condition: %w", id, err)
	}
	allowed, ok := out.(bool)
	if !ok {
		return false, fmt.Errorf("content: type %q condition returned %T, not a bool", id, out)
	}
	return allowed, nil
}

func sortedIDs[V any](m map[string]V) []api.ContentID {
	keys := sortedKeys(m)
	ids := make([]api.ContentID, len(keys))
	for i, key := range keys {
		ids[i] = api.ContentID(key)
	}
	return ids
}

// checkTypes checks materials, damage types, items, factions and template packs, compiles type
// conditions, and records one-sided hostility as warnings.
func (r *Registry) checkTypes() []error {
	var problems []error
	text := func(path string, ref api.ContentID) {
		if _, ok := r.bundle.Text[string(ref)]; !ok {
			problems = append(problems, fmt.Errorf("%s: text reference %q does not exist", path, ref))
		}
	}
	for _, key := range sortedKeys(r.bundle.Materials) {
		material := r.bundle.Materials[key]
		lists := []struct {
			field string
			ids   []api.ContentID
		}{{"weak_to", material.WeakTo}, {"immune_to", material.ImmuneTo}}
		for _, list := range lists {
			for i, id := range list.ids {
				if _, ok := r.bundle.Types[string(id)]; !ok {
					problems = append(
						problems,
						fmt.Errorf("materials[%q].%s[%d]: unknown type %q", key, list.field, i, id),
					)
				}
			}
		}
		for _, id := range material.WeakTo {
			if slices.Contains(material.ImmuneTo, id) {
				problems = append(problems, fmt.Errorf("materials[%q]: %q is both a weakness and an immunity", key, id))
			}
		}
	}
	for _, key := range sortedKeys(r.bundle.Types) {
		def := r.bundle.Types[key]
		path := fmt.Sprintf("types[%q]", key)
		text(path+".name", def.Name)
		if def.Consumes != nil {
			if _, ok := r.bundle.Items[string(*def.Consumes)]; !ok {
				problems = append(problems, fmt.Errorf("%s.consumes: unknown item %q", path, *def.Consumes))
			}
		}
		if def.Condition == "" {
			continue
		}
		program, err := expr.Compile(def.Condition, expr.Env(TypeEnv{}), expr.AsBool())
		if err != nil {
			problems = append(problems, fmt.Errorf("%s.condition: %w", path, err))
			continue
		}
		r.types[api.ContentID(key)] = program
	}
	for _, key := range sortedKeys(r.bundle.Items) {
		def := r.bundle.Items[key]
		path := fmt.Sprintf("items[%q]", key)
		text(path+".name", def.Name)
		if def.Part != nil {
			if def.Kind != "piece" {
				problems = append(
					problems,
					fmt.Errorf("%s.part: only a piece becomes a part, not a %s", path, def.Kind),
				)
			}
			if _, ok := r.bundle.Parts[string(*def.Part)]; !ok {
				problems = append(problems, fmt.Errorf("%s.part: unknown part %q", path, *def.Part))
			}
		}
		if !slices.Contains(ItemKinds, def.Kind) {
			problems = append(
				problems,
				fmt.Errorf("%s.kind: unknown kind %q (want one of %v)", path, def.Kind, ItemKinds),
			)
		}
	}
	for _, key := range sortedKeys(r.bundle.Statuses) {
		text(fmt.Sprintf("statuses[%q].name", key), r.bundle.Statuses[key].Name)
	}
	for _, key := range sortedKeys(r.bundle.Factions) {
		def := r.bundle.Factions[key]
		path := fmt.Sprintf("factions[%q]", key)
		text(path+".name", def.Name)
		for i, other := range def.HostileTo {
			target, ok := r.bundle.Factions[string(other)]
			switch {
			case !ok:
				problems = append(problems, fmt.Errorf("%s.hostile_to[%d]: unknown faction %q", path, i, other))
			case !slices.Contains(target.HostileTo, api.ContentID(key)):
				r.warnings = append(
					r.warnings,
					fmt.Sprintf("%s is hostile to %q, which is not hostile back", path, other),
				)
			}
		}
	}
	for _, key := range sortedKeys(r.bundle.Templates) {
		for i, stack := range r.bundle.Templates[key].Inventory {
			path := fmt.Sprintf("templates[%q].inventory[%d]", key, i)
			if _, ok := r.bundle.Items[string(stack.Item)]; !ok {
				problems = append(problems, fmt.Errorf("%s.item: unknown item %q", path, stack.Item))
			}
			if stack.Count < 1 {
				problems = append(problems, fmt.Errorf("%s.count: must be at least 1, got %d", path, stack.Count))
			}
		}
	}
	return problems
}
