package agents_test

import (
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/actions"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/agents"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/clock"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/effects"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game/gametest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/knowledge"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

const (
	clayTemplate    = api.ContentID("template.creature.moving_clay")
	remnantTemplate = api.ContentID("template.creature.remnant")
	beastTemplate   = api.ContentID("template.creature.beast")
)

// familyGame starts a game with the party far from where the families are spawned, and the family
// templates in names run by the mind controller. spawn gets their cells from openPair.
func familyGame(
	t *testing.T,
	mutate func(*content.Bundle),
	minded []api.ContentID,
	spawn func(w *sim.World, here, there api.Cell) error,
) (*game.Game, *sim.World) {
	t.Helper()
	var base api.Cell
	g, w, _ := gametest.New(t, 1, func(b *content.Bundle) {
		for _, id := range minded {
			template := b.Templates[string(id)]
			template.Actor = &content.ActorDef{Controller: agents.MindController}
			b.Templates[string(id)] = template
		}
		b.Agents.ThinkTicks = 2
		if mutate != nil {
			mutate(b)
		}
	}, func(cfg *game.Config) {
		cfg.Spawn = func(w *sim.World, _ *clock.Clock) error {
			here, there := openPair(t, w)
			base = here
			return spawn(w, here, there)
		}
	})
	query := ecs.NewFilter2[comp.MemberOf, comp.Pos](w.ECS).Query()
	for query.Next() {
		_, pos := query.Get()
		pos.At = base
		pos.At.Hex.R += 24
	}
	return g, w
}

func waitFor(t *testing.T, g *game.Game, limit int, what string, done func(seen []api.Event) bool) {
	t.Helper()
	var seen []api.Event
	for range limit {
		if done(seen) {
			return
		}
		events, err := g.Apply(api.Wait{Source: g.Current(), Until: api.WaitEvent})
		if err != nil {
			t.Fatal(err)
		}
		seen = append(seen, events...)
	}
	if !done(seen) {
		t.Fatalf("after %d waits: %s", limit, what)
	}
}

func TestMovingClayWalksToACorpseItSawAndConsumesIt(t *testing.T) {
	var clay, beast api.UID
	g, w := familyGame(t, nil, []api.ContentID{clayTemplate}, func(w *sim.World, here, there api.Cell) error {
		body, err := w.Spawn(clayTemplate, &comp.Pos{At: here})
		if err != nil {
			return err
		}
		clay = w.UIDOf(body)
		other, err := w.Spawn(beastTemplate, &comp.Pos{At: there})
		beast = w.UIDOf(other)
		return err
	})
	corpse, _ := w.ByUID(beast)
	kill := contenttest.Do("damage", "parts", "self:torso", "amount", "1000")
	if err := effects.Apply(&effects.Context{World: w, Source: corpse}, []content.EffectRef{kill}); err != nil {
		t.Fatal(err)
	}
	if _, err := w.Bus.Drain(w); err != nil {
		t.Fatal(err)
	}
	body, _ := w.ByUID(clay)
	ecs.NewMap[comp.Drives](w.ECS).Get(body).Levels["drive.the_dead"] = 4
	waitFor(t, g, 300, "the corpse is still there", func([]api.Event) bool {
		_, there := w.ByUID(beast)
		return !there
	})
	if level := ecs.NewMap[comp.Drives](w.ECS).Get(body).Levels["drive.the_dead"]; level >= 4 {
		t.Errorf("want of the dead is still %v after consuming", level)
	}
}

// farCell finds a standable cell at the same depth, 3 to 5 hexes from the cell, with a route to it.
func farCell(tb testing.TB, w *sim.World, from api.Cell) api.Cell {
	tb.Helper()
	rules := actions.StepRules(w)
	for _, cell := range w.Map.CellsIn(from.Hex, 5, from.Depth, from.Depth) {
		if hexmap.CellDistance(from, cell) < 3 {
			continue
		}
		if ok, _ := w.Map.Standable(cell); !ok {
			continue
		}
		if _, found, err := w.Map.Path(from, cell, rules, nil, 200); err == nil && found {
			return cell
		}
	}
	tb.Fatal("no cell to walk to")
	return api.Cell{}
}

func TestARemnantMovesToGroundItBelievesUndecayed(t *testing.T) {
	var remnant api.UID
	var promised api.Cell
	g, w := familyGame(t, func(b *content.Bundle) {
		b.World.Geography = "geography.deep"
		geography := b.Geographies["geography.deep"]
		decay := *geography.Decay
		decay.Level = "1"
		geography.Decay = &decay
		b.Geographies["geography.deep"] = geography
	}, []api.ContentID{remnantTemplate}, func(w *sim.World, here, _ api.Cell) error {
		promised = farCell(t, w, here)
		body, err := w.Spawn(remnantTemplate, &comp.Pos{At: here})
		remnant = w.UIDOf(body)
		return err
	})
	body, _ := w.ByUID(remnant)
	cfg := w.Content.Agents()
	knowledge.Learn(w, body, api.Fact{
		Cell: &promised, Claim: cfg.GroundClaim, Value: "0", Origin: cfg.SeenOrigin, Confidence: 1,
	})
	ecs.NewMap[comp.Drives](w.ECS).Get(body).Levels["drive.old_ground"] = 3
	waitFor(t, g, 300, "the remnant never reached the ground it believed undecayed", func([]api.Event) bool {
		return ecs.NewMap[comp.Pos](w.ECS).Get(body).At == promised
	})
}

func TestHostileFamiliesFightWhenTheyMeet(t *testing.T) {
	var clay, beast api.UID
	g, _ := familyGame(t, nil, nil, func(w *sim.World, here, there api.Cell) error {
		body, err := w.Spawn(clayTemplate, &comp.Pos{At: here})
		if err != nil {
			return err
		}
		clay = w.UIDOf(body)
		other, err := w.Spawn(beastTemplate, &comp.Pos{At: there})
		beast = w.UIDOf(other)
		return err
	})
	waitFor(t, g, 200, "no clay and beast swung at each other", func(seen []api.Event) bool {
		for _, e := range seen {
			if started, ok := e.(api.ActionStarted); ok && started.Action == "action.strike" &&
				(started.Source == clay && started.Target == beast || started.Source == beast && started.Target == clay) {
				return true
			}
		}
		return false
	})
}
