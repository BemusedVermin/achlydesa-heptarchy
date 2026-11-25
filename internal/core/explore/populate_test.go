package explore_test

import (
	"cmp"
	"slices"
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/explore"
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
	const id = "geography.large"
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

func TestBodiesSpawnOnceWhenAChunkFirstLoads(t *testing.T) {
	fx := newFixture(t, "", populatedRoad)
	atStart := dummies(fx.w)
	if len(atStart) == 0 {
		t.Fatal("no bodies spawned on the generated ground around the start")
	}
	for range 40 {
		if !fx.option("step/e").Enabled {
			break
		}
		fx.stepLeader("e")
	}
	afterWalk := dummies(fx.w)
	if len(afterWalk) <= len(atStart) {
		t.Fatalf("%d bodies after walking out, want more than the %d at the start", len(afterWalk), len(atStart))
	}

	populated := explore.Populated(fx.w)
	fx.w.Map.Unload(func(hexmap.ChunkKey) bool { return false })
	fx.toLeader()
	for _, dir := range []string{"w", "e"} {
		if fx.option("step/" + dir).Enabled {
			fx.step(dir)
			break
		}
	}
	for _, body := range dummies(fx.w) {
		if !slices.Contains(afterWalk, body) && slices.Contains(populated, hexmap.KeyOf(body.At)) {
			t.Errorf("body %v appeared in a chunk that was already populated", body)
		}
	}
	for _, body := range afterWalk {
		if !slices.Contains(dummies(fx.w), body) {
			t.Errorf("body %v is gone after its chunk reloaded", body)
		}
	}

	_, w, _ := gametest.Reload(t, fx.cfg, fx.g)
	if got, want := dummies(w), dummies(fx.w); !slices.Equal(got, want) {
		t.Errorf("bodies after a reload = %v, want %v", got, want)
	}
	if got, want := explore.Populated(w), explore.Populated(fx.w); !slices.Equal(got, want) {
		t.Errorf("populated chunks after a reload = %v, want %v", got, want)
	}
}

func TestTheSameSeedSpawnsTheSameBodies(t *testing.T) {
	first := newFixture(t, "", populatedRoad)
	second := newFixture(t, "", populatedRoad)
	for range 20 {
		first.stepLeader("e")
		second.stepLeader("e")
	}
	got, want := dummies(second.w), dummies(first.w)
	if len(want) == 0 || !slices.Equal(got, want) {
		t.Errorf("bodies = %v, want %v, and some", got, want)
	}
}
