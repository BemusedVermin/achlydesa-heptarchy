package sim

import (
	"strings"
	"testing"
)

type pingEvent struct{ N int }

func (pingEvent) EventKind() string { return "test.ping" }

type pongEvent struct{}

func (pongEvent) EventKind() string { return "test.pong" }

type loopEvent struct{}

func (loopEvent) EventKind() string { return "test.loop" }

func TestBusDeliversFIFOIncludingCascades(t *testing.T) {
	w := NewWorld(1)
	var seen []string
	Subscribe(w.Bus, "first", func(w *World, e pingEvent) {
		seen = append(seen, "first")
		if e.N == 1 {
			w.Bus.Publish(pongEvent{})
		}
	})
	Subscribe(w.Bus, "second", func(_ *World, _ pingEvent) { seen = append(seen, "second") })
	Subscribe(w.Bus, "pong", func(_ *World, _ pongEvent) { seen = append(seen, "pong") })

	w.Bus.Publish(pingEvent{N: 1})
	w.Bus.Publish(pingEvent{N: 2})
	delivered, err := w.Bus.Drain(w)
	if err != nil {
		t.Fatal(err)
	}

	wantSeen := "first,second,first,second,pong"
	if got := strings.Join(seen, ","); got != wantSeen {
		t.Errorf("subscriber order = %s, want %s", got, wantSeen)
	}
	wantKinds := []string{"test.ping", "test.ping", "test.pong"}
	if len(delivered) != len(wantKinds) {
		t.Fatalf("delivered %d events, want %d", len(delivered), len(wantKinds))
	}
	for i, kind := range wantKinds {
		if got := delivered[i].EventKind(); got != kind {
			t.Errorf("delivered[%d] = %s, want %s", i, got, kind)
		}
	}
}

func TestBusReturnsEventsWithoutSubscribers(t *testing.T) {
	w := NewWorld(1)
	w.Bus.Publish(pongEvent{})
	delivered, err := w.Bus.Drain(w)
	if err != nil || len(delivered) != 1 {
		t.Fatalf("Drain() = %v, %v; want one event", delivered, err)
	}
}

func TestBusRunawayCascadeErrors(t *testing.T) {
	w := NewWorld(1)
	Subscribe(w.Bus, "forever", func(w *World, _ loopEvent) { w.Bus.Publish(loopEvent{}) })
	w.Bus.Publish(loopEvent{})
	delivered, err := w.Bus.Drain(w)
	if err == nil {
		t.Fatal("Drain() returned nil error for an endless cascade")
	}
	if len(delivered) != MaxCascade {
		t.Errorf("delivered %d events, want %d", len(delivered), MaxCascade)
	}
	// The queue is cleared, so the world is usable afterwards.
	if _, err := w.Bus.Drain(w); err != nil {
		t.Errorf("Drain() after a failed drain: %v", err)
	}
}
