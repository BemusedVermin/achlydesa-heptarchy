package hexmap

import "github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"

// Bounds is the box of cells a world holds: hexes in odd-r offset columns and rows (the layout of
// [FromOffsetOddR], so a width is a column count), and depths, all inclusive.
type Bounds struct {
	MinCol, MaxCol, MinRow, MaxRow int
	MinDepth, MaxDepth             int
}

// Contains reports whether c lies inside the bounds.
func (b Bounds) Contains(c api.Cell) bool {
	col, row := ToOffsetOddR(c.Hex)
	return col >= b.MinCol && col <= b.MaxCol &&
		row >= b.MinRow && row <= b.MaxRow &&
		c.Depth >= b.MinDepth && c.Depth <= b.MaxDepth
}
