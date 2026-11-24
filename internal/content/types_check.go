package content

import (
	"fmt"
	"slices"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// Type returns the type with the given ID.
func (r *Registry) Type(id api.ContentID) (Type, bool) {
	def, ok := r.bundle.Types[string(id)]
	return def, ok
}

// TypeIDs lists every type ID in sorted order.
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
func (r *Registry) Warnings() []string { return slices.Clone(r.warnings) }

func sortedIDs[V any](m map[string]V) []api.ContentID {
	keys := sortedKeys(m)
	ids := make([]api.ContentID, len(keys))
	for i, key := range keys {
		ids[i] = api.ContentID(key)
	}
	return ids
}

// checkTypes checks materials, types, items, factions and template inventories, compiles type
// conditions, and records one-sided hostility as warnings.
func (r *Registry) checkTypes() []error {
	p := r.problems()
	for _, key := range sortedKeys(r.bundle.Materials) {
		p.text(fmt.Sprintf("materials[%q].name", key), r.bundle.Materials[key].Name)
	}
	for _, key := range sortedKeys(r.bundle.Types) {
		def := r.bundle.Types[key]
		path := fmt.Sprintf("types[%q]", key)
		p.text(path+".name", def.Name)
		p.add(r.compileMods(path, api.ContentID(key), def.Mods)...)
	}
	for _, key := range sortedKeys(r.bundle.Items) {
		def := r.bundle.Items[key]
		path := fmt.Sprintf("items[%q]", key)
		p.text(path+".name", def.Name)
		if def.Part != nil {
			ref(p, r.bundle.Parts, path+".part", "part", *def.Part)
		}
		p.add(r.compileMods(path, api.ContentID(key), def.Mods)...)
	}
	for _, key := range sortedKeys(r.bundle.Statuses) {
		p.text(fmt.Sprintf("statuses[%q].name", key), r.bundle.Statuses[key].Name)
	}
	for _, key := range sortedKeys(r.bundle.Factions) {
		def := r.bundle.Factions[key]
		path := fmt.Sprintf("factions[%q]", key)
		p.text(path+".name", def.Name)
		for i, other := range def.HostileTo {
			target, ok := ref(p, r.bundle.Factions, fmt.Sprintf("%s.hostile_to[%d]", path, i), "faction", other)
			if ok && !slices.Contains(target.HostileTo, api.ContentID(key)) {
				r.warnings = append(
					r.warnings,
					fmt.Sprintf("%s is hostile to %q, which is not hostile back", path, other),
				)
			}
		}
	}
	for _, key := range sortedKeys(r.bundle.Templates) {
		if faction := r.bundle.Templates[key].Faction; faction != nil {
			ref(p, r.bundle.Factions, fmt.Sprintf("templates[%q].faction", key), "faction", *faction)
		}
		for i, stack := range r.bundle.Templates[key].Inventory {
			path := fmt.Sprintf("templates[%q].inventory[%d]", key, i)
			ref(p, r.bundle.Items, path+".item", "item", stack.Item)
			if stack.Count < 1 {
				p.addf("%s.count: must be at least 1, got %d", path, stack.Count)
			}
		}
	}
	return p.errs
}

// checkCombat checks the numbers of alert and death.
func (r *Registry) checkCombat() []error {
	cfg := r.bundle.Combat
	p := r.problems()
	if cfg.CalmTicks < 0 {
		p.addf("combat.calm_ticks: must not be negative, got %d", cfg.CalmTicks)
	}
	if cfg.SightRadius < 0 || cfg.SightDepth < 0 {
		p.addf("combat: sight_radius and sight_depth must not be negative, got %d and %d",
			cfg.SightRadius, cfg.SightDepth)
	}
	if program, err := r.Compile(cfg.AIScore, ScoreEnv{}, expr.AsFloat64()); err != nil {
		p.addf("combat.ai_score: %w", err)
	} else {
		r.aiScore = program
	}
	if program, err := r.Compile(cfg.Aims, GrantEnv{}, expr.AsBool()); err != nil {
		p.addf("combat.aims: %w", err)
	} else {
		r.aims = program
	}
	if cfg.AttackTag == "" {
		p.addf("combat.attack_tag: required")
	}
	if cfg.CorpseTag == "" {
		p.addf("combat.corpse_tag: required")
	}
	if step, ok := ref(p, r.bundle.Actions, "combat.step_action", "action", cfg.StepAction); ok &&
		step.Target != TargetDirection {
		p.addf("combat.step_action: %q must target a direction", cfg.StepAction)
	}
	if cfg.MaterialClaim != "" {
		ref(p, r.bundle.Claims, "combat.material_claim", "claim", cfg.MaterialClaim)
	}
	return p.errs
}

// compiledMods holds a record's mod formulas, compiled against ActionEnv. Nil means none.
type compiledMods struct {
	requires, windup, recovery *vm.Program
}

// compileMods compiles the formulas of a record's mods and checks its refusal text.
func (r *Registry) compileMods(path string, id api.ContentID, mods Mods) []error {
	p := r.problems()
	compile := func(field, src string, option expr.Option) *vm.Program {
		if src == "" {
			return nil
		}
		program, err := r.Compile(src, ActionEnv{}, option)
		if err != nil {
			p.addf("%s.%s: %w", path, field, err)
		}
		return program
	}
	compiled := compiledMods{
		requires: compile("requires", mods.Requires, expr.AsBool()),
		windup:   compile("windup", mods.Windup, expr.AsInt64()),
		recovery: compile("recovery", mods.Recovery, expr.AsInt64()),
	}
	if mods.Refusal != "" {
		p.text(path+".refusal", mods.Refusal)
	}
	r.mods[id] = compiled
	return p.errs
}

// refusalText is the reason a mod refuses an action when it names none.
var refusalText api.ContentID = "text.actions.requires"

// Mods returns what the type or item changes about an action it is chosen for.
func (r *Registry) Mods(id api.ContentID) (Mods, bool) {
	if def, ok := r.bundle.Types[string(id)]; ok {
		return def.Mods, true
	}
	if def, ok := r.bundle.Items[string(id)]; ok {
		return def.Mods, true
	}
	return Mods{}, false
}

// ModsRequires evaluates the record's requires formula. When it is false, refusal is the text ID of
// the reason. A record with no formula always passes.
func (r *Registry) ModsRequires(id api.ContentID, env ActionEnv) (ok bool, refusal api.ContentID, err error) {
	program := r.mods[id].requires
	if program == nil {
		return true, "", nil
	}
	holds, err := run[bool](program, env, fmt.Sprintf("%q mods requires", id))
	if err != nil {
		return false, "", err
	}
	if holds {
		return true, "", nil
	}
	mods, _ := r.Mods(id)
	if mods.Refusal == "" {
		return false, refusalText, nil
	}
	return false, mods.Refusal, nil
}

// ModsTicks evaluates the ticks the record adds to an action's wind-up and recovery.
func (r *Registry) ModsTicks(id api.ContentID, env ActionEnv) (windup, recovery int64, err error) {
	compiled := r.mods[id]
	for _, field := range []struct {
		name    string
		program *vm.Program
		into    *int64
	}{{"windup", compiled.windup, &windup}, {"recovery", compiled.recovery, &recovery}} {
		if field.program == nil {
			continue
		}
		ticks, err := run[int64](field.program, env, fmt.Sprintf("%q mods %s", id, field.name))
		if err != nil {
			return 0, 0, err
		}
		*field.into = ticks
	}
	return windup, recovery, nil
}
