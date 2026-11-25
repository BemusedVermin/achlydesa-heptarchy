package agents_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game/gametest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

const (
	guardTemplate = "template.generic.guard"
	guardFamily   = "family.people.guards"
	guardTicks    = 40
	entrySlack    = 30
)

func (f *townField) living(template api.ContentID) []ecs.Entity {
	dead := ecs.NewMap[comp.Dead](f.w.ECS)
	var alive []ecs.Entity
	for _, body := range f.withTemplate(template) {
		if !dead.Has(body) {
			alive = append(alive, body)
		}
	}
	return alive
}

func (f *townField) kill(body ecs.Entity) {
	sim.Put(f.w, body, comp.Dead{Tick: f.w.Now})
}

func (f *townField) counted(family api.ContentID) int {
	populations := ecs.NewMap[comp.Populations](f.w.ECS).Get(f.town(townMap))
	for _, p := range populations.Of {
		if p.Family == family {
			return p.Count
		}
	}
	return 0
}

func (f *townField) replacements() []comp.Replacement {
	return ecs.NewMap[comp.Town](f.w.ECS).Get(f.town(townMap)).Replacements
}

func (f *townField) runEntrySystems() {
	f.t.Helper()
	if err := f.w.Systems.Run(f.w, sim.PerEntry); err != nil {
		f.t.Fatal(err)
	}
}

func TestAKilledGuardIsReplacedAtTheGateAfterTheContentTicks(t *testing.T) {
	f := newTownField(t, 1, townMap, nil)
	guards := f.living(guardTemplate)
	if len(guards) != 2 {
		t.Fatalf("%d guards at the start, want 2", len(guards))
	}
	killedAt := f.w.Now
	f.kill(guards[0])
	f.runEntrySystems()
	queued := f.replacements()
	if len(queued) != 1 || queued[0].At != killedAt+guardTicks {
		t.Fatalf("queued %v, want one guard due at tick %d", queued, killedAt+guardTicks)
	}
	for len(f.living(guardTemplate)) < 2 && f.w.Now < killedAt+2000 {
		f.wait()
	}
	now := f.living(guardTemplate)
	if len(now) != 2 {
		t.Fatalf("%d guards alive at tick %d, want 2", len(now), f.w.Now)
	}
	if f.w.Now < killedAt+guardTicks || f.w.Now > killedAt+guardTicks+entrySlack {
		t.Errorf("the replacement came at tick %d, want from %d to %d",
			f.w.Now, killedAt+guardTicks, killedAt+guardTicks+entrySlack)
	}
	arrival := now[0]
	if arrival == guards[1] {
		arrival = now[1]
	}
	if arrival == guards[0] || arrival == guards[1] {
		t.Fatalf("the replacement is not a new body")
	}
	if generic := ecs.NewMap[comp.Generic](
		f.w.ECS,
	); !generic.Has(arrival) ||
		generic.Get(arrival).Town != f.mapID(townMap) {
		t.Errorf("the replacement is not a generic of the town")
	}
	gate := f.w.Maps.Get(f.mapID(townMap)).Gates()[0].At
	if d := hexmap.CellDistance(gate, ecs.NewMap[comp.Pos](f.w.ECS).Get(arrival).At); d > 3 {
		t.Errorf("the replacement stands %d cells from the gate", d)
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
	for f.counted(guardFamily) < 2 && f.w.Now < lostAt+2000 {
		f.wait()
	}
	if got := f.counted(guardFamily); got != 2 {
		t.Fatalf("the town counts %d guards at tick %d, want 2", got, f.w.Now)
	}
	if f.w.Now < lostAt+guardTicks || f.w.Now > lostAt+guardTicks+entrySlack {
		t.Errorf("the replacement came at tick %d, want from %d to %d",
			f.w.Now, lostAt+guardTicks, lostAt+guardTicks+entrySlack)
	}
}

func TestAnEmptiedTownReplacesNobody(t *testing.T) {
	tests := []struct {
		name         string
		emptyFirst   bool
		wantQueuedAt int
	}{
		{"residents gone before the guard dies", true, 0},
		{"residents gone after a replacement was queued", false, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newTownField(t, 1, townMap, nil)
			killResidents := func() {
				for _, body := range f.living(plainResident) {
					f.kill(body)
				}
				ecs.NewMap[comp.Populations](f.w.ECS).Get(f.town(townMap)).Of = slices.DeleteFunc(
					slices.Clone(ecs.NewMap[comp.Populations](f.w.ECS).Get(f.town(townMap)).Of),
					func(p comp.Population) bool { return p.Family == "family.people" },
				)
			}
			guards := f.living(guardTemplate)
			if tt.emptyFirst {
				killResidents()
			}
			killedAt := f.w.Now
			f.kill(guards[0])
			f.runEntrySystems()
			if got := len(f.replacements()); got != tt.wantQueuedAt {
				t.Fatalf("%d replacements queued, want %d", got, tt.wantQueuedAt)
			}
			if !tt.emptyFirst {
				killResidents()
			}
			for f.w.Now < killedAt+10*guardTicks {
				f.wait()
			}
			if got := len(f.living(guardTemplate)); got != 1 {
				t.Errorf("%d guards alive at tick %d, want 1", got, f.w.Now)
			}
			if got := len(f.replacements()); got != 0 {
				t.Errorf("%d replacements still queued in an emptied town", got)
			}
		})
	}
}

func TestACrowdedGateKeepsTheReplacementQueued(t *testing.T) {
	f := newTownField(t, 1, townMap, nil)
	town := f.mapID(townMap)
	for col := range 15 {
		for row := range 8 {
			if _, err := f.w.Spawn(wellTemplate, &comp.Pos{Map: town, At: cellAt(col, row)}); err != nil {
				t.Fatal(err)
			}
		}
	}
	guards := f.living(guardTemplate)
	killedAt := f.w.Now
	f.kill(guards[0])
	for f.w.Now < killedAt+3*guardTicks {
		f.wait()
	}
	if got := len(f.living(guardTemplate)); got != 1 {
		t.Errorf("%d guards alive at tick %d, want 1", got, f.w.Now)
	}
	if got := len(f.replacements()); got != 1 {
		t.Errorf("%d replacements queued, want the one still waiting", got)
	}
}

func TestGenericsAndTheirTimersSurviveSaveAndLoad(t *testing.T) {
	g, w, cfg := gametest.New(t, 1, func(bundle *content.Bundle) { bundle.Start.Map = townMap })
	f := &townField{t: t, g: g, w: w}
	f.kill(f.living(guardTemplate)[0])
	f.runEntrySystems()
	if len(f.replacements()) != 1 {
		t.Fatalf("%d replacements queued, want 1", len(f.replacements()))
	}
	loaded, loadedWorld, warnings := gametest.Reload(t, cfg, g)
	if len(warnings) != 0 {
		t.Errorf("warnings = %v, want none", warnings)
	}
	before, err := g.Hash()
	if err != nil {
		t.Fatal(err)
	}
	after, err := loaded.Hash()
	if err != nil {
		t.Fatal(err)
	}
	if before != after {
		t.Error("hash changed across save and load")
	}
	reloaded := &townField{t: t, g: loaded, w: loadedWorld}
	if got := reloaded.replacements(); !slices.Equal(got, f.replacements()) {
		t.Errorf("replacements after load = %v, want %v", got, f.replacements())
	}
	if got := len(reloaded.living(guardTemplate)); got != 1 {
		t.Errorf("%d guards alive after load, want 1", got)
	}
	generics := 0
	query := ecs.NewFilter1[comp.Generic](loadedWorld.ECS).Query()
	for query.Next() {
		generics++
	}
	if generics == 0 {
		t.Error("no body carries comp.Generic after load")
	}
}

func TestGenericsAreChecked(t *testing.T) {
	tests := []struct {
		name string
		edit func(*content.MapDef)
		want string
	}{
		{"no residents", func(m *content.MapDef) { m.Region.Town.Residents = "" }, "needs residents"},
		{"an unknown template", func(m *content.MapDef) {
			m.Region.Town.Generics[0].Template = "template.nope"
		}, "unknown template"},
		{"a template listed twice", func(m *content.MapDef) {
			m.Region.Town.Generics[1].Template = m.Region.Town.Generics[0].Template
		}, "listed twice"},
		{"a count below one", func(m *content.MapDef) { m.Region.Town.Generics[0].Count = 0 }, "count"},
		{"a template in a shared family", func(m *content.MapDef) {
			m.Region.Town.Generics[0].Template = hauler
		}, "exactly one family"},
		{"a template in no family", func(m *content.MapDef) {
			m.Region.Town.Generics[0].Template = "template.person.reaper"
		}, "exactly one family"},
		{"no ticks", func(m *content.MapDef) { m.Region.Town.Generics[0].Ticks = 0 }, "ticks"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bundle := contenttest.SampleBundle(t)
			contenttest.Edit(bundle.Maps, townMap, tt.edit)
			if tt.name == "a template in a shared family" {
				contenttest.Edit(bundle.Families, "family.people.haulers", func(def *content.FamilyDef) {
					def.Templates = append(def.Templates, content.Weighted{Template: plainResident, Weight: 1})
				})
			}
			_, err := content.New(bundle)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error = %v, want one mentioning %q", err, tt.want)
			}
		})
	}
}
