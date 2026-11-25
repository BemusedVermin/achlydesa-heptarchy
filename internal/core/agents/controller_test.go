package agents_test

import (
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game/gametest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

const (
	drinker   = "template.test.drinker"
	wander    = "behavior.test.wander"
	toWater   = "behavior.test.to_water"
	idle      = "behavior.test.idle"
	waterRow  = 3
	waterCol  = 8
	waterMark = "water"
)

// waterCell is where the hall's water mark is put.
var waterCell = api.Cell{Hex: hexmap.FromOffsetOddR(waterCol, waterRow)}

// agentField is the test hall with a drinker where the moving clay stood and a water mark, after
// mutate has edited the content further.
type agentField struct {
	t   *testing.T
	g   *game.Game
	w   *sim.World
	cfg game.Config
}

func newAgentField(t *testing.T, seed uint64, mutate func(*content.Bundle)) *agentField {
	t.Helper()
	f := &agentField{t: t}
	f.g, f.w, f.cfg = gametest.New(t, seed, func(bundle *content.Bundle) {
		contenttest.Edit(bundle.Maps, "map.test.hall", func(hall *content.MapDef) {
			hall.Spawns = map[string]api.ContentID{"c": drinker}
			hall.Marks["w"] = waterMark
			layer := &hall.Layers[0]
			row := []rune(layer.Rows[waterRow])
			if row[waterCol] != '.' {
				t.Fatalf("hall row %d column %d is %q, want open floor", waterRow, waterCol, row[waterCol])
			}
			row[waterCol] = 'w'
			layer.Rows = append([]string(nil), layer.Rows...)
			layer.Rows[waterRow] = string(row)
		})
		contenttest.Edit(bundle.Needs, "need.water", func(need *content.NeedDef) {
			need.Rise = "0.5"
			need.Thresholds = nil
		})
		if mutate != nil {
			mutate(bundle)
		}
	})
	return f
}

func (f *agentField) drinker() ecs.Entity {
	f.t.Helper()
	query := ecs.NewFilter1[comp.Identity](f.w.ECS).Query()
	var found []ecs.Entity
	for query.Next() {
		if query.Get().Template == drinker {
			found = append(found, query.Entity())
		}
	}
	if len(found) != 1 {
		f.t.Fatalf("found %d drinkers, want 1", len(found))
	}
	return found[0]
}

func (f *agentField) at(body ecs.Entity) api.Cell { return ecs.NewMap[comp.Pos](f.w.ECS).Get(body).At }

func (f *agentField) doing(body ecs.Entity) api.ContentID {
	doing := ecs.NewMap[comp.Doing](f.w.ECS)
	if !doing.Has(body) {
		return ""
	}
	return doing.Get(body).Behavior
}

// wait lets the deciding party member wait for the next entry on the timeline.
func (f *agentField) wait() {
	f.t.Helper()
	events, err := f.g.Apply(api.Wait{Source: f.g.Current(), Until: api.WaitEvent})
	if err != nil {
		f.t.Fatal(err)
	}
	for _, e := range events {
		if rejected, ok := e.(api.Rejected); ok {
			f.t.Fatalf("wait rejected: %+v", rejected)
		}
	}
}

// waitUntil waits until the tick is at least until.
func (f *agentField) waitUntil(until int64) {
	f.t.Helper()
	for f.w.Now < until {
		f.wait()
	}
}

func TestADrinkerWandersThenWalksToWaterOnceThirsty(t *testing.T) {
	f := newAgentField(t, 1, nil)
	body := f.drinker()
	start := f.at(body)
	f.waitUntil(19)
	if got := f.doing(body); got != wander {
		t.Errorf("before thirst the drinker is doing %q, want %q", got, wander)
	}
	if f.at(body) == start {
		t.Error("the drinker did not move before it was thirsty")
	}
	if f.at(body) == waterCell {
		t.Fatal("the drinker found the water by chance; move the mark")
	}
	sawToWater := false
	for f.w.Now < 2000 && f.at(body) != waterCell {
		f.wait()
		sawToWater = sawToWater || f.doing(body) == toWater
	}
	if f.at(body) != waterCell {
		t.Fatalf("at tick %d the drinker is at %v, want the water mark %v", f.w.Now, f.at(body), waterCell)
	}
	if !sawToWater {
		t.Error("the drinker reached the water without going for it")
	}
	for range 5 {
		f.wait()
	}
	if f.at(body) != waterCell {
		t.Errorf("the drinker left the water for %v", f.at(body))
	}
}

func TestTheSameSeedGivesTheSamePositions(t *testing.T) {
	positions := func(seed uint64) map[api.UID]api.Cell {
		f := newAgentField(t, seed, nil)
		f.waitUntil(200)
		out := map[api.UID]api.Cell{}
		query := ecs.NewFilter1[comp.Pos](f.w.ECS).Query()
		for query.Next() {
			out[f.w.UIDOf(query.Entity())] = query.Get().At
		}
		return out
	}
	first, second := positions(7), positions(7)
	if len(first) != len(second) {
		t.Fatalf("%d bodies, then %d", len(first), len(second))
	}
	for uid, cell := range first {
		if second[uid] != cell {
			t.Errorf("body %d at %v, then at %v", uid, cell, second[uid])
		}
	}
}

func TestKeepBonusStopsSwitchingBetweenEqualBehaviors(t *testing.T) {
	switches := func(bonus float64) int {
		f := newAgentField(t, 3, func(bundle *content.Bundle) {
			bundle.Agents.KeepBonus = bonus
			contenttest.Edit(bundle.Templates, drinker, func(template *content.Template) {
				template.Behaviors = []api.ContentID{wander, idle}
			})
		})
		body := f.drinker()
		n := 0
		last := f.doing(body)
		for range 60 {
			f.wait()
			now := f.doing(body)
			if last != "" && now != last {
				n++
			}
			last = now
		}
		return n
	}
	if n := switches(0); n == 0 {
		t.Fatal("without a keep bonus the drinker never switched; the test shows nothing")
	}
	if n := switches(0.1); n != 0 {
		t.Errorf("with a keep bonus the drinker switched %d times, want 0", n)
	}
}

func TestAMoveTowardABodyStopsNextToIt(t *testing.T) {
	f := newAgentField(t, 1, func(bundle *content.Bundle) {
		bundle.Behaviors["behavior.test.to_person"] = content.BehaviorDef{
			Name: "text.behavior.test.to_person", Do: "move_toward",
			Target: &content.TargetDef{Kind: content.BehaviorTargetBody, Tags: []string{"graftable"}},
		}
		bundle.Text["text.behavior.test.to_person"] = "going to a person"
		contenttest.Edit(bundle.Templates, drinker, func(template *content.Template) {
			template.Behaviors = []api.ContentID{"behavior.test.to_person"}
		})
	})
	body := f.drinker()
	nearest := func() int {
		best := -1
		query := ecs.NewFilter2[comp.Pos, comp.Identity](f.w.ECS).Query()
		for query.Next() {
			pos, identity := query.Get()
			if identity.Template == drinker || pos.Map != ecs.NewMap[comp.Pos](f.w.ECS).Get(body).Map {
				continue
			}
			if d := hexmap.CellDistance(f.at(body), pos.At); best < 0 || d < best {
				best = d
			}
		}
		return best
	}
	for f.w.Now < 1000 && nearest() > 1 {
		f.wait()
	}
	if d := nearest(); d > 1 {
		t.Fatalf("at tick %d the drinker is %d from the nearest person, want next to one", f.w.Now, d)
	}
	for range 20 {
		if !ecs.NewMap[comp.Doing](f.w.ECS).Has(body) {
			break
		}
		f.wait()
	}
	if d := nearest(); d > 1 {
		t.Errorf("the drinker moved %d away from the person it reached", d)
	}
	if ecs.NewMap[comp.Doing](f.w.ECS).Has(body) {
		t.Error("the drinker still holds its finished behavior")
	}
}

func TestABehaviorWithNoTargetIsNotChosen(t *testing.T) {
	f := newAgentField(t, 1, func(bundle *content.Bundle) {
		bundle.Behaviors["behavior.test.to_foe"] = content.BehaviorDef{
			Name: "text.behavior.test.to_foe", Do: "move_toward", Consider: []string{"2"},
			Target: &content.TargetDef{Kind: content.BehaviorTargetBody, Hostile: new(true)},
		}
		bundle.Text["text.behavior.test.to_foe"] = "going for a foe"
		contenttest.Edit(bundle.Templates, drinker, func(template *content.Template) {
			template.Behaviors = []api.ContentID{"behavior.test.to_foe", idle}
		})
	})
	body := f.drinker()
	for range 10 {
		f.wait()
		if got := f.doing(body); got != idle {
			t.Fatalf("at tick %d the drinker is doing %q, want %q", f.w.Now, got, idle)
		}
	}
}

func TestDoingSurvivesASave(t *testing.T) {
	f := newAgentField(t, 1, nil)
	f.waitUntil(30)
	body := f.drinker()
	want := *ecs.NewMap[comp.Doing](f.w.ECS).Get(body)
	wantHash, err := f.g.Hash()
	if err != nil {
		t.Fatal(err)
	}
	loaded, world, _ := gametest.Reload(t, f.cfg, f.g)
	gotHash, err := loaded.Hash()
	if err != nil {
		t.Fatal(err)
	}
	if gotHash != wantHash {
		t.Error("the loaded game hashes differently")
	}
	f.g, f.w = loaded, world
	if got := *ecs.NewMap[comp.Doing](f.w.ECS).Get(f.drinker()); got != want {
		t.Errorf("doing after a reload = %+v, want %+v", got, want)
	}
}
