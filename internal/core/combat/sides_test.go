package combat_test

import (
	"strconv"
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
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

// field is a running game on one of the sample maps, with the sample party and its creature.
type field struct {
	t      *testing.T
	g      *game.Game
	w      *sim.World
	events []api.Event
}

func newField(t *testing.T, mapID api.ContentID, mutate func(*content.Bundle)) *field {
	t.Helper()
	return newFieldInstalling(t, mapID, mutate, nil)
}

// newFieldInstalling is newField with extra registrations made while the game installs.
func newFieldInstalling(
	t *testing.T, mapID api.ContentID, mutate func(*content.Bundle), install func(*sim.World) error,
) *field {
	t.Helper()
	bundle := contenttest.SampleBundle(t)
	bundle.Start.Map = mapID
	contenttest.StepOnlyParty(&bundle)
	withoutTownAndDeepBodies(&bundle)
	if mutate != nil {
		mutate(&bundle)
	}
	reg := contenttest.Registry(t, bundle)
	f := &field{t: t}
	f.g, f.w, _ = gametest.Start(
		t,
		game.Config{Seed: 1, Content: reg, Install: func(w *sim.World, _ *clock.Clock) error {
			w.Bus.SubscribeAll("test.events", func(_ *sim.World, e api.Event) { f.events = append(f.events, e) })
			if install != nil {
				if err := install(w); err != nil {
					return err
				}
			}
			return timeline.RegisterResolver(w, "test.swing", timeline.Resolver{
				Resolve: func(*sim.World, ecs.Entity) error { return nil },
			})
		}},
	)
	return f
}

func (f *field) view() api.View {
	f.t.Helper()
	v, err := f.g.View()
	if err != nil {
		f.t.Fatal(err)
	}
	return v
}

// option applies the current view's option, and returns the events of the apply.
func (f *field) option(id string) []api.Event {
	f.t.Helper()
	for _, o := range f.view().Options {
		if o.ID == id {
			if !o.Enabled {
				f.t.Fatalf("option %q is disabled: %s", id, o.Reason)
			}
			events, err := f.g.Apply(o.Command)
			if err != nil {
				f.t.Fatal(err)
			}
			return events
		}
	}
	f.t.Fatalf("no option %q", id)
	return nil
}

func (f *field) party() *comp.Party {
	f.t.Helper()
	uid := f.view().Party.UID
	e, _ := f.w.ByUID(uid)
	return ecs.NewMap[comp.Party](f.w.ECS).Get(e)
}

func (f *field) members() []ecs.Entity {
	f.t.Helper()
	e, _ := f.w.ByUID(f.view().Party.UID)
	return anatomy.Members(f.w, e)
}

// creature is the sample creature.
func (f *field) creature() ecs.Entity {
	f.t.Helper()
	mapID := f.view().Party.Map
	positions := ecs.NewMap[comp.Pos](f.w.ECS)
	query := ecs.NewFilter1[comp.Identity](f.w.ECS).Query()
	for query.Next() {
		if query.Get().Template == "template.creature.moving_clay" && positions.Get(query.Entity()).Map == mapID {
			e := query.Entity()
			query.Close()
			return e
		}
	}
	f.t.Fatal("no creature")
	return ecs.Entity{}
}

// moveParty places every member in a cell.
func (f *field) moveParty(at api.Cell) {
	f.t.Helper()
	positions := ecs.NewMap[comp.Pos](f.w.ECS)
	for _, m := range f.members() {
		positions.Get(m).At = at
	}
}

func TestWalkingIntoSightOfTheHallsCreatureRaisesTheAlert(t *testing.T) {
	f := newField(t, "map.test.hall", nil)
	before := f.view().Party
	if before.Alert || before.Control != api.ControlFollow {
		t.Fatalf("party at the start = alert %v, control %q; want calm and following", before.Alert, before.Control)
	}
	var started *api.AlertStarted
	for range 6 {
		for _, e := range f.option("step/e") {
			if s, ok := e.(api.AlertStarted); ok {
				started = &s
			}
		}
		if started != nil {
			break
		}
	}
	if started == nil {
		t.Fatal("six steps east never raised the alert")
	}
	creatureUID := f.w.UIDOf(f.creature())
	if started.Hostile != creatureUID {
		t.Errorf("alert names %d, want the creature %d", started.Hostile, creatureUID)
	}
	after := f.view().Party
	if !after.Alert || after.Control != api.ControlIndividual {
		t.Errorf("party after the alert = alert %v, control %q; want alert and individual", after.Alert, after.Control)
	}
	if !has(f.events, always[api.ControlChanged]) {
		t.Error("no ControlChanged was published for the switch")
	}
}

func TestThereIsNoSightThroughRock(t *testing.T) {
	f := newField(t, "map.test.descent", nil)
	creature := f.creature()
	at := *ecs.NewMap[comp.Pos](f.w.ECS).Get(creature)
	if at.At.Depth != 2 {
		t.Fatalf("creature at depth %d, want 2: the soop under the hall", at.At.Depth)
	}
	above := api.Cell{Hex: at.At.Hex, Depth: 0}
	f.moveParty(above)
	for range 5 {
		f.option("wait/event")
		if f.view().Party.Alert {
			t.Fatal("the party raised an alert standing over the creature, with rock between")
		}
	}
	if has(f.events, always[api.AlertStarted]) {
		t.Error("AlertStarted was published")
	}
	f.moveParty(api.Cell{Hex: api.Hex{Q: at.At.Hex.Q + 2, R: at.At.Hex.R}, Depth: 2})
	f.option("wait/event")
	if !f.view().Party.Alert {
		t.Error("no alert with the creature two cells away on its own level")
	}
}

func TestTheAlertEndsAfterTheCalmSpell(t *testing.T) {
	for _, tt := range []struct {
		name       string
		individual bool
		wantMode   string
	}{
		{"lead and follow resume", false, api.ControlFollow},
		{"a chosen mode stays", true, api.ControlIndividual},
	} {
		t.Run(tt.name, func(t *testing.T) {
			f := newField(t, "map.test.hall", nil)
			if tt.individual {
				f.option("control/individual")
			}
			near := *ecs.NewMap[comp.Pos](f.w.ECS).Get(f.creature())
			f.moveParty(api.Cell{Hex: api.Hex{Q: near.At.Hex.Q - 3, R: near.At.Hex.R}})
			f.option("wait/event")
			f.option("wait/event")
			if !f.view().Party.Alert {
				t.Fatal("no alert beside the creature")
			}
			far := api.Cell{Hex: api.Hex{Q: near.At.Hex.Q - 12, R: near.At.Hex.R}}
			f.moveParty(far)
			f.events = nil
			calm := f.w.Content.Combat().CalmTicks
			partyEntity, _ := f.w.ByUID(f.view().Party.UID)
			start := ecs.NewMap[comp.Calm](f.w.ECS).Get(partyEntity).Since
			for range 400 {
				if has(f.events, always[api.AlertEnded]) {
					break
				}
				f.option("wait/event")
			}
			if !has(f.events, always[api.AlertEnded]) {
				t.Fatal("the alert never ended")
			}
			if waited := f.w.Now - start; waited < calm {
				t.Errorf("the alert ended after %d ticks, want at least the calm spell of %d", waited, calm)
			}
			if got := f.view().Party; got.Alert || got.Control != tt.wantMode {
				t.Errorf("party after = alert %v, control %q; want calm and %q", got.Alert, got.Control, tt.wantMode)
			}
		})
	}
}

// damage wears the named slot's parts of the body down by an amount, as an effect would.
func (f *field) damage(source, target ecs.Entity, slot string) {
	f.t.Helper()
	ctx := &effects.Context{World: f.w, Source: source, Target: target}
	refs := []content.EffectRef{{Do: "damage", Params: map[string]string{"parts": "target:" + slot, "amount": "99"}}}
	if err := effects.Apply(ctx, refs); err != nil {
		f.t.Fatal(err)
	}
	if _, err := f.w.Bus.Drain(f.w); err != nil {
		f.t.Fatal(err)
	}
}

func TestLosingEveryVitalPartKillsABodyAndLeavesACorpse(t *testing.T) {
	f := newField(t, "map.test.hall", nil)
	creature, hero := f.creature(), f.members()[0]
	f.events = nil
	f.damage(hero, creature, "head")
	if has(f.events, always[api.BodyDied]) {
		t.Fatal("the creature died with its core intact")
	}
	f.damage(hero, creature, "torso")
	var died *api.BodyDied
	for _, e := range f.events {
		if d, ok := e.(api.BodyDied); ok {
			died = &d
		}
	}
	if died == nil || died.Body != f.w.UIDOf(creature) {
		t.Fatalf("BodyDied = %+v, want the creature", died)
	}
	if !ecs.NewMap[comp.Dead](f.w.ECS).Has(creature) {
		t.Error("the creature has no Dead component")
	}
	if ecs.NewMap[comp.Actor](f.w.ECS).Has(creature) {
		t.Error("the corpse is still on the timeline")
	}
	cfg := f.w.Content.Combat()
	look := ecs.NewMap[comp.Appearance](f.w.ECS).Get(creature)
	if look.Glyph != cfg.CorpseGlyph || look.Color != cfg.CorpseColor {
		t.Errorf("corpse looks like %+v, want %q in %q", *look, cfg.CorpseGlyph, cfg.CorpseColor)
	}
	if !ecs.NewMap[comp.Pos](f.w.ECS).Has(creature) {
		t.Error("the corpse left the map")
	}
	for range 40 {
		f.option("wait/event")
	}
	if ecs.NewMap[comp.Actor](f.w.ECS).Has(creature) || f.g.Current() == f.w.UIDOf(creature) {
		t.Error("the corpse decided again")
	}
	if f.view().Party.Alert {
		t.Error("a corpse keeps the alert up")
	}
}

func TestDeathCancelsPendingActions(t *testing.T) {
	f := newField(t, "map.test.hall", nil)
	creature, hero := f.creature(), f.members()[0]
	uid := f.w.UIDOf(creature)
	if _, err := timeline.Schedule(
		f.w,
		comp.Pending{Kind: "test.swing", Actor: uid, Channel: uid, Lands: f.w.Now + 500},
	); err != nil {
		t.Fatal(err)
	}
	if len(pendingOf(f.w, uid)) != 1 {
		t.Fatal("the swing was not scheduled")
	}
	f.events = nil
	f.damage(hero, creature, "torso")
	if left := len(pendingOf(f.w, uid)); left != 0 {
		t.Errorf("a dead body has %d pending actions", left)
	}
}

func TestAProtectedMemberIsDownedAndALostPartyEndsTheGame(t *testing.T) {
	f := newField(t, "map.test.hall", nil)
	members := f.members()
	f.events = nil
	f.damage(members[1], members[1], "torso")
	if !has(f.events, always[api.BodyDowned]) || has(f.events, always[api.BodyDied]) {
		t.Fatalf("events = %v, want the protected member downed, not dead", f.events)
	}
	if has(f.events, always[api.PartyWiped]) {
		t.Fatal("the party was lost with a member standing")
	}
	if !ecs.NewMap[comp.Downed](f.w.ECS).Has(members[1]) || ecs.NewMap[comp.Actor](f.w.ECS).Has(members[1]) {
		t.Error("the downed member is not downed and off the timeline")
	}
	f.damage(members[0], members[0], "torso")
	if !has(f.events, always[api.PartyWiped]) {
		t.Fatal("no PartyWiped after every member went down")
	}
	for _, e := range f.events {
		if lost, ok := e.(api.PartyWiped); ok && lost.Party != f.view().Party.UID {
			t.Errorf("PartyWiped names %d, want the party %d", lost.Party, f.view().Party.UID)
		}
	}
	events, err := f.g.Apply(api.Wait{Source: 1, Until: api.WaitEvent})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("events after the loss = %v, want one rejection", events)
	}
	if rejected, ok := events[0].(api.Rejected); !ok || rejected.Reason.ID != "text.game.lost" {
		t.Errorf("event = %+v, want a rejection with text.game.lost", events[0])
	}
}

func TestAnOutOfFightMemberCannotLead(t *testing.T) {
	f := newField(t, "map.test.hall", nil)
	members := f.members()
	f.damage(members[0], members[1], "torso")
	uid := f.w.UIDOf(members[1])
	for _, o := range f.view().Options {
		if o.ID == "lead/"+strconv.FormatUint(uint64(uid), 10) {
			t.Fatalf("the downed member is offered the lead: %+v", o)
		}
	}
	events, err := f.g.Apply(api.Lead{Party: f.view().Party.UID, Member: uid})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 1 {
		t.Fatalf("events = %v, want one rejection", events)
	}
	if rejected, ok := events[0].(api.Rejected); !ok || rejected.Reason.ID != "text.explore.out_of_fight" {
		t.Errorf("event = %+v, want a rejection with text.explore.out_of_fight", events[0])
	}
}

func TestTheLeadPassesOnWhenTheLeaderGoesDown(t *testing.T) {
	f := newField(t, "map.test.hall", nil)
	members := f.members()
	f.damage(members[1], members[0], "torso")
	state := f.party()
	if state.Leader != f.w.UIDOf(members[1]) || state.Order[0] != state.Leader {
		t.Errorf("leader = %d, order = %v; want the standing member to lead", state.Leader, state.Order)
	}
}

// pendingOf lists the pending actions of the body.
func pendingOf(w *sim.World, uid api.UID) []ecs.Entity {
	var out []ecs.Entity
	query := ecs.NewFilter1[comp.Pending](w.ECS).Query()
	for query.Next() {
		if query.Get().Actor == uid {
			out = append(out, query.Entity())
		}
	}
	return out
}
