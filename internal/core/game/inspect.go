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

// PartOfUID returns a copy of the part with the UID, including its true material, which the view
// hides until the party learns it. It is a read-only accessor for the harness and tests, not for
// frontends.
func (g *Game) PartOfUID(uid api.UID) (comp.Part, bool) {
	e, ok := g.world.ByUID(uid)
	if !ok {
		return comp.Part{}, false
	}
	parts := ecs.NewMap[comp.Part](g.world.ECS)
	if !parts.Has(e) {
		return comp.Part{}, false
	}
	return *parts.Get(e), true
}
