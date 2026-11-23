package hexmap

import (
	"slices"
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

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
	// An open room of radius 3 at depth 0, walled by rock beyond that, over a rock floor.
	m := NewMap(1, "room", FuncSource(func(c api.Cell) Terrain {
		if c.Depth != 0 || hexDistance(c.Hex, api.Hex{}) > 3 {
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
		if hexDistance(c.Hex, api.Hex{}) == 4 && !seen[c] {
			t.Errorf("wall cell %v not visible", c)
		}
	}
	for _, c := range m.CellsIn(api.Hex{}, 6, 0, 0) {
		if hexDistance(c.Hex, api.Hex{}) >= 5 && seen[c] {
			t.Errorf("cell %v behind the wall is visible", c)
		}
	}
}

func TestVisibleShaftAndRock(t *testing.T) {
	// Solid rock from depth 1; a shaft at (1,0) down to depth 4; a tunnel at depth 4 elsewhere.
	m := NewMap(1, "shaft", FuncSource(func(c api.Cell) Terrain {
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
	// A ramp descending east: floor rises to depth 1 at q<=0, then steps down one per hex.
	m := NewMap(1, "ramp", FuncSource(func(c api.Cell) Terrain {
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
	if !slices.IsSortedFunc(first, compareCells) {
		t.Error("result is not sorted by Depth, R, Q")
	}
	for range 100 {
		again, err := m.Visible(cell(0, 0, 0), 5, 2)
		if err != nil || !slices.Equal(first, again) {
			t.Fatalf("repeat call differs: %v", err)
		}
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

func hexDistance(a, b api.Hex) int {
	dq, dr := a.Q-b.Q, a.R-b.R
	return (absInt(dq) + absInt(dr) + absInt(dq+dr)) / 2
}

func TestVisibleRemembersUntilAnEdit(t *testing.T) {
	m := NewMap(1, "room", FuncSource(func(c api.Cell) Terrain {
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
	first[0] = cell(9, 9, 9) // a caller changing its copy must not change what the map remembers
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
