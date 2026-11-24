package content

import (
	"fmt"
	"regexp"
	"slices"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// keyPattern is the shape of an action key and of a choice provider name: lower-case words.
var keyPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

// tagPattern is the shape of a channel tag: lower-case words, optionally joined by colons.
var tagPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*(:[a-z][a-z0-9_]*)*$`)

// compiledAction holds an action's formulas, compiled against ActionEnv.
type compiledAction struct {
	requires, connects *vm.Program // nil when the action has none
	windup, recovery   *vm.Program
}

// Action returns the action with the given ID.
func (r *Registry) Action(id api.ContentID) (ActionDef, bool) {
	def, ok := r.bundle.Actions[string(id)]
	return def, ok
}

// ActionIDs lists every action ID in sorted order.
func (r *Registry) ActionIDs() []api.ContentID { return sortedIDs(r.bundle.Actions) }

func (r *Registry) actionTicks(
	id api.ContentID, name string, pick func(compiledAction) *vm.Program, env ActionEnv,
) (int64, error) {
	compiled, ok := r.actions[id]
	if !ok {
		return 0, fmt.Errorf("content: unknown action %q", id)
	}
	return run[int64](pick(compiled), env, fmt.Sprintf("action %q %s", id, name))
}

// ActionWindup is the ticks the action takes to land. 0 lands at once.
func (r *Registry) ActionWindup(id api.ContentID, env ActionEnv) (int64, error) {
	return r.actionTicks(id, "windup", func(c compiledAction) *vm.Program { return c.windup }, env)
}

// ActionRecovery is the ticks the action's channels stay busy after it lands.
func (r *Registry) ActionRecovery(id api.ContentID, env ActionEnv) (int64, error) {
	return r.actionTicks(id, "recovery", func(c compiledAction) *vm.Program { return c.recovery }, env)
}

func (r *Registry) actionBool(
	id api.ContentID, name string, pick func(compiledAction) *vm.Program, env ActionEnv,
) (value, present bool, err error) {
	compiled, ok := r.actions[id]
	if !ok {
		return false, false, fmt.Errorf("content: unknown action %q", id)
	}
	program := pick(compiled)
	if program == nil {
		return false, false, nil
	}
	value, err = run[bool](program, env, fmt.Sprintf("action %q %s", id, name))
	return value, true, err
}

// ActionRequires evaluates the action's requires formula. An action with none always passes.
func (r *Registry) ActionRequires(id api.ContentID, env ActionEnv) (bool, error) {
	value, present, err := r.actionBool(id, "requires", func(c compiledAction) *vm.Program { return c.requires }, env)
	if err != nil {
		return false, err
	}
	return value || !present, nil
}

// ActionConnects evaluates the action's connects formula. present is false when the action has none,
// meaning the default applies: the target is still in range (and sight, if required).
func (r *Registry) ActionConnects(id api.ContentID, env ActionEnv) (connects, present bool, err error) {
	return r.actionBool(id, "connects", func(c compiledAction) *vm.Program { return c.connects }, env)
}

// checkActions validates every action and compiles its formulas.
func (r *Registry) checkActions() []error {
	p := r.problems()
	keys := map[string]string{}
	for _, id := range sortedKeys(r.bundle.Actions) {
		def := r.bundle.Actions[id]
		path := fmt.Sprintf("actions[%q]", id)
		p.text(path+".name", def.Name)
		other, dup := keys[def.Key]
		switch {
		case !keyPattern.MatchString(def.Key):
			p.addf("%s.key: %q must be lower-case words", path, def.Key)
		case dup:
			p.addf("%s.key: %q is already the key of %s", path, def.Key, other)
		default:
			keys[def.Key] = id
		}
		p.add(checkActionShape(path, def)...)
		compiled, errs := r.compileAction(path, def)
		p.add(errs...)
		if len(errs) == 0 {
			r.actions[api.ContentID(id)] = compiled
		}
	}
	return p.errs
}

func checkActionShape(path string, def ActionDef) []error {
	var problems []error
	if !slices.Contains(targetKinds(), def.Target) {
		problems = append(problems, fmt.Errorf("%s.target: %q is not one of %v", path, def.Target, targetKinds()))
	}
	if def.Range < 0 {
		problems = append(problems, fmt.Errorf("%s.range: must not be negative, got %d", path, def.Range))
	}
	if len(def.Uses) == 0 {
		problems = append(problems, fmt.Errorf("%s.uses: an action must occupy at least one channel", path))
	}
	for i, use := range def.Uses {
		if !tagPattern.MatchString(use.Tag) {
			problems = append(problems, fmt.Errorf("%s.uses[%d].tag: %q is not a channel tag", path, i, use.Tag))
		}
		if use.Count < 0 {
			problems = append(
				problems,
				fmt.Errorf("%s.uses[%d].count: must not be negative, got %d", path, i, use.Count),
			)
		}
	}
	seen := map[string]bool{}
	for i, choice := range def.Choices {
		switch {
		case !keyPattern.MatchString(choice):
			problems = append(problems, fmt.Errorf("%s.choices[%d]: %q must be lower-case words", path, i, choice))
		case seen[choice]:
			problems = append(problems, fmt.Errorf("%s.choices[%d]: %q is listed twice", path, i, choice))
		}
		seen[choice] = true
	}
	return problems
}

func (r *Registry) compileAction(path string, def ActionDef) (compiledAction, []error) {
	var problems []error
	compile := func(field, src string, required bool, option expr.Option) *vm.Program {
		if src == "" {
			if required {
				problems = append(problems, fmt.Errorf("%s.%s: required", path, field))
			}
			return nil
		}
		program, err := r.Compile(src, ActionEnv{}, option)
		if err != nil {
			problems = append(problems, fmt.Errorf("%s.%s: %w", path, field, err))
			return nil
		}
		return program
	}
	compiled := compiledAction{
		requires: compile("requires", def.Requires, false, expr.AsBool()),
		connects: compile("connects", def.Connects, false, expr.AsBool()),
		windup:   compile("windup", def.Windup, true, expr.AsInt64()),
		recovery: compile("recovery", def.Recovery, true, expr.AsInt64()),
	}
	return compiled, problems
}

// checkTemplateActions reports template action lists that name actions that do not exist.
func (r *Registry) checkTemplateActions() []error {
	p := r.problems()
	for _, key := range sortedKeys(r.bundle.Templates) {
		for i, id := range r.bundle.Templates[key].Actions {
			ref(p, r.bundle.Actions, fmt.Sprintf("templates[%q].actions[%d]", key, i), "action", id)
		}
	}
	return p.errs
}

// EffectSites lists every place in content that holds effects, so game.New can check them against
// the registered primitives.
func (r *Registry) EffectSites() []EffectSite {
	var sites []EffectSite
	for _, id := range sortedKeys(r.bundle.Actions) {
		def := r.bundle.Actions[id]
		if len(def.OnStart) > 0 {
			sites = append(sites, EffectSite{Path: fmt.Sprintf("actions[%q].on_start", id), Refs: def.OnStart})
		}
		if len(def.OnLand) > 0 {
			sites = append(sites, EffectSite{Path: fmt.Sprintf("actions[%q].on_land", id), Refs: def.OnLand})
		}
	}
	for _, id := range sortedKeys(r.bundle.Types) {
		sites = append(sites, modSites(fmt.Sprintf("types[%q]", id), r.bundle.Types[id].Mods)...)
	}
	for _, id := range sortedKeys(r.bundle.Items) {
		sites = append(sites, modSites(fmt.Sprintf("items[%q]", id), r.bundle.Items[id].Mods)...)
	}
	for _, site := range r.TriggerSites() {
		if len(site.Trigger.Do) > 0 {
			sites = append(sites, EffectSite{Path: site.Path + ".do", Refs: site.Trigger.Do})
		}
	}
	return sites
}

func modSites(path string, mods Mods) []EffectSite {
	var sites []EffectSite
	if len(mods.OnStart) > 0 {
		sites = append(sites, EffectSite{Path: path + ".on_start", Refs: mods.OnStart})
	}
	if len(mods.OnLand) > 0 {
		sites = append(sites, EffectSite{Path: path + ".on_land", Refs: mods.OnLand})
	}
	return sites
}
