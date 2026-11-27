package content

import (
	"fmt"
	"slices"
	"strings"
	"text/template"

	"github.com/expr-lang/expr/vm"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// Registry is the immutable, validated content. Build one with [Load] or [New]. It is safe for concurrent use.
type Registry struct {
	bundle Bundle
	hash   string
	// moveCosts is keyed by terrain ID.
	moveCosts map[api.ContentID]*vm.Program
	// decays is keyed by geography ID; a geography without decay has no entry.
	decays    map[api.ContentID]*vm.Program
	stepTicks *vm.Program
	// climbCost is nil when party.climb_cost is empty.
	climbCost *vm.Program
	// templates is parsed text, keyed by text ID.
	templates map[api.ContentID]*template.Template
	actions   map[api.ContentID]compiledAction
	speed     *vm.Program
	aiScore   *vm.Program
	aims      *vm.Program
	records   map[api.ContentID]Record
	// offers has no entry for a choice without an offer formula.
	offers map[string]*vm.Program
	mods   map[api.ContentID]compiledMods
	drives map[api.ContentID]compiledDrive
	// weight is nil when agents.weight is empty.
	weight      *vm.Program
	warnings    []string
	collections []collection
	// rulesUsed is filled in while the registry loads, to warn about rules nothing uses.
	rulesUsed map[string]bool
}

// Hash is a hex SHA-256 of the whole bundle; equal content gives an equal hash.
func (r *Registry) Hash() string { return r.hash }

// TemplateName is the text that names a body of the template: "text.template.<ID without its
// prefix>", or "text.template.unknown" when the template ID is empty.
func (r *Registry) TemplateName(template api.ContentID) api.Text {
	if template == "" {
		return api.Text{ID: "text.template.unknown"}
	}
	return api.Text{ID: api.ContentID("text.template." + strings.TrimPrefix(string(template), "template."))}
}

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

// Origin returns the fact source with the given ID.
func (r *Registry) Origin(id api.ContentID) (Origin, bool) {
	source, ok := r.bundle.Origins[string(id)]
	return source, ok
}

// Terrain returns the terrain with the given ID.
func (r *Registry) Terrain(id api.ContentID) (Terrain, bool) {
	terrain, ok := r.bundle.Terrains[string(id)]
	return terrain, ok
}

// TerrainIDs lists every terrain ID in sorted order.
func (r *Registry) TerrainIDs() []api.ContentID { return sortedIDs(r.bundle.Terrains) }

// MoveCost is the ticks it costs to walk on top of the terrain. A terrain with no formula costs 0.
func (r *Registry) MoveCost(terrain api.ContentID, env MoveEnv) (int64, error) {
	if _, ok := r.bundle.Terrains[string(terrain)]; !ok {
		return 0, fmt.Errorf("content: unknown terrain %q", terrain)
	}
	program, ok := r.moveCosts[terrain]
	if !ok {
		return 0, nil
	}
	return run[int64](program, env, fmt.Sprintf("terrain %q move_cost", terrain))
}

// DecayLevel is how decayed the geography's generated ground is for the env, by its decay formula. A
// geography without decay is always at level 0.
func (r *Registry) DecayLevel(geography api.ContentID, env DecayEnv) (int64, error) {
	program, ok := r.decays[geography]
	if !ok {
		return 0, nil
	}
	return run[int64](program, env, fmt.Sprintf("geography %q decay level", geography))
}

// Calendar returns the calendar. Phases is a copy.
func (r *Registry) Calendar() Calendar {
	cal := r.bundle.Calendar
	cal.Phases = slices.Clone(cal.Phases)
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

// ClimbCost is the extra ticks a step up costs a body of the given speed.
func (r *Registry) ClimbCost(env SpeedEnv) (int64, error) {
	if r.climbCost == nil {
		return 0, nil
	}
	return run[int64](r.climbCost, env, "party.climb_cost")
}

// Knowledge returns the numbers of every knower's memory.
func (r *Registry) Knowledge() KnowledgeConfig { return r.bundle.Knowledge }

// Agents returns the numbers and claims of agents' minds.
func (r *Registry) Agents() AgentsConfig { return r.bundle.Agents }

// Combat returns the numbers of alert and death.
func (r *Registry) Combat() CombatConfig { return r.bundle.Combat }

// Timeline returns the numbers of the shared timeline.
func (r *Registry) Timeline() TimelineConfig { return r.bundle.Timeline }

// StepTicks is the base ticks a step keeps its channels busy for an entity with env's speed.
func (r *Registry) StepTicks(env SpeedEnv) (int64, error) {
	return run[int64](r.stepTicks, env, "timeline.step_ticks")
}

// Speed is a body's speed, from the totals of its functioning parts' grants.
func (r *Registry) Speed(env GrantEnv) (int, error) {
	speed, err := run[int64](r.speed, env, "timeline.speed")
	return int(speed), err
}

// EffectSite is a place in content that holds effects, named by its path for error messages.
type EffectSite struct {
	Path string      `json:"path"`
	Refs []EffectRef `json:"refs"`
}

// AIScore rates an intent for a hunting creature, by combat.ai_score.
func (r *Registry) AIScore(env ScoreEnv) (float64, error) {
	return run[float64](r.aiScore, env, "combat.ai_score")
}

// Aims says whether a body with the grant totals can pick its target, by combat.aims.
func (r *Registry) Aims(env GrantEnv) (bool, error) {
	return run[bool](r.aims, env, "combat.aims")
}
