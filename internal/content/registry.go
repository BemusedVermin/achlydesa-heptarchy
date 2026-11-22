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
	turnTicks *vm.Program                          // compiled turns.turn_ticks formula
	templates map[api.ContentID]*template.Template // parsed text, by text ID
}

// Hash is a hex SHA-256 of the whole bundle; equal content gives an equal hash.
func (r *Registry) Hash() string { return r.hash }

// Material returns the material with the given ID.
func (r *Registry) Material(id api.ContentID) (Material, bool) {
	material, ok := r.bundle.Materials[string(id)]
	return material, ok
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

// Party returns the numbers every party shares.
func (r *Registry) Party() PartyConfig { return r.bundle.Party }

// Turns returns the numbers of the shared timeline.
func (r *Registry) Turns() TurnsConfig { return r.bundle.Turns }

// TurnTicks is the ticks from one turn to the next for an entity with env's speed, before extra costs.
func (r *Registry) TurnTicks(env TurnEnv) (int64, error) {
	out, err := expr.Run(r.turnTicks, env)
	if err != nil {
		return 0, fmt.Errorf("content: turns.turn_ticks: %w", err)
	}
	ticks, ok := out.(int64)
	if !ok {
		return 0, fmt.Errorf("content: turns.turn_ticks returned %T, not an integer", out)
	}
	return ticks, nil
}
