package worldgen_test

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pgregory.net/rapid"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/explore"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/worldgen"
)

var update = flag.Bool("update", false, "rewrite the golden files")

func caveSource(tb testing.TB, seed, exportCount uint64) (*worldgen.Source, hexmap.Terrain) {
	tb.Helper()
	reg := contenttest.Sample(tb)
	fixed, err := explore.NewWorldSource(reg, 0)
	if err != nil {
		tb.Fatal(err)
	}
	open, err := explore.Terrain(reg, reg.World().Open)
	if err != nil {
		tb.Fatal(err)
	}
	stages := worldgen.NewStages()
	if err := worldgen.RegisterCaves(stages); err != nil {
		tb.Fatal(err)
	}
	geography, _ := reg.Geography("geography.deep")
	geography.Stages = geography.Stages[:1]
	src, err := worldgen.NewSource(worldgen.Config{
		Seed: seed, ExportCount: exportCount, Bounds: fixed.Bounds(),
		Geography: geography, Stages: stages, Fixed: fixed, Open: open,
	})
	if err != nil {
		tb.Fatal(err)
	}
	return src, open
}

func drawChunk(tb testing.TB, src hexmap.Source, key hexmap.ChunkKey, open hexmap.Terrain) string {
	tb.Helper()
	chunk, err := src.Chunk(key)
	if err != nil {
		tb.Fatal(err)
	}
	var out strings.Builder
	for depth := range hexmap.ChunkSize {
		for r := range hexmap.ChunkSize {
			for q := range hexmap.ChunkSize {
				if chunk.Get(q, r, depth) == open {
					out.WriteByte('.')
				} else {
					out.WriteByte('#')
				}
			}
			out.WriteByte('\n')
		}
		out.WriteByte('\n')
	}
	return out.String()
}

func TestCavesGolden(t *testing.T) {
	src, open := caveSource(t, 1, 0)
	got := drawChunk(t, src, hexmap.ChunkKey{Q: 2, R: 1}, open)
	path := filepath.Join("testdata", "caves_chunk.golden")
	if *update {
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != string(want) {
		t.Errorf("chunk differs from %s; run with -update if the change is intended", path)
	}
	if !strings.Contains(got, ".") || !strings.Contains(got, "#") {
		t.Error("chunk is all open or all solid")
	}
}

func TestCavesKeepSurfaceSolid(t *testing.T) {
	src, _ := caveSource(t, 1, 0)
	fixed, err := explore.NewWorldSource(contenttest.Sample(t), 0)
	if err != nil {
		t.Fatal(err)
	}
	for q := range 4 {
		for r := range 2 {
			key := hexmap.ChunkKey{Q: q, R: r}
			chunk, err := src.Chunk(key)
			if err != nil {
				t.Fatal(err)
			}
			base, err := fixed.Chunk(key)
			if err != nil {
				t.Fatal(err)
			}
			for y := range hexmap.ChunkSize {
				for x := range hexmap.ChunkSize {
					if chunk.Get(x, y, 0) != base.Get(x, y, 0) {
						t.Fatalf("chunk %v cell (%d,%d) is carved at depth 0", key, x, y)
					}
				}
			}
		}
	}
}

func TestCavesIgnoreChunkOrder(t *testing.T) {
	var keys []hexmap.ChunkKey
	for r := range 3 {
		for q := range 4 {
			keys = append(keys, hexmap.ChunkKey{Q: q, R: r})
		}
	}
	reference, open := caveSource(t, 7, 2)
	want := map[hexmap.ChunkKey]string{}
	for _, key := range keys {
		want[key] = drawChunk(t, reference, key, open)
	}
	rapid.Check(t, func(rt *rapid.T) {
		src, _ := caveSource(t, 7, 2)
		order := rapid.Permutation(keys).Draw(rt, "order")
		for _, key := range order {
			if got := drawChunk(t, src, key, open); got != want[key] {
				rt.Fatalf("chunk %v differs when built in order %v", key, order)
			}
		}
	})
}
