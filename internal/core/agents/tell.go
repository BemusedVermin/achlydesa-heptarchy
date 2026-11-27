package agents

import (
	"cmp"
	"slices"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/actions"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/effects"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/knowledge"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

func installTell(w *sim.World) error {
	return effects.Register(w, effects.Primitive{Name: "tell", Apply: applyTell})
}

// applyTell has the acting body tell its target the fact it picks for it. It does nothing when the
// target is not a living agent or the speaker has nothing left to tell.
func applyTell(ctx *effects.Context, _ effects.Params) error {
	w := ctx.World
	if ctx.Source == (ecs.Entity{}) || ctx.Target == (ecs.Entity{}) || !isListener(w, ctx.Target) {
		return nil
	}
	if told, ok := toTell(w, ctx.Source, ctx.Target); ok {
		knowledge.Learn(w, ctx.Target, told)
	}
	return nil
}

// isListener says whether the body is a living agent with a mind.
func isListener(w *sim.World, body ecs.Entity) bool {
	return w.Alive(body) && !ecs.NewMap[comp.Dead](w.ECS).Has(body) && ecs.NewMap[comp.Mind](w.ECS).Has(body)
}

// toTell is the fact the speaker would pass to the hearer: the one it is surest of, then the newest,
// among those the hearer does not already hold with at least the confidence the told copy would
// have. It comes back as the hearer would learn it, with the told origin and the content's share of
// the confidence, and is otherwise unchanged. It is false when the speaker has nothing to tell.
func toTell(w *sim.World, speaker, hearer ecs.Entity) (api.Fact, bool) {
	cfg := w.Content.Agents()
	known := ecs.NewMap[comp.Knowledge](w.ECS)
	if !known.Has(speaker) {
		return api.Fact{}, false
	}
	hearerUID := w.UIDOf(hearer)
	var best api.Fact
	found := false
	for _, fact := range known.Get(speaker).Facts {
		if fact.Subject == hearerUID {
			continue
		}
		told := fact
		told.Origin, told.Confidence = cfg.ToldOrigin, fact.Confidence*cfg.ToldConfidence
		if held, has := knowledge.Best(w, hearer, fact.Subject, fact.Claim); has && held.Confidence >= told.Confidence {
			continue
		}
		if !found || cmp.Or(
			cmp.Compare(best.Confidence, told.Confidence),
			cmp.Compare(best.Tick, told.Tick),
		) < 0 {
			best, found = told, true
		}
	}
	return best, found
}

// chatPartner is the lowest-UID living agent the body can chat to now: within the chat action's
// reach, and with something to hear.
func chatPartner(w *sim.World, body ecs.Entity) (ecs.Entity, bool) {
	cfg := w.Content.Agents()
	if cfg.ChatAction == "" {
		return ecs.Entity{}, false
	}
	var others []ecs.Entity
	query := ecs.NewFilter1[comp.Mind](w.ECS).Query()
	for query.Next() {
		if other := query.Entity(); other != body && isListener(w, other) {
			others = append(others, other)
		}
	}
	slices.SortFunc(others, func(a, b ecs.Entity) int { return cmp.Compare(w.UIDOf(a), w.UIDOf(b)) })
	for _, other := range others {
		if _, tells := toTell(w, body, other); !tells {
			continue
		}
		if ok, _ := actions.Validate(w, actions.Intent{Actor: body, Action: cfg.ChatAction, Target: other}); ok {
			return other, true
		}
	}
	return ecs.Entity{}, false
}

// startChat starts the chat action at a partner in reach, and reports false when there is none.
func startChat(w *sim.World, body ecs.Entity) (bool, error) {
	partner, ok := chatPartner(w, body)
	if !ok {
		return false, nil
	}
	err := actions.Start(w, actions.Intent{Actor: body, Action: w.Content.Agents().ChatAction, Target: partner})
	return err == nil, err
}
