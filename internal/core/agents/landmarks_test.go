package agents

import (
	"slices"
	"testing"

	"github.com/mlange-42/ark/ecs"
	"pgregory.net/rapid"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/effects"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

const (
	well       = api.ContentID("template.source.well")
	waterGiven = api.ContentID("affordance.water")
	theDead    = api.ContentID("affordance.the_dead")
)

func landmarkWorld(tb testing.TB) *sim.World {
	tb.Helper()
	w := sightWorld(tb)
	if err := effects.Install(w); err != nil {
		tb.Fatal(err)
	}
	if err := Install(w); err != nil {
		tb.Fatal(err)
	}
	return w
}

func known(tb testing.TB, w *sim.World, body ecs.Entity, tag api.ContentID) []Landmark {
	tb.Helper()
	landmarks, err := Known(w, body, tag)
	if err != nil {
		tb.Fatal(err)
	}
	return landmarks
}

func TestAnAgentKnowsAWellOnlyOnceItHasSeenIt(t *testing.T) {
	w := landmarkWorld(t)
	wellAt := at(4, 2, 0)
	spring, err := w.Spawn(well, &comp.Pos{At: wellAt})
	if err != nil {
		t.Fatal(err)
	}
	if got := ecs.NewMap[comp.Affordance](w.ECS).Get(spring).Tags; !slices.Equal(got, []api.ContentID{waterGiven}) {
		t.Fatalf("a well gives %v, want water", got)
	}
	unseeing := spawnBody(w, at(-12, 0, 0))
	looker := spawnBody(w, at(0, 2, 0))
	for _, body := range []ecs.Entity{unseeing, looker} {
		if err := sense(w, body); err != nil {
			t.Fatal(err)
		}
	}
	if got := known(t, w, unseeing, waterGiven); len(got) != 0 {
		t.Errorf("an agent that never saw the well knows %v", got)
	}
	want := []Landmark{{Subject: w.UIDOf(spring), At: wellAt}}
	if got := known(t, w, looker, waterGiven); !slices.Equal(got, want) {
		t.Errorf("an agent that saw the well knows %v, want %v", got, want)
	}
	if got := known(t, w, looker, theDead); len(got) != 0 {
		t.Errorf("the well is known to give the dead: %v", got)
	}
}

func TestKnownListsTheNearestLandmarkFirstByWhereItWasSeen(t *testing.T) {
	w := landmarkWorld(t)
	far, err := w.Spawn(well, &comp.Pos{At: at(0, 6, 0)})
	if err != nil {
		t.Fatal(err)
	}
	near, err := w.Spawn(well, &comp.Pos{At: at(0, -2, 0)})
	if err != nil {
		t.Fatal(err)
	}
	looker := spawnBody(w, at(0, 0, 0))
	if err := sense(w, looker); err != nil {
		t.Fatal(err)
	}
	ecs.NewMap[comp.Pos](w.ECS).Get(near).At = at(-3, 5, 0)
	got := known(t, w, looker, waterGiven)
	if len(got) != 2 || got[0].Subject != w.UIDOf(near) || got[1].Subject != w.UIDOf(far) {
		t.Errorf("known wells = %v, want the one seen nearer first", got)
	}
}

func TestAMarkWithNoBodyIsSeenAsALandmark(t *testing.T) {
	w := landmarkWorld(t)
	grave := w.NewIdentified("place.test.den")
	ecs.NewMap[comp.Pos](w.ECS).Add(grave, &comp.Pos{At: at(1, 1, 0)})
	ecs.NewMap[comp.Affordance](w.ECS).Add(grave, &comp.Affordance{Tags: []api.ContentID{theDead}})
	looker := spawnBody(w, at(0, 0, 0))
	if err := sense(w, looker); err != nil {
		t.Fatal(err)
	}
	want := []Landmark{{Subject: w.UIDOf(grave), At: at(1, 1, 0)}}
	if got := known(t, w, looker, theDead); !slices.Equal(got, want) {
		t.Errorf("known graves = %v, want %v", got, want)
	}
}

func TestARingHoldsExactlyTheCellsAtItsRadius(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		center := at(rapid.IntRange(-20, 20).Draw(rt, "q"), rapid.IntRange(-20, 20).Draw(rt, "r"), 3)
		radius := rapid.IntRange(0, 9).Draw(rt, "radius")
		got := appendRing(nil, center, radius)
		slices.SortFunc(got, hexmap.CompareCells)
		var want []api.Cell
		for _, cell := range (&hexmap.Map{}).CellsIn(center.Hex, radius, 3, 3) {
			if hexmap.Distance(center.Hex, cell.Hex) == radius {
				want = append(want, cell)
			}
		}
		if !slices.Equal(got, want) {
			rt.Fatalf("ring of radius %d = %v, want %v", radius, got, want)
		}
	})
}

func TestSenseRecordsTheGroundUnderfootAndTheNearestOfEachOtherDecay(t *testing.T) {
	bundle := contenttest.SampleBundle(t)
	bundle.World.Geography = "geography.deep"
	w := sightWorld(t)
	w.Content = contenttest.Registry(t, bundle)
	w.ExportCount = 2
	looker := spawnBody(w, at(0, 0, 0))
	if err := sense(w, looker); err != nil {
		t.Fatal(err)
	}
	ground, err := KnownGround(w, looker)
	if err != nil {
		t.Fatal(err)
	}
	want := []Ground{{At: at(0, 0, 4), Level: 1}, {At: at(0, 0, 0), Level: 2}, {At: at(0, 0, -4), Level: 3}}
	if !slices.Equal(ground, want) {
		t.Errorf("known ground = %v, want %v", ground, want)
	}
}
