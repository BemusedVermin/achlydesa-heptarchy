package agents_test

import (
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/actions"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/agents"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/clock"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game/gametest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/inventory"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/knowledge"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

const (
	porter        = api.ContentID("template.test.porter")
	porterPlanner = "test.porter"
	takeOp        = api.ContentID("operator.take")
	drinkOp       = api.ContentID("operator.consume")
	waterItem     = api.ContentID("item.water")
)

var porterAt = api.Cell{Hex: api.Hex{Q: 5, R: 40}, Depth: 3}

func porterContent(b *content.Bundle) {
	b.Templates[string(porter)] = content.Template{
		Name:    new(api.ContentID("text.template.test.porter")),
		Actor:   &content.ActorDef{Controller: agents.MindController},
		Mind:    &content.MindDef{Planner: porterPlanner},
		Body:    new(api.ContentID("body.human")),
		Actions: []api.ContentID{"action.take", "action.put", "action.consume"},
		Faction: new(api.ContentID("faction.party")),
		Drives:  map[string]float64{string(water): 1},
	}
	b.Text["text.template.test.porter"] = "Porter"
	well := b.Templates["template.source.well"]
	well.Source.Refill = "0"
	b.Templates["template.source.well"] = well
	b.Agents.ThinkTicks = 1 << 40
}

// porterPlan plans the steps it is given, aimed at the well, claiming what a take claims.
type porterPlan struct {
	fx    *porterFixture
	steps []api.ContentID
}

func (p porterPlan) Plan(_ agents.Snapshot, goal api.ContentID) (agents.Plan, bool) {
	plan := agents.Plan{Goal: goal}
	for _, op := range p.steps {
		step := agents.Step{Operator: op, Target: &agents.Target{Subject: p.fx.well, At: porterAt}}
		plan.Steps = append(plan.Steps, step)
		if claim, ok, err := agents.TakeClaim(p.fx.w, step); err == nil && ok {
			plan.Claims = append(plan.Claims, claim)
		}
	}
	return plan, true
}

type porterFixture struct {
	t       *testing.T
	g       *game.Game
	w       *sim.World
	well    api.UID
	porters []api.UID
}

// newPorterFixture is a well holding stock water with porters beside it, each set to run steps.
func newPorterFixture(t *testing.T, porters, stock int, steps ...api.ContentID) *porterFixture {
	t.Helper()
	fx := &porterFixture{t: t}
	fx.g, fx.w, _ = gametest.New(t, 1, porterContent, func(cfg *game.Config) {
		cfg.Install = func(w *sim.World, _ *clock.Clock) error {
			return agents.RegisterPlanner(w, porterPlanner, porterPlan{fx: fx, steps: steps})
		}
		cfg.Spawn = func(w *sim.World, _ *clock.Clock) error {
			well, err := w.Spawn("template.source.well", &comp.Pos{At: porterAt})
			if err != nil {
				return err
			}
			ecs.NewMap[comp.Source](w.ECS).Get(well).Stock = stock
			fx.well = w.UIDOf(well)
			for range porters {
				body, err := w.Spawn(porter, &comp.Pos{At: porterAt})
				if err != nil {
					return err
				}
				ecs.NewMap[comp.Mind](w.ECS).Get(body).Goal = "goal.test.porter"
				cfg := w.Content.Agents()
				for claim, value := range map[api.ContentID]string{
					cfg.GivesClaim: "affordance.water",
					cfg.AtClaim:    knowledge.CellValue(porterAt),
				} {
					knowledge.Learn(w, body, api.Fact{
						Subject: fx.well, Claim: claim, Value: value, Origin: cfg.SeenOrigin, Confidence: 1,
					})
				}
				fx.porters = append(fx.porters, w.UIDOf(body))
			}
			return nil
		}
	})
	return fx
}

func (fx *porterFixture) entity(uid api.UID) ecs.Entity {
	fx.t.Helper()
	e, ok := fx.w.ByUID(uid)
	if !ok {
		fx.t.Fatalf("no entity %d", uid)
	}
	return e
}

func (fx *porterFixture) stock() int {
	fx.t.Helper()
	return ecs.NewMap[comp.Source](fx.w.ECS).Get(fx.entity(fx.well)).Stock
}

func (fx *porterFixture) carried(uid api.UID) int {
	fx.t.Helper()
	return inventory.Count(fx.w, fx.entity(uid), waterItem, 0)
}

func (fx *porterFixture) wait(times int) {
	fx.t.Helper()
	for range times {
		if _, err := fx.g.Apply(api.Wait{Source: fx.g.Current(), Until: api.WaitEvent}); err != nil {
			fx.t.Fatal(err)
		}
	}
}

func TestABodyTakesFromASourceUntilItIsEmpty(t *testing.T) {
	fx := newPorterFixture(t, 1, 3, takeOp)
	fx.wait(60)
	if got := fx.stock(); got != 0 {
		t.Errorf("well stock = %d, want it emptied", got)
	}
	if got := fx.carried(fx.porters[0]); got != 3 {
		t.Errorf("porter carries %d water, want the well's 3", got)
	}
}

func TestTwoAgentsCannotClaimTheSameLastUnit(t *testing.T) {
	fx := newPorterFixture(t, 2, 1, takeOp)
	fx.wait(20)
	first, second := fx.porters[0], fx.porters[1]
	if got := fx.carried(first); got != 1 {
		t.Errorf("lower UID carries %d, want the last unit", got)
	}
	if got := fx.carried(second); got != 0 {
		t.Errorf("higher UID carries %d, want none", got)
	}
	if claims := ecs.NewMap[comp.Claims](fx.w.ECS); claims.Has(fx.entity(second)) {
		t.Errorf("higher UID holds claims %v on a stock already claimed", claims.Get(fx.entity(second)).Held)
	}
}

func TestAPutIntoAFullContainerIsRefusedWithAReason(t *testing.T) {
	fx := newPorterFixture(t, 1, 20, takeOp)
	body := fx.entity(fx.porters[0])
	inventory.Add(fx.w, body, comp.ItemStack{Item: waterItem, Count: 1})
	put := actions.Intent{
		Actor:   body,
		Action:  "action.put",
		Target:  fx.entity(fx.well),
		Choices: map[string]string{"carried": string(waterItem)},
	}
	ok, reason := actions.Validate(fx.w, put)
	if ok || reason.ID != "text.actions.no_room" {
		t.Errorf("put into a full well: ok = %v, reason = %q; want refused as no_room", ok, reason.ID)
	}
	if got := fx.w.Content.Render(reason); got == "" || got[0] == '[' {
		t.Errorf("refusal renders as %q, want text", got)
	}
}

func TestConsumingWaterMeetsThirstThroughTheItem(t *testing.T) {
	fx := newPorterFixture(t, 1, 0, drinkOp)
	body := fx.entity(fx.porters[0])
	inventory.Add(fx.w, body, comp.ItemStack{Item: waterItem, Count: 1})
	levels := ecs.NewMap[comp.Drives](fx.w.ECS).Get(body).Levels
	levels[water] = 20
	fx.wait(20)
	if got := fx.carried(fx.porters[0]); got != 0 {
		t.Errorf("porter still carries %d water", got)
	}
	if got := levels[water]; got >= 20 {
		t.Errorf("thirst = %v after drinking, want it lowered from 20", got)
	}
}
