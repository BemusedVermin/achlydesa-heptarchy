package content

import (
	"fmt"
	"slices"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// Affordance returns the affordance with the given ID.
func (r *Registry) Affordance(id api.ContentID) (Affordance, bool) {
	def, ok := r.bundle.Affordances[string(id)]
	return def, ok
}

// checkAffordances checks that every affordance is named, and that templates and places give only
// known affordances, each once.
func (r *Registry) checkAffordances() []error {
	p := r.problems()
	for _, key := range sortedKeys(r.bundle.Affordances) {
		p.text(fmt.Sprintf("affordances[%q].name", key), r.bundle.Affordances[key].Name)
	}
	for _, key := range sortedKeys(r.bundle.Templates) {
		r.checkGives(p, fmt.Sprintf("templates[%q].affordances", key), r.bundle.Templates[key].Affordances)
	}
	for _, key := range sortedKeys(r.bundle.Places) {
		def := r.bundle.Places[key]
		for _, char := range sortedKeys(def.Affordances) {
			path := fmt.Sprintf("places[%q].affordances[%q]", key, char)
			if len(def.Affordances[char]) == 0 {
				p.addf("%s: at least one affordance is required", path)
			}
			r.checkGives(p, path, def.Affordances[char])
		}
	}
	return p.errs
}

func (r *Registry) checkGives(p *problems, path string, ids []api.ContentID) {
	for i, id := range ids {
		ref(p, r.bundle.Affordances, fmt.Sprintf("%s[%d]", path, i), "affordance", id)
		if slices.Index(ids, id) < i {
			p.addf("%s[%d]: %q is listed twice", path, i, id)
		}
	}
}
