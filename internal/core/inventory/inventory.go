package inventory

import (
	"cmp"
	"errors"
	"fmt"
	"slices"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// Install registers the spawner that gives templates their starting pack.
func Install(w *sim.World) error {
	if w.Content == nil {
		return errors.New("inventory: install: world has no content")
	}
	return w.RegisterSpawner(sim.Spawner{Name: "inventory", Apply: spawnInventory})
}

func spawnInventory(w *sim.World, holder ecs.Entity, t content.Template) error {
	for i, def := range t.Inventory {
		if _, ok := w.Content.Item(def.Item); !ok {
			return fmt.Errorf("inventory %d: unknown item %q", i, def.Item)
		}
		Add(w, holder, comp.ItemStack{Item: def.Item, Count: def.Count})
	}
	return nil
}

func compareKey(a, b comp.ItemStack) int {
	return cmp.Or(cmp.Compare(a.Item, b.Item), cmp.Compare(a.BoundTo, b.BoundTo))
}

// Add puts the stack in the holder's pack, merging with the stack of the same item and owner and
// adding comp.Inventory if the holder has none. A stack with no items in it changes nothing.
func Add(w *sim.World, holder ecs.Entity, s comp.ItemStack) {
	if s.Count <= 0 || !w.ECS.Alive(holder) {
		return
	}
	packs := ecs.NewMap[comp.Inventory](w.ECS)
	if !packs.Has(holder) {
		packs.Add(holder, &comp.Inventory{})
	}
	pack := packs.Get(holder)
	at, found := slices.BinarySearchFunc(pack.Stacks, s, compareKey)
	if found {
		pack.Stacks[at].Count += s.Count
		return
	}
	pack.Stacks = slices.Insert(pack.Stacks, at, s)
}

// Take removes n of the item from the pack. It returns false, and changes nothing, if the pack
// holds fewer than n. Taking nothing (n <= 0) succeeds.
func Take(w *sim.World, holder ecs.Entity, item api.ContentID, boundTo api.UID, n int) bool {
	if n <= 0 {
		return true
	}
	packs := ecs.NewMap[comp.Inventory](w.ECS)
	if !w.ECS.Alive(holder) || !packs.Has(holder) {
		return false
	}
	pack := packs.Get(holder)
	at, found := slices.BinarySearchFunc(pack.Stacks, comp.ItemStack{Item: item, BoundTo: boundTo}, compareKey)
	if !found || pack.Stacks[at].Count < n {
		return false
	}
	pack.Stacks[at].Count -= n
	if pack.Stacks[at].Count == 0 {
		pack.Stacks = slices.Delete(pack.Stacks, at, at+1)
	}
	return true
}

// Count is how many of the item the holder carries.
func Count(w *sim.World, holder ecs.Entity, item api.ContentID, boundTo api.UID) int {
	packs := ecs.NewMap[comp.Inventory](w.ECS)
	if !w.ECS.Alive(holder) || !packs.Has(holder) {
		return 0
	}
	pack := packs.Get(holder)
	at, found := slices.BinarySearchFunc(pack.Stacks, comp.ItemStack{Item: item, BoundTo: boundTo}, compareKey)
	if !found {
		return 0
	}
	return pack.Stacks[at].Count
}
