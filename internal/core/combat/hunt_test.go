package combat_test

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
)

var update = flag.Bool("update", false, "rewrite the golden files")

// huntField is the hall with a hunting crawler, and a party that can fight.
func huntField(t *testing.T, mutate func(*content.Bundle)) *field {
	t.Helper()
	return newField(t, "map.test.hall", func(b *content.Bundle) {
		crawler := b.Templates["template.creature.moving_clay"]
		crawler.Actor = &content.ActorDef{Controller: "hunt"}
		crawler.Actions = []api.ContentID{"action.step", "action.strike"}
		crawler.Lists = map[string][]string{"types": {"type.cut"}}
		b.Templates["template.creature.moving_clay"] = crawler
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
	events, err := f.g.Apply(api.Wait{Actor: f.view().Timeline.Current, Until: api.WaitEvent})
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
		if s, ok := e.(api.ActionStarted); ok && s.Actor == actor {
			return s, true
		}
	}
	return api.ActionStarted{}, false
}

func (f *field) moves(actor api.UID) int {
	n := 0
	for _, e := range f.events {
		if m, ok := e.(api.Moved); ok && m.Actor == actor {
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

func (f *field) crawlerStarted(uid api.UID) bool {
	_, ok := f.started(uid)
	return ok
}

func TestAHunterClosesInOverSeveralStepsAndBites(t *testing.T) {
	f := huntField(t, nil)
	creature := f.creature()
	uid := f.w.UIDOf(creature)
	f.moveParty(east(f.at(creature), -5))
	if !f.until(300, func() bool { return f.crawlerStarted(uid) }) {
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

func TestThreatsListTheSwingUntilItLandsOrMisses(t *testing.T) {
	f := huntField(t, nil)
	creature := f.creature()
	uid := f.w.UIDOf(creature)
	f.moveParty(east(f.at(creature), -5))
	if !f.until(300, func() bool { return f.crawlerStarted(uid) }) {
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

func TestSteppingOutOfReachBeforeTheSwingLandsMakesItMiss(t *testing.T) {
	f := huntField(t, nil)
	creature := f.creature()
	uid := f.w.UIDOf(creature)
	f.moveParty(east(f.at(creature), -4))
	if !f.until(300, func() bool { return f.crawlerStarted(uid) }) {
		t.Fatal("the creature never attacked")
	}
	swing, _ := f.started(uid)
	// Let the target, whoever decides for it, step away. Both members are the player's now.
	f.until(10, func() bool { return f.view().Timeline.Current == swing.Target })
	if f.view().Timeline.Current != swing.Target {
		t.Fatal("the target never got a turn before the swing landed")
	}
	if f.w.Now >= swing.Lands {
		t.Fatalf("the target's turn came at %d, after the swing landed at %d", f.w.Now, swing.Lands)
	}
	f.option("step/w")
	f.until(20, func() bool { return f.w.Now > swing.Lands })
	var missed bool
	for _, e := range f.events {
		if m, ok := e.(api.ActionMissed); ok && m.Actor == uid && m.Tick == swing.Lands {
			missed = true
		}
		if d, ok := e.(api.PartDamaged); ok && d.Source == uid && d.Tick == swing.Lands {
			t.Errorf("the swing landed anyway: %+v", d)
		}
	}
	if !missed {
		t.Error("the swing never missed")
	}
}

func TestACreatureWithItsLegsSeveredFallsBehindAndTheAlertEnds(t *testing.T) {
	// A short sight, so the hall is big enough to walk out of it.
	f := huntField(t, func(b *content.Bundle) { b.Combat.SightRadius = 4 })
	creature := f.creature()
	f.damage(f.members()[0], creature, "legs", "99")
	start := f.at(creature)
	f.place(f.members()[0], east(start, -3))
	f.place(f.members()[1], east(start, -2))
	f.events = nil
	ended := f.until(600, func() bool {
		for _, e := range f.events {
			if _, ok := e.(api.AlertEnded); ok {
				return true
			}
		}
		for _, o := range f.view().Options {
			if o.ID == "step/w" && o.Enabled {
				if _, err := f.g.Apply(o.Command); err != nil {
					t.Fatal(err)
				}
				return false
			}
		}
		return false
	})
	if !ended {
		t.Fatal("the alert never ended")
	}
	if crawled, walked := f.moves(f.w.UIDOf(creature)), f.moves(f.memberUIDs()[0]); crawled >= walked {
		t.Errorf("the legless creature took %d steps to the walker's %d: it did not fall behind", crawled, walked)
	}
}

func TestHostileFactionsFightEachOtherOutOfSight(t *testing.T) {
	f := huntField(t, func(b *content.Bundle) {
		hall := b.Maps["map.test.hall"]
		hall.Spawns = nil
		b.Maps["map.test.hall"] = hall
		b.Factions["faction.a"] = content.Faction{Name: "text.faction.deep", HostileTo: []api.ContentID{"faction.b"}}
		b.Factions["faction.b"] = content.Faction{Name: "text.faction.deep", HostileTo: []api.ContentID{"faction.a"}}
		for _, side := range []string{"a", "b"} {
			faction := api.ContentID("faction." + side)
			crawler := b.Templates["template.creature.moving_clay"]
			crawler.Faction = &faction
			b.Templates["template.test.side_"+side] = crawler
			b.Text["text.template.test.side_"+side] = side
		}
	})
	member := f.members()[0]
	origin := f.at(member)
	mapID := ecs.NewMap[comp.Pos](f.w.ECS).Get(member).Map
	var fighters []ecs.Entity
	for i, side := range []string{"a", "b"} {
		e, err := f.w.Spawn(api.ContentID("template.test.side_"+side), &comp.Pos{Map: mapID, At: east(origin, -8+i*3)})
		if err != nil {
			t.Fatal(err)
		}
		fighters = append(fighters, e)
	}
	f.moveParty(east(origin, 5))
	f.events = nil
	a, b := f.w.UIDOf(fighters[0]), f.w.UIDOf(fighters[1])
	damaged := f.until(400, func() bool {
		for _, e := range f.events {
			if d, ok := e.(api.PartDamaged); ok && (d.Source == a && d.Body == b || d.Source == b && d.Body == a) {
				return true
			}
		}
		return false
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
	// Out of reach: the creature is across the hall.
	head := f.partOf(creature, "head")
	far := api.Do{
		Actor: f.w.UIDOf(member), Action: "action.strike", Target: f.w.UIDOf(creature), TargetPart: f.w.UIDOf(head),
		Choices: map[string]string{"type": "type.cut"},
	}
	events, err := f.g.Apply(far)
	if err != nil {
		t.Fatal(err)
	}
	if !rejectedWithReason(events) {
		t.Errorf("a strike at something out of reach was not rejected with a reason: %+v", events)
	}
	// Close in, and the option's own command starts.
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
	if !has(events, func(e api.ActionStarted) bool { return e.Actor == strike.Actor }) {
		t.Fatalf("the option's own command did not start: %+v", events)
	}
	// The arm it used is busy: the same command is refused until it is free.
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
	// A creature that stays put, so the sweep is not cut short by its bite.
	f := huntField(t, func(b *content.Bundle) {
		crawler := b.Templates["template.creature.moving_clay"]
		crawler.Actor = &content.ActorDef{Controller: "idle"}
		b.Templates["template.creature.moving_clay"] = crawler
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
	count := 0
	for _, e := range f.events {
		if _, ok := e.(api.AlertStarted); ok {
			count++
		}
	}
	if count != 1 {
		t.Errorf("AlertStarted published %d times for one alert, want 1", count)
	}
}

func TestAScriptedEncounterMatchesItsGoldenLog(t *testing.T) {
	f := huntField(t, nil)
	creature := f.creature()
	f.moveParty(east(f.at(creature), -4))
	f.until(60, func() bool { return f.w.Now >= 120 })
	var lines []string
	for _, e := range f.events {
		switch e.(type) {
		case api.Ready, api.PhaseStarted:
			continue
		}
		data, err := json.Marshal(e)
		if err != nil {
			t.Fatal(err)
		}
		lines = append(lines, fmt.Sprintf("%s %s", e.EventKind(), data))
	}
	got := strings.Join(lines, "\n") + "\n"
	path := filepath.Join("testdata", "hunt.golden")
	if *update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != string(want) {
		t.Errorf("the encounter differs from %s; run go test ./internal/core/combat -update\n%s", path, got)
	}
}
