package hexmap

import "github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"

// Footprint is what a fixed place shows the generated ground: the cells its author marked as openings, in
// [CompareCells] order.
type Footprint struct {
	Place    api.ContentID
	Openings []api.Cell
}
