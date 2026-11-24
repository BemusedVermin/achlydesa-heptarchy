package sim

import (
	"fmt"
	"maps"
	"slices"
	"sync"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// Handler applies a command. It publishes events on w.Bus and returns the
// command's time cost in ticks. An invalid player command publishes
// api.Rejected and returns 0, nil; errors are for bugs only.
type Handler[C api.Command] func(w *World, cmd C) (cost int64, err error)

// Commands routes each command kind to its handler. Safe for concurrent use;
// the lock is never held while a handler runs.
type Commands struct {
	mu       sync.RWMutex
	handlers map[string]func(w *World, cmd api.Command) (int64, error)
}

// NewCommands returns an empty registry.
func NewCommands() *Commands {
	return &Commands{handlers: make(map[string]func(*World, api.Command) (int64, error))}
}

// Handle registers h for command type C and registers C with w.Codec.
// A duplicate kind is an error.
func Handle[C api.Command](w *World, h Handler[C]) error {
	var zero C
	kind := zero.CommandKind()
	w.Commands.mu.Lock()
	defer w.Commands.mu.Unlock()
	if _, dup := w.Commands.handlers[kind]; dup {
		return fmt.Errorf("sim: duplicate command handler for %q", kind)
	}
	if err := api.RegisterCommand[C](w.Codec); err != nil {
		return fmt.Errorf("sim: handle %q: %w", kind, err)
	}
	w.Commands.handlers[kind] = func(w *World, cmd api.Command) (int64, error) {
		typed, ok := cmd.(C)
		if !ok {
			return 0, fmt.Errorf("sim: command %q has type %T, handler wants %T", kind, cmd, zero)
		}
		return h(w, typed)
	}
	return nil
}

// Dispatch runs the handler for cmd's kind. An unknown kind is an error.
func (c *Commands) Dispatch(w *World, cmd api.Command) (cost int64, err error) {
	c.mu.RLock()
	handler, ok := c.handlers[cmd.CommandKind()]
	c.mu.RUnlock()
	if !ok {
		return 0, fmt.Errorf("sim: no handler for command kind %q", cmd.CommandKind())
	}
	return handler(w, cmd)
}

// Kinds returns the registered command kinds, sorted.
func (c *Commands) Kinds() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return slices.Sorted(maps.Keys(c.handlers))
}
