package agents_test

import (
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/agents"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/clock"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game/gametest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

const (
	fighter    = api.ContentID("template.test.fighter")
	foe        = api.ContentID("template.test.foe")
	safety     = "drive.safety"
	fightPlans = "test.fight"
)

// fightPlanner plans one engage for the fight goal and one flee for the flee goal.
type fightPlanner struct{}

func (fightPlanner) Plan(_ agents.Snapshot, goal api.ContentID) (agents.Plan, bool) {
	switch goal {
	case "goal.fight":
		return agents.Plan{Goal: goal, Steps: []agents.Step{{Operator: "operator.engage"}}}, true
	case "goal.flee":
		return agents.Plan{Goal: goal, Steps: []agents.Step{{Operator: "operator.flee"}}}, true
	}
	return agents.Plan{}, false
}

type fightFixture struct {
	t       *testing.T
	g       *game.Game
	w       *sim.World
	fighter api.UID
	foe     api.UID
}

func newFightFixture(t *testing.T, terror float64) *fightFixture {
	t.Helper()
	fx := &fightFixture{t: t}
	fx.g, fx.w, _ = gametest.New(t, 1, func(b *content.Bundle) {
		clay := b.Templates["template.creature.moving_clay"]
		mind := clay
		mind.Actor = &content.ActorDef{Controller: agents.MindController}
		mind.Mind = &content.MindDef{Planner: fightPlans}
		mind.Drives = map[string]float64{safety: 2}
		mind.Triggers = append([]content.Trigger{{
			On: "combat.part_damaged", Role: content.RoleTarget,
			Do: []content.EffectRef{{
				Do: "satisfy", Params: map[string]string{"bodies": "target", "drive": safety, "amount": "-2"},
			}},
		}}, clay.Triggers...)
		b.Templates[string(fighter)] = mind
		enemy := clay
		enemy.Actor = &content.ActorDef{Controller: timeline.IdleController}
		enemy.Faction = new(api.ContentID("faction.party"))
		b.Templates[string(foe)] = enemy
		b.Agents.ThinkTicks = 2
		b.Text["text.template.test.fighter"], b.Text["text.template.test.foe"] = "Fighter", "Foe"
		mind.Name, enemy.Name = new(
			api.ContentID("text.template.test.fighter"),
		), new(
			api.ContentID("text.template.test.foe"),
		)
		b.Templates[string(fighter)], b.Templates[string(foe)] = mind, enemy
	}, func(cfg *game.Config) {
		cfg.Install = func(w *sim.World, _ *clock.Clock) error {
			return agents.RegisterPlanner(w, fightPlans, fightPlanner{})
		}
		cfg.Spawn = func(w *sim.World, _ *clock.Clock) error {
			here, there := openPair(t, w)
			body, err := w.Spawn(fighter, &comp.Pos{At: here})
			if err != nil {
				return err
			}
			ecs.NewMap[comp.Drives](w.ECS).Get(body).Levels[safety] = terror
			fx.fighter = w.UIDOf(body)
			other, err := w.Spawn(foe, &comp.Pos{At: there})
			fx.foe = w.UIDOf(other)
			return err
		}
	})
	query := ecs.NewFilter2[comp.MemberOf, comp.Pos](fx.w.ECS).Query()
	for query.Next() {
		_, pos := query.Get()
		pos.At = ecs.NewMap[comp.Pos](fx.w.ECS).Get(fx.entity(fx.fighter)).At
		pos.At.Hex.R += 24
	}
	return fx
}

func (fx *fightFixture) entity(uid api.UID) ecs.Entity {
	fx.t.Helper()
	e, ok := fx.w.ByUID(uid)
	if !ok {
		fx.t.Fatalf("no entity %d", uid)
	}
	return e
}

func (fx *fightFixture) controller() string {
	return timeline.ControllerOf(fx.w, fx.entity(fx.fighter))
}

func (fx *fightFixture) gap() int {
	positions := ecs.NewMap[comp.Pos](fx.w.ECS)
	return hexmap.CellDistance(positions.Get(fx.entity(fx.fighter)).At, positions.Get(fx.entity(fx.foe)).At)
}

func (fx *fightFixture) goal() api.ContentID {
	return ecs.NewMap[comp.Mind](fx.w.ECS).Get(fx.entity(fx.fighter)).Goal
}

// run waits up to limit times, collecting what the fighter did, and stops once stop holds.
func (fx *fightFixture) run(limit int, stop func() bool) (struck bool) {
	fx.t.Helper()
	for range limit {
		if stop() {
			return struck
		}
		events, err := fx.g.Apply(api.Wait{Source: fx.g.Current(), Until: api.WaitEvent})
		if err != nil {
			fx.t.Fatal(err)
		}
		for _, e := range events {
			if started, ok := e.(api.ActionStarted); ok && started.Source == fx.fighter &&
				started.Action == "action.strike" {
				struck = true
			}
		}
	}
	return struck
}

func TestAMindWithAHostileInSightHandsTheBodyToTheHunter(t *testing.T) {
	fx := newFightFixture(t, 0)
	struck := fx.run(200, func() bool { return fx.controller() == agents.FightController })
	if fx.controller() != agents.FightController {
		t.Fatalf("controller = %q, want %q", fx.controller(), agents.FightController)
	}
	fx.run(200, func() bool { return struck })
	if fx.goal() != "goal.fight" {
		t.Errorf("goal = %q, want goal.fight", fx.goal())
	}
}

func TestATerrifiedMindRunsInsteadOfFighting(t *testing.T) {
	fx := newFightFixture(t, 8)
	start := fx.gap()
	struck := fx.run(60, func() bool { return fx.gap() > start+2 })
	if fx.gap() <= start {
		t.Fatalf("gap went from %d to %d, want the fighter farther away", start, fx.gap())
	}
	if struck || fx.controller() != agents.MindController || fx.goal() != "goal.flee" {
		t.Errorf("struck %v, controller %q, goal %q; want a flight with no strike", struck, fx.controller(), fx.goal())
	}
}

func TestTheMindDecidesAgainWhenTheFightIsOver(t *testing.T) {
	fx := newFightFixture(t, 0)
	fx.run(200, func() bool { return fx.controller() == agents.FightController })
	if fx.controller() != agents.FightController {
		t.Fatal("the mind never handed over the fight")
	}
	sim.Put(fx.w, fx.entity(fx.foe), comp.Dead{Tick: fx.w.Now})
	fx.run(200, func() bool { return fx.controller() == agents.MindController && fx.goal() != "goal.fight" })
	if fx.controller() != agents.MindController || fx.goal() == "goal.fight" {
		t.Errorf("after the fight: controller %q, goal %q; want the mind with another goal", fx.controller(), fx.goal())
	}
}

// openPair finds two standable cells two hexes apart that see each other, with room to back away.
func openPair(tb testing.TB, w *sim.World) (api.Cell, api.Cell) {
	tb.Helper()
	for depth := range 8 {
		for q := range 30 {
			for r := 20; r < 60; r++ {
				a := api.Cell{Hex: api.Hex{Q: q, R: r}, Depth: depth}
				b := api.Cell{Hex: api.Hex{Q: q + 2, R: r}, Depth: depth}
				c := api.Cell{Hex: api.Hex{Q: q - 2, R: r}, Depth: depth}
				okA, _ := w.Map.Standable(a)
				okB, _ := w.Map.Standable(b)
				okC, _ := w.Map.Standable(c)
				seen, _ := w.Map.Sees(a, b, 8, 6)
				if okA && okB && okC && seen {
					return a, b
				}
			}
		}
	}
	tb.Fatal("no open ground")
	return api.Cell{}, api.Cell{}
}
