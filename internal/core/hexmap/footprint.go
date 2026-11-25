package hexmap

import "github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"

// Footprint is where one fixed place is open: the cells of the place that are not solid, in [CompareCells]
// order.
type Footprint struct {
	Place api.ContentID
	Open  []api.Cell
}
