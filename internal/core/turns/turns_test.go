package turns_test

import (
	"errors"
	"reflect"
	"sort"
	"testing"

	"github.com/mlange-42/ark/ecs"
	"pgregory.net/rapid"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/clock"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/turns"
)

const sampleDir = "../../../content/data"

// tester is the part of testing.T that rapid.T shares.
type tester interface {
	Helper()
	Fatal(args ...any)
	Fatalf(format string, args ...any)
}

var errStop = errors.New("test: stop")

// turn is one turn taken by a recording controller.
type turn struct {
	Tick int64
	UID  api.UID
}

// fixture is a world with the clock and turns installed, and a controller "record" that logs every
// turn it takes and stops the run once the clock passes stopAfter.
type fixture struct {
	w         *sim.World
	clk       *clock.Clock
	taken     []turn
	stopAfter int64
	speeds    map[api.UID]int
	phases    []int64 // the tick at which each PerPhase system ran
	extra     func(w *sim.World, actor ecs.Entity)
}

func newFixture(t tester, seed uint64) *fixture {
	t.Helper()
	reg, err := content.Load(sampleDir)
	if err != nil {
		t.Fatal(err)
	}
	fx := &fixture{w: sim.NewWorld(seed), speeds: map[api.UID]int{}, stopAfter: 1 << 60}
	fx.w.Content = reg
	fx.clk = clock.New(reg.Calendar())
	if err := clock.Install(fx.w, fx.clk, reg); err != nil {
		t.Fatal(err)
	}
	if err := turns.Install(fx.w, fx.clk); err != nil {
		t.Fatal(err)
	}
	turns.SetSpeed(fx.w, func(w *sim.World, e ecs.Entity) int {
		if speed, ok := fx.speeds[w.UIDOf(e)]; ok {
			return speed
		}
		return reg.Turns().DefaultSpeed
	})
	if err := turns.RegisterController(fx.w, "record", recorder{fx}); err != nil {
		t.Fatal(err)
	}
	err = fx.w.Systems.Register(sim.System{Name: "test.phase", Cadence: sim.PerPhase, Run: func(w *sim.World) error {
		fx.phases = append(fx.phases, w.Now)
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	if err := fx.w.Systems.Freeze(); err != nil {
		t.Fatal(err)
	}
	return fx
}

type recorder struct{ fx *fixture }

func (r recorder) TakeTurn(w *sim.World, actor ecs.Entity) error {
	if w.Now > r.fx.stopAfter {
		return errStop
	}
	r.fx.taken = append(r.fx.taken, turn{w.Now, w.UIDOf(actor)})
	if r.fx.extra != nil {
		r.fx.extra(w, actor)
	}
	return nil
}

// spawn adds a member template and switches it to the controller, returning its entity.
func (fx *fixture) spawn(t tester, controller string, speed int) ecs.Entity {
	t.Helper()
	fx.speeds[fx.w.NextUID()] = speed // the spawner reads the speed, so set it before spawning
	e, err := fx.w.Spawn("template.person.reaper", nil)
	if err != nil {
		t.Fatal(err)
	}
	// The template says "player"; the controller doesn't affect when the turn comes.
	ecs.NewMap[comp.Turn](fx.w.ECS).Get(e).Controller = controller
	return e
}

func (fx *fixture) ticks(t tester, speed int) int64 {
	t.Helper()
	ticks, err := fx.w.Content.TurnTicks(content.TurnEnv{Speed: speed})
	if err != nil {
		t.Fatal(err)
	}
	return ticks
}

// run runs until the stop controller fires, which is the expected end of every AI-only test.
func (fx *fixture) run(t tester) {
	t.Helper()
	if _, err := turns.Run(fx.w, fx.clk); !errors.Is(err, errStop) {
		t.Fatalf("Run error = %v, want the test stop", err)
	}
}

// reference replays the timeline with a plain sort, as the spec says the heap must match.
func reference(ticks map[api.UID]int64, until int64) []turn {
	next := map[api.UID]int64{}
	for uid, step := range ticks {
		next[uid] = step
	}
	var out []turn
	for {
		var best api.UID
		for uid, tick := range next {
			if best == 0 || tick < next[best] || (tick == next[best] && uid < best) {
				best = uid
			}
		}
		if next[best] > until {
			return out
		}
		out = append(out, turn{next[best], best})
		next[best] += ticks[best]
	}
}

func TestFasterActorTakesMoreTurns(t *testing.T) {
	fx := newFixture(t, 1)
	fast := fx.spawn(t, "record", 2)
	slow := fx.spawn(t, "record", 1)
	fx.stopAfter = 60
	fx.run(t)

	counts := map[api.UID]int{}
	for _, taken := range fx.taken {
		counts[taken.UID]++
	}
	if got, want := counts[fx.w.UIDOf(fast)], 6; got != want {
		t.Errorf("fast actor took %d turns, want %d", got, want)
	}
	if got, want := counts[fx.w.UIDOf(slow)], 4; got != want {
		t.Errorf("slow actor took %d turns, want %d", got, want)
	}
	want := reference(map[api.UID]int64{fx.w.UIDOf(fast): 10, fx.w.UIDOf(slow): 15}, 60)
	if !reflect.DeepEqual(fx.taken, want) {
		t.Errorf("order = %v\nwant    %v", fx.taken, want)
	}
}

func TestExtraDelaysOnlyTheNextTurn(t *testing.T) {
	fx := newFixture(t, 1)
	e := fx.spawn(t, "record", 2)
	fx.stopAfter = 60
	first := true
	fx.extra = func(w *sim.World, actor ecs.Entity) {
		if first {
			turns.AddExtra(w, actor, 7)
			first = false
		}
	}
	fx.run(t)
	uid := fx.w.UIDOf(e)
	want := []turn{{10, uid}, {27, uid}, {37, uid}, {47, uid}, {57, uid}}
	if !reflect.DeepEqual(fx.taken, want) {
		t.Errorf("turns = %v, want %v", fx.taken, want)
	}
}

func TestRunStopsAtPlayer(t *testing.T) {
	fx := newFixture(t, 1)
	ai := fx.spawn(t, "record", 2)
	player := fx.spawn(t, turns.PlayerController, 1)
	fx.stopAfter = 1000

	current, err := turns.Run(fx.w, fx.clk)
	if err != nil || current != player {
		t.Fatalf("Run = %v, %v; want the player %v", current, err, player)
	}
	if fx.w.Now != 15 {
		t.Errorf("now = %d, want 15", fx.w.Now)
	}
	// The AI actor (speed 2, every 10 ticks) took its turn at 10 first.
	if want := []turn{{10, fx.w.UIDOf(ai)}}; !reflect.DeepEqual(fx.taken, want) {
		t.Errorf("AI turns = %v, want %v", fx.taken, want)
	}
	again, err := turns.Run(fx.w, fx.clk)
	if err != nil || again != player || len(fx.taken) != 1 || fx.w.Now != 15 {
		t.Errorf("second Run = %v, %v (turns %v, now %d); want the same open turn", again, err, fx.taken, fx.w.Now)
	}
	if now, ok := turns.Current(fx.w); !ok || now != player {
		t.Errorf("Current = %v, %v", now, ok)
	}
}

func TestEndTurnReschedules(t *testing.T) {
	fx := newFixture(t, 1)
	player := fx.spawn(t, turns.PlayerController, 2)
	if _, err := turns.Run(fx.w, fx.clk); err != nil {
		t.Fatal(err)
	}
	if steps, actions := turns.Remaining(fx.w, player); steps != 1 || actions != 0 {
		t.Fatalf("budget = %d, %d; want 1, 0", steps, actions)
	}
	if err := turns.Spend(fx.w, player, true, 0); err != nil {
		t.Fatal(err)
	}
	if steps, _ := turns.Remaining(fx.w, player); steps != 0 {
		t.Errorf("steps left = %d after stepping", steps)
	}
	if err := turns.Spend(fx.w, player, true, 0); err == nil {
		t.Error("a second step should be over budget")
	}
	if err := turns.Spend(fx.w, player, false, 1); err == nil {
		t.Error("an action should be over a budget of none")
	}
	if err := turns.EndTurn(fx.w, player); err != nil {
		t.Fatal(err)
	}
	if err := turns.EndTurn(fx.w, player); err == nil {
		t.Error("ending a closed turn should fail")
	}
	if steps, _ := turns.Remaining(fx.w, player); steps != 0 {
		t.Error("a closed turn has no budget")
	}
	if _, err := turns.Run(fx.w, fx.clk); err != nil || fx.w.Now != 20 {
		t.Errorf("after ending: now = %d, err = %v; want tick 20", fx.w.Now, err)
	}
	if steps, _ := turns.Remaining(fx.w, player); steps != 1 {
		t.Errorf("the budget should reset; steps = %d", steps)
	}
}

func TestBudgetHook(t *testing.T) {
	fx := newFixture(t, 1)
	player := fx.spawn(t, turns.PlayerController, 2)
	turns.SetBudget(fx.w, func(*sim.World, ecs.Entity) (int, int) { return 1, 2 })
	if _, err := turns.Run(fx.w, fx.clk); err != nil {
		t.Fatal(err)
	}
	if err := turns.Spend(fx.w, player, false, 2); err != nil {
		t.Fatal(err)
	}
	if _, actions := turns.Remaining(fx.w, player); actions != 0 {
		t.Errorf("actions left = %d, want 0", actions)
	}
	if err := turns.Spend(fx.w, player, false, 1); err == nil {
		t.Error("a third arm use should be over budget")
	}
}

func TestWaitAcrossPhaseBoundary(t *testing.T) {
	fx := newFixture(t, 1)
	player := fx.spawn(t, turns.PlayerController, 2)
	if _, err := turns.Run(fx.w, fx.clk); err != nil {
		t.Fatal(err)
	}
	view, err := fx.w.Views.Build(fx.w)
	if err != nil {
		t.Fatal(err)
	}
	var wait api.Command
	var labels []string
	for _, option := range view.Options {
		labels = append(labels, option.Label)
		if option.ID == "wait" {
			wait = option.Command
		}
	}
	if want := []string{"End turn", "Wait until the next phase"}; !reflect.DeepEqual(labels, want) {
		t.Fatalf("option labels = %v, want %v", labels, want)
	}
	// At tick 10 of a 20-tick dawn there are 10 ticks left in the phase.
	if want := (api.Wait{Actor: fx.w.UIDOf(player), Ticks: 10}); wait != want {
		t.Fatalf("wait option = %+v, want %+v", wait, want)
	}
	if _, err := fx.w.Commands.Dispatch(fx.w, wait); err != nil {
		t.Fatal(err)
	}
	if len(fx.phases) != 0 {
		t.Fatalf("phase systems ran before the clock moved: %v", fx.phases)
	}
	if _, err := turns.Run(fx.w, fx.clk); err != nil {
		t.Fatal(err)
	}
	// Next = 10 + 10 (a turn) + 10 (the wait) = 30; dawn ends at 20.
	if fx.w.Now != 30 || !reflect.DeepEqual(fx.phases, []int64{20}) {
		t.Errorf("now = %d, phase systems ran at %v; want 30 and [20]", fx.w.Now, fx.phases)
	}
}

func TestCommandsRejectedUnlessCurrentPlayer(t *testing.T) {
	fx := newFixture(t, 1)
	first := fx.spawn(t, turns.PlayerController, 2)
	second := fx.spawn(t, turns.PlayerController, 1)
	if _, err := turns.Run(fx.w, fx.clk); err != nil {
		t.Fatal(err)
	}
	if current, _ := turns.Current(fx.w); current != first {
		t.Fatal("the faster actor should be current")
	}
	uid := fx.w.UIDOf(second)
	for _, cmd := range []api.Command{
		api.End{Actor: uid},
		api.Wait{Actor: uid, Ticks: 5},
		api.End{Actor: 999},
		api.Wait{Actor: fx.w.UIDOf(first), Ticks: 0},
		api.Wait{Actor: fx.w.UIDOf(first), Ticks: -3},
	} {
		if _, err := fx.w.Commands.Dispatch(fx.w, cmd); err != nil {
			t.Fatalf("%T: %v", cmd, err)
		}
		events, err := fx.w.Bus.Drain(fx.w)
		if err != nil {
			t.Fatal(err)
		}
		if len(events) != 1 || events[0].EventKind() != api.RejectedKind {
			t.Errorf("%+v: events = %v, want one rejection", cmd, events)
		}
		if current, _ := turns.Current(fx.w); current != first {
			t.Errorf("%+v changed whose turn it is", cmd)
		}
	}
	// The current actor can end its turn, and that publishes TurnEnded.
	if _, err := fx.w.Commands.Dispatch(fx.w, api.End{Actor: fx.w.UIDOf(first)}); err != nil {
		t.Fatal(err)
	}
	if _, ok := turns.Current(fx.w); ok {
		t.Error("the turn should be closed")
	}
}

func TestViewListsUpcoming(t *testing.T) {
	fx := newFixture(t, 1)
	var uids []api.UID
	for range 10 {
		uids = append(uids, fx.w.UIDOf(fx.spawn(t, "record", 2)))
	}
	view, err := fx.w.Views.Build(fx.w)
	if err != nil {
		t.Fatal(err)
	}
	if view.Turns == nil || view.Turns.Current != 0 || !reflect.DeepEqual(view.Turns.Upcoming, uids[:8]) {
		t.Errorf("turns view = %+v, want the first eight UIDs %v", view.Turns, uids[:8])
	}
}

func TestRunWithoutPlayerGivesUp(t *testing.T) {
	fx := newFixture(t, 1)
	fx.spawn(t, turns.IdleController, 2)
	if _, err := turns.Run(fx.w, fx.clk); err == nil {
		t.Fatal("expected an error")
	}
	count := 0
	query := ecs.NewFilter1[comp.Turn](fx.w.ECS).Query()
	for query.Next() {
		count++
		// Turn k happens at tick 10*k, so MaxAITurns turns end with the next at 10*(MaxAITurns+1).
		if got, want := query.Get().Next, int64(10*(turns.MaxAITurns+1)); got != want {
			t.Errorf("next = %d, want %d", got, want)
		}
	}
	if count != 1 {
		t.Errorf("%d actors on the timeline", count)
	}
}

func TestRunWithNobody(t *testing.T) {
	fx := newFixture(t, 1)
	if _, err := turns.Run(fx.w, fx.clk); err == nil {
		t.Error("an empty timeline should be an error")
	}
}

func TestUnknownController(t *testing.T) {
	fx := newFixture(t, 1)
	if err := turns.RegisterController(fx.w, turns.PlayerController, recorder{fx}); err == nil {
		t.Error("the player controller name is reserved")
	}
	if err := turns.RegisterController(fx.w, "record", recorder{fx}); err == nil {
		t.Error("a duplicate controller name should fail")
	}
	e := fx.spawn(t, "nobody", 2)
	if _, err := turns.Run(fx.w, fx.clk); err == nil {
		t.Errorf("entity %v has an unregistered controller; Run should fail", e)
	}
}

func TestSameSeedSameOrder(t *testing.T) {
	play := func() []turn {
		fx := newFixture(t, 7)
		for _, speed := range []int{0, 1, 2, 3, 2, 1, 0, 2} {
			fx.spawn(t, "record", speed)
		}
		fx.stopAfter = 1 << 60
		fx.extra = func(*sim.World, ecs.Entity) {
			if len(fx.taken) == 1000 {
				fx.stopAfter = 0 // the next turn stops the run
			}
		}
		fx.run(t)
		return fx.taken
	}
	first, second := play(), play()
	if len(first) != 1000 || !reflect.DeepEqual(first, second) {
		t.Errorf("two runs differ or are short: %d and %d turns", len(first), len(second))
	}
}

func TestHeapMatchesSort(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		fx := newFixture(t, 1)
		speeds := rapid.SliceOfN(rapid.IntRange(0, 3), 1, 12).Draw(t, "speeds")
		ticks := map[api.UID]int64{}
		for _, speed := range speeds {
			uid := fx.w.UIDOf(fx.spawn(t, "record", speed))
			ticks[uid] = fx.ticks(t, speed)
		}
		fx.stopAfter = 500
		fx.run(t)
		want := reference(ticks, 500)
		if !reflect.DeepEqual(fx.taken, want) {
			t.Fatalf("heap order differs from sorted order for speeds %v:\n%v\n%v", speeds, fx.taken, want)
		}
		if !sort.SliceIsSorted(fx.taken, func(i, j int) bool { return fx.taken[i].Tick < fx.taken[j].Tick }) {
			t.Fatalf("turns went back in time: %v", fx.taken)
		}
	})
}
