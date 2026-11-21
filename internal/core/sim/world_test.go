package sim

import "testing"

func TestNewUIDCountsFromOne(t *testing.T) {
	w := NewWorld(1)
	for want := 1; want <= 3; want++ {
		if got := w.NewUID(); int(got) != want {
			t.Fatalf("NewUID() = %d, want %d", got, want)
		}
	}
	if got := w.NextUID(); got != 4 {
		t.Fatalf("NextUID() = %d, want 4", got)
	}
	w.SetNextUID(100)
	if got := w.NewUID(); got != 100 {
		t.Fatalf("NewUID() after SetNextUID(100) = %d", got)
	}
}

func TestNewWorldWiresRegistries(t *testing.T) {
	w := NewWorld(7)
	if w.ECS == nil || w.RNG == nil || w.Bus == nil || w.Systems == nil ||
		w.Commands == nil || w.Views == nil || w.Codec == nil {
		t.Fatalf("NewWorld left a field nil: %+v", w)
	}
	if w.RNG.Seed() != 7 {
		t.Fatalf("seed = %d, want 7", w.RNG.Seed())
	}
}
