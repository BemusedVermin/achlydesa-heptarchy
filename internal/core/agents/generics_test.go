package agents_test

import (
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

const (
	guardTemplate = "template.generic.guard"
	guardFamily   = "family.people.guards"
	guardTicks    = 40
)

func (f *townField) livingGuards() []ecs.Entity {
	dead := ecs.NewMap[comp.Dead](f.w.ECS)
	var alive []ecs.Entity
	for _, body := range f.withTemplate(guardTemplate) {
		if !dead.Has(body) {
			alive = append(alive, body)
		}
	}
	return alive
}

func (f *townField) countedGuards() int {
	populations := ecs.NewMap[comp.Populations](f.w.ECS).Get(f.town(townMap))
	for _, p := range populations.Of {
		if p.Family == guardFamily {
			return p.Count
		}
	}
	return 0
}

func onlyGuardsAreResidents(bundle *content.Bundle) {
	contenttest.Edit(bundle.Maps, townMap, func(town *content.MapDef) {
		town.Region.Town.Residents = guardFamily
		town.Region.Town.Generics = town.Region.Town.Generics[:1]
		town.Region.Town.Generics[0].Count = 1
		town.Region.Populations[guardFamily] = 1
	})
}

func TestAKilledGuardIsReplacedAfterTheContentTicks(t *testing.T) {
	f := newTownField(t, 1, townMap, nil)
	guards := f.livingGuards()
	if len(guards) != 2 {
		t.Fatalf("%d guards at the start, want 2", len(guards))
	}
	killedAt := f.w.Now
	sim.Put(f.w, guards[0], comp.Dead{Tick: killedAt})
	for len(f.livingGuards()) < 2 && f.w.Now < killedAt+2000 {
		f.wait()
	}
	if got := len(f.livingGuards()); got != 2 {
		t.Fatalf("%d guards alive at tick %d, want 2", got, f.w.Now)
	}
	if elapsed := f.w.Now - killedAt; elapsed < guardTicks {
		t.Errorf("the replacement came after %d ticks, want at least %d", elapsed, guardTicks)
	}
}

func TestAGuardLostWhileTheTownIsOffScreenIsReplacedInTheCount(t *testing.T) {
	f := newTownField(t, 1, "map.play.hub", nil)
	populations := ecs.NewMap[comp.Populations](f.w.ECS).Get(f.town(townMap))
	for i := range populations.Of {
		if populations.Of[i].Family == guardFamily {
			populations.Of[i].Count--
		}
	}
	lostAt := f.w.Now
	for f.countedGuards() < 2 && f.w.Now < lostAt+2000 {
		f.wait()
	}
	if got := f.countedGuards(); got != 2 {
		t.Fatalf("the town counts %d guards at tick %d, want 2", got, f.w.Now)
	}
	if elapsed := f.w.Now - lostAt; elapsed < guardTicks {
		t.Errorf("the replacement came after %d ticks, want at least %d", elapsed, guardTicks)
	}
}

func TestAnEmptiedTownReplacesNobody(t *testing.T) {
	f := newTownField(t, 1, townMap, onlyGuardsAreResidents)
	guards := f.livingGuards()
	if len(guards) != 1 {
		t.Fatalf("%d guards at the start, want 1", len(guards))
	}
	killedAt := f.w.Now
	sim.Put(f.w, guards[0], comp.Dead{Tick: killedAt})
	for f.w.Now < killedAt+10*guardTicks {
		f.wait()
	}
	if got := len(f.livingGuards()); got != 0 {
		t.Errorf("%d guards alive at tick %d, want none", got, f.w.Now)
	}
	if got := f.countedGuards(); got != 0 {
		t.Errorf("the town counts %d guards, want none", got)
	}
	replacements := ecs.NewMap[comp.Town](f.w.ECS).Get(f.town(townMap)).Replacements
	if len(replacements) != 0 {
		t.Errorf("%d replacements queued in an emptied town", len(replacements))
	}
}

func TestGenericsNeedAFamilyOfTheirOwnAndTicks(t *testing.T) {
	bundle := contenttest.SampleBundle(t)
	contenttest.Edit(bundle.Maps, townMap, func(town *content.MapDef) {
		town.Region.Town.Generics = []content.GenericDef{{Template: plainResident, Count: 1}}
	})
	if _, err := content.New(bundle); err == nil {
		t.Error("a generic with no ticks passed validation")
	}
}
