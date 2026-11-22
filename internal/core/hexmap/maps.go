package hexmap

import (
	"fmt"
	"slices"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// Maps is a registry of maps by ID. It is owned by the world and changed only by its goroutine.
type Maps struct {
	byID map[api.MapID]*Map
}

// NewMaps returns an empty registry.
func NewMaps() *Maps { return &Maps{byID: map[api.MapID]*Map{}} }

// Add registers a map under its own ID. IDs are unique and not zero.
func (ms *Maps) Add(m *Map) error {
	if m.ID() == 0 {
		return fmt.Errorf("hexmap: map %q has no ID", m.Name())
	}
	if _, dup := ms.byID[m.ID()]; dup {
		return fmt.Errorf("hexmap: map ID %d already registered", m.ID())
	}
	ms.byID[m.ID()] = m
	return nil
}

// Get returns the map with the ID, or nil.
func (ms *Maps) Get(id api.MapID) *Map { return ms.byID[id] }

// IDs lists the registered map IDs in order.
func (ms *Maps) IDs() []api.MapID {
	ids := make([]api.MapID, 0, len(ms.byID))
	for id := range ms.byID {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	return ids
}
