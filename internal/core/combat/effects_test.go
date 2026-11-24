package combat_test

import (
	"slices"
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/clock"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/effects"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game/gametest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/inventory"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/knowledge"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

// here is where the test bodies stand.
var here = comp.Pos{Map: 1, At: api.Cell{Hex: api.Hex{Q: 2, R: 1}, Depth: 3}}

// rig is a game with the sample party, a crawler standing at here, and a piece in the inventory.
type rig struct {
	g       *game.Game
	w       *sim.World
	party   ecs.Entity
	reaper  ecs.Entity
	crawler ecs.Entity
	events  []api.Event
}

func newRig(t *testing.T) *rig { return newRigWith(t, nil) }

// newRigWith is newRig with the bundle adjusted first.
func newRigWith(t *testing.T, mutate func(*content.Bundle)) *rig {
	t.Helper()
	bundle := contenttest.SampleBundle(t)
	contenttest.NoStart(&bundle)
	armPart := api.ContentID("part.human_arm")
	bundle.Items["item.test_arm"] = content.Item{Name: "text.item.test_arm", Part: &armPart}
	bundle.Statuses["status.braced"] = content.Status{Name: "text.status.braced"}
	for id, integrity := range map[string]int{
		"part.human_head": 8, "part.human_arm": 6, "part.human_legs": 10, "part.human_torso": 12,
	} {
		part := bundle.Parts[id]
		part.Integrity = integrity
		bundle.Parts[id] = part
	}
	bundle.Text["text.item.test_arm"] = "spare arm"
	bundle.Text["text.status.braced"] = "braced"
	bundle.Templates["template.test.crawler"] = content.Template{Body: ptr(api.ContentID("body.test.crawler"))}
	if mutate != nil {
		mutate(&bundle)
	}
	reg := contenttest.Registry(t, bundle)
	r := &rig{}
	r.g, r.w, _ = gametest.Start(t, game.Config{
		Seed: 1, Content: reg,
		Install: func(w *sim.World, _ *clock.Clock) error {
			w.Bus.SubscribeAll("test.events", func(_ *sim.World, e api.Event) { r.events = append(r.events, e) })
			return timeline.RegisterResolver(w, "test.swing", timeline.Resolver{
				Resolve: func(*sim.World, ecs.Entity) error { return nil },
			})
		},
		Spawn: func(w *sim.World, _ *clock.Clock) error {
			var err error
			if r.party, err = w.Spawn("template.party.main", nil); err != nil {
				return err
			}
			r.reaper = anatomy.Members(w, r.party)[0]
			r.crawler, err = w.Spawn("template.test.crawler", &here)
			return err
		},
	})
	pos := here
	ecs.NewMap[comp.Pos](r.w.ECS).Add(r.reaper, &pos)
	return r
}

func ptr[T any](v T) *T { return &v }

// apply runs the effects and returns the events they published.
func (r *rig) apply(t *testing.T, ctx *effects.Context, refs ...content.EffectRef) []api.Event {
	t.Helper()
	r.events = nil
	if err := effects.Apply(ctx, refs); err != nil {
		t.Fatal(err)
	}
	if _, err := r.w.Bus.Drain(r.w); err != nil {
		t.Fatal(err)
	}
	return r.events
}

// ctx is the reaper acting on the target body.
func (r *rig) ctx(target ecs.Entity) *effects.Context {
	return &effects.Context{World: r.w, Source: r.reaper, Target: target}
}

// partsOf lists the body's attached parts of a slot.
func (r *rig) partsOf(body ecs.Entity, slot string) []ecs.Entity {
	var out []ecs.Entity
	parts := ecs.NewMap[comp.Part](r.w.ECS)
	for _, part := range anatomy.Parts(r.w, body) {
		if parts.Get(part).Slot == slot {
			out = append(out, part)
		}
	}
	return out
}

func (r *rig) integrity(part ecs.Entity) int {
	return ecs.NewMap[comp.Part](r.w.ECS).Get(part).Integrity
}

// kinds lists the event kinds in order, keeping only those that start with prefix.
func kinds(events []api.Event, prefix string) []string {
	var out []string
	for _, e := range events {
		if k := e.EventKind(); len(k) >= len(prefix) && k[:len(prefix)] == prefix {
			out = append(out, k)
		}
	}
	return out
}

func TestDamage(t *testing.T) {
	tests := []struct {
		name       string
		amount     string
		wantEvents []string
		wantLeft   int
	}{
		{"wears an arm down", "4", []string{api.PartDamagedKind, api.PartDamagedKind}, 2},
		{"past zero severs, damage first", "9", []string{
			api.PartDamagedKind, api.PartSeveredKind, api.PartDamagedKind, api.PartSeveredKind,
		}, 0},
		{"zero does nothing", "0", nil, 6},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newRig(t)
			target := anatomy.Members(r.w, r.party)[1]
			events := r.apply(t, r.ctx(target), contenttest.Do("damage", "parts", "target:arm", "amount", tt.amount))
			if got := kinds(events, "combat."); !slices.Equal(got, tt.wantEvents) {
				t.Errorf("events = %v, want %v", got, tt.wantEvents)
			}
			arms := r.partsOf(target, "arm")
			severed := tt.wantLeft == 0
			if severed && len(arms) != 0 {
				t.Errorf("arms attached = %d, want none", len(arms))
			}
			if !severed {
				if len(arms) != 2 {
					t.Fatalf("arms attached = %d, want 2", len(arms))
				}
				for _, arm := range arms {
					if got := r.integrity(arm); got != tt.wantLeft {
						t.Errorf("arm integrity = %d, want %d", got, tt.wantLeft)
					}
				}
			}
			for _, e := range events {
				if s, ok := e.(api.PartSevered); ok && s.Cause != api.SeveredWorn {
					t.Errorf("cause = %q, want worn", s.Cause)
				}
			}
		})
	}
}

func TestDamageSeversWhereTheBodyStands(t *testing.T) {
	r := newRig(t)
	maw := r.partsOf(r.crawler, "head")[0]
	ctx := r.ctx(r.crawler)
	ctx.TargetPart = maw
	r.apply(t, ctx, contenttest.Do("damage", "parts", "target_part", "amount", "5"))
	if got := r.partsOf(r.crawler, "head"); len(got) != 0 {
		t.Fatal("the maw is still attached")
	}
	if pos := ecs.NewMap[comp.Pos](r.w.ECS); !pos.Has(maw) || *pos.Get(maw) != here {
		t.Error("the worn-off maw is not where the crawler stands")
	}
}

func TestSever(t *testing.T) {
	tests := []struct {
		name      string
		when      string
		wantArms  int
		wantCause bool
	}{
		{"holds", "true", 1, true},
		{"false is a no-op", "false", 2, false},
		{"integrity decides", "target.integrity > 99", 2, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := newRig(t)
			target := anatomy.Members(r.w, r.party)[1]
			arm := r.partsOf(target, "arm")[0]
			ecs.NewMap[comp.Pos](r.w.ECS).Add(target, &here)
			ctx := r.ctx(target)
			ctx.TargetPart = arm
			events := r.apply(t, ctx, contenttest.Do("sever", "parts", "target_part", "when", tt.when))
			if got := len(r.partsOf(target, "arm")); got != tt.wantArms {
				t.Errorf("arms = %d, want %d", got, tt.wantArms)
			}
			gotEvent := slices.Contains(kinds(events, "combat."), api.PartSeveredKind)
			if gotEvent != tt.wantCause {
				t.Errorf("severed event = %v, want %v", gotEvent, tt.wantCause)
			}
			if !tt.wantCause {
				return
			}
			if s := events[0].(api.PartSevered); s.Cause != api.SeveredCut || s.Part != r.w.UIDOf(arm) {
				t.Errorf("event = %+v", s)
			}
			if pos := ecs.NewMap[comp.Pos](r.w.ECS); !pos.Has(arm) || *pos.Get(arm) != here {
				t.Error("the severed arm is not where the body stands")
			}
			if got := r.integrity(arm); got != 6 {
				t.Errorf("a cut arm's integrity = %d, want it kept at 6", got)
			}
		})
	}
}

func TestSeveringAnArmMidWindupCancelsTheSwing(t *testing.T) {
	r := newRig(t)
	arms := r.partsOf(r.reaper, "arm")
	if _, err := timeline.Schedule(r.w, comp.Pending{
		Kind: "test.swing", Actor: r.w.UIDOf(r.reaper), Channel: r.w.UIDOf(arms[0]), Lands: r.w.Now + 5,
	}); err != nil {
		t.Fatal(err)
	}
	other, err := timeline.Schedule(r.w, comp.Pending{
		Kind: "test.swing", Actor: r.w.UIDOf(r.reaper), Channel: r.w.UIDOf(arms[1]), Lands: r.w.Now + 6,
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx := r.ctx(r.reaper)
	ctx.TargetPart = arms[0]
	events := r.apply(t, ctx, contenttest.Do("sever", "parts", "target_part"))
	var canceled []api.ActionCanceled
	for _, e := range events {
		if c, ok := e.(api.ActionCanceled); ok {
			canceled = append(canceled, c)
		}
	}
	if len(canceled) != 1 || canceled[0].Kind != "test.swing" || canceled[0].Source != r.w.UIDOf(r.reaper) {
		t.Fatalf("canceled = %+v, want the one swing by the reaper", canceled)
	}
	if !r.w.ECS.Alive(other) || len(timeline.PendingOf(r.w, r.w.UIDOf(r.reaper))) != 1 {
		t.Error("the other arm's swing should still be pending")
	}
}

func TestAddStatusExpiresByTick(t *testing.T) {
	r := newRig(t)
	statuses := ecs.NewMap[comp.Statuses](r.w.ECS)
	events := r.apply(
		t,
		r.ctx(r.crawler),
		contenttest.Do("add_status", "bodies", "target", "status", "status.braced", "ticks", "10"),
	)
	if len(events) != 1 || events[0].(api.StatusAdded).Until != r.w.Now+10 {
		t.Fatalf("events = %+v", events)
	}
	r.apply(
		t,
		r.ctx(r.crawler),
		contenttest.Do("add_status", "bodies", "target", "status", "status.braced", "ticks", "3"),
	)
	if got := statuses.Get(r.crawler).Active; len(got) != 1 || got[0].Until != r.w.Now+10 {
		t.Fatalf("statuses = %+v, want one ending at %d", got, r.w.Now+10)
	}
	run := func() {
		t.Helper()
		if err := r.w.Systems.Run(r.w, sim.PerEntry); err != nil {
			t.Fatal(err)
		}
	}
	r.w.Now += 9
	run()
	if !statuses.Has(r.crawler) {
		t.Fatal("the status ended early")
	}
	r.w.Now++
	run()
	if statuses.Has(r.crawler) {
		t.Error("the status outlived its tick")
	}
}

func TestStatusesAndPartFactsReachFormulas(t *testing.T) {
	r := newRig(t)
	r.apply(
		t,
		r.ctx(r.crawler),
		contenttest.Do("add_status", "bodies", "target", "status", "status.braced", "ticks", "10"),
	)
	core := r.partsOf(r.crawler, "torso")[0]
	ctx := r.ctx(r.crawler)
	ctx.TargetPart = core
	ctx.Choices = map[string]string{"type": "type.soak"}
	ecs.NewMap[comp.Pos](r.w.ECS).Get(r.reaper).At.Depth = 1
	formula := `("status.braced" in target.statuses ? 1 : 0) + (rule.weak ? 2 : 0) + (rule.immune ? 100 : 0) + ` +
		`(target.locked ? 100 : 0) + (target.graft ? 100 : 0) + height * 2 + (target.integrity == target.max_integrity ? 0 : 100)`
	r.apply(t, ctx, contenttest.Do("damage", "parts", "target_part", "amount", formula))
	if got := r.integrity(core); got != 8-7 {
		t.Errorf("integrity = %d, want 1", got)
	}
}

func TestAttachPartFromTheInventory(t *testing.T) {
	r := newRig(t)
	target := anatomy.Members(r.w, r.party)[1]
	attach := contenttest.Do("attach_part", "from", "inventory", "item", "item.test_arm")
	before := len(timeline.Channels(r.w, target))
	if ok, reason := effects.Check(
		r.ctx(target),
		[]content.EffectRef{attach},
	); ok ||
		reason.ID != "text.combat.inventory_short" {
		t.Fatalf("check on an empty inventory = %v, %+v", ok, reason)
	}
	if events := r.apply(t, r.ctx(target), attach); len(events) != 0 {
		t.Fatalf("an impossible attach published %v", events)
	}
	inventory.Add(r.w, r.party, comp.ItemStack{Item: "item.test_arm", Count: 1})
	events := r.apply(t, r.ctx(target), attach)
	if len(events) != 1 || events[0].(api.PartAttached).From != api.AttachedFromInventory {
		t.Fatalf("events = %+v", events)
	}
	if got := len(timeline.Channels(r.w, target)); got != before+1 {
		t.Errorf("channels = %d, want %d", got, before+1)
	}
	if got := inventory.Count(r.w, r.party, "item.test_arm", 0); got != 0 {
		t.Errorf("pieces left = %d, want 0", got)
	}
	part, ok := r.w.ByUID(events[0].(api.PartAttached).Part)
	if !ok || !ecs.NewMap[comp.Graft](r.w.ECS).Has(part) {
		t.Error("the attached piece is not a graft")
	}
}

func TestAttachPartFromSevered(t *testing.T) {
	r := newRig(t)
	arm := r.partsOf(r.reaper, "arm")[0]
	ctx := r.ctx(r.reaper)
	ctx.TargetPart = arm
	r.apply(t, ctx, contenttest.Do("sever", "parts", "target_part"))
	attach := contenttest.Do("attach_part", "from", "severed")
	if ok, reason := effects.Check(r.ctx(r.crawler), []content.EffectRef{attach}); !ok {
		t.Fatalf("check = %v, %+v", ok, reason)
	}
	events := r.apply(t, r.ctx(r.crawler), attach)
	if len(events) != 1 || events[0].(api.PartAttached).Part != r.w.UIDOf(arm) {
		t.Fatalf("events = %+v", events)
	}
	if got := r.partsOf(r.crawler, "arm"); len(got) != 1 || got[0] != arm {
		t.Errorf("the crawler's arms = %v, want the severed one", got)
	}
	if ecs.NewMap[comp.Pos](r.w.ECS).Has(arm) {
		t.Error("an attached part kept a position")
	}
	if ecs.NewMap[comp.Graft](r.w.ECS).Has(arm) || !ecs.NewMap[comp.Reattached](r.w.ECS).Has(arm) {
		t.Error("a reattached limb should be marked Reattached and not be a graft")
	}
	if ok, reason := effects.Check(r.ctx(r.crawler), []content.EffectRef{attach}); ok || reason.ID == "" {
		t.Errorf("nothing is left to attach, but check = %v, %+v", ok, reason)
	}
}

func TestRevealWritesAFactPerPart(t *testing.T) {
	r := newRig(t)
	r.apply(t, r.ctx(r.crawler), contenttest.Do(
		"reveal", "claim", "claim.material", "value", "= target.material.id", "origin", "origin.sweep",
	))
	for _, part := range anatomy.Parts(r.w, r.crawler) {
		fact, ok := knowledge.Best(r.w, r.party, r.w.UIDOf(part), "claim.material")
		want := ecs.NewMap[comp.Part](r.w.ECS).Get(part).Material
		if !ok || fact.Value != string(want) || fact.Origin != "origin.sweep" {
			t.Errorf("fact = %+v, %v; want %s from a sweep", fact, ok, want)
		}
	}
	if got := len(ecs.NewMap[comp.Knowledge](r.w.ECS).Get(r.party).Facts); got != 3 {
		t.Errorf("facts = %d, want one per part (3)", got)
	}
}

func TestConsume(t *testing.T) {
	r := newRig(t)
	consume := []content.EffectRef{contenttest.Do("consume", "item", "item.coal")}
	if ok, _ := effects.Check(r.ctx(r.crawler), consume); !ok {
		t.Fatal("the inventory has coal, but check refused")
	}
	r.apply(t, r.ctx(r.crawler), consume...)
	if got := inventory.Count(r.w, r.party, "item.coal", 0); got != 2 {
		t.Errorf("coal = %d, want 2", got)
	}
	r.apply(t, r.ctx(r.crawler), consume...)
	r.apply(t, r.ctx(r.crawler), consume...)
	ok, reason := effects.Check(r.ctx(r.crawler), consume)
	if ok || reason.ID != "text.combat.inventory_short" || reason.Args["item"] != "coal" {
		t.Errorf("check on a short inventory = %v, %+v", ok, reason)
	}
	r.apply(t, r.ctx(r.crawler), consume...)
	if got := inventory.Count(r.w, r.party, "item.coal", 0); got != 0 {
		t.Errorf("coal = %d, a short inventory must not go negative", got)
	}
}

func TestSpawnFromPart(t *testing.T) {
	r := newRig(t)
	maw := r.partsOf(r.crawler, "head")[0]
	ctx := r.ctx(r.crawler)
	ctx.TargetPart = maw
	spawn := contenttest.Do("spawn_from_part", "faction", "faction.wild", "controller", "idle")
	r.apply(t, ctx, spawn)
	if ecs.NewMap[comp.Actor](r.w.ECS).Has(maw) {
		t.Fatal("an attached part became a body")
	}
	r.apply(t, ctx, contenttest.Do("sever", "parts", "target_part"), spawn)
	partOf := ecs.NewMap[comp.PartOf](r.w.ECS)
	if !partOf.Has(maw) {
		t.Fatal("the severed maw was not made a body")
	}
	body := partOf.GetRelation(maw)
	if got := anatomy.Parts(r.w, body); len(got) != 1 || got[0] != maw {
		t.Errorf("the new body's parts = %v, want just the maw", got)
	}
	if faction := ecs.NewMap[comp.Allegiance](r.w.ECS).Get(body).Faction; faction != "faction.wild" {
		t.Errorf("faction = %q", faction)
	}
	if controller := ecs.NewMap[comp.Actor](r.w.ECS).Get(body).Controller; controller != "idle" {
		t.Errorf("controller = %q", controller)
	}
	if pos := ecs.NewMap[comp.Pos](r.w.ECS); !pos.Has(body) || *pos.Get(body) != here || pos.Has(maw) {
		t.Error("the new body should stand where the part lay, and the part should lose its position")
	}
	if got := len(timeline.Channels(r.w, body)); got != 2 {
		t.Errorf("channels = %d, want the body and the maw", got)
	}
}

func TestRaiseMaxIntegrityKeepsIntegrity(t *testing.T) {
	r := newRig(t)
	target := anatomy.Members(r.w, r.party)[1]
	arms := r.partsOf(target, "arm")
	parts := ecs.NewMap[comp.Part](r.w.ECS)
	before := *parts.Get(arms[0])
	r.apply(t, r.ctx(target), contenttest.Do("raise_max_integrity", "parts", "target:arm", "amount", "3"))
	for _, arm := range arms {
		if got := parts.Get(arm).MaxIntegrity; got != before.MaxIntegrity+3 {
			t.Errorf("max integrity = %d, want %d", got, before.MaxIntegrity+3)
		}
		if got := parts.Get(arm).Integrity; got != before.Integrity {
			t.Errorf("integrity = %d, want %d (unchanged)", got, before.Integrity)
		}
	}
}

func TestCombatEventsNameTheirSource(t *testing.T) {
	r := newRig(t)
	target := anatomy.Members(r.w, r.party)[1]
	reaper := r.w.UIDOf(r.reaper)
	events := r.apply(t, r.ctx(target),
		contenttest.Do("damage", "parts", "target:arm", "amount", "1"),
		contenttest.Do("add_status", "bodies", "target", "status", "status.braced", "ticks", "3"),
		contenttest.Do("sever", "parts", "target:arm"),
	)
	seen := 0
	for _, e := range events {
		subjected, ok := e.(api.Subjected)
		if !ok {
			continue
		}
		seen++
		if got := subjected.Subjects()[api.SubjectSource]; got != reaper {
			t.Errorf("%s source = %d, want the reaper %d", e.EventKind(), got, reaper)
		}
	}
	if seen == 0 {
		t.Fatal("no combat events carried subjects")
	}
}
