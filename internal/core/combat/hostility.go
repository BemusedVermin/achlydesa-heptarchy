package combat

import (
	"slices"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// Hostile reports whether two bodies are enemies: either one's faction lists the other's. A body
// with no allegiance is nobody's enemy.
func Hostile(w *sim.World, a, b ecs.Entity) bool {
	allegiances := ecs.NewMap[comp.Allegiance](w.ECS)
	if !w.Alive(a) || !w.Alive(b) || !allegiances.Has(a) || !allegiances.Has(b) {
		return false
	}
	factionA, factionB := allegiances.Get(a).Faction, allegiances.Get(b).Faction
	return lists(w, factionA, factionB) || lists(w, factionB, factionA)
}

// lists reports whether the faction names the other among those it is hostile to.
func lists(w *sim.World, faction, other api.ContentID) bool {
	def, ok := w.Content.Faction(faction)
	return ok && slices.Contains(def.HostileTo, other)
}

// nameOf is the text of a body's name, "text.template.<template without its prefix>".
func nameOf(w *sim.World, e ecs.Entity) string {
	return w.Content.Render(w.Content.TemplateName(w.TemplateOf(e)))
}
