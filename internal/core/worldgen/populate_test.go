package worldgen_test

import (
	"slices"
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/worldgen"
)

func populatedSource(t *testing.T, mind *content.SpawnMind) *worldgen.Source {
	t.Helper()
	f := newFixture(t)
	src, err := worldgen.NewSource(worldgen.Config{
		Seed:   3,
		Bounds: f.fixed.Bounds(),
		Geography: content.GeographyDef{
			Stages: []content.StageDef{{Stage: "scatter"}},
			Spawns: []content.GeographySpawn{{
				Template: "template.test.dummy", Weight: 1, MaxDepth: 63, Density: 0.01, Mind: mind,
			}},
		},
		Stages: f.stages,
		Fixed:  f.fixed,
	})
	if err != nil {
		t.Fatal(err)
	}
	return src
}

func TestPopulateAllDoesNotDependOnTheWorkers(t *testing.T) {
	src := populatedSource(t, nil)
	one, err := src.PopulateAll(1)
	if err != nil {
		t.Fatal(err)
	}
	many, err := src.PopulateAll(7)
	if err != nil {
		t.Fatal(err)
	}
	if len(one) == 0 || !slices.Equal(one, many) {
		t.Errorf("1 worker gave %d spawns, 7 workers gave %d, want the same and some", len(one), len(many))
	}
}

func TestSpawnsTakeThePlannerOnlyByChance(t *testing.T) {
	for _, test := range []struct {
		chance float64
		want   string
	}{{0, ""}, {1, "goap"}} {
		spawns, err := populatedSource(t, &content.SpawnMind{Planner: "goap", Chance: test.chance}).PopulateAll(0)
		if err != nil {
			t.Fatal(err)
		}
		if len(spawns) == 0 {
			t.Fatal("no spawns")
		}
		for _, spawn := range spawns {
			if spawn.Planner != test.want {
				t.Fatalf("chance %v: planner %q, want %q", test.chance, spawn.Planner, test.want)
			}
		}
	}
}
