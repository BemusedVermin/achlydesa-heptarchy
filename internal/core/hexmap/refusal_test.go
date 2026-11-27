package hexmap

import (
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

func TestRefusal(t *testing.T) {
	east := cell(1, 0, 0)
	tests := []struct {
		name  string
		raise []api.Cell
		dig   []api.Cell
		want  Refusal
	}{
		{name: "wall", raise: []api.Cell{east, cell(1, 0, -1), cell(1, 0, -2)}, want: Refusal{Kind: RefusedSolid}},
		{
			name:  "ledge too high",
			raise: []api.Cell{east, cell(1, 0, -1)},
			want:  Refusal{Kind: RefusedTooHigh, Height: 2},
		},
		{name: "no headroom", raise: []api.Cell{east, cell(0, 0, -1)}, want: Refusal{Kind: RefusedSolid}},
		{name: "drop too far", dig: shaft(9), want: Refusal{Kind: RefusedTooFar}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := floorWorld()
			raise(t, m, tt.raise...)
			dig(t, m, tt.dig...)
			got, err := m.Refusal(cell(0, 0, 0), api.DirE, defaultStepRules())
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("Refusal = %+v, want %+v", got, tt.want)
			}
		})
	}
}
