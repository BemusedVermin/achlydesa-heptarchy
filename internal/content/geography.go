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

// checkGeographies checks that every stage is named, every spawn is sound and that the world's geography exists. Whether a
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
	for _, key := range sortedKeys(r.bundle.Geographies) {
		for i, spawn := range r.bundle.Geographies[key].Spawns {
			path := fmt.Sprintf("geographies[%q].spawns[%d]", key, i)
			ref(p, r.bundle.Templates, path+".template", "template", spawn.Template)
			if spawn.Weight <= 0 {
				p.addf("%s.weight: must be above 0", path)
			}
			if spawn.Density < 0 || spawn.Density > 1 {
				p.addf("%s.density: must be between 0 and 1", path)
			}
			if spawn.MinDepth > spawn.MaxDepth {
				p.addf("%s: min_depth is above max_depth", path)
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
