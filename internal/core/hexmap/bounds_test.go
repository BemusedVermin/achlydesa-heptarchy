package hexmap

import (
	"errors"
	"runtime"
	"testing"

	"pgregory.net/rapid"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

var edgeRock = Terrain{ID: "terrain.edge", Solid: true, Opaque: true}

func offsetCell(col, row, depth int) api.Cell {
	return api.Cell{Hex: FromOffsetOddR(col, row), Depth: depth}
}

func TestOffsetOddRRoundTrips(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		col, row := rapid.IntRange(-500, 500).Draw(rt, "col"), rapid.IntRange(-500, 500).Draw(rt, "row")
		if gotCol, gotRow := ToOffsetOddR(FromOffsetOddR(col, row)); gotCol != col || gotRow != row {
			rt.Fatalf("(%d, %d) came back as (%d, %d)", col, row, gotCol, gotRow)
		}
	})
}

func TestBoundsContains(t *testing.T) {
	b := Bounds{MinCol: 0, MaxCol: 9, MinRow: 0, MaxRow: 4, MinDepth: -1, MaxDepth: 2}
	tests := []struct {
		name string
		cell api.Cell
		want bool
	}{
		{"corner low", offsetCell(0, 0, -1), true},
		{"corner high", offsetCell(9, 4, 2), true},
		{"odd row last column", offsetCell(9, 3, 0), true},
		{"odd row past last column", offsetCell(10, 3, 0), false},
		{"even row before first column", offsetCell(-1, 2, 0), false},
		{"row above", offsetCell(3, -1, 0), false},
		{"row below", offsetCell(3, 5, 0), false},
		{"too shallow", offsetCell(3, 2, -2), false},
		{"too deep", offsetCell(3, 2, 3), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := b.Contains(tt.cell); got != tt.want {
				t.Errorf("Contains(%v) = %v, want %v", tt.cell, got, tt.want)
			}
		})
	}
}

func checkerSource(c api.Cell) Terrain {
	if (c.Hex.Q+2*c.Hex.R+3*c.Depth)%5 == 0 {
		return rock
	}
	return air
}

func TestBoundedAtMatchesSourceInsideAndEdgeOutside(t *testing.T) {
	rapid.Check(t, func(rt *rapid.T) {
		minCol, minRow := rapid.IntRange(-40, 20).Draw(rt, "minCol"), rapid.IntRange(-40, 20).Draw(rt, "minRow")
		minDepth := rapid.IntRange(-20, 10).Draw(rt, "minDepth")
		b := Bounds{
			MinCol: minCol, MaxCol: minCol + rapid.IntRange(0, 40).Draw(rt, "width"),
			MinRow: minRow, MaxRow: minRow + rapid.IntRange(0, 40).Draw(rt, "height"),
			MinDepth: minDepth, MaxDepth: minDepth + rapid.IntRange(0, 20).Draw(rt, "depths"),
		}
		m := NewMap(1, "bounded", FuncSource(checkerSource), Prisms{})
		m.SetBounds(b, edgeRock)
		coord := rapid.IntRange(-70, 70)
		c := offsetCell(coord.Draw(rt, "col"), coord.Draw(rt, "row"), coord.Draw(rt, "depth"))

		got, err := m.At(c)
		if err != nil {
			rt.Fatalf("At(%v): %v", c, err)
		}
		want := edgeRock
		if b.Contains(c) {
			want = checkerSource(c)
		}
		if got != want {
			rt.Fatalf("At(%v) = %v, want %v (inside: %v)", c, got, want, b.Contains(c))
		}

		err = m.Set(c, air)
		if b.Contains(c) != (err == nil) {
			rt.Fatalf("Set(%v) inside=%v gave error %v", c, b.Contains(c), err)
		}
		if err != nil && !errors.Is(err, ErrOutOfBounds) {
			rt.Fatalf("Set(%v) error %v is not ErrOutOfBounds", c, err)
		}
	})
}

func TestUniformChunkStoresNoSlots(t *testing.T) {
	chunk := NewChunk(air)
	chunk.Set(3, 4, 5, air)
	if chunk.cells != nil {
		t.Fatal("setting the fill terrain allocated slots")
	}
	if got := chunk.Get(15, 15, 15); got != air {
		t.Errorf("Get on a uniform chunk = %v, want air", got)
	}
	chunk.Set(3, 4, 5, rock)
	if chunk.cells == nil {
		t.Fatal("setting another terrain allocated no slots")
	}
	if got := chunk.Get(3, 4, 5); got != rock {
		t.Errorf("Get(3, 4, 5) = %v, want rock", got)
	}
	if got := chunk.Get(4, 4, 5); got != air {
		t.Errorf("Get(4, 4, 5) = %v, want air", got)
	}
	if uniform, _ := FuncSource(
		func(api.Cell) Terrain { return rock },
	).Chunk(ChunkKey{1, -2, 3}); uniform.cells != nil {
		t.Error("a uniform FuncSource chunk allocated slots")
	}
}

func TestSteppingOffTheEdgeIsRefused(t *testing.T) {
	m := floorWorld()
	m.SetBounds(Bounds{MinCol: 0, MaxCol: 5, MinRow: 0, MaxRow: 5, MinDepth: -3, MaxDepth: 5}, edgeRock)
	rules := defaultStepRules()
	lastCol := offsetCell(5, 2, 0)
	beyond := offsetCell(6, 2, 0)

	if ok, err := m.Standable(beyond); err != nil || ok {
		t.Errorf("Standable(beyond the edge) = %v, %v; want false", ok, err)
	}
	if ok, err := m.Standable(lastCol); err != nil || !ok {
		t.Fatalf("Standable(last column) = %v, %v; want true", ok, err)
	}
	if step, ok, err := m.Step(lastCol, api.DirE, rules); err != nil || ok {
		t.Errorf("Step east off the edge = %v, %v, %v; want refused", step, ok, err)
	}
	if route, ok, err := m.Path(offsetCell(0, 2, 0), beyond, rules, nil, 500); err != nil || ok {
		t.Errorf("Path beyond the edge = %v, %v, %v; want none", route, ok, err)
	}
	if _, ok, err := m.Path(offsetCell(0, 2, 0), lastCol, rules, nil, 500); err != nil || !ok {
		t.Errorf("Path to the last column = %v, %v; want a route", ok, err)
	}
	reached, err := m.Reachable(offsetCell(0, 2, 0), rules, 50, nil)
	if err != nil {
		t.Fatal(err)
	}
	for c := range reached {
		if !m.bounds.Contains(c) {
			t.Errorf("Reachable reached %v outside the bounds", c)
		}
	}
}

func TestLargeUniformWorldStaysSmall(t *testing.T) {
	const cols, rows, depths, limit = 400, 400, 60, 5 << 20
	m := NewMap(1, "large", FuncSource(func(api.Cell) Terrain { return rock }), Prisms{})
	m.SetBounds(Bounds{MaxCol: cols - 1, MaxRow: rows - 1, MaxDepth: depths - 1}, edgeRock)

	var before, after runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&before)
	for depth := 0; depth < depths; depth += ChunkSize {
		for row := range rows {
			for col := range cols {
				if _, err := m.At(offsetCell(col, row, depth)); err != nil {
					t.Fatal(err)
				}
			}
		}
	}
	runtime.ReadMemStats(&after)

	allocated := after.TotalAlloc - before.TotalAlloc
	t.Logf("%d chunks, %d bytes allocated", len(m.chunks), allocated)
	if allocated >= limit {
		t.Errorf("touching every chunk allocated %d bytes, want under %d", allocated, limit)
	}
}
