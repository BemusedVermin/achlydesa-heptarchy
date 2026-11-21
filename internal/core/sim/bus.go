package sim

import (
	"fmt"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// MaxCascade is the most events one Drain may deliver before it gives up.
const MaxCascade = 10_000

type subscriber struct {
	name string
	fn   func(w *World, e api.Event)
}

// Bus queues events and delivers them first in, first out. It is separate from
// Ark's own event system, which reports ECS changes.
type Bus struct {
	subscribers map[string][]subscriber
	queue       []api.Event
}

// NewBus returns an empty bus.
func NewBus() *Bus {
	return &Bus{subscribers: make(map[string][]subscriber)}
}

// Subscribe registers fn for events of type E. name identifies the subscriber
// in diagnostics. Subscribers of one kind run in the order they subscribed.
func Subscribe[E api.Event](b *Bus, name string, fn func(w *World, e E)) {
	var zero E
	kind := zero.EventKind()
	b.subscribers[kind] = append(b.subscribers[kind], subscriber{
		name: name,
		fn: func(w *World, e api.Event) {
			if typed, ok := e.(E); ok {
				fn(w, typed)
			}
		},
	})
}

// Publish queues e for the next Drain.
func (b *Bus) Publish(e api.Event) { b.queue = append(b.queue, e) }

// Drain delivers queued events, including ones published by subscribers during
// delivery, until the queue is empty. It returns every delivered event in
// delivery order, and an error once more than MaxCascade events are delivered
// in one drain. Events with no subscribers are still returned.
func (b *Bus) Drain(w *World) ([]api.Event, error) {
	var delivered []api.Event
	for len(b.queue) > 0 {
		if len(delivered) >= MaxCascade {
			next := b.queue[0].EventKind()
			b.queue = nil
			return delivered, fmt.Errorf("sim: event cascade exceeded %d events (next was %q)", MaxCascade, next)
		}
		e := b.queue[0]
		b.queue = b.queue[1:]
		delivered = append(delivered, e)
		for _, sub := range b.subscribers[e.EventKind()] {
			sub.fn(w, e)
		}
	}
	b.queue = nil
	return delivered, nil
}
