package agents_test

import (
	"slices"
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/agents"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
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

var sampleFamilies = []api.ContentID{
	"family.moving_clay", "family.remnant", "family.soul_feeder", "family.archon_wardens",
}

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

func (s *sampleWorld) counts(name string) map[api.ContentID]int {
	s.t.Helper()
	region, ok := agents.RegionOf(s.w, mapNamed(s.t, s.w, name))
	if !ok {
		s.t.Fatalf("no region for %q", name)
	}
	counts := map[api.ContentID]int{}
	for _, p := range ecs.NewMap[comp.Populations](s.w.ECS).Get(region).Of {
		counts[p.Family] = p.Count
	}
	return counts
}

func (s *sampleWorld) allCounts() map[string]map[api.ContentID]int {
	s.t.Helper()
	all := map[string]map[api.ContentID]int{}
	for _, name := range []string{sampleShallow, sampleDeepA, sampleDeepB, sampleLake} {
		all[name] = s.counts(name)
	}
	return all
}

func (s *sampleWorld) enter(name string) {
	s.t.Helper()
	party, ok := anatomy.ControlledParty(s.w)
	if !ok {
		s.t.Fatal("no controlled party")
	}
	to := mapNamed(s.t, s.w, name)
	positions := ecs.NewMap[comp.Pos](s.w.ECS)
	members := anatomy.Members(s.w, party)
	from := positions.Get(members[0]).Map
	var at api.Cell
	for _, gate := range s.w.Maps.Get(to).Gates() {
		at = gate.At
	}
	for _, member := range members {
		*positions.Get(member) = comp.Pos{Map: to, At: at}
	}
	s.w.Bus.Publish(api.PartyCrossed{Party: s.w.UIDOf(party), From: from, To: to})
	if _, err := s.w.Bus.Drain(s.w); err != nil {
		s.t.Fatal(err)
	}
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

func TestSampleWorldPopulationsDriftBetweenRegions(t *testing.T) {
	s := newSampleWorld(t, 1)
	before := s.allCounts()
	s.waitPhases(10 * phasesPerDay)
	after := s.allCounts()
	moved := false
	for name, counts := range after {
		for _, family := range sampleFamilies {
			moved = moved || counts[family] != before[name][family]
		}
	}
	if !moved {
		t.Fatal("no monster population changed in ten days")
	}
	if after[sampleDeepB]["family.remnant"] >= before[sampleDeepB]["family.remnant"] {
		t.Errorf("remnants in the lower deep: %d before, %d after, want fewer",
			before[sampleDeepB]["family.remnant"], after[sampleDeepB]["family.remnant"])
	}
}

func TestEnteringTheSampleDeepPromotesItsFamilies(t *testing.T) {
	s := newSampleWorld(t, 1)
	s.waitPhases(2 * phasesPerDay)
	want := s.counts(sampleDeepA)
	total := 0
	for _, count := range want {
		total += count
	}
	if total == 0 {
		t.Fatal("the upper deep is empty before the party arrives")
	}
	s.events = nil
	s.enter(sampleDeepA)
	var promoted api.Promoted
	for _, e := range s.events {
		if p, ok := e.(api.Promoted); ok && p.Map == mapNamed(t, s.w, sampleDeepA) {
			promoted = p
		}
	}
	got := map[api.ContentID]int{}
	for _, fc := range promoted.Families {
		got[fc.Family] = fc.Count
	}
	for family, count := range want {
		if count > 0 && got[family] != count {
			t.Errorf("promoted %d of %s, want %d", got[family], family, count)
		}
	}
	bodies := 0
	query := ecs.NewFilter2[comp.Pos, comp.Family](s.w.ECS).Query()
	deepA := mapNamed(t, s.w, sampleDeepA)
	for query.Next() {
		if pos, _ := query.Get(); pos.Map == deepA {
			bodies++
		}
	}
	if bodies != total {
		t.Errorf("%d family bodies in the upper deep, want %d", bodies, total)
	}
	if slices.ContainsFunc(sampleFamilies, func(f api.ContentID) bool { return s.counts(sampleDeepA)[f] != 0 }) {
		t.Error("counts remain after promotion")
	}
}
