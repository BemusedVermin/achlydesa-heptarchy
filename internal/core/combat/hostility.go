package combat

import (
	"slices"
	"strings"

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

// outOfFight reports whether a body is dead or downed.
func outOfFight(w *sim.World, body ecs.Entity) bool {
	return ecs.NewMap[comp.Dead](w.ECS).Has(body) || ecs.NewMap[comp.Downed](w.ECS).Has(body)
}

// nameOf is the text of a body's name, "text.template.<template without its prefix>".
func nameOf(w *sim.World, e ecs.Entity) string {
	identities := ecs.NewMap[comp.Identity](w.ECS)
	if !w.Alive(e) || !identities.Has(e) {
		return w.Content.Render(api.Text{ID: "text.template.unknown"})
	}
	template := strings.TrimPrefix(string(identities.Get(e).Template), "template.")
	return w.Content.Render(api.Text{ID: api.ContentID("text.template." + template)})
}
