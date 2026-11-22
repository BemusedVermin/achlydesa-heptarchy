package game

import (
	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
)

// PositionOf returns where the entity with the UID stands, including its depth, which the view
// leaves out. It is a read-only accessor for the harness and tests, not for frontends.
func (g *Game) PositionOf(uid api.UID) (comp.Pos, bool) {
	e, ok := g.world.ByUID(uid)
	if !ok {
		return comp.Pos{}, false
	}
	positions := ecs.NewMap[comp.Pos](g.world.ECS)
	if !positions.Has(e) {
		return comp.Pos{}, false
	}
	return *positions.Get(e), true
}
