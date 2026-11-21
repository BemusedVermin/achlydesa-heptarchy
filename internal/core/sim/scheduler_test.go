package sim

import (
	"errors"
	"slices"
	"strings"
	"testing"
)

func noop(*World) error { return nil }

func TestSchedulerOrderIsStable(t *testing.T) {
	want := []string{"a.a", "b.b", "c.c", "d.d"}
	for run := range 100 {
		w := NewWorld(1)
		// Register in a different order each run; the result must not change.
		systems := []System{
			{Name: "d.d", After: []string{"b.b", "c.c"}, Run: noop},
			{Name: "c.c", After: []string{"a.a"}, Run: noop},
			{Name: "b.b", After: []string{"a.a"}, Run: noop},
			{Name: "a.a", Run: noop},
		}
		if run%2 == 1 {
			slices.Reverse(systems)
		}
		for _, sys := range systems {
			if err := w.Systems.Register(sys); err != nil {
				t.Fatal(err)
			}
		}
		if err := w.Systems.Freeze(); err != nil {
			t.Fatal(err)
		}
		if got := w.Systems.Order(PerTurn); !slices.Equal(got, want) {
			t.Fatalf("run %d: order = %v, want %v", run, got, want)
		}
	}
}

func TestSchedulerFreezeErrors(t *testing.T) {
	tests := []struct {
		name    string
		systems []System
		wantIn  []string
	}{
		{
			name: "cycle",
			systems: []System{
				{Name: "x.one", After: []string{"x.two"}, Run: noop},
				{Name: "x.two", After: []string{"x.one"}, Run: noop},
			},
			wantIn: []string{"x.one", "x.two"},
		},
		{
			name:    "unknown after",
			systems: []System{{Name: "x.one", After: []string{"x.ghost"}, Run: noop}},
			wantIn:  []string{"x.one", "x.ghost"},
		},
		{
			name: "cross cadence",
			systems: []System{
				{Name: "x.one", Cadence: PerDay, Run: noop},
				{Name: "x.two", Cadence: PerTurn, After: []string{"x.one"}, Run: noop},
			},
			wantIn: []string{"x.one", "x.two"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := NewWorld(1)
			for _, sys := range tt.systems {
				if err := w.Systems.Register(sys); err != nil {
					t.Fatal(err)
				}
			}
			err := w.Systems.Freeze()
			if err == nil {
				t.Fatal("Freeze() = nil, want error")
			}
			for _, name := range tt.wantIn {
				if !strings.Contains(err.Error(), name) {
					t.Errorf("error %q does not name %q", err, name)
				}
			}
		})
	}
}

func TestSchedulerLifecycleErrors(t *testing.T) {
	w := NewWorld(1)
	if err := w.Systems.Run(w, PerTurn); err == nil {
		t.Error("Run before Freeze should error")
	}
	if err := w.Systems.Register(System{Name: "a.a", Run: noop}); err != nil {
		t.Fatal(err)
	}
	if err := w.Systems.Register(System{Name: "a.a", Run: noop}); err == nil {
		t.Error("duplicate name should error")
	}
	if err := w.Systems.Freeze(); err != nil {
		t.Fatal(err)
	}
	if err := w.Systems.Register(System{Name: "b.b", Run: noop}); err == nil {
		t.Error("Register after Freeze should error")
	}
}

func TestSchedulerRunDrainsBusAndStopsOnError(t *testing.T) {
	w := NewWorld(1)
	var trace []string
	Subscribe(w.Bus, "trace", func(_ *World, _ pongEvent) { trace = append(trace, "event") })
	boom := errors.New("boom")
	register := func(sys System) {
		t.Helper()
		if err := w.Systems.Register(sys); err != nil {
			t.Fatal(err)
		}
	}
	register(System{Name: "a.emit", Run: func(w *World) error {
		trace = append(trace, "a")
		w.Bus.Publish(pongEvent{})
		return nil
	}})
	register(System{Name: "b.fail", After: []string{"a.emit"}, Run: func(*World) error {
		trace = append(trace, "b")
		return boom
	}})
	register(System{Name: "c.never", After: []string{"b.fail"}, Run: func(*World) error {
		trace = append(trace, "c")
		return nil
	}})
	if err := w.Systems.Freeze(); err != nil {
		t.Fatal(err)
	}
	err := w.Systems.Run(w, PerTurn)
	if !errors.Is(err, boom) {
		t.Fatalf("Run() = %v, want wrapped boom", err)
	}
	if got := strings.Join(trace, ","); got != "a,event,b" {
		t.Errorf("trace = %s, want a,event,b", got)
	}
}
