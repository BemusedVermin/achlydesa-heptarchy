package hexmap

import (
	"slices"
	"testing"

	"pgregory.net/rapid"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

func TestSeesAgreesWithVisible(t *testing.T) {
	m := NewMap(FuncSource(checkerSource), Prisms{})
	rapid.Check(t, func(rt *rapid.T) {
		from := cell(
			rapid.IntRange(-10, 10).Draw(rt, "q"),
			rapid.IntRange(-10, 10).Draw(rt, "r"),
			rapid.IntRange(-3, 3).Draw(rt, "depth"),
		)
		radius, depthRange := rapid.IntRange(0, 5).Draw(rt, "radius"), rapid.IntRange(0, 2).Draw(rt, "depthRange")
		visible, err := m.Visible(from, radius, depthRange)
		if err != nil {
			rt.Fatal(err)
		}
		for _, to := range m.CellsIn(from.Hex, radius+1, from.Depth-depthRange-1, from.Depth+depthRange+1) {
			sees, err := m.Sees(from, to, radius, depthRange)
			if err != nil {
				rt.Fatal(err)
			}
			if want := slices.Contains(visible, to); sees != want {
				rt.Fatalf("Sees(%v, %v) = %v, Visible says %v", from, to, sees, want)
			}
		}
	})
}

func visibleSet(t *testing.T, m *Map, from api.Cell, radius, depthRange int) map[api.Cell]bool {
	t.Helper()
	cells, err := m.Visible(from, radius, depthRange)
	if err != nil {
		t.Fatal(err)
	}
	set := make(map[api.Cell]bool, len(cells))
	for _, c := range cells {
		set[c] = true
	}
	return set
}

func TestVisibleOpenRoomAndWalls(t *testing.T) {
	m := NewMap(FuncSource(func(c api.Cell) Terrain {
		if c.Depth != 0 || Distance(c.Hex, api.Hex{}) > 3 {
			return rock
		}
		return air
	}), Prisms{})
	seen := visibleSet(t, m, cell(0, 0, 0), 6, 0)
	for _, c := range m.CellsIn(api.Hex{}, 3, 0, 0) {
		if !seen[c] {
			t.Errorf("room cell %v not visible", c)
		}
	}
	for _, c := range m.CellsIn(api.Hex{}, 4, 0, 0) {
		if Distance(c.Hex, api.Hex{}) == 4 && !seen[c] {
			t.Errorf("wall cell %v not visible", c)
		}
	}
	for _, c := range m.CellsIn(api.Hex{}, 6, 0, 0) {
		if Distance(c.Hex, api.Hex{}) >= 5 && seen[c] {
			t.Errorf("cell %v behind the wall is visible", c)
		}
	}
}

func TestVisibleShaftAndRock(t *testing.T) {
	m := NewMap(FuncSource(func(c api.Cell) Terrain {
		if c.Depth < 1 {
			return air
		}
		return rock
	}), Prisms{})
	dig(t, m, cell(1, 0, 1), cell(1, 0, 2), cell(1, 0, 3), cell(1, 0, 4))
	dig(t, m, cell(-3, 0, 4))
	seen := visibleSet(t, m, cell(0, 0, 0), 4, 4)
	for depth := 1; depth <= 4; depth++ {
		if !seen[cell(1, 0, depth)] {
			t.Errorf("shaft cell at depth %d not visible", depth)
		}
	}
	if seen[cell(-3, 0, 4)] {
		t.Error("tunnel through solid rock is visible")
	}
}

func TestVisibleRampDown(t *testing.T) {
	m := NewMap(FuncSource(func(c api.Cell) Terrain {
		floor := 1 + max(0, c.Hex.Q)
		if c.Depth >= floor {
			return rock
		}
		return air
	}), Prisms{})
	seen := visibleSet(t, m, cell(0, 0, 0), 4, 4)
	for q := 1; q <= 4; q++ {
		if !seen[cell(q, 0, q)] {
			t.Errorf("ramp cell %v not visible", cell(q, 0, q))
		}
	}
}

func TestVisibleSortedAndStable(t *testing.T) {
	m := floorWorld()
	raise(t, m, cell(1, 0, 0), cell(-2, 1, 0))
	first, err := m.Visible(cell(0, 0, 0), 5, 2)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.IsSortedFunc(first, CompareCells) {
		t.Error("result is not sorted by Depth, R, Q")
	}
	again, err := m.Visible(cell(0, 0, 0), 5, 2)
	if err != nil || !slices.Equal(first, again) {
		t.Fatalf("repeat call differs: %v", err)
	}
}

func BenchmarkVisible(b *testing.B) {
	m := floorWorld()
	for _, c := range m.CellsIn(api.Hex{Q: 2, R: -1}, 1, 0, 0) {
		_ = m.Set(c, rock)
	}
	b.ReportAllocs()
	for b.Loop() {
		if _, err := m.Visible(cell(0, 0, 0), 6, 6); err != nil {
			b.Fatal(err)
		}
	}
}

func TestVisibleRemembersUntilAnEdit(t *testing.T) {
	m := NewMap(FuncSource(func(c api.Cell) Terrain {
		if c.Depth != 0 {
			return rock
		}
		return air
	}), Prisms{})
	from := cell(0, 0, 0)
	first, err := m.Visible(from, 4, 0)
	if err != nil {
		t.Fatal(err)
	}
	want := slices.Clone(first)
	first[0] = cell(9, 9, 9)
	again, err := m.Visible(from, 4, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(again, want) {
		t.Fatal("a remembered result changed when the caller edited its copy")
	}

	wall := cell(1, 0, 0)
	if !slices.Contains(again, cell(3, 0, 0)) {
		t.Fatal("setup: the open room should show (3,0)")
	}
	if err := m.Set(wall, rock); err != nil {
		t.Fatal(err)
	}
	after, err := m.Visible(from, 4, 0)
	if err != nil {
		t.Fatal(err)
	}
	fresh, err := m.computeVisible(from, 4, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(after, fresh) {
		t.Error("after an edit, Visible differs from a fresh computation")
	}
	if slices.Contains(after, cell(3, 0, 0)) {
		t.Error("the new wall at (1,0) should hide (3,0)")
	}
}
