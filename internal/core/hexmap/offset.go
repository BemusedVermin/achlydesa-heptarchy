package hexmap

import (
	geom "github.com/gravitton/geometry"
	hex "github.com/gravitton/hexagon"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// FromOffsetOddR converts a column and row of an odd-r offset layout (pointy-top rows, the odd ones
// shifted right by half a hex) to a hex. Authored map rows use this layout, one character per cell.
func FromOffsetOddR(col, row int) api.Hex {
	h := hex.OffsetOddR.From(geom.Pt(col, row))
	return api.Hex{Q: h.Q, R: h.R}
}
