package agents

import (
	"fmt"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/effects"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/inventory"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// noStockReason is the text ID of the refusal when no town stock is in reach.
const noStockReason api.ContentID = "text.agents.no_stock"

func takeStockPrimitive() effects.Primitive {
	return effects.Primitive{
		Name: "take_stock", Params: []string{"item"},
		Kinds:         map[string]effects.Kind{"item": effects.KindText},
		VerifyContent: verifyItem, Apply: applyTakeStock, Check: checkTakeStock,
	}
}

func verifyItem(reg *content.Registry, raw map[string]string) error {
	if item, ok := raw["item"]; ok {
		if _, known := reg.Item(api.ContentID(item)); !known {
			return fmt.Errorf("unknown item %q", item)
		}
	}
	return nil
}

// stockInReach is the town whose stock holds the item, when the body stands next to one of the
// town's stock marks.
func stockInReach(w *sim.World, body ecs.Entity, item api.ContentID) (ecs.Entity, bool) {
	positions := ecs.NewMap[comp.Pos](w.ECS)
	if !w.Alive(body) || !positions.Has(body) {
		return ecs.Entity{}, false
	}
	pos := *positions.Get(body)
	town, ok := TownOf(w, pos.Map)
	if !ok || inventory.Count(w, town, item, 0) < 1 {
		return ecs.Entity{}, false
	}
	def, _ := w.Content.Map(ecs.NewMap[comp.Region](w.ECS).Get(town).Def)
	m := w.Maps.Get(pos.Map)
	for _, mark := range def.Region.Town.StockMarks {
		if cell, found := m.Mark(mark); found && hexmap.CellDistance(pos.At, cell) <= 1 {
			return town, true
		}
	}
	return ecs.Entity{}, false
}

func checkTakeStock(ctx *effects.Context, p effects.Params) (bool, api.Text) {
	item, err := p.String("item")
	if err != nil {
		return false, api.Text{ID: noStockReason}
	}
	if _, ok := stockInReach(ctx.World, ctx.Source, api.ContentID(item)); !ok {
		return false, api.Text{ID: noStockReason}
	}
	return true, api.Text{}
}

func applyTakeStock(ctx *effects.Context, p effects.Params) error {
	item, err := p.String("item")
	if err != nil {
		return err
	}
	town, ok := stockInReach(ctx.World, ctx.Source, api.ContentID(item))
	if !ok {
		return nil
	}
	inventory.Take(ctx.World, town, api.ContentID(item), 0, 1)
	return nil
}

// stockOfMap counts, per item ID, what the stock of the map's town holds; it is empty when the map
// is not a town.
func stockOfMap(w *sim.World, id api.MapID) map[string]int {
	town, ok := TownOf(w, id)
	if !ok || !ecs.NewMap[comp.Inventory](w.ECS).Has(town) {
		return nil
	}
	out := map[string]int{}
	for _, stack := range ecs.NewMap[comp.Inventory](w.ECS).Get(town).Stacks {
		out[string(stack.Item)] += stack.Count
	}
	return out
}
