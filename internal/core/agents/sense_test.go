package agents

import (
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/knowledge"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

var wall = api.Hex{Q: 2, R: 0}

func sightWorld(tb testing.TB) *sim.World {
	tb.Helper()
	w := sim.NewWorld(1)
	w.Content = contenttest.Sample(tb)
	w.Map = hexmap.NewMap(hexmap.FuncSource(func(c api.Cell) hexmap.Terrain {
		if c.Hex == wall {
			return hexmap.Terrain{ID: "terrain.wall", Solid: true, Opaque: true}
		}
		return hexmap.Terrain{ID: "terrain.air"}
	}), hexmap.Prisms{})
	return w
}

func spawnBody(w *sim.World, at api.Cell) ecs.Entity {
	body := w.NewIdentified("template.test")
	ecs.NewMap[comp.Pos](w.ECS).Add(body, &comp.Pos{At: at})
	ecs.NewMap[comp.Body](w.ECS).Add(body, &comp.Body{})
	return body
}

func at(q, r, depth int) api.Cell { return api.Cell{Hex: api.Hex{Q: q, R: r}, Depth: depth} }

func TestABodySeesOnlyInsideItsRadiusAndLineOfSight(t *testing.T) {
	w := sightWorld(t)
	cfg := w.Content.Agents()
	radius := cfg.SightRadius
	looker := spawnBody(w, at(0, 0, 0))
	near := spawnBody(w, at(0, 3, 0))
	edge := spawnBody(w, at(-radius, 0, 0))
	beyond := spawnBody(w, at(-radius-1, 0, 0))
	behindWall := spawnBody(w, at(4, 0, 0))
	tooDeep := spawnBody(w, at(0, 1, cfg.SightDepth+1))
	corpse := spawnBody(w, at(0, -2, 0))
	ecs.NewMap[comp.Dead](w.ECS).Add(corpse, &comp.Dead{})

	if err := sense(w, looker); err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name       string
		body       ecs.Entity
		where      api.Cell
		seen, dead bool
	}{
		{"near", near, at(0, 3, 0), true, false},
		{"at the edge of the radius", edge, at(-radius, 0, 0), true, false},
		{"beyond the radius", beyond, api.Cell{}, false, false},
		{"behind a wall", behindWall, api.Cell{}, false, false},
		{"deeper than sight", tooDeep, api.Cell{}, false, false},
		{"dead", corpse, at(0, -2, 0), true, true},
	} {
		uid := w.UIDOf(test.body)
		fact, seen := knowledge.Best(w, looker, uid, cfg.AtClaim)
		if seen != test.seen || (seen && fact.Value != knowledge.CellValue(test.where)) {
			t.Errorf("%s: at fact = %+v, %v; want seen %v at %v", test.name, fact, seen, test.seen, test.where)
		}
		if _, dead := knowledge.Best(w, looker, uid, cfg.DeadClaim); dead != test.dead {
			t.Errorf("%s: dead fact known = %v, want %v", test.name, dead, test.dead)
		}
	}
	if got := knowledge.About(w, looker, w.UIDOf(looker)); len(got) != 0 {
		t.Errorf("the looker learned about itself: %v", got)
	}
}

func TestANewSightingReplacesTheOldAndAnUnseenMoveIsNotKnown(t *testing.T) {
	w := sightWorld(t)
	cfg := w.Content.Agents()
	looker := spawnBody(w, at(0, 0, 0))
	walker := spawnBody(w, at(0, 2, 0))
	uid := w.UIDOf(walker)
	if err := sense(w, looker); err != nil {
		t.Fatal(err)
	}
	positions := ecs.NewMap[comp.Pos](w.ECS)
	positions.Get(walker).At = at(5, 0, 0)
	w.Now = 10
	if err := sense(w, looker); err != nil {
		t.Fatal(err)
	}
	if fact, _ := knowledge.Best(
		w,
		looker,
		uid,
		cfg.AtClaim,
	); fact.Value != knowledge.CellValue(at(0, 2, 0)) ||
		fact.Tick != 0 {
		t.Errorf("after an unseen move the belief is %+v, want the old cell at tick 0", fact)
	}
	positions.Get(walker).At = at(1, 1, 0)
	w.Now = 20
	if err := sense(w, looker); err != nil {
		t.Fatal(err)
	}
	facts := knowledge.About(w, looker, uid)
	if len(facts) != 1 || facts[0].Value != knowledge.CellValue(at(1, 1, 0)) || facts[0].Tick != 20 {
		t.Errorf("after a new sighting the beliefs are %+v, want one at (1,1,0) from tick 20", facts)
	}
}

func TestReSensingAnUnchangedScenePublishesNothing(t *testing.T) {
	w := sightWorld(t)
	looker := spawnBody(w, at(0, 0, 0))
	spawnBody(w, at(0, 2, 0))
	corpse := spawnBody(w, at(0, -2, 0))
	ecs.NewMap[comp.Dead](w.ECS).Add(corpse, &comp.Dead{})
	var learned int
	w.Bus.SubscribeAll("test", func(_ *sim.World, e api.Event) {
		if _, ok := e.(api.FactLearned); ok {
			learned++
		}
	})
	for _, look := range []struct {
		name string
		tick int64
		want int
	}{
		{"first look", 0, 3},
		{"second look at the same scene", 10, 0},
	} {
		learned = 0
		w.Now = look.tick
		if err := sense(w, looker); err != nil {
			t.Fatal(err)
		}
		if _, err := w.Bus.Drain(w); err != nil {
			t.Fatal(err)
		}
		if learned != look.want {
			t.Errorf("%s published %d FactLearned, want %d", look.name, learned, look.want)
		}
	}
}

func BenchmarkThousandAgentsSenseOnce(b *testing.B) {
	w := sightWorld(b)
	var bodies []ecs.Entity
	for i := range 1000 {
		bodies = append(bodies, spawnBody(w, at(i%32*2, i/32*2, i/7%4)))
	}
	b.ResetTimer()
	for b.Loop() {
		for _, body := range bodies {
			if err := sense(w, body); err != nil {
				b.Fatal(err)
			}
			if _, err := w.Bus.Drain(w); err != nil {
				b.Fatal(err)
			}
		}
	}
}
