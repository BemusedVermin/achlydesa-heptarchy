package timeline_test

import (
	"reflect"
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

// swing is a resolver that logs each landing as "<tick>:<seq>".
type swing struct {
	log    []string
	events []api.Event
}

// pendingFixture has a player at its first decision (tick 0), a "swing" resolver and event capture.
func pendingFixture(t *testing.T) (*fixture, ecs.Entity, *swing) {
	t.Helper()
	fx := newFixture(t, 1)
	player := fx.spawn(t, timeline.PlayerController, 2)
	log := &swing{}
	err := timeline.RegisterResolver(fx.w, "test.swing", timeline.Resolver{
		Resolve: func(w *sim.World, e ecs.Entity) error {
			p := ecs.NewMap[comp.Pending](w.ECS).Get(e)
			log.log = append(log.log, tickSeq(w.Now, p.Seq))
			return nil
		},
		Label: func(*sim.World, ecs.Entity) string { return "a swing" },
	})
	if err != nil {
		t.Fatal(err)
	}
	sim.Subscribe(
		fx.w.Bus,
		"test.landed",
		func(_ *sim.World, e api.ActionLanded) { log.events = append(log.events, e) },
	)
	sim.Subscribe(
		fx.w.Bus,
		"test.canceled",
		func(_ *sim.World, e api.ActionCanceled) { log.events = append(log.events, e) },
	)
	if _, err := timeline.Run(fx.w, fx.clk); err != nil || fx.w.Now != 0 {
		t.Fatalf("first Run: now = %d, err = %v", fx.w.Now, err)
	}
	return fx, player, log
}

func tickSeq(tick int64, seq uint64) string {
	return string(rune('0'+tick/10)) + string(rune('0'+tick%10)) + ":" + string(rune('0'+seq))
}

func schedule(t *testing.T, fx *fixture, actor, target api.UID, lands int64) ecs.Entity {
	t.Helper()
	e, err := timeline.Schedule(
		fx.w,
		comp.Pending{Kind: "test.swing", Actor: actor, Channel: actor, Target: target, Lands: lands},
	)
	if err != nil {
		t.Fatal(err)
	}
	return e
}

func TestLandingsComeBeforeDecisionsAndGoBySeq(t *testing.T) {
	fx, player, log := pendingFixture(t)
	uid := fx.w.UIDOf(player)
	schedule(t, fx, uid, 0, 12)
	schedule(t, fx, uid, 0, 12)
	if err := timeline.Occupy(fx.w, fx.bodyChannel(player), 12); err != nil {
		t.Fatal(err)
	}
	if _, err := timeline.Run(fx.w, fx.clk); err != nil || fx.w.Now != 12 {
		t.Fatalf("Run: now = %d, err = %v; want the player at 12", fx.w.Now, err)
	}
	if want := []string{"12:1", "12:2"}; !reflect.DeepEqual(log.log, want) {
		t.Errorf("landings = %v, want %v (both before the decision at 12)", log.log, want)
	}
	if len(timeline.PendingOf(fx.w, uid)) != 0 {
		t.Error("landed actions should be gone")
	}
	if len(log.events) != 2 {
		t.Errorf("events = %v, want two ActionLanded", log.events)
	}
}

func TestScheduleRejectsBadInput(t *testing.T) {
	fx, player, _ := pendingFixture(t)
	uid := fx.w.UIDOf(player)
	if _, err := timeline.Schedule(fx.w, comp.Pending{Kind: "nope", Actor: uid, Lands: 5}); err == nil {
		t.Error("a kind without a resolver should fail")
	}
	if _, err := timeline.Schedule(fx.w, comp.Pending{Kind: "test.swing", Actor: uid, Lands: 0}); err == nil {
		t.Error("landing now should fail")
	}
	err := timeline.RegisterResolver(
		fx.w,
		"test.swing",
		timeline.Resolver{Resolve: func(*sim.World, ecs.Entity) error { return nil }},
	)
	if err == nil {
		t.Error("a duplicate resolver should fail")
	}
}

func TestCancelNeverResolves(t *testing.T) {
	fx, player, log := pendingFixture(t)
	uid := fx.w.UIDOf(player)
	pending := schedule(t, fx, uid, 0, 5)
	timeline.Cancel(fx.w, pending, "text.test.reason")
	if _, err := fx.w.Bus.Drain(fx.w); err != nil {
		t.Fatal(err)
	}
	if len(timeline.PendingOf(fx.w, uid)) != 0 || len(log.log) != 0 {
		t.Error("a canceled action should be gone and never resolve")
	}
	if len(log.events) != 1 {
		t.Fatalf("events = %v, want one ActionCanceled", log.events)
	}
	if got := log.events[0].(api.ActionCanceled); got.Reason.ID != "text.test.reason" || got.Source != uid {
		t.Errorf("canceled = %+v", got)
	}
	if err := timeline.Occupy(fx.w, fx.bodyChannel(player), 8); err != nil {
		t.Fatal(err)
	}
	if _, err := timeline.Run(fx.w, fx.clk); err != nil || fx.w.Now != 8 || len(log.log) != 0 {
		t.Errorf("now = %d, landings = %v, err = %v; want 8 and none", fx.w.Now, log.log, err)
	}
}

func TestDespawnedActorCancelsWithGone(t *testing.T) {
	fx, player, log := pendingFixture(t)
	other := fx.spawn(t, "idle", 2)
	otherUID := fx.w.UIDOf(other)
	schedule(t, fx, otherUID, fx.w.UIDOf(player), 5)
	fx.w.Despawn(other)
	if err := timeline.Occupy(fx.w, fx.bodyChannel(player), 8); err != nil {
		t.Fatal(err)
	}
	if _, err := timeline.Run(fx.w, fx.clk); err != nil || fx.w.Now != 8 {
		t.Fatalf("Run: now = %d, err = %v", fx.w.Now, err)
	}
	if len(log.log) != 0 {
		t.Errorf("landings = %v, want none", log.log)
	}
	if len(log.events) != 1 {
		t.Fatalf("events = %v, want one ActionCanceled", log.events)
	}
	if got, ok := log.events[0].(api.ActionCanceled); !ok || got.Reason.ID != "text.timeline.canceled.gone" ||
		got.Tick != 5 {
		t.Errorf("event = %+v, want canceled as gone at tick 5", log.events[0])
	}
}

func TestViewShowsLandsUntilTheyLand(t *testing.T) {
	fx, player, _ := pendingFixture(t)
	uid := fx.w.UIDOf(player)
	schedule(t, fx, uid, 0, 6)
	if err := timeline.Occupy(fx.w, fx.bodyChannel(player), 6); err != nil {
		t.Fatal(err)
	}
	countLands := func() (n int, label string) {
		view, err := fx.w.Views.Build(fx.w)
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range view.Timeline.Entries {
			if entry.Kind == api.EntryLands {
				n, label = n+1, entry.Label
			}
		}
		return n, label
	}
	if n, label := countLands(); n != 1 || label != "a swing" {
		t.Errorf("lands entries = %d (%q), want one labelled by its resolver", n, label)
	}
	if _, err := timeline.Run(fx.w, fx.clk); err != nil {
		t.Fatal(err)
	}
	if n, _ := countLands(); n != 0 {
		t.Errorf("lands entries = %d after landing, want none", n)
	}
}

func TestSeqResumesAfterRestore(t *testing.T) {
	fx, player, _ := pendingFixture(t)
	uid := fx.w.UIDOf(player)
	schedule(t, fx, uid, 0, 9)
	schedule(t, fx, uid, 0, 7)
	restored, restoredPlayer, log := pendingFixture(t)
	for _, e := range timeline.PendingOf(fx.w, uid) {
		p := *ecs.NewMap[comp.Pending](fx.w.ECS).Get(e)
		ecs.NewMap[comp.Pending](restored.w.ECS).NewEntity(&p)
	}
	if got, want := pendingTicks(restored, restoredPlayer), []int64{7, 9}; !reflect.DeepEqual(got, want) {
		t.Fatalf("restored landings = %v, want %v", got, want)
	}
	fresh := schedule(t, restored, restored.w.UIDOf(restoredPlayer), 0, 11)
	if seq := ecs.NewMap[comp.Pending](restored.w.ECS).Get(fresh).Seq; seq != 3 {
		t.Errorf("new Seq = %d, want 3 (above the restored 1 and 2)", seq)
	}
	_ = log
}

func pendingTicks(fx *fixture, body ecs.Entity) []int64 {
	var ticks []int64
	pendings := ecs.NewMap[comp.Pending](fx.w.ECS)
	for _, e := range timeline.PendingOf(fx.w, fx.w.UIDOf(body)) {
		ticks = append(ticks, pendings.Get(e).Lands)
	}
	return ticks
}

func TestReleaseFreesABusyChannelAtOnce(t *testing.T) {
	fx, player, _ := pendingFixture(t)
	channel := fx.bodyChannel(player)
	if err := timeline.Occupy(fx.w, channel, 8); err != nil {
		t.Fatal(err)
	}
	if err := timeline.Release(fx.w, channel); err != nil {
		t.Fatal(err)
	}
	if !timeline.Free(fx.w, channel) {
		t.Error("the channel should be free after Release")
	}
	if body, ok := timeline.Current(fx.w); !ok || body != player {
		t.Error("the freed body should be deciding again at once")
	}
	if err := timeline.Release(fx.w, channel); err != nil {
		t.Errorf("releasing a free channel = %v, want nothing to do", err)
	}
}

func TestLockedTellsALockFromOwnRecovery(t *testing.T) {
	fx, player, _ := pendingFixture(t)
	channel := fx.bodyChannel(player)
	if timeline.Locked(fx.w, channel) {
		t.Error("a free channel is not locked")
	}
	if err := timeline.Occupy(fx.w, channel, 4); err != nil {
		t.Fatal(err)
	}
	if timeline.Locked(fx.w, channel) {
		t.Error("recovery from its own action is not a lock")
	}
	if err := timeline.Extend(fx.w, channel, 2); err != nil || timeline.Locked(fx.w, channel) {
		t.Errorf("Extend must not lock (err %v)", err)
	}
	if err := timeline.Lock(fx.w, channel, 3); err != nil {
		t.Fatal(err)
	}
	if !timeline.Locked(fx.w, channel) {
		t.Error("the channel should be locked after Lock")
	}
	if err := timeline.Extend(fx.w, channel, 1); err != nil || !timeline.Locked(fx.w, channel) {
		t.Errorf("an active lock survives Extend (err %v)", err)
	}
	if err := timeline.Release(fx.w, channel); err != nil || timeline.Locked(fx.w, channel) {
		t.Errorf("Release should end the lock (err %v)", err)
	}
	if err := timeline.Lock(fx.w, channel, 2); err != nil {
		t.Fatal(err)
	}
	fx.w.Now += 5
	if timeline.Locked(fx.w, channel) {
		t.Error("an ended lock is not a lock")
	}
	if err := timeline.Occupy(fx.w, channel, 2); err != nil || timeline.Locked(fx.w, channel) {
		t.Errorf("busy time after a lock ended must not read as locked (err %v)", err)
	}
}
