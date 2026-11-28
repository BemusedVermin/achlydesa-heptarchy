package agents

import (
	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/actions"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/knowledge"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// CorpsePicker is the target picker that aims an action at the nearest believed landmark that is a
// corpse. A landmark the agent believes is there but finds gone is forgotten.
const CorpsePicker = "corpse"

type corpsePicker struct{}

func (corpsePicker) Pick(w *sim.World, body ecs.Entity, step Step) (actions.Intent, bool, error) {
	op, ok := w.Content.Operator(step.Operator)
	if !ok {
		return actions.Intent{}, false, nil
	}
	marks, err := thingPicker{}.subjects(w, body, op, step)
	if err != nil {
		return actions.Intent{}, false, err
	}
	dead := ecs.NewMap[comp.Dead](w.ECS)
	for _, mark := range marks {
		thing, found := w.ByUID(mark.Subject)
		if !found || !w.Alive(thing) {
			forgetLandmark(w, body, mark.Subject)
			continue
		}
		if dead.Has(thing) {
			return actions.Intent{Target: thing}, true, nil
		}
	}
	return actions.Intent{}, false, nil
}

// forgetLandmark drops what the body believes of a landmark's whereabouts and gifts, whoever told it.
func forgetLandmark(w *sim.World, body ecs.Entity, subject api.UID) {
	cfg := w.Content.Agents()
	for _, claim := range []api.ContentID{cfg.GivesClaim, cfg.AtClaim, cfg.DeadClaim} {
		for _, origin := range []api.ContentID{cfg.SeenOrigin, cfg.ToldOrigin} {
			if claim != "" && origin != "" {
				knowledge.Forget(w, body, subject, claim, origin)
			}
		}
	}
}
