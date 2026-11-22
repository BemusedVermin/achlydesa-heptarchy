package inventory_test

import (
	"maps"
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/inventory"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

func TestUsableCountsUnboundStacksAndStacksBoundToTheTarget(t *testing.T) {
	w := sim.NewWorld(1)
	holder, target, other := newHolder(w), newHolder(w), newHolder(w)
	for _, s := range []comp.ItemStack{
		{Item: "item.coal", Count: 2},
		{Item: "item.keepsake", Count: 1, BoundTo: w.UIDOf(target)},
		{Item: "item.keepsake", Count: 4, BoundTo: w.UIDOf(other)},
	} {
		inventory.Add(w, holder, s)
	}
	tests := []struct {
		name   string
		target ecs.Entity
		want   map[string]int
	}{
		{"no target counts every stack", ecs.Entity{}, map[string]int{"item.coal": 2, "item.keepsake": 5}},
		{"the target's own keepsake counts", target, map[string]int{"item.coal": 2, "item.keepsake": 1}},
		{"someone else's does not", newHolder(w), map[string]int{"item.coal": 2}},
	}
	for _, tc := range tests {
		if got := inventory.Usable(w, holder, tc.target); !maps.Equal(got, tc.want) {
			t.Errorf("%s: Usable = %v, want %v", tc.name, got, tc.want)
		}
	}
	if got := inventory.Usable(w, newHolder(w), target); len(got) != 0 {
		t.Errorf("a body with no pack has %v usable", got)
	}
}

func TestHolderOfIsThePartyForAMember(t *testing.T) {
	w := sim.NewWorld(1)
	party, member, loner := newHolder(w), newHolder(w), newHolder(w)
	ecs.NewMap[comp.MemberOf](w.ECS).Add(member, &comp.MemberOf{}, party)
	if got := inventory.HolderOf(w, member); got != party {
		t.Errorf("HolderOf(member) = %v, want the party %v", got, party)
	}
	if got := inventory.HolderOf(w, loner); got != loner {
		t.Errorf("HolderOf(loner) = %v, want itself", got)
	}
	if got := inventory.HolderOf(w, ecs.Entity{}); !got.IsZero() {
		t.Errorf("HolderOf(zero) = %v, want zero", got)
	}
}
