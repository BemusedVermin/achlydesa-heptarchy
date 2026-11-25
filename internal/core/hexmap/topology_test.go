package hexmap

import (
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

func cell(q, r, depth int) api.Cell {
	return api.Cell{Hex: api.Hex{Q: q, R: r}, Depth: depth}
}

func TestPrismsNeighbors(t *testing.T) {
	origin := cell(2, 3, 5)
	tests := []struct {
		dir  api.Dir
		want api.Cell
	}{
		{api.DirE, cell(3, 3, 5)},
		{api.DirNE, cell(3, 2, 5)},
		{api.DirNW, cell(2, 2, 5)},
		{api.DirW, cell(1, 3, 5)},
		{api.DirSW, cell(1, 4, 5)},
		{api.DirSE, cell(2, 4, 5)},
	}
	var topo Prisms
	for _, tt := range tests {
		got := topo.Neighbor(origin, tt.dir)
		if got.Cell != tt.want {
			t.Errorf("Neighbor(%v) = %+v, want %v on the same map", tt.dir, got, tt.want)
		}
	}
	if got := topo.Above(origin); got.Cell != cell(2, 3, 4) {
		t.Errorf("Above = %+v", got)
	}
	if got := topo.Below(origin); got.Cell != cell(2, 3, 6) {
		t.Errorf("Below = %+v", got)
	}
}

func TestPrismsOppositeDirectionsCancel(t *testing.T) {
	var topo Prisms
	origin := cell(-4, 7, 2)
	for dir := api.DirE; dir.Valid(); dir++ {
		opposite := (dir + 3) % api.DirCount
		there := topo.Neighbor(origin, dir).Cell
		if back := topo.Neighbor(there, opposite).Cell; back != origin {
			t.Errorf("dir %d then %d returned to %v, want %v", dir, opposite, back, origin)
		}
	}
}
