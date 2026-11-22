package content

import (
	"fmt"
	"sort"
	"text/template"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// Registry is the immutable, validated content. Build one with [Load]. It is safe for concurrent use.
type Registry struct {
	bundle    Bundle
	hash      string
	moveCosts map[api.ContentID]*vm.Program        // compiled move_cost formulas, by terrain
	stepTicks *vm.Program                          // compiled timeline.step_ticks formula
	climbCost *vm.Program                          // compiled party.climb_cost formula, nil when empty
	templates map[api.ContentID]*template.Template // parsed text, by text ID
	actions   map[api.ContentID]compiledAction     // compiled action formulas
	types     map[api.ContentID]*vm.Program        // compiled type conditions; absent when a type has none
	typeTicks map[api.ContentID]compiledTypeTicks  // compiled type wind-up and recovery formulas
	warnings  []string
}

// Hash is a hex SHA-256 of the whole bundle; equal content gives an equal hash.
func (r *Registry) Hash() string { return r.hash }

// Material returns the material with the given ID.
func (r *Registry) Material(id api.ContentID) (Material, bool) {
	material, ok := r.bundle.Materials[string(id)]
	return material, ok
}

// Claim returns the claim with the given ID.
func (r *Registry) Claim(id api.ContentID) (Claim, bool) {
	claim, ok := r.bundle.Claims[string(id)]
	return claim, ok
}

// RenderValue turns a fact's value of the given claim into player-facing text, by the claim's value
// kind. A literal, an unknown claim or an unknown ID is returned as it is.
func (r *Registry) RenderValue(claim api.ContentID, value string) string {
	switch r.bundle.Claims[string(claim)].Value {
	case ValueMaterial:
		if material, ok := r.bundle.Materials[value]; ok {
			return r.Render(api.Text{ID: material.Name})
		}
	case ValueText:
		if _, ok := r.templates[api.ContentID(value)]; ok {
			return r.Render(api.Text{ID: api.ContentID(value)})
		}
	}
	return value
}

// Source returns the fact source with the given ID.
func (r *Registry) Source(id api.ContentID) (Source, bool) {
	source, ok := r.bundle.Sources[string(id)]
	return source, ok
}

// Terrain returns the terrain with the given ID.
func (r *Registry) Terrain(id api.ContentID) (Terrain, bool) {
	terrain, ok := r.bundle.Terrains[string(id)]
	return terrain, ok
}

// TerrainIDs lists every terrain ID in sorted order.
func (r *Registry) TerrainIDs() []api.ContentID {
	ids := make([]api.ContentID, 0, len(r.bundle.Terrains))
	for key := range r.bundle.Terrains {
		ids = append(ids, api.ContentID(key))
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

// MoveCost is the ticks it costs to walk on top of the terrain. A terrain with no formula costs 0.
func (r *Registry) MoveCost(terrain api.ContentID, env MoveEnv) (int64, error) {
	if _, ok := r.bundle.Terrains[string(terrain)]; !ok {
		return 0, fmt.Errorf("content: unknown terrain %q", terrain)
	}
	program, ok := r.moveCosts[terrain]
	if !ok {
		return 0, nil
	}
	out, err := expr.Run(program, env)
	if err != nil {
		return 0, fmt.Errorf("content: terrain %q move_cost: %w", terrain, err)
	}
	cost, ok := out.(int64)
	if !ok {
		return 0, fmt.Errorf("content: terrain %q move_cost returned %T, not an integer", terrain, out)
	}
	return cost, nil
}

// Calendar returns the calendar. Phases is a copy.
func (r *Registry) Calendar() Calendar {
	cal := r.bundle.Calendar
	cal.Phases = append([]Phase(nil), cal.Phases...)
	return cal
}

// Template returns the template with the given ID. Its pointer fields are shared with the registry: don't change them.
func (r *Registry) Template(id api.ContentID) (Template, bool) {
	template, ok := r.bundle.Templates[string(id)]
	return template, ok
}

// Part returns the part template with the given ID. Its Grants slice is shared with the registry:
// don't change it.
func (r *Registry) Part(id api.ContentID) (PartTemplate, bool) {
	part, ok := r.bundle.Parts[string(id)]
	return part, ok
}

// Body returns the body plan with the given ID. Its Parts slice is shared with the registry: don't
// change it.
func (r *Registry) Body(id api.ContentID) (BodyPlan, bool) {
	plan, ok := r.bundle.Bodies[string(id)]
	return plan, ok
}

// Party returns the numbers every party shares.
func (r *Registry) Party() PartyConfig { return r.bundle.Party }

// Timeline returns the numbers of the shared timeline.
func (r *Registry) Timeline() TimelineConfig { return r.bundle.Timeline }

// StepTicks is the base ticks a step keeps its channels busy for an entity with env's speed.
func (r *Registry) StepTicks(env SpeedEnv) (int64, error) {
	out, err := expr.Run(r.stepTicks, env)
	if err != nil {
		return 0, fmt.Errorf("content: timeline.step_ticks: %w", err)
	}
	ticks, ok := out.(int64)
	if !ok {
		return 0, fmt.Errorf("content: timeline.step_ticks returned %T, not an integer", out)
	}
	return ticks, nil
}

// EffectSite is a place in content that holds effects, named by its path for error messages.
type EffectSite struct {
	Path string      `json:"path"`
	Refs []EffectRef `json:"refs"`
}
