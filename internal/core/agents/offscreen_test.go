package agents_test

import (
	"slices"
	"testing"

	"github.com/mlange-42/ark/ecs"
	"pgregory.net/rapid"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/agents"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game/gametest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/inventory"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

const (
	denMap     = "map.play.den"
	descentMap = "map.play.descent"
	waterItem  = "item.lake_water"
)

type offscreen struct {
	t    *testing.T
	g    *game.Game
	w    *sim.World
	hub  api.MapID
	den  api.MapID
	hall api.MapID
	desc api.MapID
	town api.MapID
}

// newOffscreen starts the party in the hall, so the hub, den and descent are off screen. The hub
// is a town of clay residents; the den is a town with plenty of water.
func newOffscreen(t *testing.T, seed uint64, mutate func(*content.Bundle)) *offscreen {
	t.Helper()
	g, w, _ := gametest.New(t, seed, func(bundle *content.Bundle) {
		bundle.Start.Map = playHall
		bundle.Start.PartyAtMark = "from_hub"
		contenttest.Edit(bundle.Maps, hubMap, func(hub *content.MapDef) {
			hub.Region.Populations = map[string]int{clayFamily: 6}
			hub.Region.Town.Stock = []content.StackDef{{Item: waterItem, Count: 2}}
			hub.Region.Town.Residents = clayFamily
			hub.Region.Town.Use = []content.StackDef{{Item: waterItem, Count: 1}}
			hub.Region.Town.Exodus = 0.5
		})
		contenttest.Edit(bundle.Maps, denMap, func(den *content.MapDef) {
			den.Region = &content.RegionDef{Town: &content.TownDef{
				Stock: []content.StackDef{{Item: waterItem, Count: 100}},
			}}
		})
		contenttest.Edit(bundle.Maps, playHall, func(hall *content.MapDef) { hall.Region.Populations = nil })
		if mutate != nil {
			mutate(bundle)
		}
	})
	return &offscreen{
		t: t, g: g, w: w, hub: mapNamed(t, w, hubMap), den: mapNamed(t, w, denMap),
		hall: mapNamed(t, w, playHall), desc: mapNamed(t, w, descentMap), town: mapNamed(t, w, townMap),
	}
}

func (o *offscreen) region(on api.MapID) ecs.Entity {
	o.t.Helper()
	region, ok := agents.RegionOf(o.w, on)
	if !ok {
		o.t.Fatalf("map %d has no region", on)
	}
	return region
}

func (o *offscreen) count(on api.MapID) int {
	o.t.Helper()
	for _, p := range ecs.NewMap[comp.Populations](o.w.ECS).Get(o.region(on)).Of {
		if p.Family == clayFamily {
			return p.Count
		}
	}
	return 0
}

func (o *offscreen) set(on api.MapID, count int) {
	o.t.Helper()
	populations := ecs.NewMap[comp.Populations](o.w.ECS).Get(o.region(on))
	populations.Of = []comp.Population{{Family: clayFamily, Count: count}}
}

func (o *offscreen) total() int {
	return o.count(o.hub) + o.count(o.den) + o.count(o.desc) + o.count(o.hall) + o.count(o.town)
}

func (o *offscreen) phase() []api.Event {
	o.t.Helper()
	events, err := o.g.Apply(api.Wait{Source: o.g.Current(), Until: api.WaitPhase})
	if err != nil {
		o.t.Fatal(err)
	}
	return events
}

func TestDriftNeverChangesTheTotalCount(t *testing.T) {
	o := newOffscreen(t, 1, func(bundle *content.Bundle) {
		contenttest.Edit(bundle.Families, clayFamily, func(f *content.FamilyDef) { f.Drift = "0.4" })
		contenttest.Edit(bundle.Maps, hubMap, func(hub *content.MapDef) { hub.Region.Town.Exodus = 0 })
		contenttest.Edit(bundle.Maps, hubMap, func(hub *content.MapDef) { hub.Region.Town.Use = nil })
	})
	rapid.Check(t, func(rt *rapid.T) {
		counts := rapid.SliceOfN(rapid.IntRange(0, 40), 3, 3).Draw(rt, "counts")
		o.set(o.hub, counts[0])
		o.set(o.den, counts[1])
		o.set(o.desc, counts[2])
		before := o.total()
		for range rapid.IntRange(1, 4).Draw(rt, "phases") {
			o.phase()
		}
		if got := o.total(); got != before {
			rt.Fatalf("total %d after drifting, want %d", got, before)
		}
	})
}

func TestATownWithNoWaterRefillLosesResidentsThenEmptiesOnce(t *testing.T) {
	o := newOffscreen(t, 1, nil)
	var migrated []api.Migrated
	emptied := 0
	for range 40 {
		for _, e := range o.phase() {
			switch e := e.(type) {
			case api.Migrated:
				migrated = append(migrated, e)
			case api.TownEmptied:
				emptied++
			}
		}
	}
	if len(migrated) == 0 {
		t.Fatal("no residents migrated")
	}
	for _, m := range migrated {
		if m.From != o.hub || m.To != o.den || m.Short != waterItem {
			t.Errorf("migrated %+v, want hub to den over water", m)
		}
	}
	if emptied != 1 {
		t.Errorf("town emptied %d times, want 1", emptied)
	}
	if got := o.count(o.hub); got != 0 {
		t.Errorf("%d residents left in the hub, want 0", got)
	}
	if got := o.count(o.den); got != 6 {
		t.Errorf("%d residents in the den, want 6", got)
	}
}

func TestATownThatIsRefilledKeepsItsResidents(t *testing.T) {
	o := newOffscreen(t, 1, func(bundle *content.Bundle) {
		contenttest.Edit(bundle.Maps, hubMap, func(hub *content.MapDef) {
			hub.Region.Town.Refill = []content.StackDef{{Item: waterItem, Count: 6}}
		})
	})
	for range 10 {
		o.phase()
	}
	if got := o.count(o.hub); got != 6 {
		t.Errorf("%d residents after ten phases of enough water, want 6", got)
	}
	if got := inventory.Count(o.w, o.region(o.hub), waterItem, 0); got != 2 {
		t.Errorf("stock holds %d water, want 2", got)
	}
}

func TestDriftFavoursRegionsWithDeadBodies(t *testing.T) {
	o := newOffscreen(t, 1, func(bundle *content.Bundle) {
		contenttest.Edit(bundle.Families, clayFamily, func(f *content.FamilyDef) {
			f.Drift = "neighbour.dead > region.dead ? 0.5 : 0"
		})
		contenttest.Edit(bundle.Maps, hubMap, func(hub *content.MapDef) { hub.Region.Town.Use = nil })
	})
	ecs.NewMap2[comp.Dead, comp.Pos](o.w.ECS).NewEntity(&comp.Dead{}, &comp.Pos{Map: o.den})
	for range 6 {
		o.phase()
	}
	if got := o.count(o.desc); got != 0 {
		t.Errorf("%d clay in the descent, want 0", got)
	}
	if got := o.count(o.den); got != 6 {
		t.Errorf("%d clay in the den, want all 6", got)
	}
}

func TestAttritionKillsAndReports(t *testing.T) {
	o := newOffscreen(t, 1, func(bundle *content.Bundle) {
		contenttest.Edit(bundle.Families, clayFamily, func(f *content.FamilyDef) { f.Attrition = "2" })
		contenttest.Edit(bundle.Maps, hubMap, func(hub *content.MapDef) { hub.Region.Town.Use = nil })
	})
	events := o.phase()
	died := 0
	for _, e := range events {
		if d, ok := e.(api.PopulationDied); ok && d.Map == o.hub {
			died += d.Count
		}
	}
	if died != 2 || o.count(o.hub) != 4 {
		t.Errorf("%d died and %d remain, want 2 and 4", died, o.count(o.hub))
	}
}

func TestTheRegionTheyAreInIsNotTicked(t *testing.T) {
	o := newOffscreen(t, 1, func(bundle *content.Bundle) {
		contenttest.Edit(bundle.Families, clayFamily, func(f *content.FamilyDef) { f.Drift = "0.5"; f.Attrition = "1" })
	})
	o.set(o.hall, 4)
	hubBefore := o.count(o.hub)
	o.phase()
	if got := o.count(o.hall); got != 4 {
		t.Errorf("the party's own region holds %d, want 4", got)
	}
	if o.count(o.hub) == hubBefore {
		t.Error("the hub did not change")
	}
}

func TestTheSameSeedGivesTheSameCounts(t *testing.T) {
	counts := func() []int {
		o := newOffscreen(t, 7, func(bundle *content.Bundle) {
			contenttest.Edit(bundle.Families, clayFamily, func(f *content.FamilyDef) {
				f.Drift = "0.3"
				f.Attrition = "0.5"
			})
			contenttest.Edit(bundle.Maps, hubMap, func(hub *content.MapDef) {
				hub.Region.Town.Refill = []content.StackDef{{Item: waterItem, Count: 2}}
			})
		})
		for range 10 * len(o.w.Content.Calendar().Phases) {
			o.phase()
		}
		return []int{o.count(o.hub), o.count(o.den), o.count(o.desc)}
	}
	first, second := counts(), counts()
	if !slices.Equal(first, second) {
		t.Errorf("counts %v then %v from the same seed", first, second)
	}
}
