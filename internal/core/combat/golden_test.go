package combat_test

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/actions"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

// The golden encounters. Each is a scripted fight: the setup places bodies with the test fixtures,
// and from there every move is an option of the view, as a player's would be. The events the script
// delivers, encoded with the codec, are compared to testdata/encounters/<name>.jsonl. Regenerate
// them with -update, and read the diff.

// compareGolden writes or checks the golden file of an encounter. Ready and phase events are left
// out: they say nothing of the fight and fill the file.
func compareGolden(t *testing.T, name string, w *sim.World, events []api.Event) {
	t.Helper()
	var got bytes.Buffer
	for _, e := range events {
		switch e.(type) {
		case api.Ready, api.PhaseStarted:
			continue
		}
		line, err := w.Codec.EncodeEvent(e)
		if err != nil {
			t.Fatal(err)
		}
		got.Write(line)
		got.WriteByte('\n')
	}
	path := filepath.Join("testdata", "encounters", name+".jsonl")
	if *update {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, got.Bytes(), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%v; run go test ./internal/core/combat -update", err)
	}
	if !bytes.Equal(bytes.ReplaceAll(want, []byte("\r\n"), []byte("\n")), got.Bytes()) {
		t.Errorf("the encounter differs from %s; run go test ./internal/core/combat -update\n%s", path, got.String())
	}
}

// ringEncounter sets up a ring, records every event the script delivers and checks the golden file.
func ringEncounter(t *testing.T, name string, at map[string]int, script func(r *ring)) {
	t.Helper()
	r := newRing(t, at)
	var events []api.Event
	r.w.Bus.SubscribeAll("test.golden", func(_ *sim.World, e api.Event) { events = append(events, e) })
	script(r)
	compareGolden(t, name, r.w, events)
}

func TestGoldenSoakAndStrikeSeverClay(t *testing.T) {
	ringEncounter(t, "soak_strike_severs_clay", map[string]int{"fighter": 0, "clay": 1}, func(r *ring) {
		clay := r.body("clay")
		arm := r.part(clay, "arm")
		r.give("item.lake_water", 3, ecs.Entity{})
		r.do("strike", "type.soak", r.uid(clay), r.uid(arm))
		events := r.wait(8)
		if !has(
			events,
			func(e api.PartSevered) bool { return e.Cause == api.SeveredWorn && e.Part == r.w.UIDOf(arm) },
		) {
			t.Errorf("a Soak Strike did not wear the clay arm off: %v", events)
		}
		if r.attached(clay, "arm") != 0 {
			t.Error("the clay arm is still attached")
		}
	})
}

func TestGoldenSweepRevealsThenSeverTargets(t *testing.T) {
	ringEncounter(t, "sweep_then_sever", map[string]int{"fighter": 0, "clay": 1}, func(r *ring) {
		clay := r.body("clay")
		arm := r.part(clay, "arm")
		r.give("item.lake_water", 3, ecs.Entity{})
		r.do("sweep", "type.cut", r.uid(clay))
		events := r.wait(12)
		learned := count[api.FactLearned](events)
		if learned < len(anatomy.Parts(r.w, clay)) {
			t.Errorf("the Sweep taught %d facts, want one per part", learned)
		}
		events = r.do("sever", "type.soak", r.uid(clay), r.uid(arm))
		events = append(events, r.wait(10)...)
		if !has(events, func(e api.PartSevered) bool { return e.Cause == api.SeveredCut && e.Part == r.w.UIDOf(arm) }) {
			t.Errorf("the targeted Sever did not cut the clay arm: %v", events)
		}
	})
}

func TestGoldenPinCancelsASwingThenStandFreezesTheCrawlers(t *testing.T) {
	ringEncounter(t, "pin_then_stand", map[string]int{"fighter": 0, "ally": 1, "crawler": -1}, func(r *ring) {
		ally, crawler := r.body("ally"), r.body("crawler")
		err := actions.Start(r.w, actions.Intent{
			Actor: ally, Action: "action.sweep", Target: r.body("fighter"),
			Choices: map[string]string{"type": "type.cut"},
		})
		if err != nil {
			t.Fatal(err)
		}
		arm := r.part(ally, "arm")
		events := r.do("pin", r.uid(ally), r.uid(arm))
		events = append(events, r.wait(3)...)
		if !has(events, always[api.ActionCanceled]) {
			t.Errorf("pinning the winding arm did not cancel the swing: %v", events)
		}
		if has(events, func(e api.PartDamaged) bool { return e.Body == r.w.UIDOf(r.body("fighter")) }) {
			t.Error("the canceled swing still hurt")
		}
		events = r.do("stand")
		if !has(events, always[api.ChannelsDelayed]) || r.busyChannels(crawler) == 0 {
			t.Errorf("Stand did not freeze the crawler: %v", events)
		}
		r.wait(10)
	})
}

func TestGoldenGiveAndGraftBetweenMembers(t *testing.T) {
	ringEncounter(t, "give_and_graft", map[string]int{"fighter": 0, "victim": 1, "ally": -1}, func(r *ring) {
		victim, ally := r.body("victim"), r.body("ally")
		r.give("item.test_arm", 1, ecs.Entity{})
		before := len(timeline.Channels(r.w, victim))
		events := r.do("graft", "item.test_arm", r.uid(victim))
		events = append(events, r.wait(14)...)
		if !has(events, func(e api.PartAttached) bool { return e.From == api.AttachedFromPack }) ||
			len(timeline.Channels(r.w, victim)) != before+1 {
			t.Errorf("the graft did not fit: %v", events)
		}
		for _, c := range timeline.Channels(r.w, ally) {
			if err := timeline.Occupy(r.w, c, 30); err != nil {
				t.Fatal(err)
			}
		}
		events = r.do("give", r.uid(ally))
		if !has(events, always[api.ChannelsReadied]) {
			t.Errorf("Give readied nothing: %v", events)
		}
		r.wait(2)
	})
}

// fieldEncounter records the events of a hall fight from the moment the script starts.
func fieldEncounter(t *testing.T, name string, f *field, script func()) {
	t.Helper()
	f.events = nil
	script()
	compareGolden(t, name, f.w, f.events)
}

func TestGoldenACrawlingPartOutlivesItsBody(t *testing.T) {
	f, remnant := remnantField(t)
	fieldEncounter(t, "crawling_part_outlives_body", f, func() {
		reaper := f.members()[0]
		arm := f.partOf(remnant, "arm")
		f.apply(reaper, remnant, arm, do("sever", "parts", "target_part"))
		crawling := ecs.NewMap[comp.PartOf](f.w.ECS).GetRelation(arm)
		// The source and the hall's own creature fall: only the crawling arm is left to see.
		f.damage(reaper, remnant, "torso", "99")
		f.damage(reaper, f.creature(), "torso", "99")
		dead := ecs.NewMap[comp.Dead](f.w.ECS)
		if !dead.Has(remnant) || dead.Has(crawling) {
			t.Fatal("the remnant must have fallen and its crawling arm lived")
		}
		for range 6 {
			f.option("wait/event")
		}
		if !f.view().Party.Alert || hasEvent[api.AlertEnded](f.events) {
			t.Fatal("the alert dropped while the crawling arm still lived")
		}
		f.damage(reaper, crawling, "arm", "99")
		if !f.until(400, func() bool { return !f.view().Party.Alert }) {
			t.Error("the alert never ended after the crawling arm died")
		}
	})
}

func TestGoldenADodgeStepsOutOfReachBeforeTheSwingLands(t *testing.T) {
	f := huntField(t, nil)
	creature := f.creature()
	uid := f.w.UIDOf(creature)
	f.moveParty(east(f.at(creature), -4))
	fieldEncounter(t, "dodge", f, func() {
		if !f.until(300, func() bool { return f.crawlerStarted(uid) }) {
			t.Fatal("the creature never attacked")
		}
		swing, _ := f.started(uid)
		f.until(10, func() bool { return f.view().Timeline.Current == swing.Target })
		if f.view().Timeline.Current != swing.Target || f.w.Now >= swing.Lands {
			t.Fatal("the target had no turn before the swing landed")
		}
		f.option("step/w")
		f.until(20, func() bool { return f.w.Now > swing.Lands })
		missed := slices.ContainsFunc(f.events, func(e api.Event) bool {
			m, ok := e.(api.ActionMissed)
			return ok && m.Actor == uid && m.Tick == swing.Lands
		})
		if !missed {
			t.Error("the swing did not miss the dodging target")
		}
	})
}

func TestGoldenOutpacingACreatureWithItsLegsSevered(t *testing.T) {
	f := huntField(t, func(b *content.Bundle) { b.Combat.SightRadius = 4 })
	creature := f.creature()
	f.damage(f.members()[0], creature, "legs", "99")
	start := f.at(creature)
	f.place(f.members()[0], east(start, -3))
	f.place(f.members()[1], east(start, -2))
	fieldEncounter(t, "outpaced", f, func() {
		ended := f.until(600, func() bool {
			if hasEvent[api.AlertEnded](f.events) {
				return true
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
	})
}
