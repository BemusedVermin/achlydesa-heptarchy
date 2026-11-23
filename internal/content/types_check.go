package content

import (
	"errors"
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
func (r *Registry) Warnings() []string { return append([]string(nil), r.warnings...) }

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
	var problems []error
	text := func(path string, ref api.ContentID) {
		if _, ok := r.bundle.Text[string(ref)]; !ok {
			problems = append(problems, fmt.Errorf("%s: text reference %q does not exist", path, ref))
		}
	}
	for _, key := range sortedKeys(r.bundle.Materials) {
		text(fmt.Sprintf("materials[%q].name", key), r.bundle.Materials[key].Name)
	}
	for _, key := range sortedKeys(r.bundle.Types) {
		def := r.bundle.Types[key]
		path := fmt.Sprintf("types[%q]", key)
		text(path+".name", def.Name)
		problems = append(problems, r.compileMods(path, api.ContentID(key), def.Mods)...)
	}
	for _, key := range sortedKeys(r.bundle.Items) {
		def := r.bundle.Items[key]
		path := fmt.Sprintf("items[%q]", key)
		text(path+".name", def.Name)
		if def.Part != nil {
			if _, ok := r.bundle.Parts[string(*def.Part)]; !ok {
				problems = append(problems, fmt.Errorf("%s.part: unknown part %q", path, *def.Part))
			}
		}
		problems = append(problems, r.compileMods(path, api.ContentID(key), def.Mods)...)
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
		if faction := r.bundle.Templates[key].Faction; faction != nil {
			if _, ok := r.bundle.Factions[string(*faction)]; !ok {
				problems = append(problems, fmt.Errorf("templates[%q].faction: unknown faction %q", key, *faction))
			}
		}
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

// checkCombat checks the numbers of alert and death.
func (r *Registry) checkCombat() []error {
	cfg := r.bundle.Combat
	var problems []error
	if cfg.CalmTicks < 0 {
		problems = append(problems, fmt.Errorf("combat.calm_ticks: must not be negative, got %d", cfg.CalmTicks))
	}
	if cfg.SightRadius < 0 || cfg.SightDepth < 0 {
		problems = append(problems, fmt.Errorf(
			"combat: sight_radius and sight_depth must not be negative, got %d and %d",
			cfg.SightRadius,
			cfg.SightDepth,
		))
	}
	if program, err := r.Compile(cfg.AIScore, ScoreEnv{}, expr.AsFloat64()); err != nil {
		problems = append(problems, fmt.Errorf("combat.ai_score: %w", err))
	} else {
		r.aiScore = program
	}
	if program, err := r.Compile(cfg.Aims, GrantEnv{}, expr.AsBool()); err != nil {
		problems = append(problems, fmt.Errorf("combat.aims: %w", err))
	} else {
		r.aims = program
	}
	if cfg.AttackTag == "" {
		problems = append(problems, errors.New("combat.attack_tag: required"))
	}
	if step, ok := r.bundle.Actions[string(cfg.StepAction)]; !ok {
		problems = append(problems, fmt.Errorf("combat.step_action: unknown action %q", cfg.StepAction))
	} else if step.Target != TargetDirection {
		problems = append(problems, fmt.Errorf("combat.step_action: %q must target a direction", cfg.StepAction))
	}
	if _, ok := r.bundle.Claims[string(cfg.MaterialClaim)]; cfg.MaterialClaim != "" && !ok {
		problems = append(problems, fmt.Errorf("combat.material_claim: unknown claim %q", cfg.MaterialClaim))
	}
	return problems
}

// compiledMods holds a record's mod formulas, compiled against ActionEnv. Nil means none.
type compiledMods struct {
	requires, windup, recovery *vm.Program
}

// compileMods compiles the formulas of a record's mods and checks its refusal text.
func (r *Registry) compileMods(path string, id api.ContentID, mods Mods) []error {
	var problems []error
	compile := func(field, src string, option expr.Option) *vm.Program {
		if src == "" {
			return nil
		}
		program, err := r.Compile(src, ActionEnv{}, option)
		if err != nil {
			problems = append(problems, fmt.Errorf("%s.%s: %w", path, field, err))
		}
		return program
	}
	compiled := compiledMods{
		requires: compile("requires", mods.Requires, expr.AsBool()),
		windup:   compile("windup", mods.Windup, expr.AsInt64()),
		recovery: compile("recovery", mods.Recovery, expr.AsInt64()),
	}
	if mods.Refusal != "" {
		if _, ok := r.bundle.Text[string(mods.Refusal)]; !ok {
			problems = append(problems, fmt.Errorf("%s.refusal: text reference %q does not exist", path, mods.Refusal))
		}
	}
	r.mods[id] = compiled
	return problems
}

// refusalText is the reason a mod refuses an action when it names none.
var refusalText api.ContentID = "text.action.requires"

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

func (r *Registry) runMod(id api.ContentID, name string, program *vm.Program, env ActionEnv) (any, error) {
	out, err := expr.Run(program, env)
	if err != nil {
		return nil, fmt.Errorf("content: %q mods %s: %w", id, name, err)
	}
	return out, nil
}

// ModsRequires evaluates the record's requires formula. When it is false, refusal is the text ID of
// the reason. A record with no formula always passes.
func (r *Registry) ModsRequires(id api.ContentID, env ActionEnv) (ok bool, refusal api.ContentID, err error) {
	program := r.mods[id].requires
	if program == nil {
		return true, "", nil
	}
	out, err := r.runMod(id, "requires", program, env)
	if err != nil {
		return false, "", err
	}
	holds, isBool := out.(bool)
	if !isBool {
		return false, "", fmt.Errorf("content: %q mods requires returned %T, not a boolean", id, out)
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
		out, err := r.runMod(id, field.name, field.program, env)
		if err != nil {
			return 0, 0, err
		}
		ticks, ok := out.(int64)
		if !ok {
			return 0, 0, fmt.Errorf("content: %q mods %s returned %T, not an integer", id, field.name, out)
		}
		*field.into = ticks
	}
	return windup, recovery, nil
}
