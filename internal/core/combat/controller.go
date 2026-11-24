package combat

import (
	"slices"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/actions"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

const (
	huntController = "hunt"
	// huntPathNodes bounds the search for a way to the prey.
	huntPathNodes = 200
	// rngStream is the stream a creature draws from when it picks at random.
	rngStream = "combat"
)

// ScoreChooser scores each intent with content's ai_score and takes the best, breaking ties with
// the combat stream. Install registers it as [actions.DefaultChooser].
type ScoreChooser struct{}

// Choose implements [actions.Chooser].
func (ScoreChooser) Choose(w *sim.World, actor ecs.Entity, intents []actions.Intent) (actions.Intent, bool) {
	var best []actions.Intent
	var top float64
	for _, in := range intents {
		score, err := scoreOf(w, actor, in)
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
	if len(best) == 0 {
		return actions.Intent{}, false
	}
	if len(best) == 1 {
		return best[0], true
	}
	return best[w.RNG.Get(rngStream).IntN(len(best))], true
}

func scoreOf(w *sim.World, actor ecs.Entity, in actions.Intent) (float64, error) {
	return w.Content.AIScore(actions.ScoreEnvOf(w, actor, in))
}

// hunter is the controller of a creature that goes after the nearest hostile.
type hunter struct{}

// Decide picks the nearest visible hostile, attacks it when something reaches, otherwise steps
// toward it, otherwise waits for the next entry on the timeline.
func (hunter) Decide(w *sim.World, body ecs.Entity) error {
	prey, found, err := nearestHostile(w, body)
	if err != nil || !found {
		return err
	}
	startable := actions.Startable(w, body)
	aims, err := w.Content.Aims(content.GrantEnv{Grant: anatomy.GrantTotals(w, body)})
	if err != nil {
		return err
	}
	var attacks []actions.Intent
	for _, in := range startable {
		if !isAttack(w, in) {
			continue
		}
		if aims && !mayTarget(w, body, in.Target, prey) {
			continue
		}
		attacks = append(attacks, in)
	}
	if len(attacks) > 0 {
		in, ok := pick(w, body, attacks, aims)
		if ok {
			return actions.Start(w, in)
		}
	}
	return stepToward(w, body, prey, startable)
}

// mayTarget says whether a body that aims may swing at target: the prey, or, for a body whose
// template names its own chooser, any hostile as near as the prey, so the chooser decides between them.
func mayTarget(w *sim.World, body, target, prey ecs.Entity) bool {
	if target == prey {
		return true
	}
	if actions.ChooserName(w, body) == actions.DefaultChooser || !anatomy.Hostile(w, body, target) {
		return false
	}
	positions := ecs.NewMap[comp.Pos](w.ECS)
	if !positions.Has(target) || !positions.Has(prey) || positions.Get(target).Map != positions.Get(prey).Map {
		return false
	}
	from := positions.Get(body).At
	return hexmap.CellDistance(from, positions.Get(target).At) == hexmap.CellDistance(from, positions.Get(prey).At)
}

// pick chooses among the attacks: by the chooser when the body aims, at random when it cannot.
func pick(w *sim.World, body ecs.Entity, attacks []actions.Intent, aims bool) (actions.Intent, bool) {
	if !aims {
		return attacks[w.RNG.Get(rngStream).IntN(len(attacks))], true
	}
	chooser, ok := actions.ChooserFor(w, body)
	if !ok {
		return actions.Intent{}, false
	}
	return chooser.Choose(w, body, attacks)
}

func isAttack(w *sim.World, in actions.Intent) bool {
	def, ok := w.Content.Action(in.Action)
	return ok && slices.Contains(def.Tags, w.Content.Combat().AttackTag)
}

// nearestHostile finds the visible hostile closest by hex distance plus depth difference, then UID.
func nearestHostile(w *sim.World, body ecs.Entity) (ecs.Entity, bool, error) {
	positions := ecs.NewMap[comp.Pos](w.ECS)
	if !positions.Has(body) {
		return ecs.Entity{}, false, nil
	}
	from := *positions.Get(body)
	seen, err := seenCells(w, from)
	if err != nil {
		return ecs.Entity{}, false, err
	}
	var best ecs.Entity
	bestDistance := 0
	// In UID order, so the first of equals wins.
	for _, other := range hostileCandidates(w) {
		pos := positions.Get(other)
		if pos.Map != from.Map || !seen[pos.At] || !anatomy.Hostile(w, body, other) {
			continue
		}
		distance := hexmap.CellDistance(from.At, pos.At)
		if best.IsZero() || distance < bestDistance {
			best, bestDistance = other, distance
		}
	}
	return best, !best.IsZero(), nil
}

// stepToward takes the step along the path to the prey, when the step action is startable.
func stepToward(w *sim.World, body, prey ecs.Entity, startable []actions.Intent) error {
	stepID := w.Content.Combat().StepAction
	positions := ecs.NewMap[comp.Pos](w.ECS)
	from, to := *positions.Get(body), *positions.Get(prey)
	m := w.Maps.Get(from.Map)
	if m == nil || from.Map != to.Map {
		return nil
	}
	rules := actions.StepRules(w)
	route, found, err := m.Path(from.At, to.At, rules, nil, huntPathNodes)
	if err != nil || !found || len(route) < 2 {
		return err
	}
	dir, can, err := m.DirTo(from.At, route[1], rules)
	if err != nil || !can {
		return err
	}
	for _, in := range startable {
		if in.Action == stepID && in.Dir != nil && *in.Dir == dir {
			return actions.Start(w, in)
		}
	}
	return nil
}
