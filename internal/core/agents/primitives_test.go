package agents_test

import (
	"strings"
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/inventory"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

const hallMap = "map.test.hall"

// primField is the hall with no creatures, the given marks, and a record of every event.
type primField struct {
	*agentField
	events []api.Event
}

// hallMark puts a mark's character in the hall at a column and row of open floor.
type hallMark struct {
	char     string
	name     string
	col, row int
}

func newPrimField(t *testing.T, seed uint64, marks []hallMark, mutate func(*content.Bundle)) *primField {
	t.Helper()
	f := newAgentField(t, seed, func(bundle *content.Bundle) {
		contenttest.Edit(bundle.Maps, hallMap, func(hall *content.MapDef) {
			hall.Spawns = map[string]api.ContentID{}
			hall.Layers = append([]content.LayerDef(nil), hall.Layers...)
			layer := &hall.Layers[0]
			layer.Rows = append([]string(nil), layer.Rows...)
			for _, m := range marks {
				row := []rune(layer.Rows[m.row])
				if row[m.col] != '.' {
					t.Fatalf("hall row %d column %d is %q, want open floor", m.row, m.col, row[m.col])
				}
				row[m.col] = []rune(m.char)[0]
				layer.Rows[m.row] = string(row)
				hall.Marks[m.char] = m.name
			}
		})
		if mutate != nil {
			mutate(bundle)
		}
	})
	pf := &primField{agentField: f}
	f.w.Bus.SubscribeAll("test.record", func(_ *sim.World, e api.Event) { pf.events = append(pf.events, e) })
	return pf
}

func (f *primField) wait() {
	f.t.Helper()
	if _, err := f.g.Apply(api.Wait{Source: f.g.Current(), Until: api.WaitEvent}); err != nil {
		f.t.Fatal(err)
	}
}

func (f *primField) until(limit int, stop func() bool) bool {
	f.t.Helper()
	for range limit {
		if stop() {
			return true
		}
		f.wait()
	}
	return stop()
}

func cellAt(col, row int) api.Cell { return api.Cell{Hex: hexmap.FromOffsetOddR(col, row)} }

func (f *primField) spawn(template api.ContentID, at api.Cell) ecs.Entity {
	f.t.Helper()
	body, err := f.w.Spawn(template, &comp.Pos{Map: ecs.NewMap[comp.Pos](f.w.ECS).Get(f.leader()).Map, At: at})
	if err != nil {
		f.t.Fatal(err)
	}
	return body
}

func (f *primField) leader() ecs.Entity {
	f.t.Helper()
	party, ok := anatomy.ControlledParty(f.w)
	if !ok {
		f.t.Fatal("no controlled party")
	}
	return anatomy.Members(f.w, party)[0]
}

func (f *primField) started(actor, target api.UID, action api.ContentID) bool {
	for _, e := range f.events {
		if s, ok := e.(api.ActionStarted); ok && s.Source == actor && s.Target == target && s.Action == action {
			return true
		}
	}
	return false
}

func (f *primField) countWater(holder ecs.Entity) int {
	return inventory.Count(f.w, holder, "item.lake_water", 0)
}

func TestAConsumerWalksToACorpseAndStartsItsAction(t *testing.T) {
	f := newPrimField(t, 1, nil, nil)
	corpse := f.spawn("template.test.dummy", cellAt(17, 10))
	ecs.NewMap[comp.Actor](f.w.ECS).Remove(corpse)
	ecs.NewMap[comp.Dead](f.w.ECS).Add(corpse, &comp.Dead{})
	eater := f.spawn("template.test.eater", cellAt(12, 11))
	eaterUID, corpseUID := f.w.UIDOf(eater), f.w.UIDOf(corpse)
	if !f.until(300, func() bool { return f.started(eaterUID, corpseUID, "action.gnaw") }) {
		t.Fatalf("the eater never gnawed the corpse; it is at %v", f.at(eater))
	}
	if d := hexmap.CellDistance(f.at(eater), f.at(corpse)); d > 1 {
		t.Errorf("the eater gnawed from %d cells away", d)
	}
	if got := f.doing(eater); got == "behavior.test.gnaw_corpse" {
		t.Error("the eater still holds the behavior it finished")
	}
}

func TestAHaulerMovesExactlyItsCountFromTheWellToTheCistern(t *testing.T) {
	f := newPrimField(t, 1, []hallMark{{"q", "cistern", 6, 6}}, nil)
	well := f.spawn("template.test.well", cellAt(14, 4))
	cistern := f.spawn("template.test.cistern", cellAt(6, 6))
	hauler := f.spawn("template.test.hauler", cellAt(14, 8))
	if got := f.countWater(well); got != 5 {
		t.Fatalf("the well holds %d water, want 5", got)
	}
	if !f.until(600, func() bool { return f.countWater(cistern) > 0 }) {
		t.Fatalf("nothing reached the cistern by tick %d; the hauler is at %v", f.w.Now, f.at(hauler))
	}
	if got := f.countWater(cistern); got != 2 {
		t.Errorf("the cistern holds %d water, want 2", got)
	}
	if got := f.countWater(well); got != 3 {
		t.Errorf("the well holds %d water, want 3", got)
	}
	if got := f.countWater(hauler); got != 0 {
		t.Errorf("the hauler still carries %d water", got)
	}
}

func TestAHaulerGoesAroundAWallOfStandingBodies(t *testing.T) {
	f := newPrimField(t, 1, []hallMark{{"q", "cistern", 6, 6}}, nil)
	f.spawn("template.test.well", cellAt(14, 4))
	cistern := f.spawn("template.test.cistern", cellAt(6, 6))
	for col := 8; col <= 16; col++ {
		f.spawn("template.test.dummy", cellAt(col, 6))
	}
	hauler := f.spawn("template.test.hauler", cellAt(14, 8))
	if !f.until(800, func() bool { return f.countWater(cistern) > 0 }) {
		t.Fatalf("nothing reached the cistern by tick %d; the hauler is at %v", f.w.Now, f.at(hauler))
	}
}

func TestAFleeingBodyRunsFromAHostileUntilItIsOutOfSight(t *testing.T) {
	f := newPrimField(t, 1, nil, nil)
	leader := f.leader()
	runner := f.spawn("template.test.runner", cellAt(13, 10))
	nearest := func() int {
		best := -1
		for _, member := range anatomy.Members(f.w, f.partyOf(leader)) {
			if d := hexmap.CellDistance(f.at(runner), f.at(member)); best < 0 || d < best {
				best = d
			}
		}
		return best
	}
	start := nearest()
	sawRun := false
	last := start
	for range 200 {
		f.wait()
		running := f.doing(runner) == "behavior.test.run"
		sawRun = sawRun || running
		if running {
			if d := nearest(); d < last {
				t.Fatalf("at tick %d the runner moved from %d to %d cells from the party", f.w.Now, last, d)
			}
		}
		last = nearest()
		if sawRun && !running {
			break
		}
	}
	if !sawRun {
		t.Fatal("the runner never fled")
	}
	if last <= start || last <= f.w.Content.Combat().SightRadius {
		t.Errorf("the runner stopped %d cells from the party, having started %d away", last, start)
	}
}

func (f *primField) partyOf(body ecs.Entity) ecs.Entity {
	f.t.Helper()
	party, ok := anatomy.PartyOf(f.w, body)
	if !ok {
		f.t.Fatal("body has no party")
	}
	return party
}

func TestAGuardStaysAtItsPostUntilAHostileComesNear(t *testing.T) {
	f := newPrimField(t, 1, []hallMark{{"p", "post", 15, 10}}, nil)
	guard := f.spawn("template.test.guard", cellAt(15, 10))
	post := cellAt(15, 10)
	guardUID := f.w.UIDOf(guard)
	for range 40 {
		f.wait()
		if d := hexmap.CellDistance(f.at(guard), post); d > 2 {
			t.Fatalf("at tick %d the guard is %d cells from its post, want at most 2", f.w.Now, d)
		}
	}
	for _, e := range f.events {
		if s, ok := e.(api.ActionStarted); ok && s.Source == guardUID && s.Action != "action.step" {
			t.Fatalf("the guard started %s with nobody near its post", s.Action)
		}
	}
	intruder := f.leader()
	ecs.NewMap[comp.Pos](f.w.ECS).Get(intruder).At = cellAt(17, 10)
	intruderUID := f.w.UIDOf(intruder)
	if !f.until(300, func() bool { return f.started(guardUID, intruderUID, "action.strike") }) {
		t.Fatal("the guard never struck the intruder")
	}
}

func TestAnEngagedBodyFightsAsAHunterDoes(t *testing.T) {
	landings := func(template api.ContentID) []api.ActionStarted {
		f := newPrimField(t, 5, nil, func(bundle *content.Bundle) {
			contenttest.Edit(bundle.Maps, hallMap, func(hall *content.MapDef) {
				hall.Spawns = map[string]api.ContentID{"c": template}
			})
		})
		ecs.NewMap[comp.Pos](f.w.ECS).Get(f.leader()).At = cellAt(13, 10)
		for range 150 {
			f.wait()
		}
		var out []api.ActionStarted
		for _, e := range f.events {
			if s, ok := e.(api.ActionStarted); ok {
				out = append(out, s)
			}
		}
		return out
	}
	hunted, engaged := landings("template.creature.moving_clay"), landings("template.test.engager")
	if len(hunted) == 0 {
		t.Fatal("the hunter started nothing; the test shows nothing")
	}
	if len(hunted) != len(engaged) {
		t.Fatalf("the hunter started %d actions, the engaged body %d", len(hunted), len(engaged))
	}
	for i := range hunted {
		if hunted[i] != engaged[i] {
			t.Errorf("action %d: hunter %+v, engaged %+v", i, hunted[i], engaged[i])
		}
	}
}

func TestPrimitiveParamsAreCheckedWhenAGameStarts(t *testing.T) {
	tests := []struct {
		name string
		id   string
		edit func(*content.BehaviorDef)
		want string
	}{
		{"carry without a destination", "behavior.test.haul_water", func(def *content.BehaviorDef) {
			delete(def.Params, "to")
		}, `needs "to"`},
		{"consume without an action", "behavior.test.gnaw_corpse", func(def *content.BehaviorDef) {
			def.Params = nil
		}, `needs "action"`},
		{"flee with a param", "behavior.test.run", func(def *content.BehaviorDef) {
			def.Params = map[string]string{"speed": "2"}
		}, `takes no "speed"`},
		{"guard without a radius", "behavior.test.guard_post", func(def *content.BehaviorDef) {
			def.Params = nil
		}, `needs "radius"`},
		{"engage with an unknown param", "behavior.test.fight", func(def *content.BehaviorDef) {
			def.Params = map[string]string{"weapon": "axe"}
		}, `takes no "weapon"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bundle := contenttest.SampleBundle(t)
			contenttest.Edit(bundle.Behaviors, tt.id, tt.edit)
			_, err := game.New(game.Config{Seed: 1, Content: contenttest.Registry(t, bundle)})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error = %v, want one that mentions %q", err, tt.want)
			}
		})
	}
}

func TestEngageMayNameItsController(t *testing.T) {
	bundle := contenttest.SampleBundle(t)
	contenttest.Edit(bundle.Behaviors, "behavior.test.fight", func(def *content.BehaviorDef) {
		def.Params = map[string]string{"controller": "hunt"}
	})
	if _, err := game.New(game.Config{Seed: 1, Content: contenttest.Registry(t, bundle)}); err != nil {
		t.Fatal(err)
	}
}
