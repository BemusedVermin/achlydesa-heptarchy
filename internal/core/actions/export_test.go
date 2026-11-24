package actions

import (
	"slices"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

// IntentOf exposes the command-to-intent lookup to the tests.
var IntentOf = intentOf

// UsableChannels exposes the batch usable hook to the tests.
var UsableChannels = usableChannels

// UsableOneByOne is the slow per-channel usable test the batch hook must agree with.
func UsableOneByOne(w *sim.World, body ecs.Entity, c timeline.Channel) bool {
	ev := newEvaluator(w)
	cands, err := ev.candidates(body, false)
	if err != nil {
		return false
	}
	for _, cand := range cands {
		if !slices.ContainsFunc(
			cand.intent.Channels,
			func(used timeline.Channel) bool { return used.Entity == c.Entity },
		) {
			continue
		}
		if p, reason, err := ev.evaluate(cand.intent, false); err == nil && p != nil && reason.ID == "" {
			return true
		}
	}
	return false
}
