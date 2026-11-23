package explore_test

import (
	"slices"
	"strconv"
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/rng"
)

// party is the party's component, to read and to force.
func (fx *fixture) party() *comp.Party {
	fx.t.Helper()
	e, ok := fx.w.ByUID(fx.view().Party.UID)
	if !ok {
		fx.t.Fatal("no party entity")
	}
	return ecs.NewMap[comp.Party](fx.w.ECS).Get(e)
}

// controller is the controller of the member with the UID.
func (fx *fixture) controller(uid api.UID) string {
	fx.t.Helper()
	e, ok := fx.w.ByUID(uid)
	if !ok {
		fx.t.Fatalf("no entity %d", uid)
	}
	return ecs.NewMap[comp.Actor](fx.w.ECS).Get(e).Controller
}

// members are the UIDs of the party members, in party order.
func (fx *fixture) members() []api.UID {
	fx.t.Helper()
	var uids []api.UID
	for _, m := range fx.view().Party.Members {
		uids = append(uids, m.UID)
	}
	return uids
}

func (fx *fixture) hasOption(id string) bool {
	return slices.ContainsFunc(fx.view().Options, func(o api.Option) bool { return o.ID == id })
}

func moves(events []api.Event, actor api.UID) []api.Moved {
	var out []api.Moved
	for _, e := range events {
		if m, ok := e.(api.Moved); ok && m.Source == actor {
			out = append(out, m)
		}
	}
	return out
}

func TestTheFollowerStepsIntoTheCellTheLeaderLeftInTheSameApply(t *testing.T) {
	fx := newFixture(t, "", nil)
	members := fx.members()
	leader, follower := members[0], members[1]
	if got := fx.controller(leader); got != "player" {
		t.Fatalf("leader controller = %q, want player", got)
	}
	if got := fx.controller(follower); got != "follow" {
		t.Fatalf("follower controller = %q, want follow", got)
	}
	for range 4 {
		left := fx.pos(leader)
		events := fx.step("e")
		if got := moves(events, leader); len(got) != 1 {
			t.Fatalf("leader moves = %+v, want one: a keypress is one step", got)
		}
		if got := moves(events, follower); len(got) != 1 || got[0].To != left.At {
			t.Fatalf("follower moves = %+v, want one into %+v, the cell the leader left", got, left.At)
		}
		if got := fx.pos(follower); got.At != left.At {
			t.Fatalf("follower at %+v, want %+v", got.At, left.At)
		}
		if fx.g.Current() != leader {
			t.Fatalf("current = %d, want the leader again", fx.g.Current())
		}
	}
}

// standingApart fails the test when two members share a cell and returns the members' positions.
func (fx *fixture) standingApart() []comp.Pos {
	fx.t.Helper()
	var out []comp.Pos
	for _, uid := range fx.members() {
		pos := fx.pos(uid)
		if slices.Contains(out, pos) {
			fx.t.Fatalf("two members stand in %+v", pos)
		}
		out = append(out, pos)
	}
	return out
}

func TestDownTheDescentTheFollowerTrailsSingleFileNeverSharingACell(t *testing.T) {
	fx := newFixture(t, "map.test.descent", nil)
	members := fx.members()
	leader, follower := members[0], members[1]
	for range 10 { // out of the hall and down the slope to the chamber
		fx.step("e")
		fx.standingApart()
		if d := hexmap.Distance(fx.pos(leader).At.Hex, fx.pos(follower).At.Hex); d > 2 {
			t.Fatalf("the follower fell %d hexes behind", d)
		}
	}
	if fx.pos(leader).At.Depth != 4 {
		t.Fatalf("leader at %+v, want the chamber's depth", fx.pos(leader))
	}
	for range 6 { // the leader waits, and the follower closes up
		fx.apply(fx.option("wait/event").Command)
	}
	fx.standingApart()
	if d := hexmap.Distance(fx.pos(leader).At.Hex, fx.pos(follower).At.Hex); d != 1 {
		t.Errorf("follower %d hexes from the leader, want right behind it", d)
	}
	if fx.pos(follower).At.Depth < 3 {
		t.Errorf("follower at depth %d, want it down the slope", fx.pos(follower).At.Depth)
	}
}

func TestSteppingIntoTheFollowerSwapsThem(t *testing.T) {
	fx := newFixture(t, "", nil)
	members := fx.members()
	leader, follower := members[0], members[1]
	from, to := fx.pos(leader), fx.pos(follower)
	// The follower starts on the nearest cell with the lowest R, then Q: northwest of the leader.
	if to.At.Hex != (api.Hex{Q: from.At.Hex.Q, R: from.At.Hex.R - 1}) {
		t.Fatalf("follower starts at %v, leader at %v", to.At.Hex, from.At.Hex)
	}
	events := fx.step("nw")
	if got := fx.pos(leader); got.At != to.At {
		t.Errorf("leader at %+v, want the follower's old cell %+v", got.At, to.At)
	}
	if got := fx.pos(follower); got.At != from.At {
		t.Errorf("follower at %+v, want the leader's old cell %+v", got.At, from.At)
	}
	if got := moves(events, follower); len(got) != 1 || got[0].From != to.At || got[0].To != from.At {
		t.Errorf("follower moves = %+v, want one swap", got)
	}
	// The swap updated both trails: the follower now trails into the cell the leader left, which it
	// stands in, so the next step of the leader is followed as usual.
	fx.step("nw")
	fx.standingApart()
	if got := fx.pos(follower); got.At != to.At {
		t.Errorf("follower at %+v after the leader's next step, want %+v", got.At, to.At)
	}
}

func TestSteppingIntoAnotherBodyIsRefused(t *testing.T) {
	fx := newFixture(t, "", nil)
	leader := fx.members()[0]
	// A stranger on the cell east of the leader: not a member, so no swap.
	at := fx.pos(leader).At
	at.Hex.Q++
	if _, err := fx.w.Spawn("template.person.reaper", &comp.Pos{Map: fx.pos(leader).Map, At: at}); err != nil {
		t.Fatal(err)
	}
	o := fx.option("step/e")
	if o.Enabled || o.Reason != "Someone is in the way." {
		t.Errorf("step/e = %+v, want it refused as occupied", o)
	}
}

func TestSwitchingControlByHand(t *testing.T) {
	fx := newFixture(t, "", nil)
	members := fx.members()
	leader, follower := members[0], members[1]
	if !fx.hasOption("control/individual") || fx.hasOption("control/follow") {
		t.Fatalf("options = %+v, want control/individual only", fx.view().Options)
	}
	events := fx.apply(fx.option("control/individual").Command)
	if changed, ok := find[api.ControlChanged](events); !ok || changed.Mode != api.ControlIndividual ||
		changed.By != api.ControlByPlayer {
		t.Fatalf("control changed = %+v, %v", changed, ok)
	}
	if fx.controller(follower) != "player" || fx.controller(leader) != "player" {
		t.Errorf("controllers = %q, %q; want both player", fx.controller(leader), fx.controller(follower))
	}
	// Every member's decisions come to the player, in timeline order: the leader, then the follower.
	if fx.g.Current() != leader {
		t.Fatalf("current = %d, want the leader", fx.g.Current())
	}
	fx.step("e")
	if fx.g.Current() != follower {
		t.Fatalf("current = %d, want the follower to decide for itself", fx.g.Current())
	}
	if !fx.hasOption("control/follow") {
		t.Error("no control/follow option while a follower decides")
	}
	// Back to follow while the follower is deciding: it goes back to trailing, and the leader is next.
	events = fx.apply(fx.option("control/follow").Command)
	if changed, ok := find[api.ControlChanged](events); !ok || changed.Mode != api.ControlFollow {
		t.Fatalf("control changed = %+v, %v", changed, ok)
	}
	if fx.g.Current() != leader {
		t.Errorf("current = %d, want the leader after switching back", fx.g.Current())
	}
	if fx.controller(follower) != "follow" {
		t.Errorf("follower controller = %q, want follow", fx.controller(follower))
	}
	if got := fx.pos(follower).At; got == fx.pos(leader).At {
		t.Errorf("both stand in %+v", got)
	}
}

func TestControlRejectsAnotherPartyAndAnUnknownMode(t *testing.T) {
	fx := newFixture(t, "", nil)
	events := fx.apply(api.Control{Party: 9999, Mode: api.ControlFollow})
	if rejected, ok := find[api.Rejected](events); !ok || rejected.Reason.ID != "text.explore.bad_party" {
		t.Errorf("rejected = %+v, %v", rejected, ok)
	}
	events = fx.apply(api.Control{Party: fx.view().Party.UID, Mode: "sideways"})
	if rejected, ok := find[api.Rejected](events); !ok || rejected.Reason.ID != "text.explore.bad_mode" {
		t.Errorf("rejected = %+v, %v", rejected, ok)
	}
}

// alert forces the alert flag and runs one entry of the timeline so the system sees it.
func (fx *fixture) alert(on bool) []api.Event {
	fx.t.Helper()
	fx.party().Alert = on
	return fx.apply(fx.option("wait/event").Command)
}

func controlChanges(events []api.Event) []api.ControlChanged {
	var out []api.ControlChanged
	for _, e := range events {
		if c, ok := e.(api.ControlChanged); ok {
			out = append(out, c)
		}
	}
	return out
}

func TestAnAlertSwitchesToIndividualControlAndBack(t *testing.T) {
	fx := newFixture(t, "", nil)
	follower := fx.members()[1]
	changes := controlChanges(fx.alert(true))
	if len(changes) != 1 || changes[0].Mode != api.ControlIndividual || changes[0].By != api.ControlByAlert {
		t.Fatalf("changes = %+v, want a switch to individual by the alert", changes)
	}
	if fx.party().Control != api.ControlIndividual || fx.controller(follower) != "player" {
		t.Errorf("control = %q, follower controller %q", fx.party().Control, fx.controller(follower))
	}
	changes = controlChanges(fx.alert(false))
	if len(changes) != 1 || changes[0].Mode != api.ControlFollow {
		t.Fatalf("changes = %+v, want a switch back to follow", changes)
	}
	if fx.controller(follower) != "follow" {
		t.Errorf("follower controller = %q, want follow", fx.controller(follower))
	}
	if quiet := controlChanges(fx.alert(false)); len(quiet) != 0 {
		t.Errorf("changes = %+v with nothing changed, want none", quiet)
	}
}

func TestAModeThePlayerChoseSurvivesTheAlert(t *testing.T) {
	t.Run("individual before the alert", func(t *testing.T) {
		fx := newFixture(t, "", nil)
		fx.apply(fx.option("control/individual").Command)
		if changes := controlChanges(fx.alert(true)); len(changes) != 0 {
			t.Errorf("alert changes = %+v, want none: already individual", changes)
		}
		if changes := controlChanges(fx.alert(false)); len(changes) != 0 {
			t.Errorf("end of alert changes = %+v, want none", changes)
		}
		if got := fx.party().Control; got != api.ControlIndividual {
			t.Errorf("control = %q, want the player's individual", got)
		}
	})
	t.Run("follow during the alert", func(t *testing.T) {
		fx := newFixture(t, "", nil)
		fx.alert(true)
		fx.apply(fx.option("control/follow").Command)
		if changes := controlChanges(fx.alert(false)); len(changes) != 0 {
			t.Errorf("end of alert changes = %+v, want none", changes)
		}
		if got := fx.party().Control; got != api.ControlFollow {
			t.Fatalf("control = %q, want follow to stay", got)
		}
		changes := controlChanges(fx.alert(true))
		if len(changes) != 1 || changes[0].Mode != api.ControlIndividual {
			t.Errorf("next alert changes = %+v, want individual again", changes)
		}
	})
}

func TestThePartyViewShowsTheControlModeAndTheAlert(t *testing.T) {
	fx := newFixture(t, "", nil)
	if got := fx.view().Party; got.Control != api.ControlFollow || got.Alert {
		t.Fatalf("view = control %q, alert %v; want follow and calm", got.Control, got.Alert)
	}
	fx.apply(fx.option("control/individual").Command)
	if got := fx.view().Party; got.Control != api.ControlIndividual || got.Alert {
		t.Errorf("after explore.control, view = control %q, alert %v; want individual and calm", got.Control, got.Alert)
	}
	fx.apply(fx.option("control/follow").Command)
	fx.alert(true)
	if got := fx.view().Party; !got.Alert || got.Control != api.ControlIndividual {
		t.Errorf(
			"with the alert forced, view = control %q, alert %v; want individual and alert",
			got.Control,
			got.Alert,
		)
	}
}

func TestAnAlertChangesNothingWhenContentTurnsItOff(t *testing.T) {
	fx := newFixture(t, "", func(b *content.Bundle) { b.Party.IndividualOnAlert = false })
	if changes := controlChanges(fx.alert(true)); len(changes) != 0 {
		t.Errorf("changes = %+v, want none", changes)
	}
	if got := fx.party().Control; got != api.ControlFollow {
		t.Errorf("control = %q, want follow", got)
	}
}

func TestLeadChangesTheLeaderTheOrderAndTheControllersInEitherMode(t *testing.T) {
	for _, individual := range []bool{false, true} {
		name := "follow"
		if individual {
			name = "individual"
		}
		t.Run(name, func(t *testing.T) {
			fx := newFixture(t, "", nil)
			if individual {
				fx.apply(fx.option("control/individual").Command)
			}
			was := fx.members()
			leader, other := was[0], was[1]
			if fx.hasOption("lead/" + strconv.Itoa(int(leader))) {
				t.Error("an option to lead with the leader")
			}
			o := fx.option("lead/" + strconv.Itoa(int(other)))
			if o.Label != "Lead with Salken" {
				t.Errorf("label = %q", o.Label)
			}
			events := fx.apply(o.Command)
			if changed, ok := find[api.LeaderChanged](events); !ok || changed.Leader != other {
				t.Fatalf("leader changed = %+v, %v", changed, ok)
			}
			if got := fx.members(); !slices.Equal(got, []api.UID{other, leader}) {
				t.Errorf("order = %v, want %v", got, []api.UID{other, leader})
			}
			if got := fx.party().Leader; got != other {
				t.Errorf("leader = %d, want %d", got, other)
			}
			if got := fx.controller(other); got != "player" {
				t.Errorf("new leader controller = %q, want player", got)
			}
			want := "follow"
			if individual {
				want = "player"
			}
			if got := fx.controller(leader); got != want {
				t.Errorf("old leader controller = %q, want %q", got, want)
			}
			if events := fx.apply(api.Lead{Party: fx.view().Party.UID, Member: other}); len(events) == 0 {
				t.Error("leading again published nothing, want a rejection")
			}
			// The new leader walks and the old one trails it.
			for fx.g.Current() != other {
				fx.apply(fx.option("wait/event").Command)
			}
			fx.step("e")
			fx.standingApart()
		})
	}
}

func TestReplayingThreeHundredRandomOptionsGivesTheSameHash(t *testing.T) {
	fx := newFixture(t, "map.test.descent", nil)
	rnd := rng.New(11).Get("test.options")
	for range 300 {
		var options []api.Option
		for _, o := range fx.view().Options {
			if o.Enabled {
				options = append(options, o)
			}
		}
		if len(options) == 0 {
			t.Fatal("no enabled option")
		}
		fx.apply(options[rnd.IntN(len(options))].Command)
		fx.standingApart()
	}
	want, err := fx.g.Hash()
	if err != nil {
		t.Fatal(err)
	}
	replayed, err := game.Replay(fx.cfg, fx.g.History())
	if err != nil {
		t.Fatal(err)
	}
	if got, err := replayed.Hash(); err != nil || got != want {
		t.Errorf("replay hash = %x, %v; want %x", got, err, want)
	}
}
