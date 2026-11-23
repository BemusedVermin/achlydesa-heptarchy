package timeline_test

import (
	"errors"
	"reflect"
	"slices"
	"testing"

	"github.com/mlange-42/ark/ecs"
	"pgregory.net/rapid"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/clock"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

const sampleDir = "../../../content/data"

// tester is the part of testing.T that rapid.T shares.
type tester interface {
	Helper()
	Fatal(args ...any)
	Fatalf(format string, args ...any)
}

var errStop = errors.New("test: stop")

// decision is one decision taken by a recording controller.
type decision struct {
	Tick int64
	UID  api.UID
}

// fixture is a world with the clock and timeline installed, and a controller "record" that logs
// every decision it takes, occupies the body for one step and stops the run once the clock passes
// stopAfter.
type fixture struct {
	w         *sim.World
	clk       *clock.Clock
	taken     []decision
	stopAfter int64
	speeds    map[api.UID]int
	phases    []int64 // the tick at which each PerPhase system ran
	perTurn   int     // how many times the PerTurn system ran
	after     func(w *sim.World, body ecs.Entity)
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
	if err := timeline.Install(fx.w, fx.clk); err != nil {
		t.Fatal(err)
	}
	timeline.SetSpeed(fx.w, func(w *sim.World, e ecs.Entity) int {
		if speed, ok := fx.speeds[w.UIDOf(e)]; ok {
			return speed
		}
		return reg.Timeline().DefaultSpeed
	})
	timeline.SetVisible(fx.w, func(*sim.World, ecs.Entity) bool { return true })
	if err := timeline.RegisterController(fx.w, "record", recorder{fx}); err != nil {
		t.Fatal(err)
	}
	err = fx.w.Systems.Register(sim.System{Name: "test.phase", Cadence: sim.PerPhase, Run: func(w *sim.World) error {
		fx.phases = append(fx.phases, w.Now)
		return nil
	}})
	if err != nil {
		t.Fatal(err)
	}
	err = fx.w.Systems.Register(sim.System{Name: "test.turn", Cadence: sim.PerTurn, Run: func(*sim.World) error {
		fx.perTurn++
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

// idleController decides nothing, so the loop makes its body wait.
type idleController struct{}

func (idleController) Decide(*sim.World, ecs.Entity) error { return nil }

type recorder struct{ fx *fixture }

func (r recorder) Decide(w *sim.World, body ecs.Entity) error {
	if w.Now > r.fx.stopAfter {
		return errStop
	}
	r.fx.taken = append(r.fx.taken, decision{w.Now, w.UIDOf(body)})
	ticks, err := timeline.StepTicks(w, body)
	if err != nil {
		return err
	}
	if err := timeline.Occupy(w, timeline.Channels(w, body)[0], ticks); err != nil {
		return err
	}
	if r.fx.after != nil {
		r.fx.after(w, body)
	}
	return nil
}

// spawn adds a member template and switches it to the controller, returning its entity.
func (fx *fixture) spawn(t tester, controller string, speed int) ecs.Entity {
	t.Helper()
	fx.speeds[fx.w.NextUID()] = speed // the speed hook reads it, so set it before spawning
	e, err := fx.w.Spawn("template.person.reaper", nil)
	if err != nil {
		t.Fatal(err)
	}
	// The template says "player"; swap in the controller under test.
	ecs.NewMap[comp.Actor](fx.w.ECS).Get(e).Controller = controller
	return e
}

func (fx *fixture) stepTicks(t tester, speed int) int64 {
	t.Helper()
	ticks, err := fx.w.Content.StepTicks(content.SpeedEnv{Speed: speed})
	if err != nil {
		t.Fatal(err)
	}
	return ticks
}

// run runs until the stop controller fires, which is the expected end of every AI-only test.
func (fx *fixture) run(t tester) {
	t.Helper()
	if _, err := timeline.Run(fx.w, fx.clk); !errors.Is(err, errStop) {
		t.Fatalf("Run error = %v, want the test stop", err)
	}
}

// bodyChannel is the body's own channel.
func (fx *fixture) bodyChannel(body ecs.Entity) timeline.Channel {
	return timeline.Channels(fx.w, body)[0]
}

// reference replays the timeline with a plain sort: each body decides at 0 and then every step.
func reference(steps map[api.UID]int64, until int64) []decision {
	next := map[api.UID]int64{}
	for uid := range steps {
		next[uid] = 0
	}
	var out []decision
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
		out = append(out, decision{next[best], best})
		next[best] += steps[best]
	}
}

func TestOccupyAndExtendMoveTheNextDecision(t *testing.T) {
	fx := newFixture(t, 1)
	player := fx.spawn(t, timeline.PlayerController, 2)
	if _, err := timeline.Run(fx.w, fx.clk); err != nil || fx.w.Now != 0 {
		t.Fatalf("first Run: now = %d, err = %v; want a decision at tick 0", fx.w.Now, err)
	}
	channel := fx.bodyChannel(player)
	if err := timeline.Occupy(fx.w, channel, 6); err != nil {
		t.Fatal(err)
	}
	if timeline.Free(fx.w, channel) {
		t.Error("an occupied channel is not free")
	}
	if err := timeline.Occupy(fx.w, channel, 3); err == nil {
		t.Error("occupying a busy channel should fail")
	}
	if err := timeline.Extend(fx.w, channel, 2); err != nil {
		t.Fatal(err)
	}
	if err := timeline.Occupy(fx.w, channel, 0); err == nil {
		t.Error("occupying for no ticks should fail")
	}
	if current, err := timeline.Run(fx.w, fx.clk); err != nil || current != player || fx.w.Now != 8 {
		t.Errorf("Run = %v, %v at tick %d; want the player at tick 8", current, err, fx.w.Now)
	}
	// Extending a free channel counts from now.
	if err := timeline.Extend(fx.w, channel, 5); err != nil {
		t.Fatal(err)
	}
	if _, err := timeline.Run(fx.w, fx.clk); err != nil || fx.w.Now != 13 {
		t.Errorf("after extending a free channel: now = %d, err = %v; want 13", fx.w.Now, err)
	}
}

// twoChannels gives every body a second channel, on a part entity.
func twoChannels(t tester, fx *fixture, body ecs.Entity) timeline.Channel {
	t.Helper()
	part := ecs.NewMap[comp.Appearance](fx.w.ECS).NewEntity(&comp.Appearance{})
	timeline.SetChannels(fx.w, func(_ *sim.World, b ecs.Entity) []timeline.Channel {
		return []timeline.Channel{
			{Entity: part, Body: b, Tags: []string{"arm"}},
			{Entity: b, Tags: []string{"body"}},
		}
	})
	return timeline.Channels(fx.w, body)[1]
}

func TestSecondChannelLetsTheBodyDecideAgain(t *testing.T) {
	fx := newFixture(t, 1)
	player := fx.spawn(t, timeline.PlayerController, 2)
	arm := twoChannels(t, fx, player)
	if channels := timeline.Channels(fx.w, player); channels[0].Entity != player || channels[1].Entity != arm.Entity {
		t.Fatalf("channels = %+v; want the body first, then the arm", channels)
	}
	if _, err := timeline.Run(fx.w, fx.clk); err != nil {
		t.Fatal(err)
	}
	if err := timeline.Occupy(fx.w, fx.bodyChannel(player), 6); err != nil {
		t.Fatal(err)
	}
	current, err := timeline.Run(fx.w, fx.clk)
	if err != nil || current != player || fx.w.Now != 0 {
		t.Fatalf("Run = %v, %v at tick %d; want the same body at tick 0", current, err, fx.w.Now)
	}
	if !timeline.Free(fx.w, arm) || timeline.Free(fx.w, fx.bodyChannel(player)) {
		t.Error("the arm should be free and the body busy")
	}
	if err := timeline.Occupy(fx.w, arm, 4); err != nil {
		t.Fatal(err)
	}
	// The arm frees first, at tick 4.
	if current, err := timeline.Run(fx.w, fx.clk); err != nil || current != player || fx.w.Now != 4 {
		t.Errorf("Run = %v, %v at tick %d; want the player at tick 4", current, err, fx.w.Now)
	}
}

func TestRescheduleBringsADecisionForwardWhenAHookChanges(t *testing.T) {
	fx := newFixture(t, 1)
	ai := fx.spawn(t, timeline.IdleController, 2)
	arm := twoChannels(t, fx, ai)
	// Only the arm counts, so the body decides when the arm frees, at tick 9.
	timeline.SetUsable(
		fx.w,
		func(_ *sim.World, _ ecs.Entity, c timeline.Channel) bool { return c.Entity == arm.Entity },
	)
	if err := timeline.Occupy(fx.w, fx.bodyChannel(ai), 4); err != nil {
		t.Fatal(err)
	}
	if err := timeline.Occupy(fx.w, arm, 9); err != nil {
		t.Fatal(err)
	}
	readyAt := func() int64 {
		t.Helper()
		view, _ := fx.view(t)
		for _, entry := range view.Timeline.Entries {
			if entry.Kind == "ready" && entry.Actor == fx.w.UIDOf(ai) {
				return entry.Tick
			}
		}
		t.Fatal("the body is not on the timeline")
		return 0
	}
	if got := readyAt(); got != 9 {
		t.Fatalf("ready at %d, want 9", got)
	}
	// Now the body channel counts too. Nothing requeues the body until Reschedule says so.
	timeline.SetUsable(fx.w, func(*sim.World, ecs.Entity, timeline.Channel) bool { return true })
	if got := readyAt(); got != 9 {
		t.Fatalf("ready at %d before Reschedule, want the old 9", got)
	}
	timeline.Reschedule(fx.w, ai)
	if got := readyAt(); got != 4 {
		t.Errorf("ready at %d after Reschedule, want 4", got)
	}
	// An entity with no Actor is left alone.
	timeline.Reschedule(fx.w, arm.Entity)
}

func TestUnusableChannelsDoNotMakeABodyDecide(t *testing.T) {
	fx := newFixture(t, 1)
	player := fx.spawn(t, timeline.PlayerController, 2)
	arm := twoChannels(t, fx, player)
	timeline.SetUsable(
		fx.w,
		func(_ *sim.World, _ ecs.Entity, c timeline.Channel) bool { return c.Entity != arm.Entity },
	)
	if _, err := timeline.Run(fx.w, fx.clk); err != nil {
		t.Fatal(err)
	}
	if err := timeline.Occupy(fx.w, fx.bodyChannel(player), 6); err != nil {
		t.Fatal(err)
	}
	// The arm is free but unusable, so the body decides when the body channel frees.
	if _, err := timeline.Run(fx.w, fx.clk); err != nil || fx.w.Now != 6 {
		t.Errorf("now = %d, err = %v; want tick 6", fx.w.Now, err)
	}
	// With nothing usable, the body uses all its channels: it decides when the earliest frees.
	timeline.SetUsable(fx.w, func(*sim.World, ecs.Entity, timeline.Channel) bool { return false })
	if err := timeline.Occupy(fx.w, fx.bodyChannel(player), 9); err != nil {
		t.Fatal(err)
	}
	if err := timeline.Occupy(fx.w, arm, 5); err != nil {
		t.Fatal(err)
	}
	if _, err := timeline.Run(fx.w, fx.clk); err != nil || fx.w.Now != 11 {
		t.Errorf("now = %d, err = %v; want tick 11", fx.w.Now, err)
	}
}

// view builds the view and returns it with the wait option commands by ID.
func (fx *fixture) view(t tester) (api.View, map[string]api.Command) {
	t.Helper()
	view, err := fx.w.Views.Build(fx.w)
	if err != nil {
		t.Fatal(err)
	}
	commands := map[string]api.Command{}
	for _, option := range view.Options {
		commands[option.ID] = option.Command
	}
	return view, commands
}

func TestWaitUntilSomethingHappens(t *testing.T) {
	fx := newFixture(t, 1)
	ai := fx.spawn(t, "record", 2)
	player := fx.spawn(t, timeline.PlayerController, 2)
	if current, err := timeline.Run(fx.w, fx.clk); err != nil || current != player || fx.w.Now != 0 {
		t.Fatalf("Run = %v, %v at tick %d; want the player at tick 0", current, err, fx.w.Now)
	}
	view, commands := fx.view(t)
	var labels []string
	for _, option := range view.Options {
		labels = append(labels, option.Label)
	}
	if want := []string{"Wait until something happens", "Wait until day"}; !reflect.DeepEqual(labels, want) {
		t.Fatalf("option labels = %v, want %v", labels, want)
	}
	if _, err := fx.w.Commands.Dispatch(fx.w, commands["wait/event"]); err != nil {
		t.Fatal(err)
	}
	// The AI body occupied itself for 10 ticks, so its next decision is the next entry.
	if current, err := timeline.Run(fx.w, fx.clk); err != nil || current != player || fx.w.Now != 10 {
		t.Fatalf("Run = %v, %v at tick %d; want the player at tick 10", current, err, fx.w.Now)
	}
	want := []decision{{0, fx.w.UIDOf(ai)}, {10, fx.w.UIDOf(ai)}}
	if !reflect.DeepEqual(fx.taken, want) {
		t.Errorf("AI decisions = %v, want %v", fx.taken, want)
	}
}

func TestWaitUntilThePhaseChanges(t *testing.T) {
	fx := newFixture(t, 1)
	player := fx.spawn(t, timeline.PlayerController, 2)
	if _, err := timeline.Run(fx.w, fx.clk); err != nil {
		t.Fatal(err)
	}
	_, commands := fx.view(t)
	if _, err := fx.w.Commands.Dispatch(fx.w, commands["wait/phase"]); err != nil {
		t.Fatal(err)
	}
	if len(fx.phases) != 0 {
		t.Fatalf("phase systems ran before the clock moved: %v", fx.phases)
	}
	if current, err := timeline.Run(fx.w, fx.clk); err != nil || current != player {
		t.Fatal(current, err)
	}
	// Dawn lasts 20 ticks.
	if fx.w.Now != 20 || !reflect.DeepEqual(fx.phases, []int64{20}) {
		t.Errorf("now = %d, phase systems ran at %v; want 20 and [20]", fx.w.Now, fx.phases)
	}
	if _, err := fx.w.Commands.Dispatch(
		fx.w,
		api.Wait{Actor: fx.w.UIDOf(player), Until: api.WaitTicks, Ticks: 7},
	); err != nil {
		t.Fatal(err)
	}
	if _, err := timeline.Run(fx.w, fx.clk); err != nil || fx.w.Now != 27 {
		t.Errorf("after a 7-tick wait: now = %d, err = %v; want 27", fx.w.Now, err)
	}
}

func TestWaitCommandsRejected(t *testing.T) {
	fx := newFixture(t, 1)
	first := fx.spawn(t, timeline.PlayerController, 2)
	second := fx.spawn(t, timeline.PlayerController, 2)
	if current, err := timeline.Run(fx.w, fx.clk); err != nil || current != first {
		t.Fatal("the first player body should be current", current, err)
	}
	secondUID, firstUID := fx.w.UIDOf(second), fx.w.UIDOf(first)
	for _, cmd := range []api.Command{
		api.Wait{Actor: secondUID, Until: api.WaitEvent},
		api.Wait{Actor: 999, Until: api.WaitPhase},
		api.Wait{Actor: firstUID, Until: api.WaitTicks, Ticks: 0},
		api.Wait{Actor: firstUID, Until: api.WaitTicks, Ticks: -3},
		api.Wait{Actor: firstUID, Until: "forever"},
	} {
		if _, err := fx.w.Commands.Dispatch(fx.w, cmd); err != nil {
			t.Fatalf("%+v: %v", cmd, err)
		}
		events, err := fx.w.Bus.Drain(fx.w)
		if err != nil {
			t.Fatal(err)
		}
		if len(events) != 1 || events[0].EventKind() != api.RejectedKind {
			t.Errorf("%+v: events = %v, want one rejection", cmd, events)
		}
		if current, _ := timeline.Current(fx.w); current != first {
			t.Errorf("%+v changed who is deciding", cmd)
		}
	}
}

func TestControllerThatDoesNothingIsMadeToWait(t *testing.T) {
	fx := newFixture(t, 1)
	idler := fx.spawn(t, timeline.IdleController, 2)
	player := fx.spawn(t, timeline.PlayerController, 2)
	if current, err := timeline.Run(fx.w, fx.clk); err != nil || current != player || fx.w.Now != 0 {
		t.Fatalf("Run = %v, %v at tick %d; want the player at tick 0", current, err, fx.w.Now)
	}
	if wait := ecs.NewMap[comp.Actor](fx.w.ECS).Get(idler).WaitUntil; wait != 1 {
		t.Errorf("the idle body waits until %d, want 1", wait)
	}
}

func TestRunWithoutPlayerGivesUp(t *testing.T) {
	fx := newFixture(t, 1)
	idler := fx.spawn(t, timeline.IdleController, 2)
	if _, err := timeline.Run(fx.w, fx.clk); err == nil {
		t.Fatal("expected an error")
	}
	// Decision k is at tick k, and each makes the body wait one tick.
	if got, want := ecs.NewMap[comp.Actor](fx.w.ECS).Get(idler).WaitUntil, int64(timeline.MaxAITurns); got != want {
		t.Errorf("wait until = %d, want %d", got, want)
	}
}

func TestRunWithNobody(t *testing.T) {
	fx := newFixture(t, 1)
	if _, err := timeline.Run(fx.w, fx.clk); err == nil {
		t.Error("an empty timeline should be an error")
	}
}

func TestControllerRegistration(t *testing.T) {
	fx := newFixture(t, 1)
	if err := timeline.RegisterController(fx.w, timeline.PlayerController, recorder{fx}); err == nil {
		t.Error("the player controller name is reserved")
	}
	if err := timeline.RegisterController(fx.w, "record", recorder{fx}); err == nil {
		t.Error("a duplicate controller name should fail")
	}
	e := fx.spawn(t, "nobody", 2)
	if _, err := timeline.Run(fx.w, fx.clk); err == nil {
		t.Errorf("entity %v has an unregistered controller; Run should fail", e)
	}
}

func TestViewListsVisibleBodiesInWindow(t *testing.T) {
	fx := newFixture(t, 1)
	var bodies []ecs.Entity
	for range 5 {
		bodies = append(bodies, fx.spawn(t, "record", 2))
	}
	uids := make([]api.UID, len(bodies))
	for i, body := range bodies {
		uids[i] = fx.w.UIDOf(body)
	}
	hidden := uids[1]
	timeline.SetVisible(fx.w, func(w *sim.World, e ecs.Entity) bool { return w.UIDOf(e) != hidden })
	// Decisions after: uid0 at 12, uid1 (hidden) at 5, uid2 at 5, uid3 at 40 (outside the window), uid4 at 5.
	for i, ticks := range []int64{12, 5, 5, 40, 5} {
		if err := timeline.Occupy(fx.w, fx.bodyChannel(bodies[i]), ticks); err != nil {
			t.Fatal(err)
		}
	}
	view, _ := fx.view(t)
	if view.Timeline == nil || view.Timeline.Current != 0 || view.Timeline.Now != 0 {
		t.Fatalf("timeline view = %+v", view.Timeline)
	}
	var got []decision
	for _, entry := range view.Timeline.Entries {
		if entry.Kind != "ready" || entry.Glyph != "@" || entry.Label != "The Reaper is ready" {
			t.Errorf("entry = %+v", entry)
		}
		got = append(got, decision{entry.Tick, entry.Actor})
	}
	want := []decision{{5, uids[2]}, {5, uids[4]}, {12, uids[0]}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("entries = %v, want %v", got, want)
	}
}

func TestDefaultVisibilityIsTheParty(t *testing.T) {
	reg, err := content.Load(sampleDir)
	if err != nil {
		t.Fatal(err)
	}
	w := sim.NewWorld(1)
	w.Content = reg
	clk := clock.New(reg.Calendar())
	if err := clock.Install(w, clk, reg); err != nil {
		t.Fatal(err)
	}
	if err := timeline.Install(w, clk); err != nil {
		t.Fatal(err)
	}
	// The sample party's followers use the controller explore registers.
	if err := timeline.RegisterController(w, "follow", idleController{}); err != nil {
		t.Fatal(err)
	}
	if err := w.Systems.Freeze(); err != nil {
		t.Fatal(err)
	}
	if _, err := w.Spawn("template.party.main", nil); err != nil {
		t.Fatal(err)
	}
	if _, err := w.Spawn("template.person.salken", nil); err != nil { // not in the party
		t.Fatal(err)
	}
	view, err := w.Views.Build(w)
	if err != nil {
		t.Fatal(err)
	}
	if got := len(view.Timeline.Entries); got != 2 {
		t.Errorf("%d entries, want the 2 party members", got)
	}
}

func TestPerTurnSystemsRunAfterEveryEntry(t *testing.T) {
	fx := newFixture(t, 1)
	fx.spawn(t, "record", 2)
	fx.spawn(t, "record", 1)
	fx.stopAfter = 100
	fx.run(t)
	// Every entry but the last (which stopped the run) has been followed by a PerTurn run.
	if want := len(fx.taken); fx.perTurn != want {
		t.Errorf("PerTurn ran %d times after %d decisions", fx.perTurn, want)
	}
}

func TestSameSeedSameOrder(t *testing.T) {
	play := func() []decision {
		fx := newFixture(t, 7)
		for _, speed := range []int{0, 1, 2, 3, 2, 1, 0, 2} {
			fx.spawn(t, "record", speed)
		}
		fx.after = func(*sim.World, ecs.Entity) {
			if len(fx.taken) == 1000 {
				fx.stopAfter = -1 // the next decision stops the run
			}
		}
		fx.run(t)
		return fx.taken
	}
	first, second := play(), play()
	if len(first) != 1000 || !reflect.DeepEqual(first, second) {
		t.Errorf("two runs differ or are short: %d and %d decisions", len(first), len(second))
	}
}

func TestOrderMatchesSort(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		fx := newFixture(t, 1)
		speeds := rapid.SliceOfN(rapid.IntRange(0, 3), 1, 12).Draw(t, "speeds")
		steps := map[api.UID]int64{}
		for _, speed := range speeds {
			steps[fx.w.UIDOf(fx.spawn(t, "record", speed))] = fx.stepTicks(t, speed)
		}
		fx.stopAfter = 500
		fx.run(t)
		// The run stops at the first decision past 500, which is not recorded.
		want := reference(steps, 500)
		if !reflect.DeepEqual(fx.taken, want) {
			t.Fatalf("order differs from sorted order for speeds %v:\n%v\n%v", speeds, fx.taken, want)
		}
		if !slices.IsSortedFunc(fx.taken, func(a, b decision) int { return int(a.Tick - b.Tick) }) {
			t.Fatalf("decisions went back in time: %v", fx.taken)
		}
	})
}

func TestUsableChannelsHookDecidesWhichChannelsCount(t *testing.T) {
	for _, tc := range []struct {
		name  string
		hook  func(cs []timeline.Channel, arm ecs.Entity) []bool
		reset bool
		want  int64
	}{
		{"the arm only", func(cs []timeline.Channel, arm ecs.Entity) []bool {
			out := make([]bool, len(cs))
			for i, c := range cs {
				out[i] = c.Entity == arm
			}
			return out
		}, false, 9},
		{"neither counts, so both do", func(cs []timeline.Channel, _ ecs.Entity) []bool { return make([]bool, len(cs)) }, false, 4},
		{"a miscounting hook is ignored", func([]timeline.Channel, ecs.Entity) []bool { return nil }, false, 4},
		{"SetUsable replaces the hook", func(cs []timeline.Channel, _ ecs.Entity) []bool { return make([]bool, len(cs)) }, true, 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fx := newFixture(t, 1)
			ai := fx.spawn(t, timeline.IdleController, 2)
			arm := twoChannels(t, fx, ai)
			timeline.SetUsableChannels(fx.w, func(_ *sim.World, _ ecs.Entity, cs []timeline.Channel) []bool {
				return tc.hook(cs, arm.Entity)
			})
			if tc.reset {
				timeline.SetUsable(fx.w, func(*sim.World, ecs.Entity, timeline.Channel) bool { return true })
			}
			if err := timeline.Occupy(fx.w, fx.bodyChannel(ai), 4); err != nil {
				t.Fatal(err)
			}
			if err := timeline.Occupy(fx.w, arm, 9); err != nil {
				t.Fatal(err)
			}
			timeline.Reschedule(fx.w, ai)
			view, _ := fx.view(t)
			for _, entry := range view.Timeline.Entries {
				if entry.Kind == "ready" && entry.Actor == fx.w.UIDOf(ai) {
					if entry.Tick != tc.want {
						t.Errorf("ready at %d, want %d", entry.Tick, tc.want)
					}
					return
				}
			}
			t.Fatal("the body is not on the timeline")
		})
	}
}
