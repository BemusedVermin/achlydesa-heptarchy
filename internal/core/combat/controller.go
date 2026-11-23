package combat

import (
	"fmt"
	"slices"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/actions"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/effects"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/inventory"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/knowledge"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

const (
	// huntController is the name of the controller of a creature that hunts the party.
	huntController = "hunt"
	// scoreChooser is the name the content-scored chooser is registered under, and the one the
	// hunt controller asks for.
	scoreChooser = "score"
	// huntPathNodes bounds the search for a way to the prey.
	huntPathNodes = 200
	// rngStream is the stream a creature draws from when it picks at random.
	rngStream = "combat"
)

// Chooser picks one of the intents a body could start. Choosers are registered by name, so a
// smarter one can replace the scored one later.
type Chooser interface {
	// Choose returns the intent to start, or false when it wants none.
	Choose(w *sim.World, actor ecs.Entity, intents []actions.Intent) (actions.Intent, bool)
}

// ScoreChooser scores each intent with content's ai_score and takes the best, breaking ties with
// the combat stream.
type ScoreChooser struct{}

// Choose implements [Chooser].
func (ScoreChooser) Choose(w *sim.World, actor ecs.Entity, intents []actions.Intent) (actions.Intent, bool) {
	var best []actions.Intent
	var top float64
	for _, in := range intents {
		score, err := scoreOf(w, actor, in)
		if err != nil {
			continue // a formula that cannot rate this intent leaves it out
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

// scoreOf rates one intent by the content formula.
func scoreOf(w *sim.World, actor ecs.Entity, in actions.Intent) (float64, error) {
	var sourcePart ecs.Entity
	for _, c := range in.Channels {
		if !c.Body.IsZero() && c.Entity != c.Body {
			sourcePart = c.Entity
			break
		}
	}
	env := content.ScoreEnv{
		Source: effects.SideOf(w, actor, sourcePart), Target: effects.SideOf(w, in.Target, in.TargetPart),
		Chosen: effects.ChosenRecords(w, in.Choices), Pack: inventory.Usable(w, actor, in.Target),
		Windup: actions.Windup(w, in),
	}
	env.TargetIsParty = ecs.NewMap[comp.MemberOf](w.ECS).Has(in.Target) && partyControlled(w, in.Target)
	if parts := ecs.NewMap[comp.Part](w.ECS); !in.TargetPart.IsZero() && parts.Has(in.TargetPart) {
		env.TargetIntegrity = parts.Get(in.TargetPart).Integrity
		_, env.ImmuneKnown = knowledge.Best(
			w, inventory.HolderOf(w, actor), w.UIDOf(in.TargetPart), w.Content.Combat().MaterialClaim,
		)
	}
	positions := ecs.NewMap[comp.Pos](w.ECS)
	if positions.Has(actor) && positions.Has(in.Target) {
		env.Height = positions.Get(in.Target).At.Depth - positions.Get(actor).At.Depth
	}
	return w.Content.AIScore(env)
}

// partyControlled says whether the body belongs to the party the player commands.
func partyControlled(w *sim.World, body ecs.Entity) bool {
	party, ok := controlledParty(w)
	return ok && ecs.NewMap[comp.MemberOf](w.ECS).GetRelation(body) == party
}

// RegisterChooser makes a chooser available by name. Names are unique.
func RegisterChooser(w *sim.World, name string, c Chooser) error {
	st := stateOf(w)
	if _, dup := st.choosers[name]; dup {
		return fmt.Errorf("combat: chooser %q already registered", name)
	}
	st.choosers[name] = c
	return nil
}

// state is what combat keeps on the world, as an Ark resource.
type state struct {
	choosers map[string]Chooser
}

func stateOf(w *sim.World) *state {
	res := ecs.NewResource[state](w.ECS)
	if res.Has() {
		return res.Get()
	}
	st := &state{choosers: map[string]Chooser{}}
	res.Add(st)
	return st
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
	cfg := w.Content.Combat()
	startable := actions.Startable(w, body)
	aims, err := w.Content.Aims(content.GrantEnv{Grant: anatomy.GrantTotals(w, body)})
	if err != nil {
		return err
	}
	var attacks []actions.Intent
	for _, in := range startable {
		if def, ok := w.Content.Action(in.Action); !ok || !slices.Contains(def.Tags, cfg.AttackTag) {
			continue
		}
		if aims && in.Target != prey {
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

// pick chooses among the attacks: by the chooser when the body aims, at random when it cannot.
func pick(w *sim.World, body ecs.Entity, attacks []actions.Intent, aims bool) (actions.Intent, bool) {
	if !aims {
		return attacks[w.RNG.Get(rngStream).IntN(len(attacks))], true
	}
	chooser, ok := stateOf(w).choosers[scoreChooser]
	if !ok {
		return actions.Intent{}, false
	}
	return chooser.Choose(w, body, attacks)
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
	for _, other := range hostileCandidates(w) { // in UID order, so the first of equals wins
		pos := positions.Get(other)
		if pos.Map != from.Map || !seen[pos.At] || !Hostile(w, body, other) {
			continue
		}
		distance := hexmap.Distance(from.At.Hex, pos.At.Hex) + abs(from.At.Depth-pos.At.Depth)
		if best.IsZero() || distance < bestDistance {
			best, bestDistance = other, distance
		}
	}
	return best, !best.IsZero(), nil
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
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
	cfg := w.Content.Party()
	rules := hexmap.StepRules{MaxStepUp: cfg.MaxStepUp, MaxStepDown: cfg.MaxStepDown, MaxDrop: cfg.MaxDrop}
	route, found, err := m.Path(from.At, to.At, rules, nil, huntPathNodes)
	if err != nil || !found || len(route) < 2 {
		return err
	}
	for _, in := range startable {
		if in.Action != stepID || in.Dir == nil {
			continue
		}
		step, can, err := m.Step(from.At, *in.Dir, rules)
		if err != nil {
			return err
		}
		if can && (step.To.Map == 0 || step.To.Map == from.Map) && step.To.Cell == route[1] {
			return actions.Start(w, in)
		}
	}
	return nil
}
