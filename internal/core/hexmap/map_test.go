package hexmap

import (
	"slices"
	"testing"

	"pgregory.net/rapid"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

var (
	air  = Terrain{ID: "terrain.air"}
	rock = Terrain{ID: "terrain.rock", Solid: true, Opaque: true}
)

// countingSource counts how often each chunk is requested.
type countingSource struct {
	inner Source
	calls map[ChunkKey]int
}

func (c *countingSource) Chunk(key ChunkKey) (*Chunk, error) {
	c.calls[key]++
	return c.inner.Chunk(key)
}

func TestChunkKeyFloorsAcrossBoundaries(t *testing.T) {
	tests := []struct {
		coord, key, local int
	}{
		{-17, -2, 15}, {-16, -1, 0}, {-15, -1, 1}, {-1, -1, 15},
		{0, 0, 0}, {1, 0, 1}, {15, 0, 15}, {16, 1, 0}, {17, 1, 1},
	}
	for _, tt := range tests {
		got := KeyOf(cell(tt.coord, tt.coord, tt.coord))
		if want := (ChunkKey{tt.key, tt.key, tt.key}); got != want {
			t.Errorf("KeyOf(%d) = %v, want %v", tt.coord, got, want)
		}
		if got := floorMod(tt.coord); got != tt.local {
			t.Errorf("floorMod(%d) = %d, want %d", tt.coord, got, tt.local)
		}
	}
}

func TestSetAtRoundTrips(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		m := NewMap(FuncSource(func(api.Cell) Terrain { return air }), Prisms{})
		coord := rapid.IntRange(-40, 40)
		c := cell(coord.Draw(rt, "q"), coord.Draw(rt, "r"), coord.Draw(rt, "depth"))
		if rapid.Bool().Draw(rt, "edge") {
			c.Hex.Q = rapid.SampledFrom([]int{-17, -16, -1, 0, 15, 16}).Draw(rt, "edgeQ")
		}
		if err := m.Set(c, rock); err != nil {
			rt.Fatalf("Set: %v", err)
		}
		got, err := m.At(c)
		if err != nil || got != rock {
			rt.Fatalf("At(%v) = %v, %v; want rock", c, got, err)
		}
		other := c
		other.Hex.Q++
		if got, _ := m.At(other); got != air {
			rt.Fatalf("At(%v) = %v; want air", other, got)
		}
	})
}

func TestChunkLoadedOnceAndEditsSurvive(t *testing.T) {
	src := &countingSource{inner: FuncSource(func(api.Cell) Terrain { return air }), calls: map[ChunkKey]int{}}
	m := NewMap(src, Prisms{})

	edited := []api.Cell{cell(5, 5, 5), cell(-1, 0, 0), cell(5, 4, 5), cell(5, 5, 4)}
	for _, c := range edited {
		if err := m.Set(c, rock); err != nil {
			t.Fatal(err)
		}
	}
	for range 3 {
		for _, c := range append(edited, cell(0, 0, 0), cell(1, 1, 1)) {
			if _, err := m.At(c); err != nil {
				t.Fatal(err)
			}
		}
	}
	for key, n := range src.calls {
		if n != 1 {
			t.Errorf("chunk %v requested %d times, want 1", key, n)
		}
	}
	for _, c := range edited {
		if got, _ := m.At(c); got != rock {
			t.Errorf("edit at %v lost: %v", c, got)
		}
	}
	got := m.Edits()
	want := []api.Cell{cell(-1, 0, 0), cell(5, 5, 4), cell(5, 4, 5), cell(5, 5, 5)}
	if len(got) != len(want) {
		t.Fatalf("Edits() has %d entries, want %d", len(got), len(want))
	}
	for i, e := range got {
		if e.Cell != want[i] {
			t.Errorf("Edits()[%d] = %v, want %v", i, e.Cell, want[i])
		}
	}
}

func TestCellsInSortedAndStable(t *testing.T) {
	m := NewMap(FuncSource(func(api.Cell) Terrain { return air }), Prisms{})
	center := api.Hex{Q: -3, R: 2}
	first := m.CellsIn(center, 2, 0, 1)
	if want := 2 * 19; len(first) != want {
		t.Fatalf("CellsIn returned %d cells, want %d", len(first), want)
	}
	if !slices.IsSortedFunc(first, CompareCells) {
		t.Error("CellsIn is not sorted by Depth, R, Q")
	}
	for _, c := range first {
		if dist := Distance(c.Hex, center); dist > 2 {
			t.Errorf("%v is %d hexes from the center", c, dist)
		}
	}
	if again := m.CellsIn(center, 2, 0, 1); !slices.Equal(first, again) {
		t.Fatal("CellsIn changed between calls")
	}
	if got := m.CellsIn(center, -1, 0, 0); got != nil {
		t.Errorf("negative radius gave %v", got)
	}
	if got := m.CellsIn(center, 1, 2, 1); got != nil {
		t.Errorf("empty depth range gave %v", got)
	}
}

func TestAnUnloadedChunkReloadsTheSameWithItsEdits(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		m := NewMap(FuncSource(func(c api.Cell) Terrain {
			if (c.Hex.Q*7+c.Hex.R*3+c.Depth)%5 == 0 {
				return rock
			}
			return air
		}), Prisms{})
		coord := rapid.IntRange(-40, 40)
		cells := make([]api.Cell, 8)
		for i := range cells {
			cells[i] = cell(coord.Draw(rt, "q"), coord.Draw(rt, "r"), coord.Draw(rt, "depth"))
		}
		edited := cells[0]
		if err := m.Set(edited, Terrain{ID: "terrain.edit"}); err != nil {
			rt.Fatal(err)
		}
		before := make([]Terrain, len(cells))
		for i, c := range cells {
			before[i], _ = m.At(c)
		}
		m.Unload(func(ChunkKey) bool { return false })
		if loaded := m.Loaded(); len(loaded) != 0 {
			rt.Fatalf("after unloading everything, loaded %v", loaded)
		}
		for i, c := range cells {
			if got, _ := m.At(c); got != before[i] {
				rt.Fatalf("At(%v) after reload = %v, want %v", c, got, before[i])
			}
		}
		if got, _ := m.At(edited); got.ID != "terrain.edit" {
			rt.Fatalf("edit at %v lost on unload: %v", edited, got)
		}
	})
}

func TestUnloadKeepsWhatKeepAccepts(t *testing.T) {
	src := &countingSource{inner: FuncSource(func(api.Cell) Terrain { return air }), calls: map[ChunkKey]int{}}
	m := NewMap(src, Prisms{})
	for _, q := range []int{0, 16, 32} {
		if _, err := m.At(cell(q, 0, 0)); err != nil {
			t.Fatal(err)
		}
	}
	m.Unload(func(k ChunkKey) bool { return k.Q != 1 })
	if got, want := m.Loaded(), []ChunkKey{{0, 0, 0}, {2, 0, 0}}; !slices.Equal(got, want) {
		t.Fatalf("loaded %v, want %v", got, want)
	}
	if _, err := m.At(cell(16, 0, 0)); err != nil {
		t.Fatal(err)
	}
	if got := src.calls[ChunkKey{1, 0, 0}]; got != 2 {
		t.Errorf("chunk 1 loaded %d times, want 2", got)
	}
}

func TestSlotOfAndCellAreInverse(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		coord := rapid.IntRange(-100, 100)
		c := cell(coord.Draw(rt, "q"), coord.Draw(rt, "r"), coord.Draw(rt, "depth"))
		if got := KeyOf(c).Cell(SlotOf(c)); got != c {
			rt.Fatalf("KeyOf(%v).Cell(SlotOf) = %v", c, got)
		}
	})
}
