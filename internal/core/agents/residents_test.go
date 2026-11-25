package agents_test

import (
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/agents"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/clock"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game/gametest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/inventory"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

const (
	townMap       = "map.test.town"
	hauler        = "template.resident.hauler"
	plainResident = "template.resident.plain"
	wellTemplate  = "template.town.well"
)

type townField struct {
	t *testing.T
	g *game.Game
	w *sim.World
}

func newTownField(t *testing.T, seed uint64, start api.ContentID, mutate func(*content.Bundle)) *townField {
	t.Helper()
	g, w, _ := gametest.New(t, seed, func(bundle *content.Bundle) {
		bundle.Start.Map = start
		if mutate != nil {
			mutate(bundle)
		}
	})
	return &townField{t: t, g: g, w: w}
}

func (f *townField) wait() []api.Event {
	f.t.Helper()
	events, err := f.g.Apply(api.Wait{Source: f.g.Current(), Until: api.WaitEvent})
	if err != nil {
		f.t.Fatal(err)
	}
	return events
}

func (f *townField) phase() string {
	calendar := f.w.Content.Calendar()
	return string(calendar.Phases[clock.New(calendar).Phase(f.w.Now)].ID)
}

func (f *townField) withTemplate(template api.ContentID) []ecs.Entity {
	var found []ecs.Entity
	query := ecs.NewFilter1[comp.Identity](f.w.ECS).Query()
	for query.Next() {
		if query.Get().Template == template {
			found = append(found, query.Entity())
		}
	}
	return f.w.SortByUID(found)
}

func (f *townField) one(template api.ContentID) ecs.Entity {
	f.t.Helper()
	found := f.withTemplate(template)
	if len(found) != 1 {
		f.t.Fatalf("found %d bodies of %s, want 1", len(found), template)
	}
	return found[0]
}

func (f *townField) town(name string) ecs.Entity {
	f.t.Helper()
	for _, id := range f.w.Maps.IDs() {
		if f.w.Maps.Get(id).Name() == name {
			town, ok := agents.TownOf(f.w, id)
			if !ok {
				f.t.Fatalf("%s is not a town", name)
			}
			return town
		}
	}
	f.t.Fatalf("no map %q", name)
	return ecs.Entity{}
}

func (f *townField) mapID(name string) api.MapID {
	f.t.Helper()
	for _, id := range f.w.Maps.IDs() {
		if f.w.Maps.Get(id).Name() == name {
			return id
		}
	}
	f.t.Fatalf("no map %q", name)
	return 0
}

func thirstyQuickly(bundle *content.Bundle) {
	contenttest.Edit(bundle.Needs, "need.water", func(need *content.NeedDef) { need.Rise = "0.1" })
	contenttest.Edit(bundle.Maps, townMap, func(town *content.MapDef) {
		town.Region.Town.Stock = []content.StackDef{{Item: waterItem, Count: 3}}
	})
}

func TestAHaulerFillsTheCisternByDayAndRestsAtNight(t *testing.T) {
	f := newTownField(t, 1, townMap, nil)
	town := f.town(townMap)
	well := f.one(wellTemplate)
	haulerBody := f.one(hauler)
	wellStart := inventory.Count(f.w, well, waterItem, 0)
	stockStart := inventory.Count(f.w, town, waterItem, 0)
	cistern, _ := f.w.Maps.Get(f.mapID(townMap)).Mark("cistern")
	positions := ecs.NewMap[comp.Pos](f.w.ECS)
	wellCell := positions.Get(well).At

	lastWell := wellStart
	visitedWell, visitedCistern := false, false
	delivered := false
	for f.w.Now < 240 {
		f.wait()
		at := positions.Get(haulerBody).At
		visitedWell = visitedWell || hexmap.CellDistance(at, wellCell) <= 1
		visitedCistern = visitedCistern || (visitedWell && hexmap.CellDistance(at, cistern) <= 1)
		now := inventory.Count(f.w, well, waterItem, 0)
		if now < lastWell && f.phase() == "phase.night" {
			t.Fatalf("the hauler drew from the well at tick %d, in the night", f.w.Now)
		}
		lastWell = now
		delivered = delivered || inventory.Count(f.w, town, waterItem, 0) > stockStart
	}
	if !visitedWell || !visitedCistern {
		t.Errorf("hauler visited the well: %v, then the cistern: %v", visitedWell, visitedCistern)
	}
	if !delivered {
		t.Errorf("the stock never rose above %d", stockStart)
	}
	if lastWell >= wellStart {
		t.Errorf("the well still holds %d, from %d", lastWell, wellStart)
	}
}

func TestResidentsLeaveThroughTheGateWhenWaterRunsOut(t *testing.T) {
	f := newTownField(t, 1, townMap, thirstyQuickly)
	well := f.one(wellTemplate)
	inventory.Take(f.w, well, waterItem, 0, inventory.Count(f.w, well, waterItem, 0))
	hub := f.town("map.play.hub")
	populations := ecs.NewMap[comp.Populations](f.w.ECS)
	countPeople := func() int {
		total := 0
		for _, p := range populations.Get(hub).Of {
			if p.Family == "family.people" || p.Family == "family.people.haulers" {
				total += p.Count
			}
		}
		return total
	}
	residents := func() int { return len(f.withTemplate(hauler)) + len(f.withTemplate(plainResident)) }
	if residents() != 3 {
		t.Fatalf("%d residents at the start, want 3", residents())
	}
	var migrated []api.Migrated
	for range 3000 {
		if residents() == 0 {
			break
		}
		for _, e := range f.wait() {
			if m, ok := e.(api.Migrated); ok {
				migrated = append(migrated, m)
			}
		}
	}
	if residents() != 0 {
		t.Fatalf("%d residents still in town at tick %d", residents(), f.w.Now)
	}
	if got := countPeople(); got != 3 {
		t.Errorf("the hub counts %d people, want 3", got)
	}
	from, to := f.mapID(townMap), f.mapID("map.play.hub")
	if len(migrated) != 3 {
		t.Errorf("%d Migrated events, want 3", len(migrated))
	}
	for _, m := range migrated {
		if m.From != from || m.To != to || m.Count != 1 || m.Short != "" {
			t.Errorf("migrated %+v, want one resident from %d to %d", m, from, to)
		}
	}
}

func TestResidentsWalkOntoThePartysMapWhenTheyLeaveToIt(t *testing.T) {
	f := newTownField(t, 1, "map.play.hub", thirstyQuickly)
	well := f.one(wellTemplate)
	inventory.Take(f.w, well, waterItem, 0, inventory.Count(f.w, well, waterItem, 0))
	hub := f.mapID("map.play.hub")
	town := f.mapID(townMap)
	for i, template := range []api.ContentID{hauler, plainResident, plainResident} {
		if _, err := f.w.Spawn(template, &comp.Pos{Map: town, At: cellAt(3+i, 3)}); err != nil {
			t.Fatal(err)
		}
	}
	onHub := func() int {
		count := 0
		for _, template := range []api.ContentID{hauler, plainResident} {
			for _, body := range f.withTemplate(template) {
				if ecs.NewMap[comp.Pos](f.w.ECS).Get(body).Map == hub {
					count++
				}
			}
		}
		return count
	}
	for range 3000 {
		if onHub() == 3 {
			break
		}
		f.wait()
	}
	if got := onHub(); got != 3 {
		t.Errorf("%d residents reached the hub, want 3", got)
	}
}

func TestATownDayReplaysToTheSameHash(t *testing.T) {
	run := func() [32]byte {
		f := newTownField(t, 7, townMap, nil)
		for f.w.Now < 300 {
			f.wait()
		}
		hash, err := f.g.Hash()
		if err != nil {
			t.Fatal(err)
		}
		return hash
	}
	if a, b := run(), run(); a != b {
		t.Errorf("hashes differ: %x and %x", a, b)
	}
}
