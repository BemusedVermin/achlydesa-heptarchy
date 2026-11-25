package content

import (
	"fmt"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// Family returns the family with the given ID. Its slices are shared with the registry: don't
// change them.
func (r *Registry) Family(id api.ContentID) (FamilyDef, bool) {
	def, ok := r.bundle.Families[string(id)]
	return def, ok
}

// FamilyIDs lists every family ID in sorted order.
func (r *Registry) FamilyIDs() []api.ContentID { return sortedIDs(r.bundle.Families) }

// checkFamilies checks each family's name, templates and weights, and the populations maps start
// with.
func (r *Registry) checkFamilies() []error {
	p := r.problems()
	for _, key := range sortedKeys(r.bundle.Families) {
		def := r.bundle.Families[key]
		path := fmt.Sprintf("families[%q]", key)
		p.text(path+".name", def.Name)
		if len(def.Templates) == 0 {
			p.addf("%s.templates: at least one is required", path)
		}
		for i, weighted := range def.Templates {
			at := fmt.Sprintf("%s.templates[%d]", path, i)
			ref(p, r.bundle.Templates, at+".template", "template", weighted.Template)
			if weighted.Weight < 1 {
				p.addf("%s.weight: must be at least 1, got %d", at, weighted.Weight)
			}
		}
	}
	for _, key := range sortedKeys(r.bundle.Maps) {
		region := r.bundle.Maps[key].Region
		if region == nil {
			continue
		}
		for _, family := range sortedKeys(region.Populations) {
			path := fmt.Sprintf("maps[%q].region.populations[%q]", key, family)
			ref(p, r.bundle.Families, path, "family", api.ContentID(family))
			if region.Populations[family] < 0 {
				p.addf("%s: must not be negative, got %d", path, region.Populations[family])
			}
		}
	}
	return p.errs
}
