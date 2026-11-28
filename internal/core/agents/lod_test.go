package agents_test

import (
	"slices"
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
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/inventory"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/knowledge"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/worldgen"
)

const (
	farPlanner     = "test.far"
	carryPlanner   = "test.far.carry"
	farCarrier     = api.ContentID("template.test.far.carrier")
	farAgent       = api.ContentID("template.test.far")
	farDrinkOp     = api.ContentID("operator.test.drink")
	farDrinkAction = api.ContentID("action.test.drink")
	waterMark      = api.ContentID("affordance.water")
	drinkGoal      = api.ContentID("goal.drink")
)

var (
	partyAt  = api.Cell{Hex: api.Hex{Q: 5, R: 5}}
	farStart = api.Cell{Hex: api.Hex{Q: 200, R: 200}}
	farWell  = api.Cell{Hex: api.Hex{Q: 260, R: 200}}
	farBed   = api.Cell{Hex: api.Hex{Q: 262, R: 204}}
)

// drinkPlanner drinks when it believes water is near, and walks to the well first when it only knows of
// it. A carrying planner drinks by taking water from a well it does not name, then consuming it.
type drinkPlanner struct {
	well  *api.UID
	carry bool
}

func (p drinkPlanner) Plan(snapshot agents.Snapshot, goal api.ContentID) (agents.Plan, bool, error) {
	if goal != drinkGoal {
		return agents.Plan{}, false, nil
	}
	drink := []agents.Step{{Operator: farDrinkOp}}
	if p.carry {
		drink = []agents.Step{{Operator: takeOp}, {Operator: drinkOp}}
	}
	switch {
	case slices.Contains(snapshot.Beliefs, agents.Fact{Kind: content.FactNear, Tag: waterMark}):
		return agents.Plan{Steps: drink}, true, nil
	case slices.Contains(snapshot.Beliefs, agents.Fact{Kind: content.FactKnown, Tag: waterMark}):
		walk := agents.Step{Operator: goToOp, Target: &agents.Target{Subject: *p.well, At: farWell}}
		return agents.Plan{Steps: append([]agents.Step{walk}, drink...)}, true, nil
	}
	return agents.Plan{}, false, nil
}

func farContent(b *content.Bundle) {
	b.Actions[string(farDrinkAction)] = content.ActionDef{
		Name: "text.action.test.drink", Key: "drink", Uses: []content.Requirement{{Tag: "body", Count: 1}},
		Target: content.TargetSelf, Windup: "0", Recovery: "1",
		OnLand: []content.EffectRef{{Do: "satisfy", Params: map[string]string{
			"bodies": "self", "drive": string(water), "amount": "100",
		}}},
	}
	b.Operators[string(farDrinkOp)] = content.Operator{
		Pre:  []content.Condition{{Fact: content.FactNear, Tag: waterMark}},
		Eff:  []content.Condition{{Fact: content.FactDriveMet, Tag: water}},
		Cost: 1, Does: content.Does{Action: farDrinkAction, Target: agents.SelfPicker}, Kind: "test",
	}
	b.Templates[string(farAgent)] = content.Template{
		Name:    new(api.ContentID("text.template.test.far")),
		Actor:   &content.ActorDef{Controller: agents.MindController},
		Mind:    &content.MindDef{Planner: farPlanner},
		Actions: []api.ContentID{"action.step", farDrinkAction},
		Faction: new(api.ContentID("faction.party")),
		Drives:  map[string]float64{string(water): 1},
	}
	b.Text["text.action.test.drink"] = "Drink"
	b.Text["text.template.test.far"] = "Traveller"
	b.Templates[string(farCarrier)] = content.Template{
		Name:    new(api.ContentID("text.template.test.far.carrier")),
		Actor:   &content.ActorDef{Controller: agents.MindController},
		Mind:    &content.MindDef{Planner: carryPlanner},
		Body:    new(api.ContentID("body.human")),
		Actions: []api.ContentID{"action.step", "action.take", "action.consume"},
		Faction: new(api.ContentID("faction.party")),
		Drives:  map[string]float64{string(water): 1},
	}
	b.Text["text.template.test.far.carrier"] = "Carrier"
	thirst := b.Drives[string(water)]
	thirst.Thresholds = nil
	b.Drives[string(water)] = thirst
	well := b.Templates["template.source.well"]
	well.Source.Refill = "0"
	b.Templates["template.source.well"] = well
}

type farFixture struct {
	tb             testing.TB
	g              *game.Game
	w              *sim.World
	cfg            game.Config
	agent          api.UID
	well, farOther api.UID
}

// floorGround is an open floor at depth 0 over solid clay.
func floorGround() hexmap.Source {
	return hexmap.FuncSource(func(c api.Cell) hexmap.Terrain {
		if c.Depth > 0 {
			return hexmap.Terrain{ID: "terrain.packed_clay", Solid: true}
		}
		return hexmap.Terrain{ID: "terrain.air"}
	})
}

// newFarFixture is the party at partyAt and a thirsty agent of the template at farStart that knows of the
// well at farWell, with a second well at farBed it has never seen.
func newFarFixture(tb testing.TB, seed uint64, template api.ContentID) *farFixture {
	tb.Helper()
	fx := &farFixture{tb: tb}
	fx.g, fx.w, fx.cfg = gametest.New(tb, seed, farContent, func(cfg *game.Config) {
		cfg.Install = func(w *sim.World, _ *clock.Clock) error {
			setGround(w, floorGround())
			if err := agents.RegisterPlanner(w, farPlanner, drinkPlanner{well: &fx.well}); err != nil {
				return err
			}
			return agents.RegisterPlanner(w, carryPlanner, drinkPlanner{well: &fx.well, carry: true})
		}
		cfg.Spawn = func(w *sim.World, _ *clock.Clock) error {
			placeParty(w, partyAt)
			well, err := w.Spawn("template.source.well", &comp.Pos{At: farWell})
			if err != nil {
				return err
			}
			fx.well = w.UIDOf(well)
			other, err := w.Spawn("template.source.well", &comp.Pos{At: farBed})
			if err != nil {
				return err
			}
			fx.farOther = w.UIDOf(other)
			body, err := w.Spawn(template, &comp.Pos{At: farStart})
			if err != nil {
				return err
			}
			fx.agent = w.UIDOf(body)
			teach(w, body, fx.well, farWell)
			ecs.NewMap[comp.Drives](w.ECS).Get(body).Levels[water] = 10
			return nil
		}
	})
	return fx
}

// setGround makes the ground the world's map and chunk graph, over a world wider than the fixture's.
func setGround(w *sim.World, ground hexmap.Source) {
	w.Map = hexmap.NewMap(ground, hexmap.Prisms{})
	graph := ecs.NewResource[worldgen.Graph](w.ECS)
	graph.Remove()
	graph.Add(worldgen.NewGraph(worldgen.GraphConfig{
		Source: ground,
		Bounds: hexmap.Bounds{MaxCol: 599, MaxRow: 399, MaxDepth: 7},
		Edge:   hexmap.Terrain{ID: "terrain.packed_clay", Solid: true},
		Rules:  actions.StepRules(w),
	}))
}

// placeParty puts every party member at the cell.
func placeParty(w *sim.World, at api.Cell) {
	query := ecs.NewFilter2[comp.MemberOf, comp.Pos](w.ECS).Query()
	for query.Next() {
		_, pos := query.Get()
		pos.At = at
	}
}

// teach has the body know that the landmark gives water and stands at the cell.
func teach(w *sim.World, body ecs.Entity, landmark api.UID, at api.Cell) {
	cfg := w.Content.Agents()
	for claim, value := range map[api.ContentID]string{cfg.GivesClaim: string(waterMark), cfg.AtClaim: knowledge.CellValue(at)} {
		knowledge.Learn(w, body, api.Fact{
			Subject: landmark, Claim: claim, Value: value, Origin: cfg.SeenOrigin, Confidence: 1,
		})
	}
}

func (fx *farFixture) body() ecs.Entity {
	fx.tb.Helper()
	e, ok := fx.w.ByUID(fx.agent)
	if !ok {
		fx.tb.Fatal("the agent is gone")
	}
	return e
}

func (fx *farFixture) wait(ticks int64) {
	fx.tb.Helper()
	if _, err := fx.g.Apply(api.Wait{Source: fx.g.Current(), Until: api.WaitTicks, Ticks: ticks}); err != nil {
		fx.tb.Fatal(err)
	}
}

func (fx *farFixture) thirst() float64 {
	return ecs.NewMap[comp.Drives](fx.w.ECS).Get(fx.body()).Levels[water]
}

// route waits until the agent is far and on its way, and returns the stops ahead of it.
func (fx *farFixture) route() []comp.FarHop {
	fx.tb.Helper()
	fars := ecs.NewMap[comp.Far](fx.w.ECS)
	for range 20 {
		if fars.Has(fx.body()) && len(fars.Get(fx.body()).Hops) > 0 {
			return slices.Clone(fars.Get(fx.body()).Hops)
		}
		fx.wait(1)
	}
	fx.tb.Fatal("the agent never set off along the chunk graph")
	return nil
}

func TestAFarAgentTravelsTheGraphToAWellItKnowsAndDrinks(t *testing.T) {
	fx := newFarFixture(t, 1, farAgent)
	hops := fx.route()
	if got := hops[len(hops)-1].At; got != farWell {
		t.Fatalf("route ends at %v, want the well at %v", got, farWell)
	}
	for range 100 {
		if fx.thirst() < 1 {
			break
		}
		fx.wait(50)
	}
	if level := fx.thirst(); level >= 1 {
		t.Fatalf("thirst %v after travelling, want it met", level)
	}
	if at := ecs.NewMap[comp.Pos](fx.w.ECS).Get(fx.body()).At; at != farWell {
		t.Errorf("agent drank at %v, want at the well %v", at, farWell)
	}
	if _, ok := knowledge.Best(fx.w, fx.body(), fx.farOther, fx.w.Content.Agents().GivesClaim); !ok {
		t.Error("agent reached the well's chunk without learning of the other well there")
	}
	if !ecs.NewMap[comp.Parked](fx.w.ECS).Has(fx.body()) {
		t.Error("a far agent is on the timeline")
	}
}

func TestTheSameSeedGivesAFarAgentTheSameArrivalTicksSavedOrNot(t *testing.T) {
	fx := newFarFixture(t, 7, farAgent)
	first := fx.route()
	if again := newFarFixture(t, 7, farAgent).route(); !slices.Equal(first, again) {
		t.Fatalf("the same seed routed %v then %v", first, again)
	}
	loaded, _, _ := gametest.Reload(t, fx.cfg, fx.g)
	for _, g := range []*game.Game{fx.g, loaded} {
		for range 30 {
			if _, err := g.Apply(api.Wait{Source: g.Current(), Until: api.WaitTicks, Ticks: 50}); err != nil {
				t.Fatal(err)
			}
		}
	}
	kept, err := fx.g.Hash()
	if err != nil {
		t.Fatal(err)
	}
	reloaded, err := loaded.Hash()
	if err != nil {
		t.Fatal(err)
	}
	if kept != reloaded {
		t.Fatal("a far agent's journey went differently after a save and load")
	}
}

func TestAFarAgentThePartyComesNearStandsOnGroundAndDecidesAsABody(t *testing.T) {
	fx := newFarFixture(t, 1, farAgent)
	fx.route()
	inRock := farStart
	inRock.Depth = 3
	ecs.NewMap[comp.Far](fx.w.ECS).Get(fx.body()).Hops = nil
	ecs.NewMap[comp.Pos](fx.w.ECS).Get(fx.body()).At = inRock
	nearby := farStart
	nearby.Hex.Q -= 4
	placeParty(fx.w, nearby)
	fx.wait(1)
	body := fx.body()
	at := ecs.NewMap[comp.Pos](fx.w.ECS).Get(body).At
	if ok, err := fx.w.Map.Standable(at); err != nil || !ok {
		t.Fatalf("promoted agent stands at %v, standable %v, err %v", at, ok, err)
	}
	if ecs.NewMap[comp.Far](fx.w.ECS).Has(body) || ecs.NewMap[comp.Parked](fx.w.ECS).Has(body) {
		t.Error("a promoted agent is still far or off the timeline")
	}
	if !agents.Near(fx.w, body) {
		t.Error("a promoted agent is not near")
	}
}

func TestAFarAgentTakesWaterFromTheWellItReachedAndDrinksIt(t *testing.T) {
	fx := newFarFixture(t, 1, farCarrier)
	sources := ecs.NewMap[comp.Source](fx.w.ECS)
	stock := func(uid api.UID) int {
		e, ok := fx.w.ByUID(uid)
		if !ok {
			t.Fatalf("well %d is gone", uid)
		}
		return sources.Get(e).Stock
	}
	wellStock := stock(fx.well)
	fx.route()
	for range 100 {
		if fx.thirst() < 1 {
			break
		}
		fx.wait(50)
	}
	if level := fx.thirst(); level >= 1 {
		t.Fatalf("thirst %v after travelling, want it met", level)
	}
	if got := stock(fx.well); got >= wellStock {
		t.Errorf("the well it reached holds %d, want less than its %d", got, wellStock)
	}
	if got := inventory.Count(fx.w, fx.body(), waterItem, 0); got != 0 {
		t.Errorf("agent still carries %d water, want it drunk", got)
	}
	if !ecs.NewMap[comp.Parked](fx.w.ECS).Has(fx.body()) {
		t.Error("the agent drank on the timeline, not far")
	}
}
