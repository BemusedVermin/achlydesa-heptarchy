package sim

import (
	"fmt"
	"sync"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// ViewBuilder fills in one section of the view.
type ViewBuilder struct {
	Name  string
	After []string
	Build func(w *World, v *api.View) error
}

// Views holds the view builders. Safe for concurrent use; the lock is never
// held while a builder runs.
type Views struct {
	mu       sync.RWMutex
	builders map[string]ViewBuilder
}

// NewViews returns an empty registry.
func NewViews() *Views {
	return &Views{builders: make(map[string]ViewBuilder)}
}

// Register adds a builder. A duplicate name is an error.
func (v *Views) Register(b ViewBuilder) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	if b.Build == nil {
		return fmt.Errorf("sim: view builder %q has no Build func", b.Name)
	}
	if _, dup := v.builders[b.Name]; dup {
		return fmt.Errorf("sim: duplicate view builder %q", b.Name)
	}
	v.builders[b.Name] = b
	return nil
}

// Build runs the builders in sorted order (After, then name) over a fresh view
// whose Tick is w.Now.
func (v *Views) Build(w *World) (api.View, error) {
	v.mu.RLock()
	items := make([]orderItem, 0, len(v.builders))
	builders := make(map[string]ViewBuilder, len(v.builders))
	for name, b := range v.builders {
		items = append(items, orderItem{b.Name, b.After})
		builders[name] = b
	}
	v.mu.RUnlock()
	order, err := sortByAfter("view builder", items)
	if err != nil {
		return api.View{}, err
	}
	view := api.View{Tick: w.Now}
	for _, name := range order {
		if err := builders[name].Build(w, &view); err != nil {
			return api.View{}, fmt.Errorf("sim: view builder %q: %w", name, err)
		}
	}
	return view, nil
}
