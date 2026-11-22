package hexmap

import (
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
	route, ok, err := m.Path(cell(0, 0, 0), cell(3, 0, 3), DefaultStepRules(), nil, 1000)
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

	if _, ok, err := m.Path(low, top, DefaultStepRules(), nil, 2000); err != nil || ok {
		t.Errorf("climbing a 3-cell ledge: ok=%v err=%v; want refused", ok, err)
	}
	route, ok, err := m.Path(top, low, DefaultStepRules(), nil, 2000)
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
	route, ok, err := m.Path(cell(0, 0, 0), cell(50, 0, 0), DefaultStepRules(), nil, 200)
	if err != nil || ok || route != nil {
		t.Errorf("Path = %v, %v, %v; want no route", route, ok, err)
	}

	solid := NewMap(2, "solid", FuncSource(func(api.Cell) Terrain { return rock }), Prisms{})
	if _, ok, err := solid.Path(cell(0, 0, 0), cell(5, 0, 0), DefaultStepRules(), nil, 100); err != nil || ok {
		t.Errorf("all-solid map: ok=%v err=%v", ok, err)
	}
}

func TestPathDeterministic(t *testing.T) {
	m := floorWorld()
	raise(t, m, cell(1, 0, 0), cell(1, -1, 0), cell(2, -1, 0))
	first, ok, err := m.Path(cell(0, 0, 0), cell(4, 0, 0), DefaultStepRules(), nil, 1000)
	if err != nil || !ok {
		t.Fatal(ok, err)
	}
	for range 100 {
		again, _, _ := m.Path(cell(0, 0, 0), cell(4, 0, 0), DefaultStepRules(), nil, 1000)
		if len(again) != len(first) {
			t.Fatal("route length changed")
		}
		for i := range first {
			if first[i] != again[i] {
				t.Fatalf("route differs at %d", i)
			}
		}
	}
}
