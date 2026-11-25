package content

import (
	"fmt"
	"math"

	"github.com/expr-lang/expr"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// RecruitWorth is what a body of the hirable template brings, by its worth formula.
func (r *Registry) RecruitWorth(template api.ContentID, env BodyEnv) (float64, error) {
	program, ok := r.worths[template]
	if !ok {
		return 0, fmt.Errorf("content: template %q is not hirable", template)
	}
	return run[float64](program, env, fmt.Sprintf("template %q recruit.worth", template))
}

// RecruitPrice scales each want's count by worth and rounds up, so a price is never below the base
// count's share.
func RecruitPrice(wants []StackDef, worth float64) []StackDef {
	price := make([]StackDef, 0, len(wants))
	for _, want := range wants {
		price = append(price, StackDef{Item: want.Item, Count: int(math.Ceil(float64(want.Count) * worth))})
	}
	return price
}

// checkRecruits checks the items a hirable template wants and compiles its worth formula.
func (r *Registry) checkRecruits() []error {
	p := r.problems()
	for _, key := range sortedKeys(r.bundle.Templates) {
		def := r.bundle.Templates[key].Recruit
		if def == nil {
			continue
		}
		path := fmt.Sprintf("templates[%q].recruit", key)
		if len(def.Wants) == 0 {
			p.addf("%s.wants: a recruit wants at least one item", path)
		}
		for _, list := range []struct {
			field  string
			stacks []StackDef
		}{{"wants", def.Wants}, {"upkeep", def.Upkeep}} {
			for i, stack := range list.stacks {
				at := fmt.Sprintf("%s.%s[%d]", path, list.field, i)
				ref(p, r.bundle.Items, at+".item", "item", stack.Item)
				if stack.Count < 1 {
					p.addf("%s.count: must be at least 1, got %d", at, stack.Count)
				}
			}
		}
		program, err := r.Compile(def.Worth, BodyEnv{}, expr.AsFloat64())
		if err != nil {
			p.addf("%s.worth: %w", path, err)
			continue
		}
		r.worths[api.ContentID(key)] = program
	}
	return p.errs
}
