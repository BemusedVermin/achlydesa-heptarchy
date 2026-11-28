package agents_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/agents"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/clock"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game/gametest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

const (
	thinker     = api.ContentID("template.test.thinker")
	ponder      = api.ContentID("action.test.ponder")
	ponderOp    = api.ContentID("operator.test.ponder")
	thinkGoal   = api.ContentID("goal.test.think")
	onePlanner  = "test.one"
	planSteps   = 3
	ponderTicks = 3
)

// thinkingContent adds a thinker run by the mind controller, and a ponder action it takes while it is
// thirsty, landing after 2 ticks and recovering for 1.
func thinkingContent(b *content.Bundle) {
	b.Actions[string(ponder)] = content.ActionDef{
		Name: "text.action.test.ponder", Key: "ponder", Uses: []content.Requirement{{Tag: "body", Count: 1}},
		Target: content.TargetSelf, Windup: "2", Recovery: "1",
	}
	if b.Operators == nil {
		b.Operators = map[string]content.Operator{}
	}
	b.Operators[string(ponderOp)] = content.Operator{
		Pre:  []content.Condition{{Fact: content.FactDriveMet, Tag: water, Not: true}},
		Eff:  []content.Condition{{Fact: content.FactDriveMet, Tag: water}},
		Cost: 1, Does: content.Does{Action: ponder, Target: agents.SelfPicker}, Kind: "test",
	}
	b.Templates[string(thinker)] = content.Template{
		Name:    new(api.ContentID("text.template.test.thinker")),
		Actor:   &content.ActorDef{Controller: agents.MindController},
		Mind:    &content.MindDef{Planner: onePlanner},
		Actions: []api.ContentID{ponder},
		Faction: new(api.ContentID("faction.party")),
		Drives:  map[string]float64{string(water): 1, "drive.rest": 1},
	}
	b.Agents.ThinkTicks = 1 << 40
	for id, place := range b.Places {
		place.Spawns = nil
		b.Places[id] = place
	}
	b.Text["text.action.test.ponder"] = "Ponder"
	b.Text["text.template.test.thinker"] = "Thinker"
}

// oneOperator is the trivial planner: three ponder steps while the ponder operator's preconditions
// hold, claiming the whole well.
type oneOperator struct{ fx *mindFixture }

func (p oneOperator) Plan(s agents.Snapshot, goal api.ContentID) (agents.Plan, bool, error) {
	p.fx.calls++
	op, ok := s.Content.Operator(ponderOp)
	if !ok || !agents.Holds(op, s.Beliefs) {
		return agents.Plan{}, false, nil
	}
	return agents.Plan{
		Goal:   goal,
		Steps:  slices.Repeat([]agents.Step{{Operator: ponderOp}}, planSteps),
		Claims: []comp.Claim{{Target: p.fx.well}},
	}, true, nil
}

// mindFixture is a game with a well and thirsty thinkers that all plan to ponder by it.
type mindFixture struct {
	t        *testing.T
	g        *game.Game
	w        *sim.World
	cfg      game.Config
	well     api.UID
	thinkers []api.UID
	calls    int
	starts   []api.ActionStarted
}

func newMindFixture(t *testing.T, thinkers int) *mindFixture {
	t.Helper()
	fx := &mindFixture{t: t}
	fx.g, fx.w, fx.cfg = gametest.New(t, 1, thinkingContent, func(cfg *game.Config) {
		cfg.Install = func(w *sim.World, _ *clock.Clock) error {
			return agents.RegisterPlanner(w, onePlanner, oneOperator{fx})
		}
		cfg.Spawn = func(w *sim.World, _ *clock.Clock) error { return fx.spawn(w, thinkers) }
	})
	return fx
}

func (fx *mindFixture) spawn(w *sim.World, thinkers int) error {
	at := &comp.Pos{At: api.Cell{Hex: api.Hex{Q: 5, R: 40}, Depth: 3}}
	well, err := w.Spawn("template.source.well", at)
	if err != nil {
		return err
	}
	fx.well = w.UIDOf(well)
	for range thinkers {
		body, err := w.Spawn(thinker, at)
		if err != nil {
			return err
		}
		ecs.NewMap[comp.Drives](w.ECS).Get(body).Levels[water] = 3
		ecs.NewMap[comp.Mind](w.ECS).Get(body).Goal = thinkGoal
		fx.thinkers = append(fx.thinkers, w.UIDOf(body))
	}
	return nil
}

func (fx *mindFixture) entity(uid api.UID) ecs.Entity {
	fx.t.Helper()
	e, ok := fx.w.ByUID(uid)
	if !ok {
		fx.t.Fatalf("no entity %d", uid)
	}
	return e
}

func (fx *mindFixture) mind(uid api.UID) comp.Mind {
	fx.t.Helper()
	return *ecs.NewMap[comp.Mind](fx.w.ECS).Get(fx.entity(uid))
}

func (fx *mindFixture) claims(uid api.UID) []comp.Claim {
	fx.t.Helper()
	claims := ecs.NewMap[comp.Claims](fx.w.ECS)
	if e := fx.entity(uid); claims.Has(e) {
		return claims.Get(e).Held
	}
	return nil
}

// wait lets the deciding body wait for the next entry, and records the ponders that started.
func (fx *mindFixture) wait() {
	fx.t.Helper()
	fx.starts = append(fx.starts, waitOnce(fx.t, fx.g)...)
}

// until waits, at most limit times, until stop holds, and fails the test if it never does.
func (fx *mindFixture) until(limit int, what string, stop func() bool) {
	fx.t.Helper()
	for range limit {
		if stop() {
			return
		}
		fx.wait()
	}
	if !stop() {
		fx.t.Fatalf("after %d waits, still not %s", limit, what)
	}
}

func waitOnce(tb testing.TB, g *game.Game) []api.ActionStarted {
	tb.Helper()
	events, err := g.Apply(api.Wait{Source: g.Current(), Until: api.WaitEvent})
	if err != nil {
		tb.Fatal(err)
	}
	var starts []api.ActionStarted
	for _, e := range events {
		if started, ok := e.(api.ActionStarted); ok && started.Action == ponder {
			starts = append(starts, started)
		}
	}
	return starts
}

func TestAPlanRunsStepByStepAndIsMadeAgainWhenDone(t *testing.T) {
	fx := newMindFixture(t, 1)
	fx.until(40, "four ponders started", func() bool { return len(fx.starts) >= planSteps+1 })
	for i := 1; i < planSteps; i++ {
		if gap := fx.starts[i].Tick - fx.starts[i-1].Tick; gap != ponderTicks {
			t.Errorf("step %d started %d ticks after the one before, want %d", i, gap, ponderTicks)
		}
	}
	if fx.calls != 2 {
		t.Errorf("the planner was asked %d times by the first step of the second plan, want 2", fx.calls)
	}
	if mind := fx.mind(fx.thinkers[0]); mind.Step != 1 || len(mind.Plan.Steps) != planSteps {
		t.Errorf("mind is at step %d of %d, want 1 of %d", mind.Step, len(mind.Plan.Steps), planSteps)
	}
	if got, want := fx.claims(fx.thinkers[0]), []comp.Claim{{Target: fx.well}}; !slices.Equal(got, want) {
		t.Errorf("claims = %v, want %v", got, want)
	}
}

func TestAPlanThatStopsHoldingIsDropped(t *testing.T) {
	cases := []struct {
		name     string
		breakIt  func(fx *mindFixture)
		replaced bool
	}{
		{"a lost claim", func(fx *mindFixture) { fx.w.Despawn(fx.entity(fx.well)) }, false},
		{"an export", func(fx *mindFixture) { fx.w.ExportCount++ }, true},
		{"a contradicted belief", func(fx *mindFixture) {
			ecs.NewMap[comp.Drives](fx.w.ECS).Get(fx.entity(fx.thinkers[0])).Levels[water] = 0
		}, false},
		{"a changed goal", func(fx *mindFixture) {
			ecs.NewMap[comp.Mind](fx.w.ECS).Get(fx.entity(fx.thinkers[0])).Goal = "goal.test.other"
		}, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fx := newMindFixture(t, 1)
			fx.until(20, "pondering", func() bool { return len(fx.starts) == 1 })
			tc.breakIt(fx)
			fx.until(20, "planning again", func() bool { return fx.calls >= 2 })
			if len(fx.starts) > 2 {
				t.Errorf(
					"%d ponders started before the plan was made again, want the old plan dropped after 1",
					len(fx.starts),
				)
			}
			mind := fx.mind(fx.thinkers[0])
			if got := len(mind.Plan.Steps) > 0; got != tc.replaced {
				t.Errorf("has a plan after it was made again = %v, want %v", got, tc.replaced)
			}
			if !tc.replaced && fx.claims(fx.thinkers[0]) != nil {
				t.Errorf("claims %v outlived the plan", fx.claims(fx.thinkers[0]))
			}
		})
	}
}

func TestAClashingClaimGoesToTheLowerUID(t *testing.T) {
	fx := newMindFixture(t, 2)
	fx.until(20, "pondering", func() bool { return len(fx.starts) >= 1 })
	fx.until(20, "the second thinker planning", func() bool { return fx.calls >= 3 })
	first, second := fx.thinkers[0], fx.thinkers[1]
	if fx.mind(first).Plan.Steps == nil || fx.claims(first) == nil {
		t.Error("the lower UID did not win the claim")
	}
	if fx.mind(second).Plan.Steps != nil || fx.claims(second) != nil {
		t.Errorf("the higher UID holds a plan %v and claims %v on a claimed well",
			fx.mind(second).Plan, fx.claims(second))
	}
	for _, started := range fx.starts {
		if started.Source != first {
			t.Errorf("body %d pondered without the claim", started.Source)
		}
	}
}

func TestADelegatedMindWaitsUntilItsDelegateResumes(t *testing.T) {
	fx := newMindFixture(t, 1)
	fx.until(20, "pondering", func() bool { return len(fx.starts) == 1 })
	body := fx.entity(fx.thinkers[0])
	if err := timeline.DelegateTo(fx.w, body, timeline.IdleController); err != nil {
		t.Fatal(err)
	}
	for range 10 {
		fx.wait()
	}
	if len(fx.starts) != 1 || fx.mind(fx.thinkers[0]).Step != 1 {
		t.Fatalf("a delegated mind started %d ponders and reached step %d, want 1 and 1",
			len(fx.starts), fx.mind(fx.thinkers[0]).Step)
	}
	timeline.Resume(fx.w, body)
	fx.until(20, "finishing the plan", func() bool { return len(fx.starts) == planSteps })
	if fx.calls != 1 {
		t.Errorf("the planner was asked %d times, want the first plan kept through the hand-off", fx.calls)
	}
}

func TestARecipeWhoseWhenFailsStopsTheStep(t *testing.T) {
	const broken = "recipe.test.broken"
	bundle := contenttest.SampleBundle(t)
	thinkingContent(&bundle)
	bundle.Templates[string(thinker)].Mind.Planner = agents.HTNPlanner
	bundle.Goals[string(thinkGoal)] = content.Goal{Weight: "0", When: "false"}
	bundle.Recipes[broken] = content.Recipe{
		Goal: thinkGoal, When: `has["nothing"] % has["nothing"] == 0`,
		Steps: []content.RecipeStep{{Operator: ponderOp}},
	}
	g, err := game.New(game.Config{
		Seed: 1, Content: contenttest.Registry(t, bundle),
		Spawn: func(w *sim.World, _ *clock.Clock) error { return (&mindFixture{t: t}).spawn(w, 1) },
	})
	if err != nil {
		t.Fatal(err)
	}
	for range 20 {
		if _, err = g.Apply(api.Wait{Source: g.Current(), Until: api.WaitEvent}); err != nil {
			break
		}
	}
	if err == nil || !strings.Contains(err.Error(), broken) {
		t.Fatalf("error = %v, want one naming %s", err, broken)
	}
}

func TestAMindSavedMidPlanFinishesTheSameWay(t *testing.T) {
	fx := newMindFixture(t, 1)
	fx.until(20, "pondering", func() bool { return len(fx.starts) == 1 })
	if step := fx.mind(fx.thinkers[0]).Step; step != 1 {
		t.Fatalf("saving at step %d, want 1", step)
	}
	loaded, _, _ := gametest.Reload(t, fx.cfg, fx.g)
	var kept, reloaded []api.ActionStarted
	for range 12 {
		kept = append(kept, waitOnce(t, fx.g)...)
		reloaded = append(reloaded, waitOnce(t, loaded)...)
	}
	if len(kept) < planSteps {
		t.Fatalf("only %d ponders started after the save, want the plan finished", len(kept))
	}
	if !slices.EqualFunc(kept, reloaded, func(a, b api.ActionStarted) bool {
		return a.Source == b.Source && a.Tick == b.Tick && a.Lands == b.Lands
	}) {
		t.Errorf("after the load the ponders went %v, want %v", reloaded, kept)
	}
	want, err := fx.g.Hash()
	if err != nil {
		t.Fatal(err)
	}
	if got, err := loaded.Hash(); err != nil || got != want {
		t.Errorf("hash after the load = %x, %v; want %x", got, err, want)
	}
}
