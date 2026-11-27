package hexmap

import (
	"slices"

	hex "github.com/gravitton/hexagon"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// Visible returns every cell within radius hexes and depthRange depths of from that a straight line from from's
// center reaches, sorted by Depth, R, Q. from itself is included.
//
// A cell is visible when the line crosses no opaque cell other than the target, so walls are visible and what is
// behind them is not. Depth is interpolated linearly along the hex line; where it is fractional, the line is clear
// if either the sequence of cells above or the sequence below is clear.
func (m *Map) Visible(from api.Cell, radius, depthRange int) ([]api.Cell, error) {
	key := sightKey{from: from, radius: radius, depthRange: depthRange}
	if cells, ok := m.sights[key]; ok {
		return slices.Clone(cells), nil
	}
	cells, err := m.computeVisible(from, radius, depthRange)
	if err != nil {
		return nil, err
	}
	if m.sights == nil || len(m.sights) >= maxSights {
		m.sights = map[sightKey][]api.Cell{}
	}
	m.sights[key] = slices.Clone(cells)
	return cells, nil
}

// Sees reports whether to is in [Map.Visible] from from, without working out every visible cell: it is
// cheaper when only a few cells are asked about, as when an agent looks at the bodies around it.
func (m *Map) Sees(from, to api.Cell, radius, depthRange int) (bool, error) {
	if Distance(from.Hex, to.Hex) > radius || max(from.Depth-to.Depth, to.Depth-from.Depth) > depthRange {
		return false, nil
	}
	return m.clearLine(from, to)
}

// sightKey is what a Visible result depends on, besides the terrain.
type sightKey struct {
	from               api.Cell
	radius, depthRange int
}

// maxSights bounds the Visible results a map remembers; past it the memory starts over.
const maxSights = 4096

func (m *Map) computeVisible(from api.Cell, radius, depthRange int) ([]api.Cell, error) {
	candidates := m.CellsIn(from.Hex, radius, from.Depth-depthRange, from.Depth+depthRange)
	out := make([]api.Cell, 0, len(candidates))
	for _, target := range candidates {
		visible, err := m.clearLine(from, target)
		if err != nil {
			return nil, err
		}
		if visible {
			out = append(out, target)
		}
	}
	return out, nil
}

// clearLine reports whether no opaque cell lies strictly between from and to along the sight line.
func (m *Map) clearLine(from, to api.Cell) (bool, error) {
	line := hex.Pt(from.Hex.Q, from.Hex.R).LineTo(hex.Pt(to.Hex.Q, to.Hex.R))
	steps := len(line) - 1
	dd := to.Depth - from.Depth
	if steps == 0 {
		return m.clearColumn(from, to)
	}
	// Cells at the floor / ceiling of the interpolated depth
	lowClear, highClear := true, true
	for i := 1; i < steps; i++ {
		num := from.Depth*steps + dd*i
		low := floorDiv(num, steps)
		high := low
		if num%steps != 0 {
			high = low + 1
		}
		at := api.Hex{Q: line[i].Q, R: line[i].R}
		lowOpaque, err := m.opaque(api.Cell{Hex: at, Depth: low})
		if err != nil {
			return false, err
		}
		highOpaque := lowOpaque
		if high != low {
			if highOpaque, err = m.opaque(api.Cell{Hex: at, Depth: high}); err != nil {
				return false, err
			}
		}
		lowClear = lowClear && !lowOpaque
		highClear = highClear && !highOpaque
		if !lowClear && !highClear {
			return false, nil
		}
	}
	return lowClear || highClear, nil
}

// clearColumn checks the cells strictly between two cells sharing a hex.
func (m *Map) clearColumn(from, to api.Cell) (bool, error) {
	lo, hi := min(from.Depth, to.Depth), max(from.Depth, to.Depth)
	for depth := lo + 1; depth < hi; depth++ {
		opaque, err := m.opaque(api.Cell{Hex: from.Hex, Depth: depth})
		if err != nil || opaque {
			return false, err
		}
	}
	return true, nil
}

func (m *Map) opaque(c api.Cell) (bool, error) {
	t, err := m.At(c)
	return t.Opaque, err
}

// Distance is the number of hex steps between two cells of a layer, ignoring depth.
func Distance(a, b api.Hex) int {
	return hex.Pt(a.Q, a.R).DistanceTo(hex.Pt(b.Q, b.R))
}

// CellDistance is the hex distance between two cells plus their difference in depth.
func CellDistance(a, b api.Cell) int {
	return Distance(a.Hex, b.Hex) + max(a.Depth-b.Depth, b.Depth-a.Depth)
}
