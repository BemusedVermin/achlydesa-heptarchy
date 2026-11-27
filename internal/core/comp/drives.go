package comp

import "github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"

// Drives is what a body wants, keyed by drive ID: how unmet each drive is now, and the weight this
// individual puts on it. Weights are drawn once at spawn and never redrawn. Use the agents package
// rather than changing the maps.
type Drives struct {
	// Levels are 0 when a drive is met, and never below.
	Levels  map[api.ContentID]float64 `json:"levels"`
	Weights map[api.ContentID]float64 `json:"weights"`
}
