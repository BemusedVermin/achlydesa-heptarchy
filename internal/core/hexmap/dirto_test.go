package hexmap

import (
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

func TestCellDistanceAddsTheDepthDifference(t *testing.T) {
	tests := []struct {
		name string
		a, b [3]int
		want int
	}{
		{"same cell", [3]int{0, 0, 0}, [3]int{0, 0, 0}, 0},
		{"flat", [3]int{0, 0, 0}, [3]int{2, 0, 0}, 2},
		{"depth only", [3]int{0, 0, 0}, [3]int{0, 0, -3}, 3},
		{"both", [3]int{0, 0, 1}, [3]int{2, 0, 4}, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, b := cell(tt.a[0], tt.a[1], tt.a[2]), cell(tt.b[0], tt.b[1], tt.b[2])
			if got := CellDistance(a, b); got != tt.want {
				t.Fatalf("CellDistance = %d, want %d", got, tt.want)
			}
			if got := CellDistance(b, a); got != tt.want {
				t.Fatalf("CellDistance reversed = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestDirTo(t *testing.T) {
	from := cell(0, 0, 0)
	rules := defaultStepRules()

	tests := []struct {
		name   string
		raise  []api.Cell
		to     [3]int
		wantOK bool
	}{
		{"neighbour on open ground", nil, [3]int{1, 0, 0}, true},
		{"too far for one step", nil, [3]int{5, 0, 0}, false},
		{"blocked neighbour", []api.Cell{cell(1, 0, 0)}, [3]int{1, 0, 0}, false},
		{"same cell", nil, [3]int{0, 0, 0}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := floorWorld()
			raise(t, m, tt.raise...)
			to := cell(tt.to[0], tt.to[1], tt.to[2])
			dir, ok, err := m.DirTo(from, to, rules)
			if err != nil {
				t.Fatal(err)
			}
			if ok != tt.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tt.wantOK)
			}
			if !ok {
				return
			}
			step, stepped, err := m.Step(from, dir, rules)
			if err != nil || !stepped || step.To.Cell != to {
				t.Fatalf("Step(%v) = %+v, %v, %v; want a step to %v", dir, step, stepped, err, to)
			}
		})
	}
}
