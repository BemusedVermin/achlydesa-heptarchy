package agents

import (
	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// publishIfEmptied publishes [api.TownEmptied] when the map is the controlled party's town, the
// family is the town's residents and none of that family is left in it. It is called only as a
// resident leaves or dies, so each emptying is reported once; off screen the tick reports it.
func publishIfEmptied(w *sim.World, id api.MapID, family api.ContentID) {
	if at, onScreen := partyMap(w); !onScreen || at != id {
		return
	}
	town, ok := TownOf(w, id)
	if !ok {
		return
	}
	def, _ := w.Content.Map(ecs.NewMap[comp.Region](w.ECS).Get(town).Def)
	if def.Region == nil || def.Region.Town == nil || def.Region.Town.Residents != family {
		return
	}
	if residentsOf(w, town, id, def.Region.Town) == 0 {
		w.Bus.Publish(api.TownEmptied{Map: id, Family: family})
	}
}

func onBodyDied(w *sim.World, e api.BodyDied) {
	body, ok := w.ByUID(e.Body)
	if !ok {
		return
	}
	families, positions := ecs.NewMap[comp.Family](w.ECS), ecs.NewMap[comp.Pos](w.ECS)
	if !families.Has(body) || !positions.Has(body) {
		return
	}
	publishIfEmptied(w, positions.Get(body).Map, families.Get(body).ID)
}
