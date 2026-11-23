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

// Install registers the spawner that gives templates their starting inventory.
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

// Add puts the stack in the holder's inventory, merging with the stack of the same item and owner and
// adding comp.Inventory if the holder has none. A stack with no items in it changes nothing.
func Add(w *sim.World, holder ecs.Entity, s comp.ItemStack) {
	if s.Count <= 0 || !w.ECS.Alive(holder) {
		return
	}
	inventories := ecs.NewMap[comp.Inventory](w.ECS)
	if !inventories.Has(holder) {
		inventories.Add(holder, &comp.Inventory{})
	}
	held := inventories.Get(holder)
	at, found := slices.BinarySearchFunc(held.Stacks, s, compareKey)
	if found {
		held.Stacks[at].Count += s.Count
		return
	}
	held.Stacks = slices.Insert(held.Stacks, at, s)
}

// Take removes n of the item from the inventory. It returns false, and changes nothing, if the inventory
// holds fewer than n. Taking nothing (n <= 0) succeeds.
func Take(w *sim.World, holder ecs.Entity, item api.ContentID, boundTo api.UID, n int) bool {
	if n <= 0 {
		return true
	}
	inventories := ecs.NewMap[comp.Inventory](w.ECS)
	if !w.ECS.Alive(holder) || !inventories.Has(holder) {
		return false
	}
	held := inventories.Get(holder)
	at, found := slices.BinarySearchFunc(held.Stacks, comp.ItemStack{Item: item, BoundTo: boundTo}, compareKey)
	if !found || held.Stacks[at].Count < n {
		return false
	}
	held.Stacks[at].Count -= n
	if held.Stacks[at].Count == 0 {
		held.Stacks = slices.Delete(held.Stacks, at, at+1)
	}
	return true
}

// HolderOf is who keeps the body's inventory: its party when it is a member, otherwise itself. A body
// that is gone has no holder (the zero entity).
func HolderOf(w *sim.World, body ecs.Entity) ecs.Entity {
	if body.IsZero() || !w.ECS.Alive(body) {
		return ecs.Entity{}
	}
	members := ecs.NewMap[comp.MemberOf](w.ECS)
	if members.Has(body) {
		return members.GetRelation(body)
	}
	return body
}

// Usable counts, per item, the stacks in the inventory of the body's holder that can be used on the
// target: unbound stacks, and stacks bound to the target. With no target every stack counts.
func Usable(w *sim.World, body, target ecs.Entity) map[string]int {
	out := map[string]int{}
	holder := HolderOf(w, body)
	inventories := ecs.NewMap[comp.Inventory](w.ECS)
	if holder.IsZero() || !inventories.Has(holder) {
		return out
	}
	var targetUID api.UID
	if !target.IsZero() && w.ECS.Alive(target) {
		targetUID = w.UIDOf(target)
	}
	for _, stack := range inventories.Get(holder).Stacks {
		if target.IsZero() || stack.BoundTo == 0 || stack.BoundTo == targetUID {
			out[string(stack.Item)] += stack.Count
		}
	}
	return out
}

// Count is how many of the item the holder carries.
func Count(w *sim.World, holder ecs.Entity, item api.ContentID, boundTo api.UID) int {
	inventories := ecs.NewMap[comp.Inventory](w.ECS)
	if !w.ECS.Alive(holder) || !inventories.Has(holder) {
		return 0
	}
	held := inventories.Get(holder)
	at, found := slices.BinarySearchFunc(held.Stacks, comp.ItemStack{Item: item, BoundTo: boundTo}, compareKey)
	if !found {
		return 0
	}
	return held.Stacks[at].Count
}
