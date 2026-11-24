package hexmap

import (
	"slices"
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

func TestPathDownRamp(t *testing.T) {
	// Floor descends one per hex going east.
	m := NewMap(1, "ramp", FuncSource(func(c api.Cell) Terrain {
		if c.Depth >= 1+max(0, c.Hex.Q) {
			return rock
		}
		return air
	}), Prisms{})
	route, ok, err := m.Path(cell(0, 0, 0), cell(3, 0, 3), defaultStepRules(), nil, 1000)
	if err != nil || !ok {
		t.Fatalf("Path = %v, %v, %v; want a route", route, ok, err)
	}
	if route[0] != cell(0, 0, 0) || route[len(route)-1] != cell(3, 0, 3) || len(route) != 4 {
		t.Errorf("route = %v", route)
	}
}

func TestPathLedgeAndDrop(t *testing.T) {
	// A three-cell ledge at q>=1: standing at depth -3 above rock from depth -2 down... model it as a floor that
	// is three cells higher east of q=0.
	high := FuncSource(func(c api.Cell) Terrain {
		floor := 1
		if c.Hex.Q >= 1 {
			floor = -2
		}
		if c.Depth >= floor {
			return rock
		}
		return air
	})
	m := NewMap(1, "ledge", high, Prisms{})
	top, low := cell(1, 0, -3), cell(0, 0, 0)

	if _, ok, err := m.Path(low, top, defaultStepRules(), nil, 2000); err != nil || ok {
		t.Errorf("climbing a 3-cell ledge: ok=%v err=%v; want refused", ok, err)
	}
	route, ok, err := m.Path(top, low, defaultStepRules(), nil, 2000)
	if err != nil || !ok {
		t.Fatalf("dropping down the ledge: %v %v", ok, err)
	}
	if route[len(route)-1] != low {
		t.Errorf("route = %v", route)
	}
}

func TestPathGivesUp(t *testing.T) {
	// Open infinite floor, target buried in rock: no route, but the search must still end.
	m := floorWorld()
	raise(t, m, cell(50, 0, 0))
	route, ok, err := m.Path(cell(0, 0, 0), cell(50, 0, 0), defaultStepRules(), nil, 200)
	if err != nil || ok || route != nil {
		t.Errorf("Path = %v, %v, %v; want no route", route, ok, err)
	}

	solid := NewMap(2, "solid", FuncSource(func(api.Cell) Terrain { return rock }), Prisms{})
	if _, ok, err := solid.Path(cell(0, 0, 0), cell(5, 0, 0), defaultStepRules(), nil, 100); err != nil || ok {
		t.Errorf("all-solid map: ok=%v err=%v", ok, err)
	}
}

func TestReachableWalksTheStepRulesWithinTheLimit(t *testing.T) {
	// Floor descends one per hex going east, so each cell east is one step down from the last.
	m := NewMap(1, "ramp", FuncSource(func(c api.Cell) Terrain {
		if c.Depth >= 1+max(0, c.Hex.Q) {
			return rock
		}
		return air
	}), Prisms{})
	reached, err := m.Reachable(cell(0, 0, 0), defaultStepRules(), 3, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := reached[cell(3, 0, 3)]; got != 3 {
		t.Errorf("steps to the ramp's third cell = %d, want 3", got)
	}
	if _, far := reached[cell(4, 0, 4)]; far {
		t.Error("a cell four steps away was reached with a limit of three")
	}
	if _, ok := reached[cell(0, 0, 0)]; !ok {
		t.Error("the start is not in the reached set")
	}
	only, err := m.Reachable(cell(0, 0, 0), defaultStepRules(), 3, func(c api.Cell) bool { return c.Hex.Q <= 1 })
	if err != nil {
		t.Fatal(err)
	}
	if _, past := only[cell(2, 0, 2)]; past {
		t.Error("allow did not stop the search at q=1")
	}
}

func TestPathDeterministic(t *testing.T) {
	m := floorWorld()
	raise(t, m, cell(1, 0, 0), cell(1, -1, 0), cell(2, -1, 0))
	first, ok, err := m.Path(cell(0, 0, 0), cell(4, 0, 0), defaultStepRules(), nil, 1000)
	if err != nil || !ok {
		t.Fatal(ok, err)
	}
	again, _, _ := m.Path(cell(0, 0, 0), cell(4, 0, 0), defaultStepRules(), nil, 1000)
	if !slices.Equal(first, again) {
		t.Fatalf("route changed between calls: %v, then %v", first, again)
	}
}
