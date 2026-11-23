package actions_test

import (
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/actions"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/clock"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/effects"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/rng"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

const sampleDir = "../../../content/data"

// opts shapes a fixture.
type opts struct {
	actions []api.ContentID // what the fighter knows; defaults to every test action and the step
	arms    int             // the fighter's arms; defaults to 3
	dummyQ  int             // where the dummy stands; defaults to 1, next to the fighter
	hooks   bool            // register a "type" choice with a validator and a modifier
	// contentChoices keeps the sample's "type" choice from content instead of a Go provider.
	contentChoices bool
	// install runs with the world before the game starts, to register primitives.
	install func(w *sim.World) error
	mutate  func(*content.Bundle)
}

// fixture is a game with a fighter (the player, with arms) and an idle dummy.
type fixture struct {
	t       *testing.T
	cfg     game.Config
	g       *game.Game
	w       *sim.World
	fighter ecs.Entity
	dummy   ecs.Entity
	arms    []ecs.Entity
	hits    int
}

var hit = []content.EffectRef{{Do: "test.hit", Params: map[string]string{}}}

func use(tag string, count int) []content.Requirement {
	return []content.Requirement{{Tag: tag, Count: count}}
}

// testBundle is the sample content plus the actions the tests need.
func testBundle(t *testing.T, o opts) content.Bundle {
	t.Helper()
	bundle, err := content.LoadBundle(sampleDir)
	if err != nil {
		t.Fatal(err)
	}
	// The fixtures place bodies by hand, so they start with no maps and no start party; a body that
	// is not on a map steps in place.
	bundle.Maps, bundle.Start = nil, content.Start{}
	// The fixture serves "type" from Go, so the sample's content choice of that name would shadow it.
	if !o.contentChoices {
		delete(bundle.Choices, "type")
	}
	known := o.actions
	if known == nil {
		known = []api.ContentID{
			"action.step", "action.test.grab", "action.test.swing", "action.test.slam", "action.test.gate",
			"action.test.load",
		}
	}
	defs := map[string]content.ActionDef{
		"grab": {Uses: use("arm", 2), Target: "none", Windup: "0", Recovery: "10", OnLand: hit},
		"swing": {
			Uses: use("arm", 1), Target: "body", Range: 2, Windup: "3", Recovery: "5", OnLand: hit,
		},
		"slam": {Uses: use("arm", 2), Target: "none", Windup: "3", Recovery: "4", OnLand: hit},
		"gate": {
			Uses: use("body", 1), Target: "self", Windup: "0", Recovery: "4",
			OnStart: []content.EffectRef{{Do: "test.gate", Params: map[string]string{}}},
		},
		"load": {Uses: use("body", 1), Target: "none", Choices: []string{"type"}, Windup: "0", Recovery: "2"},
	}
	for key, def := range defs {
		def.Name, def.Key = api.ContentID("text.action.test."+key), key
		bundle.Actions["action.test."+key] = def
		bundle.Text["text.action.test."+key] = strings.ToUpper(key[:1]) + key[1:]
	}
	bundle.Text["text.template.test.fighter"] = "Clay-thing"
	bundle.Text["text.template.test.dummy"] = "Dummy"
	bundle.Templates["template.test.fighter"] = content.Template{
		Actor: &content.ActorDef{Controller: timeline.PlayerController}, Actions: known,
		// The types a body knows innately are the sample's content choice's business.
		Data: content.Data{Lists: map[string][]string{
			"types": {"type.cut", "type.name", "type.press", "type.relic"},
		}},
	}
	bundle.Templates["template.test.dummy"] = content.Template{Actor: &content.ActorDef{Controller: "idle"}}
	if o.mutate != nil {
		o.mutate(&bundle)
	}
	return bundle
}

func newFixture(t *testing.T, seed uint64, o opts) *fixture {
	t.Helper()
	if o.arms == 0 {
		o.arms = 3
	}
	if o.dummyQ == 0 {
		o.dummyQ = 1
	}
	reg, err := content.New(testBundle(t, o))
	if err != nil {
		t.Fatal(err)
	}
	fx := &fixture{t: t}
	fx.cfg = game.Config{Seed: seed, Content: reg, Spawn: func(w *sim.World, _ *clock.Clock) error {
		fx.arms = nil
		at := func(q int) *comp.Pos { return &comp.Pos{Map: 1, At: api.Cell{Hex: api.Hex{Q: q}}} }
		var err error
		if fx.fighter, err = w.Spawn("template.test.fighter", at(0)); err != nil {
			return err
		}
		if fx.dummy, err = w.Spawn("template.test.dummy", at(o.dummyQ)); err != nil {
			return err
		}
		for range o.arms {
			arm, err := w.Spawn("template.test.dummy", nil)
			if err != nil {
				return err
			}
			ecs.NewMap[comp.Actor](w.ECS).Remove(arm) // a part is not on the timeline itself
			fx.arms = append(fx.arms, arm)
		}
		return nil
	}, Install: func(w *sim.World, _ *clock.Clock) error {
		fx.w = w
		timeline.SetVisible(w, func(*sim.World, ecs.Entity) bool { return true })
		timeline.SetChannels(w, func(_ *sim.World, body ecs.Entity) []timeline.Channel {
			channels := []timeline.Channel{{Entity: body, Tags: []string{"body"}}}
			if body == fx.fighter {
				for _, arm := range fx.arms {
					channels = append(channels, timeline.Channel{Entity: arm, Body: body, Tags: []string{"arm"}})
				}
			}
			return channels
		})
		err = effects.Register(w, effects.Primitive{
			Name: "test.hit", Apply: func(*effects.Context, effects.Params) error { fx.hits++; return nil },
		})
		if err != nil {
			return err
		}
		err = effects.Register(w, effects.Primitive{
			Name:  "test.gate",
			Apply: func(*effects.Context, effects.Params) error { return nil },
			Check: func(*effects.Context, effects.Params) (bool, api.Text) {
				return false, api.Text{ID: "text.explore.bump"}
			},
		})
		if err != nil {
			return err
		}
		if o.install != nil {
			if err := o.install(w); err != nil {
				return err
			}
		}
		if o.contentChoices {
			return nil
		}
		err = actions.RegisterChoice(w, "type", func(*sim.World, ecs.Entity) []string {
			return []string{"type.a", "type.b"}
		})
		if err != nil || !o.hooks {
			return err
		}
		fx.addHooks(w)
		return nil
	}}
	if fx.g, err = game.New(fx.cfg); err != nil {
		t.Fatal(err)
	}
	return fx
}

// addHooks registers a validator that refuses type.a and a modifier that makes type.b take 5 more
// ticks to recover.
func (fx *fixture) addHooks(w *sim.World) {
	actions.AddValidator(w, func(_ *sim.World, in actions.Intent) (bool, api.Text) {
		if in.Choices["type"] == "type.a" {
			return false, api.Text{ID: "text.actions.requires"}
		}
		return true, api.Text{}
	})
	actions.AddModifier(w, func(_ *sim.World, in actions.Intent) actions.Modifier {
		if in.Choices["type"] == "type.b" {
			return actions.Modifier{Recovery: 5}
		}
		return actions.Modifier{}
	})
}

func (fx *fixture) uid(e ecs.Entity) api.UID { return fx.w.UIDOf(e) }

func (fx *fixture) view() api.View {
	fx.t.Helper()
	view, err := fx.g.View()
	if err != nil {
		fx.t.Fatal(err)
	}
	return view
}

// option finds the first option whose ID starts with prefix.
func (fx *fixture) option(prefix string) api.Option {
	fx.t.Helper()
	for _, option := range fx.view().Options {
		if strings.HasPrefix(option.ID, prefix) {
			return option
		}
	}
	fx.t.Fatalf("no option starts with %q", prefix)
	return api.Option{}
}

func (fx *fixture) optionIDs() []string {
	var ids []string
	for _, option := range fx.view().Options {
		ids = append(ids, option.ID)
	}
	return ids
}

func (fx *fixture) apply(cmd api.Command) []api.Event {
	fx.t.Helper()
	events, err := fx.g.Apply(cmd)
	if err != nil {
		fx.t.Fatal(err)
	}
	return events
}

func (fx *fixture) wait(ticks int64) []api.Event {
	fx.t.Helper()
	return fx.apply(api.Wait{Actor: fx.uid(fx.fighter), Until: api.WaitTicks, Ticks: ticks})
}

func find[E api.Event](events []api.Event) (E, bool) {
	for _, event := range events {
		if typed, ok := event.(E); ok {
			return typed, true
		}
	}
	var zero E
	return zero, false
}

func uidSegments(fx *fixture, entities ...ecs.Entity) string {
	var out []string
	for _, e := range entities {
		out = append(out, strconv.FormatUint(uint64(fx.uid(e)), 10))
	}
	return strings.Join(out, "/")
}

func TestTwoArmActionNeedsBothFreeAndOffersEachPair(t *testing.T) {
	fx := newFixture(t, 1, opts{actions: []api.ContentID{"action.test.grab"}})
	starts := actions.Startable(fx.w, fx.fighter)
	if len(starts) != 3 {
		t.Fatalf("three free arms gave %d intents, want 3 pairs", len(starts))
	}
	a1, a2, a3 := fx.arms[0], fx.arms[1], fx.arms[2]
	wantPairs := [][]ecs.Entity{{a1, a2}, {a1, a3}, {a2, a3}}
	for i, in := range starts {
		var got []ecs.Entity
		for _, c := range in.Channels {
			got = append(got, c.Entity)
		}
		if !slices.Equal(got, wantPairs[i]) {
			t.Errorf("intent %d uses %v, want %v", i, got, wantPairs[i])
		}
	}
	wantIDs := []string{
		"grab/" + uidSegments(fx, a1, a2), "grab/" + uidSegments(fx, a1, a3), "grab/" + uidSegments(fx, a2, a3),
	}
	var gotIDs []string
	for _, id := range fx.optionIDs() {
		if strings.HasPrefix(id, "grab") {
			gotIDs = append(gotIDs, id)
		}
	}
	if !slices.Equal(gotIDs, wantIDs) {
		t.Errorf("option IDs = %v, want %v", gotIDs, wantIDs)
	}

	if err := actions.Start(fx.w, starts[0]); err != nil {
		t.Fatal(err)
	}
	for i, arm := range fx.arms {
		free := timeline.Free(fx.w, timeline.Channel{Entity: arm, Body: fx.fighter})
		if want := i == 2; free != want {
			t.Errorf("arm %d free = %v, want %v", i, free, want)
		}
	}
	if left := actions.Startable(fx.w, fx.fighter); len(left) != 0 {
		t.Errorf("one free arm left %d intents, want none", len(left))
	}
	if fx.hits != 1 {
		t.Errorf("hits = %d, want the windup-0 action to have landed once", fx.hits)
	}
}

func TestTwoArmsOfTwoOffersOneUnnamedPair(t *testing.T) {
	fx := newFixture(t, 1, opts{actions: []api.ContentID{"action.test.grab"}, arms: 2})
	if got := fx.option("grab").ID; got != "grab" {
		t.Errorf("only one pair, so the option ID = %q, want %q", got, "grab")
	}
	if err := timeline.Occupy(fx.w, timeline.Channel{Entity: fx.arms[0], Body: fx.fighter}, 5); err != nil {
		t.Fatal(err)
	}
	if left := actions.Startable(fx.w, fx.fighter); len(left) != 0 {
		t.Errorf("a busy arm left %d intents, want none", len(left))
	}
}

func TestWindupLandsLaterAndMissesWhenTargetLeaves(t *testing.T) {
	fx := newFixture(t, 1, opts{})
	started := fx.apply(fx.option("swing/").Command)
	event, ok := find[api.ActionStarted](started)
	if !ok || event.Lands != fx.w.Now+3 {
		t.Fatalf("ActionStarted = %+v, %v; want it to land at now+3 = %d", event, ok, fx.w.Now+3)
	}
	if pending := timeline.PendingOf(fx.w, fx.uid(fx.fighter)); len(pending) != 1 {
		t.Fatalf("pending actions = %d, want 1", len(pending))
	}
	if fx.hits != 0 {
		t.Fatalf("hits = %d before the windup ended", fx.hits)
	}
	ecs.NewMap[comp.Pos](fx.w.ECS).Get(fx.dummy).At.Hex = api.Hex{Q: 9}
	events := fx.wait(3)
	missed, ok := find[api.ActionMissed](events)
	if !ok {
		t.Fatalf("no ActionMissed in %v", events)
	}
	if got := fx.w.Content.Render(missed.Line); got != "Clay-thing's Swing hits empty air." {
		t.Errorf("miss line = %q", got)
	}
	if fx.hits != 0 {
		t.Errorf("hits = %d, want the swing to have missed", fx.hits)
	}
}

func TestWindupHitsWhenTargetStaysInRange(t *testing.T) {
	fx := newFixture(t, 1, opts{})
	fx.apply(fx.option("swing/").Command)
	label := ""
	for _, entry := range fx.view().Timeline.Entries {
		if entry.Kind == "lands" {
			label = entry.Label
		}
	}
	if want := "Clay-thing's Swing lands on Dummy"; label != want {
		t.Errorf("timeline label = %q, want %q", label, want)
	}
	events := fx.wait(3)
	if _, ok := find[api.ActionLanded](events); !ok || fx.hits != 1 {
		t.Errorf("landed = %v, hits = %d; want it to land and hit once", ok, fx.hits)
	}
	if _, ok := find[api.ActionMissed](events); ok {
		t.Error("a swing at a target in reach missed")
	}
}

func TestLockingOneUsedChannelCancelsTheAction(t *testing.T) {
	fx := newFixture(t, 1, opts{})
	fx.apply(fx.option("slam/").Command)
	if pending := timeline.PendingOf(fx.w, fx.uid(fx.fighter)); len(pending) != 1 {
		t.Fatalf("pending actions = %d, want 1", len(pending))
	}
	// Lock the second arm only; the action's own part is the first.
	lock := []content.EffectRef{{Do: "lock", Params: map[string]string{"channels": "target_part", "ticks": "20"}}}
	ctx := &effects.Context{World: fx.w, Source: fx.fighter, Target: fx.fighter, TargetPart: fx.arms[1]}
	if err := effects.Apply(ctx, lock); err != nil {
		t.Fatal(err)
	}
	if pending := timeline.PendingOf(fx.w, fx.uid(fx.fighter)); len(pending) != 0 {
		t.Fatalf("pending actions = %d after the lock, want 0", len(pending))
	}
	events := fx.wait(3)
	if _, ok := find[api.ActionLanded](events); ok || fx.hits != 0 {
		t.Errorf("a canceled action landed (hits = %d)", fx.hits)
	}
}

func TestFailingCheckDisablesTheOptionWithItsReason(t *testing.T) {
	fx := newFixture(t, 1, opts{})
	option := fx.option("gate")
	if option.Enabled || option.Reason != "Packed clay. No way through." {
		t.Errorf("option = enabled %v, reason %q; want disabled with the check's reason", option.Enabled, option.Reason)
	}
	ok, reason := actions.Validate(fx.w, actions.Intent{Actor: fx.fighter, Action: "action.test.gate"})
	if ok || reason.ID != "text.explore.bump" {
		t.Errorf("Validate = %v, %v; want it refused with text.explore.bump", ok, reason)
	}
	if left := actions.Startable(fx.w, fx.fighter); slices.ContainsFunc(left, func(in actions.Intent) bool {
		return in.Action == "action.test.gate"
	}) {
		t.Error("Startable lists an intent a check refuses")
	}
	events := fx.apply(option.Command)
	if rejected, ok := find[api.Rejected](events); !ok || rejected.Reason.ID != "text.explore.bump" {
		t.Errorf("Do of a refused option gave %v, want Rejected with the check's reason", events)
	}
}

func TestUsableHookLetsABodyWithAUselessArmWaitForItsBody(t *testing.T) {
	// The dummy is out of the swing's reach, so the free arms are useless and the body must wait
	// for its own channel to recover from the step.
	far := newFixture(t, 1, opts{
		actions: []api.ContentID{"action.step", "action.test.swing"}, dummyQ: 9,
	})
	start := far.w.Now
	far.apply(far.option("step/e").Command)
	if got, want := far.w.Now-start, int64(10); got != want {
		t.Errorf("with useless arms the step passed %d ticks before the next decision, want %d", got, want)
	}

	near := newFixture(t, 1, opts{actions: []api.ContentID{"action.step", "action.test.swing"}})
	start = near.w.Now
	near.apply(near.option("step/e").Command)
	if got := near.w.Now - start; got != 0 {
		t.Errorf("with a target in reach the arms let the body decide at once, but %d ticks passed", got)
	}
}

func TestEveryOptionValidatesAndABusyChannelIsRejected(t *testing.T) {
	probe := newFixture(t, 1, opts{hooks: true})
	enabled := 0
	for _, option := range probe.view().Options {
		do, ok := option.Command.(api.Do)
		if !ok || !option.Enabled {
			continue
		}
		enabled++
		fresh := newFixture(t, 1, opts{hooks: true})
		events := fresh.apply(do)
		if rejected, ok := find[api.Rejected](events); ok {
			t.Errorf("option %q was rejected: %v", option.ID, rejected.Reason)
		}
	}
	if enabled == 0 {
		t.Fatal("no enabled action options to check")
	}

	fx := newFixture(t, 1, opts{})
	do := fx.option("swing/").Command.(api.Do)
	fx.apply(do)
	events := fx.apply(do) // the same arm again
	rejected, ok := find[api.Rejected](events)
	if !ok || rejected.Reason.ID != "text.actions.channels_busy" {
		t.Errorf("a Do on a busy channel gave %v, want Rejected for busy channels", events)
	}
	other := do
	other.Channels = []api.UID{fx.uid(fx.dummy)}
	if _, ok := find[api.Rejected](fx.apply(other)); !ok {
		t.Error("a Do naming a channel that is not the actor's was accepted")
	}
	stranger := api.Do{Actor: fx.uid(fx.dummy), Action: "action.step", Dir: new(api.Dir)}
	if _, ok := find[api.Rejected](fx.apply(stranger)); !ok {
		t.Error("a Do from a body that is not deciding was accepted")
	}
}

func TestChoicesHooksAndOptionOrder(t *testing.T) {
	fx := newFixture(t, 1, opts{actions: []api.ContentID{"action.test.load"}, hooks: true})
	ids := fx.optionIDs()
	if want := []string{"load/type.a", "load/type.b"}; !slices.Equal(ids[len(ids)-2:], want) {
		t.Fatalf("last option IDs = %v, want %v", ids, want)
	}
	if option := fx.option("load/type.a"); option.Enabled || option.Reason != "Not right now." {
		t.Errorf("type.a = enabled %v, reason %q; want the validator's refusal", option.Enabled, option.Reason)
	}
	b := fx.option("load/type.b")
	if !b.Enabled || !strings.Contains(b.Label, "busy 7") {
		t.Errorf("type.b = %+v; want it enabled and busy 2+5 = 7 ticks", b)
	}
	if got := actions.Startable(fx.w, fx.fighter); len(got) != 1 || got[0].Choices["type"] != "type.b" {
		t.Errorf("Startable = %v, want only type.b", got)
	}
	bad := api.Do{
		Actor: fx.uid(fx.fighter), Action: "action.test.load", Choices: map[string]string{"type": "type.zzz"},
	}
	if rejected, ok := find[api.Rejected](fx.apply(bad)); !ok || rejected.Reason.ID != "text.actions.bad_choice" {
		t.Errorf("an unoffered choice was not rejected as bad_choice")
	}
	fx.apply(b.Command)
	if busy := fx.w.Now; busy != 7 {
		t.Errorf("after the action the next decision came at tick %d, want 7", busy)
	}
}

func TestStepOptionsAreTheSixDirections(t *testing.T) {
	fx := newFixture(t, 1, opts{actions: []api.ContentID{"action.step"}, arms: 1})
	var ids []string
	for _, id := range fx.optionIDs() {
		if strings.HasPrefix(id, "step/") {
			ids = append(ids, id)
		}
	}
	want := []string{"step/e", "step/ne", "step/nw", "step/w", "step/sw", "step/se"}
	if !slices.Equal(ids, want) {
		t.Errorf("step options = %v, want %v", ids, want)
	}
	if label := fx.option("step/nw").Label; label != "Step: northwest (lands in 0, busy 10)" {
		t.Errorf("label = %q", label)
	}
}

func TestReplayOfRandomOptionsGivesTheSameHash(t *testing.T) {
	played := newFixture(t, 7, opts{hooks: true})
	pick := rng.New(99).Get("test")
	for range 200 {
		var enabled []api.Option
		for _, option := range played.view().Options {
			if option.Enabled {
				enabled = append(enabled, option)
			}
		}
		if len(enabled) == 0 {
			t.Fatal("the player has no enabled option")
		}
		played.apply(enabled[pick.IntN(len(enabled))].Command)
	}
	want, err := played.g.Hash()
	if err != nil {
		t.Fatal(err)
	}
	replayed := newFixture(t, 7, opts{hooks: true})
	g, err := game.Replay(replayed.cfg, played.g.History())
	if err != nil {
		t.Fatal(err)
	}
	if got, err := g.Hash(); err != nil || got != want {
		t.Errorf("replay hash = %x, %v; want %x", got, err, want)
	}
	if played.hits == 0 {
		t.Error("200 random options never landed a hit; the test is not exercising actions")
	}
}

// bare removes the test actions, whose primitives only the fixture registers.
func bare(b *content.Bundle) {
	for id := range b.Actions {
		if strings.HasPrefix(string(id), "action.test.") {
			delete(b.Actions, id)
		}
	}
	fighter := b.Templates["template.test.fighter"]
	fighter.Actions = []api.ContentID{"action.step"}
	b.Templates["template.test.fighter"] = fighter
}

func TestGameNewRejectsActionsContentCannotRun(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*content.Bundle)
		want   string
	}{
		{
			name: "an unknown primitive",
			mutate: func(b *content.Bundle) {
				bare(b)
				def := b.Actions["action.step"]
				def.OnLand = []content.EffectRef{{Do: "no.such.primitive", Params: map[string]string{}}}
				b.Actions["action.step"] = def
			},
			want: `actions["action.step"].on_land[0]: unknown primitive "no.such.primitive"`,
		},
		{
			name: "a choice nothing provides",
			mutate: func(b *content.Bundle) {
				bare(b)
				def := b.Actions["action.step"]
				def.Choices = []string{"nothing"}
				b.Actions["action.step"] = def
			},
			want: `nothing serves choice "nothing"`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reg, err := content.New(testBundle(t, opts{mutate: tc.mutate}))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := game.New(
				game.Config{Seed: 1, Content: reg},
			); err == nil ||
				!strings.Contains(err.Error(), tc.want) {
				t.Errorf("game.New error = %v, want it to contain %q", err, tc.want)
			}
		})
	}
}

func TestBatchUsableHookAgreesWithTheChannelByChannelTest(t *testing.T) {
	for _, dummyQ := range []int{1, 9} {
		fx := newFixture(t, 1, opts{
			actions: []api.ContentID{"action.step", "action.test.swing"}, dummyQ: dummyQ, hooks: true,
		})
		for step := 0; step < 6; step++ {
			for _, body := range []ecs.Entity{fx.fighter, fx.dummy} {
				channels := timeline.Channels(fx.w, body)
				got := actions.UsableChannels(fx.w, body, channels)
				for i, c := range channels {
					if want := actions.UsableOneByOne(fx.w, body, c); got[i] != want {
						t.Errorf("dummyQ %d step %d channel %d: batch %v, one by one %v", dummyQ, step, i, got[i], want)
					}
				}
			}
			var next *api.Option
			for _, option := range fx.view().Options {
				if option.Enabled {
					next = &option
					break
				}
			}
			if next == nil {
				break
			}
			fx.apply(next.Command)
		}
	}
}
