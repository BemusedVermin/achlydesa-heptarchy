package combat_test

import (
	"flag"
	"slices"
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

var update = flag.Bool("update", false, "rewrite the golden files")

// huntField is the hall with a hunting moving clay, and a party that can fight.
func huntField(t *testing.T, mutate func(*content.Bundle)) *field {
	t.Helper()
	return newField(t, "place.test.hall", func(b *content.Bundle) {
		movingClay := b.Templates["template.creature.moving_clay"]
		movingClay.Actor = &content.ActorDef{Controller: "hunt"}
		movingClay.Actions = []api.ContentID{"action.step", "action.strike"}
		movingClay.Lists = map[string][]string{"types": {"type.cut"}}
		b.Templates["template.creature.moving_clay"] = movingClay
		for _, id := range []string{"template.person.reaper", "template.person.salken"} {
			person := b.Templates[id]
			person.Actions = []api.ContentID{"action.step", "action.strike", "action.sweep"}
			b.Templates[id] = person
		}
		if mutate != nil {
			mutate(b)
		}
	})
}

// place puts a body in a cell.
func (f *field) place(e ecs.Entity, at api.Cell) {
	f.t.Helper()
	ecs.NewMap[comp.Pos](f.w.ECS).Get(e).At = at
}

func (f *field) at(e ecs.Entity) api.Cell { return ecs.NewMap[comp.Pos](f.w.ECS).Get(e).At }

// east is the cell n to the east of a cell.
func east(c api.Cell, n int) api.Cell {
	c.Hex.Q += n
	return c
}

// wait lets the body that is deciding wait for the next entry on the timeline.
func (f *field) wait() []api.Event {
	f.t.Helper()
	events, err := f.g.Apply(api.Wait{Source: f.view().Timeline.Current, Until: api.WaitEvent})
	if err != nil {
		f.t.Fatal(err)
	}
	return events
}

// until waits turn by turn, for at most limit waits, until stop says so. It reports whether it did.
func (f *field) until(limit int, stop func() bool) bool {
	f.t.Helper()
	for range limit {
		if stop() {
			return true
		}
		f.wait()
	}
	return stop()
}

// started is the first action the body started, from the events so far.
func (f *field) started(actor api.UID) (api.ActionStarted, bool) {
	for _, e := range f.events {
		if s, ok := e.(api.ActionStarted); ok && s.Source == actor {
			return s, true
		}
	}
	return api.ActionStarted{}, false
}

func (f *field) moves(actor api.UID) int {
	n := 0
	for _, e := range f.events {
		if m, ok := e.(api.Moved); ok && m.Source == actor {
			n++
		}
	}
	return n
}

// partOf is the body's first part of a slot.
func (f *field) partOf(body ecs.Entity, slot string) ecs.Entity {
	f.t.Helper()
	parts := ecs.NewMap[comp.Part](f.w.ECS)
	for _, part := range anatomy.Parts(f.w, body) {
		if parts.Get(part).Slot == slot {
			return part
		}
	}
	f.t.Fatalf("no %s", slot)
	return ecs.Entity{}
}

// bodyView finds a body in the bodies view.
func (f *field) bodyView(uid api.UID) (api.BodyView, bool) {
	f.t.Helper()
	view := f.view().Bodies
	if view == nil {
		f.t.Fatal("no bodies view")
	}
	for _, b := range view.Bodies {
		if b.UID == uid {
			return b, true
		}
	}
	return api.BodyView{}, false
}

func (f *field) memberUIDs() []api.UID {
	var out []api.UID
	for _, m := range f.members() {
		out = append(out, f.w.UIDOf(m))
	}
	return out
}

func (f *field) movingClayStarted(uid api.UID) bool {
	_, ok := f.started(uid)
	return ok
}

func TestAHunterClosesInOverSeveralStepsAndBites(t *testing.T) {
	f := huntField(t, nil)
	creature := f.creature()
	uid := f.w.UIDOf(creature)
	f.moveParty(east(f.at(creature), -5))
	if !f.until(300, func() bool { return f.movingClayStarted(uid) }) {
		t.Fatal("the creature never attacked")
	}
	if got := f.moves(uid); got < 3 {
		t.Errorf("the creature took %d steps before its first swing, want several", got)
	}
	swing, _ := f.started(uid)
	if swing.Action != "action.strike" {
		t.Errorf("first action = %s, want a strike", swing.Action)
	}
	if !slices.Contains(f.memberUIDs(), swing.Target) {
		t.Errorf("the swing is aimed at %d, not a member of the party", swing.Target)
	}
}

func TestADelegatedHunterHandsTheBodyBackWithNoPreyInSight(t *testing.T) {
	f := huntField(t, func(b *content.Bundle) {
		movingClay := b.Templates["template.creature.moving_clay"]
		movingClay.Actor = &content.ActorDef{Controller: "still"}
		movingClay.Mind = nil
		b.Templates["template.creature.moving_clay"] = movingClay
	})
	creature := f.creature()
	uid := f.w.UIDOf(creature)
	if err := timeline.DelegateTo(f.w, creature, "hunt"); err != nil {
		t.Fatal(err)
	}
	f.moveParty(east(f.at(creature), -5))
	if !f.until(300, func() bool { return f.movingClayStarted(uid) }) {
		t.Fatal("the delegated hunter never attacked")
	}
	if got := timeline.ControllerOf(f.w, creature); got != "hunt" {
		t.Fatalf("controller with prey in sight = %q, want hunt", got)
	}
	f.moveParty(east(f.at(creature), -12))
	if !f.until(300, func() bool { return timeline.ControllerOf(f.w, creature) == "still" }) {
		t.Errorf(
			"controller with no prey in sight = %q, want the body's own, still",
			timeline.ControllerOf(f.w, creature),
		)
	}
}

func TestThreatsListTheSwingUntilItLandsOrMisses(t *testing.T) {
	f := huntField(t, nil)
	creature := f.creature()
	uid := f.w.UIDOf(creature)
	f.moveParty(east(f.at(creature), -5))
	if !f.until(300, func() bool { return f.movingClayStarted(uid) }) {
		t.Fatal("the creature never attacked")
	}
	swing, _ := f.started(uid)
	var threat *api.ThreatView
	for _, th := range f.view().Bodies.Threats {
		if th.Attacker == uid {
			threat = &th
		}
	}
	if threat == nil {
		t.Fatalf("no threat from the creature in %+v", f.view().Bodies.Threats)
	}
	if threat.Tick != swing.Lands || threat.Lands != swing.Lands-f.w.Now || threat.Target != swing.Target {
		t.Errorf("threat = %+v, want landing at %d on %d (now %d)", *threat, swing.Lands, swing.Target, f.w.Now)
	}
	body, _ := f.bodyView(uid)
	winding := false
	for _, c := range body.Channels {
		winding = winding || c.Windup != nil && c.Windup.Target == swing.Target
	}
	if !winding {
		t.Errorf("no part of the creature shows the wind-up: %+v", body.Channels)
	}
	threatened := func() bool {
		for _, th := range f.view().Bodies.Threats {
			if th.Attacker == uid && th.Tick == swing.Lands {
				return true
			}
		}
		return false
	}
	f.until(40, func() bool { return !threatened() })
	if threatened() {
		t.Error("the swing is still a threat after it landed")
	}
	if f.w.Now < swing.Lands {
		t.Errorf("the threat went at %d, before the swing lands at %d", f.w.Now, swing.Lands)
	}
}

func TestHostileFactionsFightEachOtherOutOfSight(t *testing.T) {
	f := huntField(t, func(b *content.Bundle) {
		hall := b.Places["place.test.hall"]
		hall.Spawns = nil
		b.Places["place.test.hall"] = hall
		b.Factions["faction.a"] = content.Faction{Name: "text.faction.a", HostileTo: []api.ContentID{"faction.b"}}
		b.Factions["faction.b"] = content.Faction{Name: "text.faction.b", HostileTo: []api.ContentID{"faction.a"}}
		b.Text["text.faction.a"], b.Text["text.faction.b"] = "a", "b"
		for _, side := range []string{"a", "b"} {
			faction := api.ContentID("faction." + side)
			movingClay := b.Templates["template.creature.moving_clay"]
			movingClay.Faction = &faction
			textID := api.ContentID("text.template.test.side_" + side)
			movingClay.Name = &textID
			b.Templates["template.test.side_"+side] = movingClay
			b.Text["text.template.test.side_"+side] = side
		}
	})
	member := f.members()[0]
	origin := f.at(member)
	var fighters []ecs.Entity
	for i, side := range []string{"a", "b"} {
		e, err := f.w.Spawn(api.ContentID("template.test.side_"+side), &comp.Pos{At: east(origin, -8+i*3)})
		if err != nil {
			t.Fatal(err)
		}
		fighters = append(fighters, e)
	}
	f.moveParty(east(origin, 5))
	f.events = nil
	a, b := f.w.UIDOf(fighters[0]), f.w.UIDOf(fighters[1])
	damaged := f.until(400, func() bool {
		return has(f.events, func(d api.PartDamaged) bool {
			return d.Source == a && d.Body == b || d.Source == b && d.Body == a
		})
	})
	if !damaged {
		t.Fatal("the two creatures never hurt each other")
	}
	if f.view().Party.Alert {
		t.Error("the party was alerted by a fight it cannot see")
	}
}

func TestOptionsAreLegalAndAHandBuiltDoIsRefused(t *testing.T) {
	f := huntField(t, nil)
	creature := f.creature()
	member := f.members()[0]
	f.until(5, func() bool { return f.view().Timeline.Current == f.w.UIDOf(member) })
	head := f.partOf(creature, "head")
	far := api.Do{
		Source: f.w.UIDOf(member), Action: "action.strike", Target: f.w.UIDOf(creature), TargetPart: f.w.UIDOf(head),
		Choices: map[string]string{"type": "type.cut"},
	}
	events, err := f.g.Apply(far)
	if err != nil {
		t.Fatal(err)
	}
	if !rejectedWithReason(events) {
		t.Errorf("a strike at something out of reach was not rejected with a reason: %+v", events)
	}
	f.place(creature, east(f.at(member), 1))
	var strike *api.Do
	for _, o := range f.view().Options {
		if do, ok := o.Command.(api.Do); ok && o.Enabled && do.Action == "action.strike" &&
			do.Target == f.w.UIDOf(creature) {
			strike = &do
			break
		}
	}
	if strike == nil {
		t.Fatalf("no strike option against the creature beside the member: %v", optionIDs(f.view()))
	}
	events, err = f.g.Apply(*strike)
	if err != nil {
		t.Fatal(err)
	}
	if !has(events, func(e api.ActionStarted) bool { return e.Source == strike.Source }) {
		t.Fatalf("the option's own command did not start: %+v", events)
	}
	f.until(3, func() bool { return f.view().Timeline.Current == f.w.UIDOf(member) })
	if f.view().Timeline.Current != f.w.UIDOf(member) {
		t.Fatal("the member never decided again during the wind-up")
	}
	events, err = f.g.Apply(*strike)
	if err != nil {
		t.Fatal(err)
	}
	if !rejectedWithReason(events) {
		t.Errorf("a strike with a busy arm was not rejected with a reason: %+v", events)
	}
}

func rejectedWithReason(events []api.Event) bool {
	return has(events, func(e api.Rejected) bool { return e.Reason.ID != "" })
}

func optionIDs(v api.View) []string {
	var ids []string
	for _, o := range v.Options {
		ids = append(ids, o.ID)
	}
	return ids
}

func TestAMaterialIsUnknownUntilASweepLands(t *testing.T) {
	f := huntField(t, func(b *content.Bundle) {
		movingClay := b.Templates["template.creature.moving_clay"]
		movingClay.Actor = &content.ActorDef{Controller: "idle"}
		b.Templates["template.creature.moving_clay"] = movingClay
	})
	creature := f.creature()
	uid := f.w.UIDOf(creature)
	member := f.members()[0]
	f.until(5, func() bool { return f.view().Timeline.Current == f.w.UIDOf(member) })
	f.place(creature, east(f.at(member), 1))
	body, ok := f.bodyView(uid)
	if !ok || len(body.Channels) == 0 {
		t.Fatalf("the creature is not in the bodies view: %+v", f.view().Bodies)
	}
	for _, c := range body.Channels {
		if c.Material != "?" {
			t.Errorf("%s shows %q before any sweep", c.Name, c.Material)
		}
	}
	self, _ := f.bodyView(f.w.UIDOf(member))
	for _, c := range self.Channels {
		if c.Material == "?" || c.Material == "" {
			t.Errorf("a member's own %s shows %q", c.Name, c.Material)
		}
	}
	for _, o := range f.view().Options {
		if do, ok := o.Command.(api.Do); ok && o.Enabled && do.Action == "action.sweep" && do.Target == uid {
			if _, err := f.g.Apply(do); err != nil {
				t.Fatal(err)
			}
			break
		}
	}
	learned := f.until(30, func() bool {
		b, _ := f.bodyView(uid)
		return len(b.Channels) > 0 && b.Channels[0].Material != "?"
	})
	if !learned {
		t.Fatal("the sweep never taught the party what the creature is made of")
	}
	body, _ = f.bodyView(uid)
	for _, c := range body.Channels {
		if c.Material == "?" {
			t.Errorf("%s still shows ? after the sweep", c.Name)
		}
	}
}

func TestOneAlertPublishesOneAlertStarted(t *testing.T) {
	f := huntField(t, nil)
	creature := f.creature()
	f.moveParty(east(f.at(creature), -4))
	f.until(20, func() bool { return f.w.Now >= 20 })
	if got := count[api.AlertStarted](f.events); got != 1 {
		t.Errorf("AlertStarted published %d times for one alert, want 1", got)
	}
}
