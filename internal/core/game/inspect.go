package game

import (
	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
)

// PositionOf returns where the entity with the UID stands, including its depth, which the view
// leaves out. It is a read-only accessor for the harness and tests, not for frontends.
func (g *Game) PositionOf(uid api.UID) (comp.Pos, bool) { return component[comp.Pos](g, uid) }

// PartOfUID returns a copy of the part with the UID, including its true material, which the view
// hides until the party learns it. It is a read-only accessor for the harness and tests, not for
// frontends.
func (g *Game) PartOfUID(uid api.UID) (comp.Part, bool) { return component[comp.Part](g, uid) }

// component returns a copy of the T of the entity with the UID.
func component[T any](g *Game, uid api.UID) (T, bool) {
	var zero T
	e, ok := g.world.ByUID(uid)
	if !ok {
		return zero, false
	}
	values := ecs.NewMap[T](g.world.ECS)
	if !values.Has(e) {
		return zero, false
	}
	return *values.Get(e), true
}
