package worldgen_test

import (
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/explore"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/worldgen"
)

func TestZZMeasure(t *testing.T) {
	if os.Getenv("ZZ") == "" {
		t.Skip()
	}
	reg := contenttest.Full(t)
	fixed, err := explore.NewWorldSource(reg)
	if err != nil {
		t.Fatal(err)
	}
	open, _ := explore.Terrain(reg, reg.World().Open)
	stages := worldgen.NewStages()
	for _, register := range []func(*worldgen.Stages) error{worldgen.RegisterCaves, worldgen.RegisterJoin, worldgen.RegisterSinks, worldgen.RegisterVerify} {
		if err := register(stages); err != nil {
			t.Fatal(err)
		}
	}
	geo, _ := reg.Geography("geography.deep")
	src, err := worldgen.NewSource(worldgen.Config{Seed: 1, Bounds: fixed.Bounds(), Geography: geo, Stages: stages, Fixed: fixed, Open: open})
	if err != nil {
		t.Fatal(err)
	}
	var stop = make(chan struct{})
	var peakHeap uint64
	go func() {
		for {
			select {
			case <-stop:
				return
			case <-time.After(50 * time.Millisecond):
				var ms runtime.MemStats
				runtime.ReadMemStats(&ms)
				peakHeap = max(peakHeap, ms.HeapAlloc)
			}
		}
	}()
	start := time.Now()
	key := hexmap.KeyOf(src.Openings()[0].Cell)
	if _, err := src.Chunk(key); err != nil {
		t.Fatal(err)
	}
	close(stop)
	t.Logf("verify in %v, peak heap ~%d MB, peak chunks %d", time.Since(start), peakHeap>>20, worldgen.PeakChunks(src))
}
