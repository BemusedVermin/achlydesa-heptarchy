package agents

import (
	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/actions"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

// chooserName is the name a template gives in actor.chooser to pick its attacks with [attackChooser].
const chooserName = "agents"

// attackChooser scores each intent with the rule agents.score_rule names, which can read needs and
// target.dying, and takes the best. Ties are broken by the agents stream.
type attackChooser struct{}

// Choose implements [actions.Chooser].
func (attackChooser) Choose(w *sim.World, actor ecs.Entity, intents []actions.Intent) (actions.Intent, bool) {
	var best []actions.Intent
	var top float64
	for _, in := range intents {
		score, err := w.Content.AgentsScore(actions.ScoreEnvOf(w, actor, in))
		if err != nil {
			// A formula that cannot rate this intent leaves it out.
			continue
		}
		switch {
		case len(best) == 0 || score > top:
			best, top = []actions.Intent{in}, score
		case score == top:
			best = append(best, in)
		}
	}
	switch len(best) {
	case 0:
		return actions.Intent{}, false
	case 1:
		return best[0], true
	}
	return best[w.RNG.Get(rngStream).IntN(len(best))], true
}

// threat is a hostile's action in flight aimed at a body.
type threat struct {
	pending ecs.Entity
	actor   ecs.Entity
	// ticks is how long until it lands.
	ticks int64
}

// threatsTo lists the actions in flight that hostiles aim at the body, soonest first.
func threatsTo(w *sim.World, body ecs.Entity) []threat {
	pendings := ecs.NewMap[comp.Pending](w.ECS)
	var out []threat
	for _, pending := range timeline.AimedAt(w, w.UIDOf(body)) {
		data := pendings.Get(pending)
		actor, ok := w.ByUID(data.Actor)
		if !ok || !anatomy.Hostile(w, body, actor) {
			continue
		}
		out = append(out, threat{pending: pending, actor: actor, ticks: max(data.Lands-w.Now, 0)})
	}
	return out
}

// threatened is the ticks until the soonest landing aimed at the body by a hostile, or -1.
func threatened(w *sim.World, body ecs.Entity) int64 {
	if threats := threatsTo(w, body); len(threats) > 0 {
		return threats[0].ticks
	}
	return -1
}
