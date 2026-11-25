package content

import (
	"fmt"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// Geography returns the geography with the given ID.
func (r *Registry) Geography(id api.ContentID) (GeographyDef, bool) {
	def, ok := r.bundle.Geographies[string(id)]
	return def, ok
}

// GeographyIDs lists every geography ID in sorted order.
func (r *Registry) GeographyIDs() []api.ContentID { return sortedIDs(r.bundle.Geographies) }

// checkGeographies checks that every stage is named and that the world's geography exists. Whether a
// stage is registered and takes its params is for worldgen to check: only it knows the stages.
func (r *Registry) checkGeographies() []error {
	p := r.problems()
	for _, key := range sortedKeys(r.bundle.Geographies) {
		for i, stage := range r.bundle.Geographies[key].Stages {
			if stage.Stage == "" {
				p.addf("%s.stage: required", GeographyPath(api.ContentID(key), i))
			}
		}
	}
	if id := r.bundle.World.Geography; id != "" {
		ref(p, r.bundle.Geographies, "world.geography", "geography", id)
	}
	return p.errs
}

// GeographyPath is where a geography's stage sits in content, for error messages.
func GeographyPath(id api.ContentID, stage int) string {
	return fmt.Sprintf("geographies[%q].stages[%d]", id, stage)
}
