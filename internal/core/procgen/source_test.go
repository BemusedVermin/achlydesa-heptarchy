package procgen_test

import (
	"errors"
	"slices"
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/explore"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/procgen"
)

type stageFunc func(ctx *procgen.Context, chunk *hexmap.Chunk) error

func (f stageFunc) Run(ctx *procgen.Context, chunk *hexmap.Chunk) error { return f(ctx, chunk) }

var noise = []hexmap.Terrain{{ID: "terrain.test.noise_a"}, {ID: "terrain.test.noise_b", Solid: true}}

func scatter(ctx *procgen.Context, chunk *hexmap.Chunk) error {
	for depth := range hexmap.ChunkSize {
		for r := range hexmap.ChunkSize {
			for q := range hexmap.ChunkSize {
				chunk.Set(q, r, depth, noise[ctx.Rand.IntN(len(noise))])
			}
		}
	}
	return nil
}

type fixture struct {
	fixed  *explore.WorldSource
	stages *procgen.Stages
}

func newFixture(t *testing.T) fixture {
	t.Helper()
	fixed, err := explore.NewWorldSource(contenttest.Sample(t))
	if err != nil {
		t.Fatal(err)
	}
	stages := procgen.NewStages()
	if err := stages.Register("scatter", stageFunc(scatter), func(params map[string]float64) error {
		if _, ok := params["unknown"]; ok {
			return errors.New("takes no param unknown")
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	return fixture{fixed: fixed, stages: stages}
}

func (f fixture) source(t *testing.T, seed, exportCount uint64) *procgen.Source {
	t.Helper()
	src, err := procgen.NewSource(procgen.Config{
		Seed:        seed,
		ExportCount: exportCount,
		Bounds:      f.fixed.Bounds(),
		Geography:   content.GeographyDef{Stages: []content.StageDef{{Stage: "scatter"}}},
		Stages:      f.stages,
		Fixed:       f.fixed,
	})
	if err != nil {
		t.Fatal(err)
	}
	return src
}

func cells(t *testing.T, src hexmap.Source, key hexmap.ChunkKey) []hexmap.Terrain {
	t.Helper()
	chunk, err := src.Chunk(key)
	if err != nil {
		t.Fatal(err)
	}
	var out []hexmap.Terrain
	for depth := range hexmap.ChunkSize {
		for r := range hexmap.ChunkSize {
			for q := range hexmap.ChunkSize {
				out = append(out, chunk.Get(q, r, depth))
			}
		}
	}
	return out
}

func TestSourceChunkDependsOnlyOnSeedExportCountAndKey(t *testing.T) {
	f := newFixture(t)
	var keys []hexmap.ChunkKey
	for depth := -1; depth <= 1; depth++ {
		for r := -1; r <= 2; r++ {
			for q := -1; q <= 5; q++ {
				keys = append(keys, hexmap.ChunkKey{Q: q, R: r, Depth: depth})
			}
		}
	}

	forward, backward := f.source(t, 1, 0), f.source(t, 1, 0)
	built := map[hexmap.ChunkKey][]hexmap.Terrain{}
	for _, key := range keys {
		built[key] = cells(t, forward, key)
	}
	for _, key := range slices.Backward(keys) {
		if !slices.Equal(cells(t, backward, key), built[key]) {
			t.Fatalf("chunk %v differs when chunks are built in reverse order", key)
		}
	}

	t.Run("fixed cells and cells outside the bounds keep the fixed layer", func(t *testing.T) {
		generated, kept := 0, 0
		for _, key := range keys {
			fixed := cells(t, f.fixed, key)
			origin := key.Origin()
			for depth := range hexmap.ChunkSize {
				for r := range hexmap.ChunkSize {
					for q := range hexmap.ChunkSize {
						cell := api.Cell{
							Hex:   api.Hex{Q: origin.Hex.Q + q, R: origin.Hex.R + r},
							Depth: origin.Depth + depth,
						}
						i := (depth*hexmap.ChunkSize+r)*hexmap.ChunkSize + q
						got := built[key][i]
						if !f.fixed.Bounds().Contains(cell) || f.fixed.Covers(cell) {
							kept++
							if got != fixed[i] {
								t.Fatalf("%v is %q, want the fixed layer's %q", cell, got.ID, fixed[i].ID)
							}
						} else if slices.Contains(noise, got) {
							generated++
						}
					}
				}
			}
		}
		if generated == 0 || kept == 0 {
			t.Fatalf("%d generated and %d kept cells, want some of each", generated, kept)
		}
	})

	t.Run("export count and world seed change a chunk", func(t *testing.T) {
		key := hexmap.ChunkKey{Q: 1, R: 1}
		if slices.Equal(cells(t, f.source(t, 1, 1), key), built[key]) {
			t.Error("chunk unchanged after an export")
		}
		if slices.Equal(cells(t, f.source(t, 2, 0), key), built[key]) {
			t.Error("chunk unchanged with another world seed")
		}
	})

	t.Run("a geography's stages are checked", func(t *testing.T) {
		tests := []struct {
			name  string
			stage content.StageDef
		}{
			{"unknown stage", content.StageDef{Stage: "missing"}},
			{"bad params", content.StageDef{Stage: "scatter", Params: map[string]float64{"unknown": 1}}},
		}
		for _, test := range tests {
			_, err := procgen.NewSource(procgen.Config{
				Bounds:    f.fixed.Bounds(),
				Geography: content.GeographyDef{Stages: []content.StageDef{test.stage}},
				Stages:    f.stages,
				Fixed:     f.fixed,
			})
			if err == nil {
				t.Errorf("%s: NewSource accepted it", test.name)
			}
		}
	})
}
