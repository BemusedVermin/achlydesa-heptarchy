package content

import (
	"fmt"
	"slices"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// Behavior returns the behavior with the given ID. Its slices, maps and target are shared with the
// registry: don't change them.
func (r *Registry) Behavior(id api.ContentID) (BehaviorDef, bool) {
	def, ok := r.bundle.Behaviors[string(id)]
	return def, ok
}

// BehaviorIDs lists every behavior ID in sorted order.
func (r *Registry) BehaviorIDs() []api.ContentID { return sortedIDs(r.bundle.Behaviors) }

// Agents returns the numbers every agent shares.
func (r *Registry) Agents() AgentsConfig { return r.bundle.Agents }

// BehaviorScore is the product of the behavior's considerations at env.
func (r *Registry) BehaviorScore(id api.ContentID, env ConsiderEnv) (float64, error) {
	programs, ok := r.behaviors[id]
	if !ok {
		return 0, fmt.Errorf("content: unknown behavior %q", id)
	}
	score := 1.0
	for i, program := range programs {
		value, err := run[float64](program, env, fmt.Sprintf("behavior %q consider[%d]", id, i))
		if err != nil {
			return 0, err
		}
		score *= value
	}
	return score, nil
}

// checkBehaviors compiles every behavior's considerations, checks its target, and checks the
// behaviors templates list. That the primitive exists and takes the params is checked when a game
// starts.
func (r *Registry) checkBehaviors() []error {
	p := r.problems()
	cfg := r.bundle.Agents
	if cfg.KeepBonus < 0 {
		p.addf("agents.keep_bonus: must not be negative, got %v", cfg.KeepBonus)
	}
	if cfg.CellRadius < 0 {
		p.addf("agents.cell_radius: must not be negative, got %d", cfg.CellRadius)
	}
	for _, key := range sortedKeys(r.bundle.Behaviors) {
		def := r.bundle.Behaviors[key]
		path := fmt.Sprintf("behaviors[%q]", key)
		p.text(path+".name", def.Name)
		if def.Do == "" {
			p.addf("%s.do: required", path)
		}
		programs := make([]*vm.Program, 0, len(def.Consider))
		for i, src := range def.Consider {
			program, err := r.Compile(src, ConsiderEnv{}, expr.AsFloat64())
			if err != nil {
				p.addf("%s.consider[%d]: %w", path, i, err)
				continue
			}
			programs = append(programs, program)
		}
		r.behaviors[api.ContentID(key)] = programs
		if def.Target != nil {
			r.checkBehaviorTarget(p, path+".target", *def.Target)
		}
	}
	for _, key := range sortedKeys(r.bundle.Templates) {
		for i, id := range r.bundle.Templates[key].Behaviors {
			ref(p, r.bundle.Behaviors, fmt.Sprintf("templates[%q].behaviors[%d]", key, i), "behavior", id)
		}
	}
	return p.errs
}

func (r *Registry) checkBehaviorTarget(p *problems, path string, target TargetDef) {
	if kinds := behaviorTargetKinds(); !slices.Contains(kinds, target.Kind) {
		p.addf("%s.kind: want one of %q, got %q", path, kinds, target.Kind)
		return
	}
	misplaced := map[string]bool{
		"tags":    target.Kind != BehaviorTargetBody && len(target.Tags) > 0,
		"dead":    target.Kind != BehaviorTargetBody && target.Dead != nil,
		"hostile": target.Kind != BehaviorTargetBody && target.Hostile != nil,
		"mark":    target.Kind != BehaviorTargetMark && target.Mark != "",
		"terrain": target.Kind != BehaviorTargetCell && target.Terrain != "",
	}
	for _, field := range sortedKeys(misplaced) {
		if misplaced[field] {
			p.addf("%s.%s: a %s target has no %s", path, field, target.Kind, field)
		}
	}
	switch target.Kind {
	case BehaviorTargetBody:
	case BehaviorTargetMark:
		if target.Mark == "" {
			p.addf("%s.mark: required for a mark target", path)
		}
	case BehaviorTargetCell:
		ref(p, r.bundle.Terrains, path+".terrain", "terrain", target.Terrain)
	}
}
