package rng

import (
	"slices"
	"testing"

	"pgregory.net/rapid"
)

func draw(s *Streams, name string, n int) []uint64 {
	r := s.Get(name)
	out := make([]uint64, n)
	for i := range out {
		out[i] = r.Uint64()
	}
	return out
}

func TestSameSeedAndNameReproduce(t *testing.T) {
	if !slices.Equal(draw(New(7), "combat", 1000), draw(New(7), "combat", 1000)) {
		t.Fatal("same seed and name gave different sequences")
	}
}

func TestCreationOrderIrrelevant(t *testing.T) {
	ab := New(7)
	a1, b1 := draw(ab, "a", 100), draw(ab, "b", 100)
	ba := New(7)
	b2, a2 := draw(ba, "b", 100), draw(ba, "a", 100)
	if !slices.Equal(a1, a2) || !slices.Equal(b1, b2) {
		t.Fatal("creation order changed the sequences")
	}
}

func TestDifferentNamesDiffer(t *testing.T) {
	s := New(7)
	if slices.Equal(draw(s, "a", 100), draw(s, "b", 100)) {
		t.Fatal("different names gave identical sequences")
	}
}

func TestStateRestoreContinues(t *testing.T) {
	original := New(42)
	draw(original, "a", 10)
	draw(original, "b", 25)
	state, err := original.State()
	if err != nil {
		t.Fatal(err)
	}
	restored := New(42)
	if err := restored.Restore(state); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"a", "b"} {
		if !slices.Equal(draw(original, name, 100), draw(restored, name, 100)) {
			t.Fatalf("stream %q diverged after restore", name)
		}
	}
}

func TestRestoreRejectsBadState(t *testing.T) {
	if err := New(1).Restore(map[string][]byte{"a": {1, 2, 3}}); err == nil {
		t.Fatal("expected error for corrupt state")
	}
}

func TestDerived(t *testing.T) {
	s := New(9)
	first := s.Derived("procgen", 1)
	again := s.Derived("procgen", 1)
	other := s.Derived("procgen", 2)
	a, b, c := first.Uint64(), again.Uint64(), other.Uint64()
	if a != b {
		t.Fatal("same inputs did not reproduce")
	}
	if a == c {
		t.Fatal("different salt did not differ")
	}
	if state, _ := s.State(); len(state) != 0 {
		t.Fatal("Derived must not register a stream")
	}
}

func TestProperties(t *testing.T) {
	rapid.Check(t, func(t *rapid.T) {
		seed := rapid.Uint64().Draw(t, "seed")
		first := rapid.StringMatching(`[a-z]{1,8}`).Draw(t, "first")
		second := rapid.StringMatching(`[a-z]{1,8}`).Draw(t, "second")

		if !slices.Equal(draw(New(seed), first, 50), draw(New(seed), first, 50)) {
			t.Fatal("same seed and name diverged")
		}
		forward := New(seed)
		f1, f2 := draw(forward, first, 50), draw(forward, second, 50)
		backward := New(seed)
		b2, b1 := draw(backward, second, 50), draw(backward, first, 50)
		if !slices.Equal(f1, b1) || !slices.Equal(f2, b2) {
			t.Fatal("creation order changed the sequences")
		}
	})
}
