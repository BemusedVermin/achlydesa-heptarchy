package agents_test

import (
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/inventory"
)

const residentFamily = "family.people"

func (f *townField) livingResidents(on api.MapID) []ecs.Entity {
	var found []ecs.Entity
	query := ecs.NewFilter2[comp.Family, comp.Pos](f.w.ECS).Without(ecs.C[comp.Dead]()).Query()
	for query.Next() {
		if family, pos := query.Get(); family.ID == residentFamily && pos.Map == on {
			found = append(found, query.Entity())
		}
	}
	return f.w.SortByUID(found)
}

func townEmptiedIn(events []api.Event) []api.TownEmptied {
	var emptied []api.TownEmptied
	for _, e := range events {
		if town, ok := e.(api.TownEmptied); ok {
			emptied = append(emptied, town)
		}
	}
	return emptied
}

func TestATownTheLastResidentWalksOutOfIsEmptiedOnceWhileThePartyIsInIt(t *testing.T) {
	f := newTownField(t, 1, townMap, thirstyQuickly)
	well := f.one(wellTemplate)
	inventory.Take(f.w, well, waterItem, 0, inventory.Count(f.w, well, waterItem, 0))
	town := f.mapID(townMap)
	var emptied []api.TownEmptied
	migrated := 0
	for range 3000 {
		if len(f.livingResidents(town)) == 0 {
			break
		}
		events := f.wait()
		emptied = append(emptied, townEmptiedIn(events)...)
		for _, e := range events {
			if m, ok := e.(api.Migrated); ok && m.From == town && m.Family == residentFamily {
				migrated++
			}
		}
	}
	if migrated == 0 {
		t.Error("no resident left through a gate")
	}
	if left := len(f.livingResidents(town)); left != 0 {
		t.Fatalf("%d residents still in town at tick %d", left, f.w.Now)
	}
	want := []api.TownEmptied{{Map: town, Family: residentFamily}}
	if len(emptied) != 1 || emptied[0] != want[0] {
		t.Errorf("TownEmptied events %+v, want %+v", emptied, want)
	}
}

func TestATownWhoseLastResidentDiesIsEmptiedOnceWhileThePartyIsInIt(t *testing.T) {
	f := newTownField(t, 1, townMap, nil)
	town := f.mapID(townMap)
	residents := f.livingResidents(town)
	if len(residents) < 2 {
		t.Fatalf("%d residents at the start, want at least 2", len(residents))
	}
	for i, body := range residents {
		f.kill(body)
		f.w.Bus.Publish(api.BodyDied{Body: f.w.UIDOf(body), Tick: f.w.Now})
		events, err := f.w.Bus.Drain(f.w)
		if err != nil {
			t.Fatal(err)
		}
		emptied := townEmptiedIn(events)
		if last := i == len(residents)-1; last {
			want := api.TownEmptied{Map: town, Family: residentFamily}
			if len(emptied) != 1 || emptied[0] != want {
				t.Errorf("after the last death, TownEmptied events %+v, want [%+v]", emptied, want)
			}
		} else if len(emptied) != 0 {
			t.Errorf("after death %d of %d, TownEmptied events %+v, want none", i+1, len(residents), emptied)
		}
	}
}
