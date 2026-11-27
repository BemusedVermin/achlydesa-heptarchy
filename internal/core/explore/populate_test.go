package explore_test

import (
	"cmp"
	"slices"
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game/gametest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

const dummyTemplate = "template.test.dummy"

type dummy struct {
	UID api.UID
	At  api.Cell
}

// populatedRoad is the road world on cave ground that starts a dummy on one in fifty open cells.
func populatedRoad(bundle *content.Bundle) {
	roadWorld(bundle)
	const id = "geography.deep"
	bundle.World.Geography = id
	geography := bundle.Geographies[id]
	geography.Stages = geography.Stages[:1]
	geography.Spawns = []content.GeographySpawn{{
		Template: dummyTemplate, Weight: 1, MinDepth: 0, MaxDepth: 7, Density: 0.02,
	}}
	bundle.Geographies[id] = geography
}

// dummies lists the dummies in the world, by UID.
func dummies(w *sim.World) []dummy {
	var out []dummy
	query := ecs.NewFilter1[comp.Pos](w.ECS).Query()
	for query.Next() {
		if e := query.Entity(); w.TemplateOf(e) == dummyTemplate {
			out = append(out, dummy{UID: w.UIDOf(e), At: query.Get().At})
		}
	}
	slices.SortFunc(out, func(a, b dummy) int { return cmp.Compare(a.UID, b.UID) })
	return out
}

func TestTheWholeWorldIsPopulatedAtStart(t *testing.T) {
	fx := newFixture(t, "", populatedRoad)
	all := dummies(fx.w)
	loaded := slices.Collect(func(yield func(hexmap.ChunkKey) bool) {
		for _, key := range fx.w.Map.Loaded() {
			if !yield(key) {
				return
			}
		}
	})
	inLoaded := 0
	for _, body := range all {
		if slices.Contains(loaded, hexmap.KeyOf(body.At)) {
			inLoaded++
		}
	}
	if len(all) < 10 || inLoaded == len(all) {
		t.Errorf("%d bodies, %d of them in loaded chunks, want over 10 and some in unloaded chunks", len(all), inLoaded)
	}
	for range 20 {
		fx.stepLeader("e")
	}
	if got := len(dummies(fx.w)); got != len(all) {
		t.Errorf("%d bodies after walking, want the %d from the start", got, len(all))
	}
}

func TestReloadingAndExportingCreateAndLoseNobody(t *testing.T) {
	fx := newFixture(t, "", populatedRoad)
	atStart := dummies(fx.w)
	_, w, _ := gametest.Reload(t, fx.cfg, fx.g)
	if got := dummies(w); !slices.Equal(got, atStart) {
		t.Errorf("%d bodies after a reload, want the %d from the start", len(got), len(atStart))
	}
	fx.w.ExportCount = 1
	fx.w.Bus.Publish(api.ExportBegan{Day: 30, Count: 1})
	if _, err := fx.w.Bus.Drain(fx.w); err != nil {
		t.Fatal(err)
	}
	after := dummies(fx.w)
	if len(after) != len(atStart) {
		t.Fatalf("%d bodies after an export, want %d", len(after), len(atStart))
	}
	for i := range after {
		if after[i].UID != atStart[i].UID {
			t.Errorf("body %d is UID %d after an export, want %d", i, after[i].UID, atStart[i].UID)
		}
	}
}

func TestTheSameSeedSpawnsTheSameBodies(t *testing.T) {
	first := newFixture(t, "", populatedRoad)
	second := newFixture(t, "", populatedRoad)
	got, want := dummies(second.w), dummies(first.w)
	if len(want) == 0 || !slices.Equal(got, want) {
		t.Errorf("bodies = %v, want %v, and some", got, want)
	}
}

// TestTheRealWorldStartsWithAboutAThousandIndividuals builds the real 400 by 400 world, which takes seconds.
func TestTheRealWorldStartsWithAboutAThousandIndividuals(t *testing.T) {
	_, w, _ := gametest.Start(t, game.Config{Seed: 1, Content: contenttest.Full(t)})
	if got := w.NextUID(); got < 500 || got > 6000 {
		t.Errorf("%d individuals at start, want between 500 and 6000 while density is tuned", got)
	}
}
