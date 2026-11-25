package agents_test

import (
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game/gametest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

const (
	sampleStart   = "start.sample"
	phasesPerDay  = 4
	sampleDeepA   = "map.sample.deep_a"
	sampleDeepB   = "map.sample.deep_b"
	sampleLake    = "map.sample.lake"
	sampleShallow = "map.sample.shallow"
)

type sampleWorld struct {
	t      *testing.T
	g      *game.Game
	w      *sim.World
	events []api.Event
}

func newSampleWorld(t *testing.T, seed uint64) *sampleWorld {
	t.Helper()
	s := &sampleWorld{t: t}
	s.g, s.w, _ = gametest.New(t, seed, func(bundle *content.Bundle) {
		bundle.Start = bundle.Starts[sampleStart]
	})
	s.w.Bus.SubscribeAll("test.record", func(_ *sim.World, e api.Event) { s.events = append(s.events, e) })
	return s
}

func (s *sampleWorld) waitPhases(n int) {
	s.t.Helper()
	for range n {
		events, err := s.g.Apply(api.Wait{Source: s.g.Current(), Until: api.WaitPhase})
		if err != nil {
			s.t.Fatal(err)
		}
		for _, e := range events {
			if rejected, ok := e.(api.Rejected); ok {
				s.t.Fatalf("command rejected: %+v", rejected)
			}
		}
	}
}

func (s *sampleWorld) living(name string) map[api.ContentID][]ecs.Entity {
	s.t.Helper()
	on := mapNamed(s.t, s.w, name)
	found := map[api.ContentID][]ecs.Entity{}
	query := ecs.NewFilter1[comp.Pos](s.w.ECS).Without(ecs.C[comp.Dead]()).Query()
	for query.Next() {
		if query.Get().Map == on {
			template := s.w.TemplateOf(query.Entity())
			found[template] = append(found[template], query.Entity())
		}
	}
	for _, bodies := range found {
		s.w.SortByUID(bodies)
	}
	return found
}

func TestTheSampleWorldRunsTenDaysWithoutErrors(t *testing.T) {
	s := newSampleWorld(t, 1)
	s.waitPhases(10 * phasesPerDay)
	view, err := s.g.View()
	if err != nil {
		t.Fatal(err)
	}
	if view.Clock == nil || view.Clock.Day < 10 {
		t.Errorf("clock after ten days = %+v", view.Clock)
	}
}

func TestTheSampleWorldsBodiesExistFromTickZero(t *testing.T) {
	s := newSampleWorld(t, 1)
	want := map[string]map[api.ContentID]int{
		sampleShallow: {
			"template.resident.plain": 2, "template.resident.hauler": 1,
			"template.generic.guard": 2, "template.generic.field_hand": 1,
		},
		sampleDeepA: {"template.creature.moving_clay": 3, "template.creature.remnant": 2},
		sampleDeepB: {
			"template.creature.remnant": 4, "template.creature.moving_clay": 2,
			"template.creature.soul_feeder": 2, "template.creature.archon_warden": 1,
		},
		sampleLake: {
			"template.resident.plain": 3, "template.generic.tamed_clay": 1,
			"template.generic.leashed_feeder": 1,
		},
	}
	for name, templates := range want {
		got := s.living(name)
		for template, count := range templates {
			if len(got[template]) != count {
				t.Errorf("%s holds %d of %s at the start, want %d", name, len(got[template]), template, count)
			}
		}
	}
}

func TestABodyOnAMapThePartyIsNotOnStillDecides(t *testing.T) {
	s := newSampleWorld(t, 1)
	positions := ecs.NewMap[comp.Pos](s.w.ECS)
	take := func() map[ecs.Entity]comp.Pos {
		out := map[ecs.Entity]comp.Pos{}
		for _, bodies := range s.living(sampleDeepB) {
			for _, body := range bodies {
				out[body] = *positions.Get(body)
			}
		}
		return out
	}
	before := take()
	if len(before) == 0 {
		t.Fatal("no body on the lower deep")
	}
	s.waitPhases(2 * phasesPerDay)
	moved := false
	for body, was := range before {
		moved = moved || !s.w.Alive(body) || *positions.Get(body) != was
	}
	if !moved {
		t.Error("no body on the lower deep moved or died in two days")
	}
}
