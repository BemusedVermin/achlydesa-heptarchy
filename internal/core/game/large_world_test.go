package game_test

import (
	"slices"
	"testing"
	"time"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game/gametest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

func startLarge(t *testing.T, seed uint64) (*game.Game, *sim.World, game.Config) {
	t.Helper()
	return gametest.Start(t, game.Config{Seed: seed, Content: contenttest.Full(t)})
}

// deepAround renders whether each cell of a block around the places in the generated deep is solid.
func deepAround(t *testing.T, w *sim.World) string {
	t.Helper()
	var out []byte
	for depth := 8; depth <= 14; depth++ {
		for row := 180; row < 220; row++ {
			for col := 150; col < 260; col++ {
				terrain, err := w.Map.At(api.Cell{Hex: api.Hex{Q: col - row/2, R: row}, Depth: depth})
				if err != nil {
					t.Fatal(err)
				}
				if terrain.Solid {
					out = append(out, '#')
				} else {
					out = append(out, '.')
				}
			}
		}
	}
	return string(out)
}

func TestLargeWorldStartsLazily(t *testing.T) {
	begin := time.Now()
	_, w, _ := startLarge(t, 1)
	if elapsed := time.Since(begin); elapsed > 5*time.Second {
		t.Errorf("start took %v", elapsed)
	}
	total := 400 * 400 * 64 / (hexmap.ChunkSize * hexmap.ChunkSize * hexmap.ChunkSize)
	if loaded := len(w.Map.Loaded()); loaded*20 > total {
		t.Errorf("loaded %d of %d chunks at start", loaded, total)
	}
}

func TestLargeWorldDeepFollowsTheSeed(t *testing.T) {
	_, a, _ := startLarge(t, 1)
	_, b, _ := startLarge(t, 1)
	_, c, _ := startLarge(t, 2)
	if deepAround(t, a) != deepAround(t, b) {
		t.Error("the same seed generated different deeps")
	}
	if deepAround(t, a) == deepAround(t, c) {
		t.Error("different seeds generated the same deep")
	}
}

func TestLargeWorldSurvivesSaveAndLoad(t *testing.T) {
	g, w, cfg := startLarge(t, 1)
	air, err := w.Map.At(api.Cell{Hex: api.Hex{Q: 0, R: 0}, Depth: 0})
	if err != nil {
		t.Fatal(err)
	}
	if err := w.Map.Set(api.Cell{Hex: api.Hex{Q: 100, R: 200}, Depth: 9}, air); err != nil {
		t.Fatal(err)
	}
	want, err := g.Hash()
	if err != nil {
		t.Fatal(err)
	}
	loaded, loadedWorld, _ := gametest.Reload(t, cfg, g)
	got, err := loaded.Hash()
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Error("loaded game hashes differently")
	}
	if !slices.Equal(loadedWorld.Map.Edits(), w.Map.Edits()) || deepAround(t, loadedWorld) != deepAround(t, w) {
		t.Error("loaded world differs")
	}
}
