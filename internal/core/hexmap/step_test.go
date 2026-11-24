package hexmap

import (
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// floorWorld is open air above depth 1 and rock from depth 1 down, so a body at depth 0 stands on solid ground.
func floorWorld() *Map {
	return NewMap(1, "floor", FuncSource(func(c api.Cell) Terrain {
		if c.Depth >= 1 {
			return rock
		}
		return air
	}), Prisms{})
}

func dig(t *testing.T, m *Map, cells ...api.Cell) {
	t.Helper()
	for _, c := range cells {
		if err := m.Set(c, air); err != nil {
			t.Fatal(err)
		}
	}
}

func raise(t *testing.T, m *Map, cells ...api.Cell) {
	t.Helper()
	for _, c := range cells {
		if err := m.Set(c, rock); err != nil {
			t.Fatal(err)
		}
	}
}

// defaultStepRules are the limits most tests step under; content sets the game's own.
func defaultStepRules() StepRules { return StepRules{MaxStepUp: 1, MaxStepDown: 1, MaxDrop: 8} }

// shaft is the cells east of the origin from depth 1 down to depth, opened to fall into.
func shaft(depth int) []api.Cell {
	cells := make([]api.Cell, 0, depth)
	for d := 1; d <= depth; d++ {
		cells = append(cells, cell(1, 0, d))
	}
	return cells
}

func TestStep(t *testing.T) {
	from := cell(0, 0, 0)
	east := cell(1, 0, 0)
	rules := defaultStepRules()

	tests := []struct {
		name   string
		raise  []api.Cell // cells made solid before the step
		dig    []api.Cell // cells opened before the step
		rules  StepRules
		want   Step
		wantOK bool
	}{
		{
			name:   "level",
			want:   Step{To: Link{Cell: east}, Kind: StepLevel},
			wantOK: true,
		},
		{
			name:   "up one",
			raise:  []api.Cell{east},
			want:   Step{To: Link{Cell: cell(1, 0, -1)}, Kind: StepUp},
			wantOK: true,
		},
		{
			name:  "up two refused",
			raise: []api.Cell{east, cell(1, 0, -1)},
		},
		{
			name:   "up two allowed with a higher limit",
			raise:  []api.Cell{east, cell(1, 0, -1)},
			rules:  StepRules{MaxStepUp: 2, MaxStepDown: 1, MaxDrop: 8},
			want:   Step{To: Link{Cell: cell(1, 0, -2)}, Kind: StepUp},
			wantOK: true,
		},
		{
			name:  "up refused without headroom",
			raise: []api.Cell{east, cell(0, 0, -1)},
		},
		{
			name:   "down one",
			dig:    []api.Cell{cell(1, 0, 1)},
			want:   Step{To: Link{Cell: cell(1, 0, 1)}, Kind: StepDown},
			wantOK: true,
		},
		{
			name:   "drop of three",
			dig:    shaft(3),
			want:   Step{To: Link{Cell: cell(1, 0, 3)}, Kind: StepDrop, Fall: 3},
			wantOK: true,
		},
		{
			name:   "drop of eight allowed",
			dig:    shaft(8),
			want:   Step{To: Link{Cell: cell(1, 0, 8)}, Kind: StepDrop, Fall: 8},
			wantOK: true,
		},
		{
			name: "drop past MaxDrop refused",
			dig:  shaft(9),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := floorWorld()
			raise(t, m, tt.raise...)
			dig(t, m, tt.dig...)
			r := tt.rules
			if r == (StepRules{}) {
				r = rules
			}
			got, ok, err := m.Step(from, api.DirE, r)
			if err != nil {
				t.Fatal(err)
			}
			if ok != tt.wantOK || (ok && got != tt.want) {
				t.Errorf("Step = %+v, %v; want %+v, %v", got, ok, tt.want, tt.wantOK)
			}
		})
	}
}

func TestStepRejectsBadDirection(t *testing.T) {
	if _, ok, err := floorWorld().Step(cell(0, 0, 0), api.Dir(9), defaultStepRules()); ok || err == nil {
		t.Errorf("Step with a bad direction = %v, %v; want an error", ok, err)
	}
}

func TestStepFollowsLinksToOtherMaps(t *testing.T) {
	m := NewMap(1, "linked", FuncSource(func(api.Cell) Terrain { return rock }), portalTopology{})
	got, ok, err := m.Step(cell(0, 0, 0), api.DirE, defaultStepRules())
	want := Step{To: Link{Map: 7, Cell: cell(40, 0, 0)}, Kind: StepLevel}
	if err != nil || !ok || got != want {
		t.Errorf("Step = %+v, %v, %v; want %+v", got, ok, err, want)
	}
}

// portalTopology sends every eastward step to map 7.
type portalTopology struct{ Prisms }

func (p portalTopology) Neighbor(at api.Cell, dir api.Dir) Link {
	if dir == api.DirE {
		return Link{Map: 7, Cell: cell(40, 0, 0)}
	}
	return p.Prisms.Neighbor(at, dir)
}

// TestStepKeepsOverlappingTunnelsApart crosses two tunnels over the same hexes at depths 0 and 4.
func TestStepKeepsOverlappingTunnelsApart(t *testing.T) {
	inTunnel := func(c api.Cell) bool {
		return c.Hex.R == 0 && c.Hex.Q >= 0 && c.Hex.Q <= 5 && (c.Depth == 0 || c.Depth == 4)
	}
	m := NewMap(1, "tunnels", FuncSource(func(c api.Cell) Terrain {
		if inTunnel(c) {
			return air
		}
		return rock
	}), Prisms{})

	for _, depth := range []int{0, 4} {
		for q := 0; q <= 5; q++ {
			from := cell(q, 0, depth)
			for dir := api.DirE; dir.Valid(); dir++ {
				step, ok, err := m.Step(from, dir, defaultStepRules())
				if err != nil {
					t.Fatal(err)
				}
				if !ok {
					continue
				}
				if step.To.Cell.Depth != depth || step.Kind != StepLevel || !inTunnel(step.To.Cell) {
					t.Errorf("step from %v dir %d left the tunnel: %+v", from, dir, step)
				}
			}
		}
	}
}
