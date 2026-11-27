package inventory

import (
	"fmt"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// RefillSystem is the per-phase system that refills every source.
const RefillSystem = "inventory.refill"

// reach is "next to": the same cell or a neighboring one. It is what next to means, not a number to
// tune; an action's own range decides how far a body walks first.
const reach = 1

func spawnSource(w *sim.World, holder ecs.Entity, t content.Template) error {
	def := t.Source
	if def == nil {
		return nil
	}
	if _, ok := w.Content.Item(def.Item); !ok {
		return fmt.Errorf("source: unknown item %q", def.Item)
	}
	template := ecs.NewMap[comp.Identity](w.ECS).Get(holder).Template
	ecs.NewMap[comp.Source](w.ECS).Add(holder, &comp.Source{
		Item: def.Item, Stock: def.Max, Max: def.Max, Refill: template,
	})
	return nil
}

// refill adds each source's refill to its stock, up to its Max. Each source refills on its own, so
// the order they are met in cannot change the outcome.
func refill(w *sim.World) error {
	query := ecs.NewFilter1[comp.Source](w.ECS).Query()
	defer query.Close()
	for query.Next() {
		source := query.Get()
		rate, err := w.Content.Refill(source.Refill, content.RefillEnv{Stock: source.Stock, Max: source.Max})
		if err != nil {
			return fmt.Errorf("inventory: refill: %w", err)
		}
		if rate < 0 {
			return fmt.Errorf("inventory: refill: template %q source.refill is %d, below 0", source.Refill, rate)
		}
		source.Stock = int(min(int64(source.Max), int64(source.Stock)+rate))
	}
	return nil
}

// Refusal is why a transfer cannot happen, or None.
type Refusal int

// Reasons a transfer is refused.
const (
	None Refusal = iota
	// OutOfReach: the body is not next to the thing it takes from or puts into.
	OutOfReach
	// Short: the giver holds fewer than the count.
	Short
	// Full: a source would go over its Max.
	Full
)

// Move carries count of the item, unbound, from one entity to another that stands next to it: on
// the same cell or a neighboring one. A source of the item gives from and takes into its stock, up
// to its Max; anything else uses the inventory. It returns false, and changes nothing, when either
// has no position, they are not next to each other, from holds fewer than count, or a source would
// go over its Max. Moving nothing (count <= 0) succeeds.
func Move(w *sim.World, from, to ecs.Entity, item api.ContentID, count int) bool {
	if from == to {
		return count <= 0
	}
	return Transfer(w, from, to, to, item, count) == None
}

// Transfer is Move for a body that carries into or out of its own holder, such as a party's
// inventory, which has no position: only an end that is not by's holder must be next to by. It
// refuses with the reason, and changes nothing, for the reasons Move does.
func Transfer(w *sim.World, from, to, by ecs.Entity, item api.ContentID, count int) Refusal {
	if reason := CheckTransfer(w, from, to, by, item, count); reason != None {
		return reason
	}
	if count <= 0 {
		return None
	}
	sources := ecs.NewMap[comp.Source](w.ECS)
	if giver := sourceOf(sources, from, item); giver != nil {
		giver.Stock -= count
	} else {
		Take(w, from, item, 0, count)
	}
	if receiver := sourceOf(sources, to, item); receiver != nil {
		receiver.Stock += count
	} else {
		Add(w, to, comp.ItemStack{Item: item, Count: count})
	}
	return None
}

// CheckTransfer says whether Transfer would succeed, and why not, without changing anything.
func CheckTransfer(w *sim.World, from, to, by ecs.Entity, item api.ContentID, count int) Refusal {
	if count <= 0 {
		return None
	}
	if from == to || !w.Alive(from) || !w.Alive(to) || !w.Alive(by) {
		return OutOfReach
	}
	positions := ecs.NewMap[comp.Pos](w.ECS)
	holder := HolderOf(w, by)
	for _, end := range []ecs.Entity{from, to} {
		if end == holder {
			continue
		}
		if !positions.Has(end) || !positions.Has(by) ||
			hexmap.CellDistance(positions.Get(by).At, positions.Get(end).At) > reach {
			return OutOfReach
		}
	}
	sources := ecs.NewMap[comp.Source](w.ECS)
	if giver := sourceOf(sources, from, item); giver != nil {
		if giver.Stock < count {
			return Short
		}
	} else if Count(w, from, item, 0) < count {
		return Short
	}
	if receiver := sourceOf(sources, to, item); receiver != nil && receiver.Stock+count > receiver.Max {
		return Full
	}
	return None
}

// sourceOf is the entity's source when it is a source of the item, otherwise nil.
func sourceOf(sources *ecs.Map[comp.Source], e ecs.Entity, item api.ContentID) *comp.Source {
	if !sources.Has(e) {
		return nil
	}
	if source := sources.Get(e); source.Item == item {
		return source
	}
	return nil
}
