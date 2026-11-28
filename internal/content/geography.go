package content

import (
	"fmt"

	"github.com/expr-lang/expr"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// Geography returns the geography with the given ID.
func (r *Registry) Geography(id api.ContentID) (GeographyDef, bool) {
	def, ok := r.bundle.Geographies[string(id)]
	return def, ok
}

// GeographyIDs lists every geography ID in sorted order.
func (r *Registry) GeographyIDs() []api.ContentID { return sortedIDs(r.bundle.Geographies) }

// checkGeographies checks that every stage is named, every spawn and landmark is sound and that the world's geography exists. Whether a
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
		r.checkDecay(p, key, r.bundle.Geographies[key].Decay)
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
			if spawn.Mind != nil {
				if spawn.Mind.Planner == "" {
					p.addf("%s.mind.planner: required", path)
				}
				if spawn.Mind.Chance < 0 || spawn.Mind.Chance > 1 {
					p.addf("%s.mind.chance: must be between 0 and 1", path)
				}
				if template, ok := r.bundle.Templates[string(spawn.Template)]; ok && template.Mind == nil {
					p.addf("%s.mind: template %q has no mind", path, spawn.Template)
				}
			}
		}
	}
	for _, key := range sortedKeys(r.bundle.Geographies) {
		r.checkLandmarks(p, key, r.bundle.Geographies[key].Landmarks)
	}
	if id := r.bundle.World.Geography; id != "" {
		ref(p, r.bundle.Geographies, "world.geography", "geography", id)
		if r.bundle.World.PocketExits < 1 {
			p.addf("world.pocket_exits: must be at least 1")
		}
	}
	return p.errs
}

func (r *Registry) checkDecay(p *problems, key string, decay *DecayDef) {
	if decay == nil {
		return
	}
	path := fmt.Sprintf("geographies[%q].decay", key)
	if len(decay.Terrains) == 0 {
		p.addf("%s.terrains: at least one terrain is required", path)
	}
	for i, id := range decay.Terrains {
		terrain, ok := r.bundle.Terrains[string(id)]
		switch {
		case !ok:
			p.addf("%s.terrains[%d]: unknown terrain %q", path, i, id)
		case !terrain.Solid:
			p.addf("%s.terrains[%d]: terrain %q must be solid", path, i, id)
		}
	}
	program, err := r.Compile(decay.Level, DecayEnv{}, expr.AsInt64())
	if err != nil {
		p.addf("%s.level: %w", path, err)
		return
	}
	r.decays[api.ContentID(key)] = program
}

func (r *Registry) checkLandmarks(p *problems, key string, landmarks []GeographyLandmark) {
	world := r.bundle.World
	for i, landmark := range landmarks {
		path := fmt.Sprintf("geographies[%q].landmarks[%d]", key, i)
		ref(p, r.bundle.Pieces, path+".piece", "piece", landmark.Piece)
		if town, ok := ref(p, r.bundle.Settlements, path+".near", "settlement", landmark.Near); ok &&
			landmark.MinDistance <= town.Radius+1 {
			p.addf("%s.min_distance: %d does not clear the settlement's radius %d and a hex of margin",
				path, landmark.MinDistance, town.Radius)
		}
		if landmark.MinDistance > landmark.MaxDistance {
			p.addf("%s: min_distance is above max_distance", path)
		}
		if landmark.MinDepth > landmark.MaxDepth {
			p.addf("%s: min_depth is above max_depth", path)
		}
		if landmark.MinDepth < world.MinDepth || landmark.MaxDepth > world.MaxDepth {
			p.addf("%s: the depth band reaches outside the world", path)
		}
	}
}

// GeographyPath is where a geography's stage sits in content, for error messages.
func GeographyPath(id api.ContentID, stage int) string {
	return fmt.Sprintf("geographies[%q].stages[%d]", id, stage)
}
