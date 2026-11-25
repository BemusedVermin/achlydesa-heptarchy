package agents_test

import (
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/agents"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game/gametest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

const (
	hubMap     = "map.play.hub"
	playHall   = "map.play.hall"
	clayFamily = "family.test.clay"
)

type lodField struct {
	t      *testing.T
	w      *sim.World
	hub    api.MapID
	hall   api.MapID
	leader ecs.Entity
}

func newLodField(t *testing.T, seed uint64, mutate func(*content.Bundle)) *lodField {
	t.Helper()
	_, w, _ := gametest.New(t, seed, func(bundle *content.Bundle) {
		bundle.Start.Map = hubMap
		if mutate != nil {
			mutate(bundle)
		}
	})
	party, ok := anatomy.ControlledParty(w)
	if !ok {
		t.Fatal("no controlled party")
	}
	leader, ok := w.ByUID(ecs.NewMap[comp.Party](w.ECS).Get(party).Leader)
	if !ok {
		t.Fatal("no leader")
	}
	return &lodField{t: t, w: w, hub: mapNamed(t, w, hubMap), hall: mapNamed(t, w, playHall), leader: leader}
}

func (f *lodField) crossTo(to api.MapID) {
	f.t.Helper()
	pos := ecs.NewMap[comp.Pos](f.w.ECS).Get(f.leader)
	from := pos.Map
	m := f.w.Maps.Get(to)
	var at api.Cell
	for _, gate := range m.Gates() {
		at = gate.At
	}
	*pos = comp.Pos{Map: to, At: at}
	party, _ := anatomy.ControlledParty(f.w)
	f.w.Bus.Publish(api.PartyCrossed{Party: f.w.UIDOf(party), From: from, To: to})
	if _, err := f.w.Bus.Drain(f.w); err != nil {
		f.t.Fatal(err)
	}
}

func (f *lodField) familyBodies(on api.MapID) []ecs.Entity {
	var out []ecs.Entity
	query := ecs.NewFilter2[comp.Pos, comp.Family](f.w.ECS).Query()
	for query.Next() {
		if pos, _ := query.Get(); pos.Map == on {
			out = append(out, query.Entity())
		}
	}
	return f.w.SortByUID(out)
}

func (f *lodField) count(on api.MapID, family api.ContentID) int {
	f.t.Helper()
	region, ok := agents.RegionOf(f.w, on)
	if !ok {
		f.t.Fatalf("map %d has no region", on)
	}
	for _, p := range ecs.NewMap[comp.Populations](f.w.ECS).Get(region).Of {
		if p.Family == family {
			return p.Count
		}
	}
	return 0
}

func TestEnteringAMapPromotesItsPopulationAndLeavingCountsItBack(t *testing.T) {
	f := newLodField(t, 1, nil)
	if got := f.count(f.hall, clayFamily); got != 3 {
		t.Fatalf("the hall starts with %d, want 3", got)
	}
	if got := len(f.familyBodies(f.hall)); got != 0 {
		t.Fatalf("%d family bodies in the hall before the party arrives", got)
	}
	f.crossTo(f.hall)
	if got := len(f.familyBodies(f.hall)); got != 3 {
		t.Errorf("%d family bodies in the hall, want 3", got)
	}
	if got := f.count(f.hall, clayFamily); got != 0 {
		t.Errorf("the count is %d while promoted, want 0", got)
	}
	f.crossTo(f.hub)
	if got := len(f.familyBodies(f.hall)); got != 0 {
		t.Errorf("%d family bodies left in the hall, want 0", got)
	}
	if got := f.count(f.hall, clayFamily); got != 3 {
		t.Errorf("the count is %d after leaving, want 3", got)
	}
	f.crossTo(f.hall)
	if got := len(f.familyBodies(f.hall)); got != 3 {
		t.Errorf("%d family bodies on the second visit, want 3", got)
	}
}

func TestAPromotedBodyActsAndStandsOutOfTheParitesSight(t *testing.T) {
	f := newLodField(t, 1, nil)
	f.crossTo(f.hall)
	actors := ecs.NewMap[comp.Actor](f.w.ECS)
	positions := ecs.NewMap[comp.Pos](f.w.ECS)
	m := f.w.Maps.Get(f.hall)
	cfg := f.w.Content.Party()
	seen, err := m.Visible(positions.Get(f.leader).At, cfg.SightRadius, cfg.SightDepth)
	if err != nil {
		t.Fatal(err)
	}
	visible := map[api.Cell]bool{}
	for _, cell := range seen {
		visible[cell] = true
	}
	for _, body := range f.familyBodies(f.hall) {
		if !actors.Has(body) {
			t.Error("a promoted body is not on the timeline")
		}
		if at := positions.Get(body).At; visible[at] {
			t.Errorf("a promoted body stands at %v in the party's sight", at)
		}
	}
}

func TestAnIndividualStaysOffTheTimelineUntilItsMapIsEnteredAgain(t *testing.T) {
	f := newLodField(t, 1, nil)
	f.crossTo(f.hall)
	bodies := f.familyBodies(f.hall)
	chosen := bodies[0]
	sim.Put(f.w, chosen, comp.Individual{})
	at := *ecs.NewMap[comp.Pos](f.w.ECS).Get(chosen)
	f.crossTo(f.hub)
	if !f.w.Alive(chosen) {
		t.Fatal("the individual was removed")
	}
	if ecs.NewMap[comp.Actor](f.w.ECS).Has(chosen) {
		t.Error("the individual is still on the timeline")
	}
	if got := *ecs.NewMap[comp.Pos](f.w.ECS).Get(chosen); got != at {
		t.Errorf("the individual moved to %v, want %v", got, at)
	}
	if got := f.count(f.hall, clayFamily); got != 2 {
		t.Errorf("the count is %d, want the 2 that were not individuals", got)
	}
	f.crossTo(f.hall)
	if !ecs.NewMap[comp.Actor](f.w.ECS).Has(chosen) {
		t.Error("the individual is not back on the timeline")
	}
	if got := len(f.familyBodies(f.hall)); got != 3 {
		t.Errorf("%d family bodies after returning, want 3", got)
	}
}

func TestDeadBodiesStayWhereTheyFell(t *testing.T) {
	f := newLodField(t, 1, nil)
	f.crossTo(f.hall)
	body := f.familyBodies(f.hall)[0]
	sim.Put(f.w, body, comp.Dead{})
	f.crossTo(f.hub)
	if !f.w.Alive(body) {
		t.Error("the dead body was removed")
	}
	if got := f.count(f.hall, clayFamily); got != 2 {
		t.Errorf("the count is %d, want 2", got)
	}
}

type placedBody struct {
	template api.ContentID
	at       api.Cell
}

func promotedBodies(f *lodField) []placedBody {
	var out []placedBody
	for _, body := range f.familyBodies(f.hall) {
		out = append(out, placedBody{f.w.TemplateOf(body), ecs.NewMap[comp.Pos](f.w.ECS).Get(body).At})
	}
	return out
}

func mixedHall(bundle *content.Bundle) {
	contenttest.Edit(bundle.Maps, playHall, func(hall *content.MapDef) {
		hall.Region = &content.RegionDef{Populations: map[string]int{"family.test.mixed": 8}}
	})
}

func TestTheSameSeedPromotesTheSameTemplatesIntoTheSameCells(t *testing.T) {
	run := func(seed uint64) []placedBody {
		f := newLodField(t, seed, mixedHall)
		f.crossTo(f.hall)
		return promotedBodies(f)
	}
	first, second := run(7), run(7)
	if len(first) != 8 {
		t.Fatalf("%d bodies, want 8", len(first))
	}
	for i := range first {
		if first[i] != second[i] {
			t.Errorf("body %d: %v then %v", i, first[i], second[i])
		}
	}
	other := run(8)
	same := true
	for i := range first {
		same = same && first[i] == other[i]
	}
	if same {
		t.Error("two seeds promoted identical bodies")
	}
}

func TestPromoteDemotePromoteGivesTheSameCounts(t *testing.T) {
	f := newLodField(t, 3, mixedHall)
	f.crossTo(f.hall)
	f.crossTo(f.hub)
	if got := f.count(f.hall, "family.test.mixed"); got != 8 {
		t.Errorf("the count is %d, want 8", got)
	}
	f.crossTo(f.hall)
	if got := len(f.familyBodies(f.hall)); got != 8 {
		t.Errorf("%d bodies, want 8", got)
	}
}

func TestTheStartingMapIsPromotedAtGameStart(t *testing.T) {
	_, w, _ := gametest.New(t, 1, func(bundle *content.Bundle) {
		bundle.Start.Map, bundle.Start.PartyAtMark = playHall, "from_hub"
	})
	query := ecs.NewFilter1[comp.Family](w.ECS).Query()
	count := 0
	for query.Next() {
		count++
	}
	if count != 3 {
		t.Errorf("%d family bodies at the start, want 3", count)
	}
}

func TestAFamilyAppearsNextToItsMarks(t *testing.T) {
	f := newLodField(t, 1, func(bundle *content.Bundle) {
		mixedHall(bundle)
		contenttest.Edit(bundle.Maps, playHall, func(hall *content.MapDef) {
			hall.Marks["l"] = "lair"
			hall.Layers = append([]content.LayerDef(nil), hall.Layers...)
			layer := &hall.Layers[0]
			layer.Rows = append([]string(nil), layer.Rows...)
			row := []rune(layer.Rows[12])
			row[3] = 'l'
			layer.Rows[12] = string(row)
		})
	})
	mark, ok := f.w.Maps.Get(f.hall).Mark("lair")
	if !ok {
		t.Fatal("the hall has no lair mark")
	}
	f.crossTo(f.hall)
	for _, body := range promotedBodies(f) {
		if body.at.Hex.Q < mark.Hex.Q-4 || body.at.Hex.Q > mark.Hex.Q+4 {
			t.Errorf("body at %v is far from the lair at %v", body.at, mark)
		}
	}
}
