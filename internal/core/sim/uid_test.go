package sim

import (
	"slices"
	"testing"

	"github.com/mlange-42/ark/ecs"
)

func TestAliveIsFalseForTheZeroEntity(t *testing.T) {
	w := NewWorld(1)
	if w.Alive(ecs.Entity{}) {
		t.Fatal("the zero entity must not be alive: callers rely on it instead of an IsZero check")
	}
}

func TestAliveFollowsDespawn(t *testing.T) {
	w := NewWorld(1)
	e := w.NewIdentified("template.test")
	if !w.Alive(e) {
		t.Fatal("a new entity is alive")
	}
	w.Despawn(e)
	if w.Alive(e) {
		t.Fatal("a despawned entity is not alive")
	}
}

func TestSortByUIDAndCompareUID(t *testing.T) {
	w := NewWorld(1)
	first := w.NewIdentified("template.test")
	second := w.NewIdentified("template.test")
	third := w.NewIdentified("template.test")

	tests := []struct {
		name string
		a, b ecs.Entity
		want int
	}{
		{"lower first", first, second, -1},
		{"equal", second, second, 0},
		{"higher first", third, first, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := w.CompareUID(tt.a, tt.b); got != tt.want {
				t.Fatalf("CompareUID = %d, want %d", got, tt.want)
			}
		})
	}

	got := w.SortByUID([]ecs.Entity{third, first, second})
	if want := []ecs.Entity{first, second, third}; !slices.Equal(got, want) {
		t.Fatalf("SortByUID = %v, want %v", got, want)
	}
}

func TestPutAddsThenOverwrites(t *testing.T) {
	type marker struct{ N int }
	w := NewWorld(1)
	e := w.NewIdentified("template.test")
	markers := ecs.NewMap[marker](w.ECS)

	tests := []struct {
		name string
		put  marker
		want marker
	}{
		{"adds when missing", marker{1}, marker{1}},
		{"overwrites when present", marker{2}, marker{2}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Put(w, e, tt.put)
			if !markers.Has(e) || *markers.Get(e) != tt.want {
				t.Fatalf("component = %v, want %v", markers.Get(e), tt.want)
			}
		})
	}
}
