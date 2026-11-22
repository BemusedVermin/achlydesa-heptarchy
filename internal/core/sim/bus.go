package sim

import (
	"fmt"
	"sync"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// MaxCascade is the most events one Drain may deliver before it gives up.
const MaxCascade = 10_000

type subscriber struct {
	name string
	fn   func(w *World, e api.Event)
}

// Bus queues events and delivers them first in, first out. It is separate from
// Ark's own event system, which reports ECS changes. Safe for concurrent use:
// the lock guards the queue and subscriber lists and is never held while a
// subscriber runs.
type Bus struct {
	mu          sync.Mutex
	draining    bool
	subscribers map[string][]subscriber
	all         []subscriber
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
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subscribers[kind] = append(b.subscribers[kind], subscriber{
		name: name,
		fn: func(w *World, e api.Event) {
			if typed, ok := e.(E); ok {
				fn(w, typed)
			}
		},
	})
}

// SubscribeAll registers fn for every event. It runs after the subscribers of the event's own
// kind, in the order the all-event subscribers subscribed. Games use it to build the message log.
func (b *Bus) SubscribeAll(name string, fn func(w *World, e api.Event)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.all = append(b.all, subscriber{name: name, fn: fn})
}

// Publish queues e for the next Drain.
func (b *Bus) Publish(e api.Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.queue = append(b.queue, e)
}

// Drain delivers queued events, including ones published by subscribers during
// delivery, until the queue is empty. It returns every delivered event in
// delivery order, and an error once more than MaxCascade events are delivered
// in one drain. Events with no subscribers are still returned.
//
// Only one drain runs at a time. A call made while another is in progress (from
// another goroutine, or from a subscriber) returns nil, nil at once; the drain
// already running delivers the queued events.
func (b *Bus) Drain(w *World) ([]api.Event, error) {
	b.mu.Lock()
	if b.draining {
		b.mu.Unlock()
		return nil, nil
	}
	b.draining = true
	b.mu.Unlock()

	var delivered []api.Event
	for {
		b.mu.Lock()
		if len(b.queue) == 0 {
			b.queue = nil
			b.draining = false
			b.mu.Unlock()
			return delivered, nil
		}
		if len(delivered) >= MaxCascade {
			next := b.queue[0].EventKind()
			b.queue = nil
			b.draining = false
			b.mu.Unlock()
			return delivered, fmt.Errorf("sim: event cascade exceeded %d events (next was %q)", MaxCascade, next)
		}
		e := b.queue[0]
		b.queue = b.queue[1:]
		subs := b.subscribers[e.EventKind()] // append-only, so these snapshots stay valid
		all := b.all
		b.mu.Unlock()

		delivered = append(delivered, e)
		for _, sub := range subs {
			sub.fn(w, e)
		}
		for _, sub := range all {
			sub.fn(w, e)
		}
	}
}
