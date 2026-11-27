package agents_test

import (
	"slices"
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
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/knowledge"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

const (
	walkerPlanner = "test.walker"
	goToOp        = api.ContentID("operator.go_to")
	wanderOp      = api.ContentID("operator.wander")
)

var walkStart = api.Cell{Hex: api.Hex{Q: 5, R: 40}, Depth: 0}

type wanderPlanner struct{}

func (wanderPlanner) Plan(_ agents.Snapshot, goal api.ContentID) (agents.Plan, bool) {
	return agents.Plan{Goal: goal, Steps: []agents.Step{{Operator: wanderOp}}}, true
}

type goToPlanner struct {
	target api.Cell
	well   *api.UID
}

func (p goToPlanner) Plan(_ agents.Snapshot, goal api.ContentID) (agents.Plan, bool) {
	step := agents.Step{Operator: goToOp, Target: &agents.Target{Subject: *p.well, At: p.target}}
	return agents.Plan{Goal: goal, Steps: []agents.Step{step}}, true
}

func walkingContent(b *content.Bundle) {
	b.Templates["template.test.walker"] = content.Template{
		Name:    new(api.ContentID("text.template.test.walker")),
		Actor:   &content.ActorDef{Controller: agents.MindController},
		Mind:    &content.MindDef{Planner: walkerPlanner},
		Actions: []api.ContentID{"action.step"},
		Faction: new(api.ContentID("faction.party")),
	}
	b.Text["text.template.test.walker"] = "Walker"
}

type walkFixture struct {
	t      *testing.T
	g      *game.Game
	w      *sim.World
	walker api.UID
}

// newWalkFixture is a walker at walkStart on an open floor, with a standing body in each blocker cell.
func newWalkFixture(t *testing.T, seed uint64, plan agents.Planner, well *api.UID, blockers ...api.Cell) *walkFixture {
	t.Helper()
	fx := &walkFixture{t: t}
	fx.g, fx.w, _ = gametest.New(t, seed, walkingContent, func(cfg *game.Config) {
		cfg.Install = func(w *sim.World, _ *clock.Clock) error {
			return agents.RegisterPlanner(w, walkerPlanner, plan)
		}
		cfg.Spawn = func(w *sim.World, _ *clock.Clock) error {
			w.Map = hexmap.NewMap(hexmap.FuncSource(func(c api.Cell) hexmap.Terrain {
				if c.Depth > 0 {
					return hexmap.Terrain{ID: "terrain.packed_clay", Solid: true}
				}
				return hexmap.Terrain{ID: "terrain.air"}
			}), hexmap.Prisms{})
			for _, cell := range blockers {
				blocker := w.NewIdentified("template.test.blocker")
				ecs.NewMap[comp.Pos](w.ECS).Add(blocker, &comp.Pos{At: cell})
				ecs.NewMap[comp.Actor](w.ECS).Add(blocker, &comp.Actor{Controller: "idle"})
			}
			spring := walkStart
			spring.Hex.Q += 5
			if well != nil {
				e, err := w.Spawn("template.source.well", &comp.Pos{At: spring})
				if err != nil {
					return err
				}
				*well = w.UIDOf(e)
			}
			body, err := w.Spawn("template.test.walker", &comp.Pos{At: walkStart})
			if err != nil {
				return err
			}
			if well != nil {
				cfg := w.Content.Agents()
				for claim, value := range map[api.ContentID]string{
					cfg.GivesClaim: "affordance.water",
					cfg.AtClaim:    knowledge.CellValue(spring),
				} {
					knowledge.Learn(w, body, api.Fact{
						Subject: *well, Claim: claim, Value: value, Origin: cfg.SeenOrigin, Confidence: 1,
					})
				}
			}
			ecs.NewMap[comp.Mind](w.ECS).Get(body).Goal = "goal.test.walk"
			fx.walker = w.UIDOf(body)
			return nil
		}
	})
	return fx
}

func (fx *walkFixture) at() api.Cell {
	fx.t.Helper()
	e, ok := fx.w.ByUID(fx.walker)
	if !ok {
		fx.t.Fatal("the walker is gone")
	}
	return ecs.NewMap[comp.Pos](fx.w.ECS).Get(e).At
}

// walk waits until the walker stands where stop says or limit waits pass, and returns each cell it
// stood in, in order.
func (fx *walkFixture) walk(limit int, stop func(api.Cell) bool) []api.Cell {
	fx.t.Helper()
	trail := []api.Cell{fx.at()}
	for range limit {
		if stop(fx.at()) {
			break
		}
		if _, err := fx.g.Apply(api.Wait{Source: fx.g.Current(), Until: api.WaitEvent}); err != nil {
			fx.t.Fatal(err)
		}
		if now := fx.at(); now != trail[len(trail)-1] {
			trail = append(trail, now)
		}
	}
	return trail
}

func TestAnAgentWalksToABelievedCellAroundAStandingBody(t *testing.T) {
	var well api.UID
	target, blocker := walkStart, walkStart
	target.Hex.Q += 5
	blocker.Hex.Q += 2
	fx := newWalkFixture(t, 1, goToPlanner{target: target, well: &well}, &well, blocker)
	trail := fx.walk(400, func(c api.Cell) bool { return hexmap.CellDistance(c, target) <= 1 })
	if got := trail[len(trail)-1]; hexmap.CellDistance(got, target) > 1 {
		t.Fatalf("walker ended at %v, want next to %v; trail %v", got, target, trail)
	}
	if slices.Contains(trail, blocker) {
		t.Errorf("walker went through the standing body at %v: %v", blocker, trail)
	}
}

func TestAnAgentWithNothingKnownWandersTheSameWayForTheSameSeed(t *testing.T) {
	const moves = 12
	wander := func(seed uint64) []api.Cell {
		fx := newWalkFixture(t, seed, wanderPlanner{}, nil)
		trail := []api.Cell{fx.at()}
		for range 400 {
			if len(trail) > moves {
				break
			}
			fx.walk(1, func(api.Cell) bool { return false })
			if now := fx.at(); now != trail[len(trail)-1] {
				trail = append(trail, now)
			}
		}
		return trail
	}
	first := wander(1)
	if len(first) <= moves {
		t.Fatalf("walker made %d moves in 400 waits, want %d", len(first)-1, moves)
	}
	for i := 2; i < len(first); i++ {
		if first[i] == first[i-2] {
			t.Errorf("walker went straight back at move %d: %v", i, first)
		}
	}
	if again := wander(1); !slices.Equal(first, again) {
		t.Errorf("the same seed wandered %v then %v", first, again)
	}
}
