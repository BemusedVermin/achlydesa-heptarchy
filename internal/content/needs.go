package content

import (
	"fmt"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// compiledNeed holds a need's formulas, compiled against NeedEnv.
type compiledNeed struct {
	rise, urge *vm.Program
}

// Need returns the need with the given ID. Its Thresholds slice is shared with the registry: don't
// change it.
func (r *Registry) Need(id api.ContentID) (NeedDef, bool) {
	def, ok := r.bundle.Needs[string(id)]
	return def, ok
}

// NeedIDs lists every need ID in sorted order.
func (r *Registry) NeedIDs() []api.ContentID { return sortedIDs(r.bundle.Needs) }

// NeedRise is how much the need's level rises this phase, by its rise formula.
func (r *Registry) NeedRise(id api.ContentID, env NeedEnv) (float64, error) {
	compiled, ok := r.needs[id]
	if !ok {
		return 0, fmt.Errorf("content: unknown need %q", id)
	}
	return run[float64](compiled.rise, env, fmt.Sprintf("need %q rise", id))
}

// NeedUrge is how pressing the need is at env's level, by its urge formula.
func (r *Registry) NeedUrge(id api.ContentID, env NeedEnv) (float64, error) {
	compiled, ok := r.needs[id]
	if !ok {
		return 0, fmt.Errorf("content: unknown need %q", id)
	}
	return run[float64](compiled.urge, env, fmt.Sprintf("need %q urge", id))
}

// checkNeeds compiles every need's formulas, checks its thresholds, and checks the needs templates
// start with. Threshold effects are checked with the other effect sites when a game starts.
func (r *Registry) checkNeeds() []error {
	p := r.problems()
	for _, key := range sortedKeys(r.bundle.Needs) {
		def := r.bundle.Needs[key]
		path := fmt.Sprintf("needs[%q]", key)
		p.text(path+".name", def.Name)
		var compiled compiledNeed
		for _, formula := range []struct {
			field, src string
			into       **vm.Program
		}{{"rise", def.Rise, &compiled.rise}, {"urge", def.Urge, &compiled.urge}} {
			program, err := r.Compile(formula.src, NeedEnv{}, expr.AsFloat64())
			if err != nil {
				p.addf("%s.%s: %w", path, formula.field, err)
				continue
			}
			*formula.into = program
		}
		r.needs[api.ContentID(key)] = compiled
		for i, threshold := range def.Thresholds {
			at := fmt.Sprintf("%s.thresholds[%d].at", path, i)
			if !levelInRange(threshold.At) {
				p.addf("%s: must be from 0 to 1, got %v", at, threshold.At)
			}
			if i > 0 && threshold.At <= def.Thresholds[i-1].At {
				p.addf("%s: thresholds must be in ascending at, got %v after %v", at, threshold.At,
					def.Thresholds[i-1].At)
			}
		}
	}
	for _, key := range sortedKeys(r.bundle.Templates) {
		levels := r.bundle.Templates[key].Needs
		for _, need := range sortedKeys(levels) {
			path := fmt.Sprintf("templates[%q].needs[%q]", key, need)
			ref(p, r.bundle.Needs, path, "need", api.ContentID(need))
			if !levelInRange(levels[need]) {
				p.addf("%s: a starting level must be from 0 to 1, got %v", path, levels[need])
			}
		}
	}
	return p.errs
}

func levelInRange(level float64) bool { return level >= 0 && level <= 1 }
