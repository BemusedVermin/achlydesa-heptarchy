package combat

import (
	"slices"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/actions"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/effects"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

// applyActAlone makes the source part, which must be a graft, start an attack of its own for its
// host. It picks through the scored chooser among the attacks that use only its channel and may
// aim at anyone, either side. It does nothing when the part is not a free graft or has no attack.
func applyActAlone(ctx *effects.Context, _ effects.Params) error {
	w := ctx.World
	part, body := ctx.SourcePart, ctx.Source
	if !w.Alive(part) || !w.Alive(body) ||
		!ecs.NewMap[comp.Graft](w.ECS).Has(part) || bodyOf(w, part) != body {
		return nil
	}
	if outOfFight(w, body) {
		return nil
	}
	cfg := w.Content.Combat()
	var own []actions.Intent
	for _, in := range actions.Startable(w, body) {
		if len(in.Channels) != 1 || in.Channels[0].Entity != part {
			continue
		}
		if def, ok := w.Content.Action(in.Action); !ok || !slices.Contains(def.Tags, cfg.AttackTag) {
			continue
		}
		own = append(own, in)
	}
	if len(own) == 0 {
		return nil
	}
	chooser, ok := stateOf(w).choosers[scoreChooser]
	if !ok {
		return nil
	}
	in, ok := chooser.Choose(w, body, own)
	if !ok {
		return nil
	}
	if !timeline.Free(w, in.Channels[0]) {
		return nil
	}
	return actions.StartAutonomous(w, in)
}
