package actions

import (
	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/effects"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/inventory"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/knowledge"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

// ScoreEnvOf is what a chooser's score formula sees of one intent the actor could start.
func ScoreEnvOf(w *sim.World, actor ecs.Entity, in Intent) content.ScoreEnv {
	env := content.ScoreEnv{
		Source: effects.BodyEnvOf(w, actor, timeline.FirstPart(in.Channels)),
		Target: effects.BodyEnvOf(w, in.Target, in.TargetPart),
		Chosen: effects.ChosenRecords(w, in.Choices), Inventory: inventory.Usable(w, actor, in.Target),
		Windup: Windup(w, in),
	}
	env.Target.IsParty = anatomy.InControlledParty(w, in.Target)
	if parts := ecs.NewMap[comp.Part](w.ECS); !in.TargetPart.IsZero() && parts.Has(in.TargetPart) {
		_, env.MaterialKnown = knowledge.Best(
			w, inventory.HolderOf(w, actor), w.UIDOf(in.TargetPart), w.Content.Combat().MaterialClaim,
		)
	}
	positions := ecs.NewMap[comp.Pos](w.ECS)
	if positions.Has(actor) && positions.Has(in.Target) {
		env.Height = positions.Get(in.Target).At.Depth - positions.Get(actor).At.Depth
	}
	return env
}
