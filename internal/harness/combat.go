package harness

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
)

// Outcomes of a fight.
const (
	OutcomeNone  = ""      // still going, or never began
	OutcomeWon   = "won"   // a hostile died and none is left in sight
	OutcomeWiped = "wiped" // every member of the party was wiped out
)

// Binder is implemented by a policy that needs the game it plays, such as to look up what a part
// is made of. The harness calls Bind once, before the first choice.
type Binder interface {
	Bind(g *game.Game)
}

// FightResult says how one seeded fight went.
type FightResult struct {
	Seed    uint64
	Outcome string
	// Ticks is how long the fight lasted: from the alert to the last hostile's death or the party's
	// fall. It is zero when the alert never came.
	Ticks int64
	// Steps is how many commands the policy sent.
	Steps int
}

// tracker follows a run's combat from its events and views.
type tracker struct {
	alertStart   int64
	started      bool
	lastTick     int64
	alertTicks   int64
	landings     int
	misses       int
	severed      int
	deaths       int
	hostiles     map[api.UID]bool
	hostileDeath int
	outcome      string
	end          int64
}

func newTracker() *tracker { return &tracker{hostiles: map[api.UID]bool{}} }

func (t *tracker) noteView(view api.View) {
	if view.Bodies == nil {
		return
	}
	for _, body := range view.Bodies.Bodies {
		if body.Hostile {
			t.hostiles[body.UID] = true
		}
	}
}

// observe takes the events one command delivered and the view after it.
func (t *tracker) observe(events []api.Event, view api.View) {
	t.noteView(view)
	for _, event := range events {
		switch e := event.(type) {
		case api.AlertStarted:
			if !t.started {
				t.started, t.alertStart = true, e.Tick
			}
			t.hostiles[e.Hostile] = true
		case api.PartDamaged:
			t.landings++
		case api.ActionMissed:
			t.misses++
		case api.PartSevered:
			t.severed++
		case api.BodyDied:
			t.deaths++
			if t.hostiles[e.Body] {
				t.hostileDeath++
				t.end = e.Tick
			}
		case api.PartyWiped:
			if t.outcome == OutcomeNone {
				t.outcome, t.end = OutcomeWiped, e.Tick
			}
		}
	}
	if view.Party != nil && view.Party.Alert {
		t.alertTicks += view.Tick - t.lastTick
	}
	t.lastTick = view.Tick
	if t.outcome == OutcomeNone && t.hostileDeath > 0 && view.Bodies != nil &&
		!slices.ContainsFunc(view.Bodies.Bodies, func(b api.BodyView) bool { return b.Hostile }) {
		t.outcome = OutcomeWon
	}
}

func (t *tracker) columns() []string {
	return []string{
		strconv.FormatInt(t.alertTicks, 10), strconv.Itoa(t.landings), strconv.Itoa(t.misses),
		strconv.Itoa(t.severed), strconv.Itoa(t.deaths), t.outcome,
	}
}

func (t *tracker) result(seed uint64, steps int) FightResult {
	result := FightResult{Seed: seed, Outcome: t.outcome, Steps: steps}
	if t.started && t.outcome != OutcomeNone {
		result.Ticks = t.end - t.alertStart
	}
	return result
}

// PlayFight plays one seeded game until its fight is decided or maxSteps commands have been sent.
func PlayFight(reg *content.Registry, seed uint64, policy Policy, maxSteps int) (FightResult, error) {
	g, err := game.New(game.Config{Seed: seed, Content: reg})
	if err != nil {
		return FightResult{}, err
	}
	if binder, ok := policy.(Binder); ok {
		binder.Bind(g)
	}
	track := newTracker()
	steps := 0
	view, err := g.View()
	if err != nil {
		return FightResult{}, err
	}
	track.noteView(view)
	for steps < maxSteps && track.outcome == OutcomeNone {
		events, ok, err := Step(g, policy)
		if err != nil {
			return FightResult{}, fmt.Errorf("step %d: %w", steps+1, err)
		}
		if !ok {
			break
		}
		steps++
		if view, err = g.View(); err != nil {
			return FightResult{}, err
		}
		track.observe(events, view)
	}
	return track.result(seed, steps), nil
}

// FightPolicy plays a fight from the options alone. The blind one takes the first attack offered
// and never Sweeps; the informed one knows what every part is made of and loads the type that
// wounds it. Both walk east until something is in sight, which is where the sample hall keeps its
// creature, and wait for a creature that is not yet in reach.
type FightPolicy struct {
	reg      *content.Registry
	g        *game.Game
	informed bool
}

// NewBlindPolicy returns the policy of a party that knows nothing of what it fights.
func NewBlindPolicy(reg *content.Registry) *FightPolicy { return &FightPolicy{reg: reg} }

// NewInformedPolicy returns the policy of a party that knows every material.
func NewInformedPolicy(reg *content.Registry) *FightPolicy {
	return &FightPolicy{reg: reg, informed: true}
}

// Bind implements [Binder].
func (p *FightPolicy) Bind(g *game.Game) { p.g = g }

// Choose implements [Policy].
func (p *FightPolicy) Choose(view api.View) (api.Option, bool) {
	hostile := map[api.UID]bool{}
	if view.Bodies != nil {
		for _, body := range view.Bodies.Bodies {
			if body.Hostile {
				hostile[body.UID] = true
			}
		}
	}
	var attacks, steps []api.Option
	var wait api.Option
	var haveWait bool
	for _, option := range view.Options {
		if !option.Enabled {
			continue
		}
		switch cmd := option.Command.(type) {
		case api.Wait:
			if !haveWait {
				wait, haveWait = option, true
			}
		case api.Do:
			switch {
			case hostile[cmd.Target]:
				attacks = append(attacks, option)
			case cmd.Dir != nil:
				steps = append(steps, option)
			}
		}
	}
	if choice, ok := p.attack(attacks); ok {
		return choice, true
	}
	if step, ok := p.closeIn(steps, hostile); ok {
		return step, true
	}
	return wait, haveWait
}

// closeIn takes the step that brings the actor nearest a hostile body in sight, or the first step
// offered when none is in sight, which in the sample hall leads east to the creature.
func (p *FightPolicy) closeIn(steps []api.Option, hostile map[api.UID]bool) (api.Option, bool) {
	if len(steps) == 0 {
		return api.Option{}, false
	}
	if len(hostile) == 0 || p.g == nil {
		return steps[0], len(hostile) == 0
	}
	best, bestDistance := api.Option{}, 0
	found := false
	for _, option := range steps {
		do := option.Command.(api.Do)
		from, ok := p.g.PositionOf(do.Source)
		if !ok {
			continue
		}
		to := hexmap.Prisms{}.Neighbor(from.At, *do.Dir).Cell
		for uid := range hostile {
			at, ok := p.g.PositionOf(uid)
			if !ok {
				continue
			}
			distance := hexmap.Distance(to.Hex, at.At.Hex)
			if !found || distance < bestDistance {
				best, bestDistance, found = option, distance, true
			}
		}
	}
	return best, found
}

func (p *FightPolicy) attack(attacks []api.Option) (api.Option, bool) {
	best, bestScore, found := api.Option{}, 0.0, false
	for _, option := range attacks {
		do := option.Command.(api.Do)
		def, ok := p.reg.Action(do.Action)
		if !ok {
			continue
		}
		if !p.informed {
			if def.Key == "sweep" || len(def.Choices) == 0 {
				continue
			}
			return option, true // the first offered
		}
		score, ok := p.score(do, def)
		if ok && (!found || score > bestScore) {
			best, bestScore, found = option, score, true
		}
	}
	return best, found
}

// score rates an attack for the informed party: a type the part is weak to, then a part the body
// cannot live without, then a form that cuts off what it hits. A swing that does nothing is out.
func (p *FightPolicy) score(do api.Do, def content.ActionDef) (float64, bool) {
	typeID, hasType := do.Choices["type"]
	if !hasType || do.TargetPart == 0 || p.g == nil {
		return 0, false
	}
	part, ok := p.g.PartOfUID(do.TargetPart)
	if !ok {
		return 0, false
	}
	material, ok := p.reg.Material(part.Material)
	if !ok {
		return 0, false
	}
	if slices.Contains(material.Lists["immune_to"], typeID) {
		return 0, false
	}
	score := 0.0
	if slices.Contains(material.Lists["weak_to"], typeID) {
		score += 10
	}
	if part.Vital {
		score += 5
	}
	if def.Key == "sever" {
		score += 2
	}
	return score, true
}
