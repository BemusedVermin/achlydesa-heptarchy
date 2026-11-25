package agents

import (
	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// publishIfEmptied publishes [api.TownEmptied] when the map is the controlled party's town, the
// template is the town's residents and none of its bodies is left in it. Call it after a body of
// the template leaves the map or dies, so each emptying is reported once.
func publishIfEmptied(w *sim.World, id api.MapID, template api.ContentID) {
	if at, onScreen := partyMap(w); !onScreen || at != id {
		return
	}
	town, ok := TownOf(w, id)
	if !ok {
		return
	}
	def, _ := w.Content.Map(ecs.NewMap[comp.Region](w.ECS).Get(town).Def)
	if def.Region == nil || def.Region.Town == nil || def.Region.Town.Residents != template {
		return
	}
	if residentsOf(w, id, def.Region.Town) == 0 {
		w.Bus.Publish(api.TownEmptied{Map: id, Template: template})
	}
}

func onBodyDied(w *sim.World, e api.BodyDied) {
	body, ok := w.ByUID(e.Body)
	if !ok || !ecs.NewMap[comp.Pos](w.ECS).Has(body) {
		return
	}
	publishIfEmptied(w, ecs.NewMap[comp.Pos](w.ECS).Get(body).Map, w.TemplateOf(body))
}
