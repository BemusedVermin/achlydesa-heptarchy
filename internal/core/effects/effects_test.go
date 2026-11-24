package effects_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/clock"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/effects"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

const mapID api.MapID = 1

// fixture is a world with a small open map, a source S at the origin and bodies around it:
//
//	B (-2,0) hostile to S, S (0,0), A (2,0) with two arms and a leg, a wall at (3,0),
//	H (5,0) hidden behind the wall, F (10,0) beyond sight.
type fixture struct {
	w               *sim.World
	clk             *clock.Clock
	s, a, b, h, f   ecs.Entity
	arm1, arm2, leg ecs.Entity
	events          []api.Event
	parts           map[ecs.Entity][]ecs.Entity
	partTags        map[ecs.Entity][]string
}

func newFixture(t *testing.T) *fixture {
	t.Helper()
	reg := contenttest.Sample(t)
	fx := &fixture{
		w: sim.NewWorld(1), parts: map[ecs.Entity][]ecs.Entity{}, partTags: map[ecs.Entity][]string{},
	}
	fx.w.Content = reg
	fx.clk = clock.New(reg.Calendar())
	for _, step := range []func() error{
		func() error { return clock.Install(fx.w, fx.clk) },
		func() error { return timeline.Install(fx.w, fx.clk) },
		func() error { return effects.Install(fx.w) },
	} {
		if err := step(); err != nil {
			t.Fatal(err)
		}
	}
	timeline.SetVisible(fx.w, func(*sim.World, ecs.Entity) bool { return true })
	timeline.SetChannels(fx.w, fx.channelsOf)
	m := hexmap.NewMap(mapID, "room", hexmap.FuncSource(func(c api.Cell) hexmap.Terrain {
		return hexmap.Terrain{ID: "terrain.air", Opaque: c.Hex == api.Hex{Q: 3, R: 0} && c.Depth == 0}
	}), hexmap.Prisms{})
	effects.SetMaps(fx.w, func(id api.MapID) *hexmap.Map {
		if id == mapID {
			return m
		}
		return nil
	})
	fx.s = fx.spawn(t, "idle", 0, 0)
	fx.a = fx.spawn(t, timeline.PlayerController, 2, 0)
	fx.b = fx.spawn(t, "idle", -2, 0)
	fx.h = fx.spawn(t, "idle", 5, 0)
	fx.f = fx.spawn(t, "idle", 10, 0)
	fx.arm1 = fx.part(t, fx.a, "arm")
	fx.arm2 = fx.part(t, fx.a, "arm")
	fx.leg = fx.part(t, fx.a, "leg")
	effects.SetHostile(fx.w, func(w *sim.World, from, to ecs.Entity) bool { return from == fx.s && to == fx.b })
	fx.w.Bus.SubscribeAll("test.events", func(_ *sim.World, e api.Event) { fx.events = append(fx.events, e) })
	err := timeline.RegisterResolver(fx.w, "test.swing", timeline.Resolver{
		Resolve: func(*sim.World, ecs.Entity) error { return nil },
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := fx.w.Systems.Freeze(); err != nil {
		t.Fatal(err)
	}
	return fx
}

func (fx *fixture) spawn(t *testing.T, controller string, q, r int) ecs.Entity {
	t.Helper()
	e, err := fx.w.Spawn("template.person.reaper", &comp.Pos{Map: mapID, At: api.Cell{Hex: api.Hex{Q: q, R: r}}})
	if err != nil {
		t.Fatal(err)
	}
	ecs.NewMap[comp.Actor](fx.w.ECS).Get(e).Controller = controller
	return e
}

// part makes an entity that is a channel of body, tagged tag.
func (fx *fixture) part(t *testing.T, body ecs.Entity, tag string) ecs.Entity {
	t.Helper()
	e, err := fx.w.Spawn("template.person.reaper", nil)
	if err != nil {
		t.Fatal(err)
	}
	ecs.NewMap[comp.Actor](fx.w.ECS).Remove(e)
	fx.parts[body] = append(fx.parts[body], e)
	fx.partTags[e] = []string{tag}
	return e
}

func (fx *fixture) channelsOf(_ *sim.World, body ecs.Entity) []timeline.Channel {
	channels := []timeline.Channel{{Entity: body, Tags: []string{"body"}}}
	for _, part := range fx.parts[body] {
		channels = append(channels, timeline.Channel{Entity: part, Body: body, Tags: fx.partTags[part]})
	}
	return channels
}

func (fx *fixture) channel(e ecs.Entity) timeline.Channel {
	for body, parts := range fx.parts {
		for _, part := range parts {
			if part == e {
				return timeline.Channel{Entity: e, Body: body, Tags: fx.partTags[e]}
			}
		}
	}
	return timeline.Channel{Entity: e, Tags: []string{"body"}}
}

// swing starts a pending action by actor that uses part and lands at tick.
func (fx *fixture) swing(t *testing.T, actor, part ecs.Entity, lands int64) {
	t.Helper()
	_, err := timeline.Schedule(fx.w, comp.Pending{
		Kind: "test.swing", Actor: fx.w.UIDOf(actor), Channel: fx.w.UIDOf(part), Lands: lands,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func (fx *fixture) ctx() *effects.Context {
	return &effects.Context{World: fx.w, Source: fx.s, Target: fx.a}
}

func (fx *fixture) apply(t *testing.T, ctx *effects.Context, refs ...content.EffectRef) []api.Event {
	t.Helper()
	fx.events = nil
	if err := effects.Apply(ctx, refs); err != nil {
		t.Fatal(err)
	}
	if _, err := fx.w.Bus.Drain(fx.w); err != nil {
		t.Fatal(err)
	}
	return fx.events
}

func (fx *fixture) uids(entities ...ecs.Entity) []api.UID {
	uids := []api.UID{}
	for _, e := range entities {
		uids = append(uids, fx.w.UIDOf(e))
	}
	return uids
}

func TestBodySelectors(t *testing.T) {
	fx := newFixture(t)
	tests := []struct {
		src  string
		want []ecs.Entity
	}{
		{"self", []ecs.Entity{fx.s}},
		{"target", []ecs.Entity{fx.a}},
		{"others", []ecs.Entity{fx.a, fx.b}},
		{"everyone", []ecs.Entity{fx.s, fx.a, fx.b}},
		{"allies", []ecs.Entity{fx.a}},
		{"enemies", []ecs.Entity{fx.b}},
	}
	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			sel, err := effects.ParseBodies(tt.src)
			if err != nil {
				t.Fatal(err)
			}
			got, err := sel.Resolve(fx.ctx())
			if err != nil {
				t.Fatal(err)
			}
			if len(got) == 0 && len(tt.want) == 0 {
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s = %v, want %v", tt.src, fx.uids(got...), fx.uids(tt.want...))
			}
		})
	}
}

func TestSightHidesWallsAndDistance(t *testing.T) {
	fx := newFixture(t)
	sel, _ := effects.ParseBodies("everyone")
	got, err := sel.Resolve(fx.ctx())
	if err != nil {
		t.Fatal(err)
	}
	for _, hidden := range []ecs.Entity{fx.h, fx.f} {
		for _, body := range got {
			if body == hidden {
				t.Errorf("body %d is behind a wall or out of range but was selected", fx.w.UIDOf(hidden))
			}
		}
	}
}

func TestSelectorsWithNoSourceOrMaps(t *testing.T) {
	fx := newFixture(t)
	sel, _ := effects.ParseBodies("others")
	if got, err := sel.Resolve(&effects.Context{World: fx.w}); err != nil || len(got) != 0 {
		t.Errorf("others with no source = %v, %v; want none", got, err)
	}
	effects.SetMaps(fx.w, nil)
	if got, err := sel.Resolve(fx.ctx()); err != nil || len(got) != 0 {
		t.Errorf("others with no maps = %v, %v; want none", got, err)
	}
}

func TestChannelSelectors(t *testing.T) {
	fx := newFixture(t)
	ctx := fx.ctx()
	ctx.Used = []timeline.Channel{fx.channel(fx.arm2)}
	ctx.TargetPart = fx.leg
	tests := []struct {
		src  string
		want []ecs.Entity
	}{
		{"target:all", []ecs.Entity{fx.a, fx.arm1, fx.arm2, fx.leg}},
		{"target:arm", []ecs.Entity{fx.arm1, fx.arm2}},
		{"target:leg", []ecs.Entity{fx.leg}},
		{"target:wing", nil},
		{"self:body", []ecs.Entity{fx.s}},
		{"others:body", []ecs.Entity{fx.a, fx.b}},
		{"used", []ecs.Entity{fx.arm2}},
		{"target_part", []ecs.Entity{fx.leg}},
	}
	for _, tt := range tests {
		t.Run(tt.src, func(t *testing.T) {
			sel, err := effects.ParseChannels(tt.src)
			if err != nil {
				t.Fatal(err)
			}
			channels, err := sel.Resolve(ctx)
			if err != nil {
				t.Fatal(err)
			}
			var got []ecs.Entity
			for _, c := range channels {
				got = append(got, c.Entity)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("%s = %v, want %v", tt.src, fx.uids(got...), fx.uids(tt.want...))
			}
		})
	}
}

func TestMalformedSelectors(t *testing.T) {
	for _, src := range []string{"", "nobody", "target:", "target:Arm", "target:two words", "used:arm", "nobody:arm", "arm"} {
		if _, err := effects.ParseChannels(src); err == nil {
			t.Errorf("ParseChannels(%q) succeeded, want an error", src)
		}
	}
	for _, src := range []string{"", "target:arm", "Self", "used", "others[", "others[]", "others[Crawls]", "[crawls]"} {
		if _, err := effects.ParseBodies(src); err == nil {
			t.Errorf("ParseBodies(%q) succeeded, want an error", src)
		}
	}
}

func TestBodySelectorNarrowsByTemplateTag(t *testing.T) {
	fx := newFixture(t)
	sel, err := effects.ParseBodies("others[nosuchtag]")
	if err != nil {
		t.Fatal(err)
	}
	if got, err := sel.Resolve(fx.ctx()); err != nil || len(got) != 0 {
		t.Errorf("others[nosuchtag] = %v, %v; want none", got, err)
	}
	sel, _ = effects.ParseBodies("others[graftable]")
	if got, err := sel.Resolve(fx.ctx()); err != nil || len(got) != 2 {
		t.Errorf("others[graftable] = %v, %v; want the two others in sight", got, err)
	}
}

func TestLockCancelsWindUpsAndDelaysDecision(t *testing.T) {
	fx := newFixture(t)
	if _, err := timeline.Run(fx.w, fx.clk); err != nil || fx.w.Now != 0 {
		t.Fatalf("first Run: now = %d, err = %v", fx.w.Now, err)
	}
	fx.swing(t, fx.a, fx.arm1, 5)
	fx.swing(t, fx.a, fx.arm2, 6)
	events := fx.apply(t, fx.ctx(), contenttest.Do("lock", "channels", "target:all", "ticks", "8"))
	if left := timeline.PendingOf(fx.w, fx.w.UIDOf(fx.a)); len(left) != 0 {
		t.Errorf("%d pending actions left, want both swings canceled", len(left))
	}
	var canceled int
	var locked *api.ChannelsLocked
	for _, e := range events {
		switch e := e.(type) {
		case api.ActionCanceled:
			canceled++
			if e.Reason.ID != "text.effects.interrupted" {
				t.Errorf("reason = %q", e.Reason.ID)
			}
		case api.ChannelsLocked:
			locked = &e
		}
	}
	if canceled != 2 {
		t.Errorf("ActionCanceled events = %d, want 2", canceled)
	}
	if locked == nil || locked.Ticks != 8 ||
		!reflect.DeepEqual(locked.Channels, fx.uids(fx.a, fx.arm1, fx.arm2, fx.leg)) {
		t.Errorf("ChannelsLocked = %+v", locked)
	}
	if body, err := timeline.Run(fx.w, fx.clk); err != nil || body != fx.a || fx.w.Now != 8 {
		t.Errorf("Run = (%v, %v) at tick %d; want the locked body at tick 8", body, err, fx.w.Now)
	}
}

func TestInterrupt(t *testing.T) {
	t.Run("does nothing to an idle body", func(t *testing.T) {
		fx := newFixture(t)
		ctx := &effects.Context{World: fx.w, Source: fx.s, Target: fx.b}
		events := fx.apply(t, ctx, contenttest.Do("interrupt", "bodies", "target", "when", "target.winding_up"))
		if len(events) != 0 {
			t.Errorf("events = %v, want none", events)
		}
	})
	t.Run("cancels a body's wind-up when the condition holds", func(t *testing.T) {
		fx := newFixture(t)
		fx.swing(t, fx.a, fx.arm1, 5)
		events := fx.apply(t, fx.ctx(), contenttest.Do("interrupt", "bodies", "target", "when", "target.winding_up"))
		if len(timeline.PendingOf(fx.w, fx.w.UIDOf(fx.a))) != 0 {
			t.Error("the swing should be canceled")
		}
		last := events[len(events)-1]
		if got, ok := last.(api.ActionsInterrupted); !ok || !reflect.DeepEqual(got.Actors, fx.uids(fx.a)) {
			t.Errorf("last event = %+v, want ActionsInterrupted naming the target", last)
		}
		if !timeline.Free(fx.w, fx.channel(fx.arm1)) {
			t.Error("interrupt should not keep the arm busy")
		}
	})
	t.Run("leaves wind-ups alone when the condition does not hold", func(t *testing.T) {
		fx := newFixture(t)
		fx.swing(t, fx.a, fx.arm1, 5)
		fx.apply(t, fx.ctx(), contenttest.Do("interrupt", "bodies", "target", "when", "now > 100"))
		if len(timeline.PendingOf(fx.w, fx.w.UIDOf(fx.a))) != 1 {
			t.Error("the swing should still be pending")
		}
	})
	t.Run("can pick channels instead of bodies", func(t *testing.T) {
		fx := newFixture(t)
		fx.swing(t, fx.a, fx.arm1, 5)
		fx.swing(t, fx.a, fx.arm2, 6)
		fx.apply(t, fx.ctx(), contenttest.Do("interrupt", "channels", "target_part"))
		if left := timeline.PendingOf(fx.w, fx.w.UIDOf(fx.a)); len(left) != 2 {
			t.Errorf("pending = %d, want 2: no target part is set", len(left))
		}
		ctx := fx.ctx()
		ctx.TargetPart = fx.arm1
		fx.apply(t, ctx, contenttest.Do("interrupt", "channels", "target_part"))
		if left := timeline.PendingOf(fx.w, fx.w.UIDOf(fx.a)); len(left) != 1 {
			t.Errorf("pending = %d, want only the other arm's swing", len(left))
		}
	})
}

func TestDelayPushesBackEveryVisibleBodyButTheSource(t *testing.T) {
	fx := newFixture(t)
	events := fx.apply(t, fx.ctx(), contenttest.Do("delay", "channels", "others:all", "ticks", "8"))
	for _, e := range []ecs.Entity{fx.a, fx.arm1, fx.arm2, fx.leg, fx.b} {
		if timeline.Free(fx.w, fx.channel(e)) {
			t.Errorf("channel of %d should be busy", fx.w.UIDOf(e))
		}
	}
	for _, e := range []ecs.Entity{fx.s, fx.h, fx.f} {
		if !timeline.Free(fx.w, fx.channel(e)) {
			t.Errorf("channel of %d should be free: the source, or out of sight", fx.w.UIDOf(e))
		}
	}
	if got, ok := events[len(events)-1].(api.ChannelsDelayed); !ok || got.Ticks != 8 || len(got.Channels) != 5 {
		t.Errorf("event = %+v, want ChannelsDelayed for five channels", events[len(events)-1])
	}
	fx.apply(t, fx.ctx(), contenttest.Do("delay", "channels", "target:body", "ticks", "4"))
	if timeline.Free(fx.w, fx.channel(fx.a)) {
		t.Error("the body channel should be busy until tick 12")
	}
}

func TestReadyFreesTheLowestBusyChannel(t *testing.T) {
	fx := newFixture(t)
	for _, e := range []ecs.Entity{fx.arm1, fx.arm2, fx.leg} {
		if err := timeline.Occupy(fx.w, fx.channel(e), 10); err != nil {
			t.Fatal(err)
		}
	}
	events := fx.apply(t, fx.ctx(), contenttest.Do("ready", "channels", "target:all", "count", "1"))
	if !timeline.Free(fx.w, fx.channel(fx.arm1)) {
		t.Error("the lowest-UID busy channel should be free")
	}
	for _, e := range []ecs.Entity{fx.arm2, fx.leg} {
		if timeline.Free(fx.w, fx.channel(e)) {
			t.Errorf("channel %d should still be busy", fx.w.UIDOf(e))
		}
	}
	if got, ok := events[len(events)-1].(api.ChannelsReadied); !ok ||
		!reflect.DeepEqual(got.Channels, fx.uids(fx.arm1)) {
		t.Errorf("event = %+v, want ChannelsReadied naming the first arm", events[len(events)-1])
	}
	fx.apply(t, fx.ctx(), contenttest.Do("ready", "channels", "target:all", "count", "5"))
	for _, e := range []ecs.Entity{fx.arm2, fx.leg} {
		if !timeline.Free(fx.w, fx.channel(e)) {
			t.Errorf("channel %d should be free after readying up to five", fx.w.UIDOf(e))
		}
	}
	if events := fx.apply(
		t,
		fx.ctx(),
		contenttest.Do("ready", "channels", "target:all", "count", "1"),
	); len(
		events,
	) != 0 {
		t.Errorf("readying nothing busy published %v", events)
	}
}

func TestReadyLeavesAChannelWithASwingWindingUp(t *testing.T) {
	fx := newFixture(t)
	for _, e := range []ecs.Entity{fx.arm1, fx.arm2} {
		if err := timeline.Occupy(fx.w, fx.channel(e), 10); err != nil {
			t.Fatal(err)
		}
	}
	fx.swing(t, fx.a, fx.arm1, 6)
	events := fx.apply(t, fx.ctx(), contenttest.Do("ready", "channels", "target:all", "count", "5"))
	if timeline.Free(fx.w, fx.channel(fx.arm1)) {
		t.Error("the channel with a swing winding up should stay busy")
	}
	if !timeline.Free(fx.w, fx.channel(fx.arm2)) {
		t.Error("the channel with no swing should be free")
	}
	if got, ok := events[len(events)-1].(api.ChannelsReadied); !ok ||
		!reflect.DeepEqual(got.Channels, fx.uids(fx.arm2)) {
		t.Errorf("event = %+v, want ChannelsReadied naming only the second arm", events[len(events)-1])
	}
}

func TestParametersAreFormulas(t *testing.T) {
	fx := newFixture(t)
	var got struct {
		ticks    int
		distance int
		winding  bool
		roll     float64
	}
	err := effects.Register(fx.w, effects.Primitive{
		Name: "probe", Params: []string{"ticks", "distance", "winding", "roll"},
		Apply: func(_ *effects.Context, p effects.Params) (err error) {
			if got.ticks, err = p.Int("ticks"); err != nil {
				return err
			}
			var d float64
			if d, err = p.Float("distance"); err != nil {
				return err
			}
			got.distance = int(d)
			if got.winding, err = p.Bool("winding"); err != nil {
				return err
			}
			got.roll, err = p.Float("roll")
			return err
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	fx.swing(t, fx.a, fx.arm1, 5)
	fx.apply(
		t,
		fx.ctx(),
		contenttest.Do(
			"probe",
			"ticks",
			"4 + 4",
			"distance",
			"distance",
			"winding",
			"target.winding_up",
			"roll",
			"roll",
		),
	)
	if got.ticks != 8 || got.distance != 2 || !got.winding || got.roll < 0 || got.roll >= 1 {
		t.Errorf("probe saw %+v, want ticks 8, distance 2, winding up and a roll in [0,1)", got)
	}
}

func TestApplyStopsAtTheFirstError(t *testing.T) {
	fx := newFixture(t)
	err := effects.Apply(fx.ctx(), []content.EffectRef{
		contenttest.Do("delay", "channels", "target:body", "ticks", "0"),
		contenttest.Do("lock", "channels", "target:body", "ticks", "5"),
	})
	if err == nil || !strings.Contains(err.Error(), "delay") {
		t.Fatalf("error = %v, want the delay to fail", err)
	}
	if !timeline.Free(fx.w, fx.channel(fx.a)) {
		t.Error("the effects after the failing one should not have run")
	}
	if err := effects.Apply(fx.ctx(), []content.EffectRef{contenttest.Do("explode")}); err == nil {
		t.Error("an unknown primitive should be an error")
	}
}

func TestCheckReportsTheFirstReason(t *testing.T) {
	fx := newFixture(t)
	why := api.Text{ID: "text.test.nope"}
	err := effects.Register(fx.w, effects.Primitive{
		Name:  "gate",
		Apply: func(*effects.Context, effects.Params) error { return nil },
		Check: func(*effects.Context, effects.Params) (bool, api.Text) { return false, why },
	})
	if err != nil {
		t.Fatal(err)
	}
	if ok, _ := effects.Check(
		fx.ctx(),
		[]content.EffectRef{contenttest.Do("lock", "channels", "used", "ticks", "1")},
	); !ok {
		t.Error("a primitive with no Check always passes")
	}
	if ok, reason := effects.Check(fx.ctx(), []content.EffectRef{contenttest.Do("gate")}); ok || reason.ID != why.ID {
		t.Errorf("Check = %v, %v; want it to fail with the gate's reason", ok, reason)
	}
}

func TestRegisterRejectsBadPrimitives(t *testing.T) {
	fx := newFixture(t)
	apply := func(*effects.Context, effects.Params) error { return nil }
	for name, p := range map[string]effects.Primitive{
		"no name":         {Apply: apply},
		"no apply":        {Name: "x"},
		"duplicate":       {Name: "lock", Apply: apply},
		"repeated param":  {Name: "y", Params: []string{"a"}, Optional: []string{"a"}, Apply: apply},
		"kind of nothing": {Name: "z", Kinds: map[string]effects.Kind{"a": effects.KindText}, Apply: apply},
	} {
		if err := effects.Register(fx.w, p); err == nil {
			t.Errorf("%s: Register succeeded, want an error", name)
		}
	}
}

func TestValidateListsEveryProblem(t *testing.T) {
	fx := newFixture(t)
	err := effects.Validate(fx.w, "actions[x].on_start", []content.EffectRef{
		contenttest.Do("lock", "channels", "target:all", "ticks", "8"),
		contenttest.Do("explode"),
		contenttest.Do("lock", "channels", "target:all"),
		contenttest.Do("lock", "channels", "target:all", "ticks", "8", "force", "9"),
		contenttest.Do("delay", "channels", "everybody:all", "ticks", "8"),
		contenttest.Do("delay", "channels", "target:all", "ticks", "8 +"),
		contenttest.Do("interrupt", "bodies", "target", "channels", "used"),
		contenttest.Do("interrupt"),
		contenttest.Do("ready", "channels", "target:all", "count", "nonsense"),
	})
	if err == nil {
		t.Fatal("Validate succeeded, want every problem")
	}
	msg := err.Error()
	for _, want := range []string{
		`actions[x].on_start[1]: unknown primitive "explode"`,
		`[2] (lock): missing parameter "ticks"`,
		`[3] (lock): unknown parameter "force"`,
		`[4] (delay).channels: bad channel selector "everybody:all"`,
		`[5] (delay).ticks:`,
		`[6] (interrupt): give exactly one of bodies or channels`,
		`[7] (interrupt): give exactly one of bodies or channels`,
		`[8] (ready).count:`,
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("error is missing %q:\n%s", want, msg)
		}
	}
	if strings.Contains(msg, "[0]") {
		t.Errorf("the valid effect was reported:\n%s", msg)
	}
}

func TestInstallRegistersEvents(t *testing.T) {
	fx := newFixture(t)
	for _, event := range []api.Event{
		api.ChannelsLocked{}, api.ActionsInterrupted{}, api.ChannelsDelayed{}, api.ChannelsReadied{},
	} {
		data, err := fx.w.Codec.EncodeEvent(event)
		if err != nil {
			t.Fatalf("%s: %v", event.EventKind(), err)
		}
		if _, err := fx.w.Codec.DecodeEvent(data); err != nil {
			t.Errorf("%s did not round-trip: %v", event.EventKind(), err)
		}
	}
}

func TestChecksDoNotDrawFromTheRNG(t *testing.T) {
	fx := newFixture(t)
	err := effects.Register(fx.w, effects.Primitive{
		Name: "gauge", Params: []string{"limit"},
		Apply: func(*effects.Context, effects.Params) error { return nil },
		Check: func(_ *effects.Context, p effects.Params) (bool, api.Text) {
			limit, err := p.Int("limit")
			return err == nil && limit > 0, api.Text{ID: "text.test.nope"}
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	before, err := fx.w.RNG.State()
	if err != nil {
		t.Fatal(err)
	}
	refs := []content.EffectRef{contenttest.Do("gauge", "limit", "distance + 1")}
	for range 2 {
		if ok, _ := effects.Check(fx.ctx(), refs); !ok {
			t.Fatal("the check should pass")
		}
	}
	after, err := fx.w.RNG.State()
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, after) {
		t.Error("checking an effect changed the RNG state")
	}
	if ok, _ := effects.Check(fx.ctx(), []content.EffectRef{contenttest.Do("gauge", "limit", "roll")}); ok {
		t.Error("a check that reads roll should fail rather than draw")
	}
	if err := effects.Validate(
		fx.w,
		"x",
		[]content.EffectRef{contenttest.Do("gauge", "limit", "roll * 2")},
	); err == nil {
		t.Error("Validate should reject a primitive with a Check reading roll")
	}
	if err := effects.Validate(fx.w, "x", refs); err != nil {
		t.Errorf("Validate rejected a check that does not read roll: %v", err)
	}
}
