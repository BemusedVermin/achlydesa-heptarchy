package inventory_test

import (
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game/gametest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/inventory"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

const (
	wellTemplate    = "template.source.well"
	cisternTemplate = "template.container.cistern"
	water           = api.ContentID("item.water")
)

func spawnAt(tb testing.TB, w *sim.World, template api.ContentID, at api.Cell) ecs.Entity {
	tb.Helper()
	e, err := w.Spawn(template, &comp.Pos{At: at})
	if err != nil {
		tb.Fatal(err)
	}
	return e
}

func sourceOf(w *sim.World, e ecs.Entity) comp.Source {
	return *ecs.NewMap[comp.Source](w.ECS).Get(e)
}

func TestSourceRefillsByItsRateAndNeverAboveMax(t *testing.T) {
	_, w, _ := gametest.New(t, 1, func(b *content.Bundle) {
		well := b.Templates[wellTemplate]
		well.Source = &content.SourceDef{Item: water, Max: 10, Refill: "max - stock >= 3 ? 3 : 4"}
		b.Templates[wellTemplate] = well
	})
	well := spawnAt(t, w, wellTemplate, api.Cell{})
	ecs.NewMap[comp.Source](w.ECS).Get(well).Stock = 2
	for _, want := range []int{5, 8, 10, 10} {
		if err := w.Systems.Run(w, sim.PerPhase); err != nil {
			t.Fatal(err)
		}
		if got := sourceOf(w, well).Stock; got != want {
			t.Fatalf("stock = %d, want %d", got, want)
		}
	}
}

func TestANegativeRefillIsAContentError(t *testing.T) {
	_, w, _ := gametest.New(t, 1, func(b *content.Bundle) {
		well := b.Templates[wellTemplate]
		well.Source = &content.SourceDef{Item: water, Max: 10, Refill: "-1"}
		b.Templates[wellTemplate] = well
	})
	spawnAt(t, w, wellTemplate, api.Cell{})
	if err := w.Systems.Run(w, sim.PerPhase); err == nil {
		t.Error("a refill of -1 ran without an error")
	}
}

func TestMove(t *testing.T) {
	_, w, _ := gametest.New(t, 1, nil)
	well := spawnAt(t, w, wellTemplate, api.Cell{})
	cistern := spawnAt(t, w, cisternTemplate, api.Cell{Hex: api.Hex{Q: 1}})
	far := spawnAt(t, w, cisternTemplate, api.Cell{Hex: api.Hex{Q: 3}})
	full := sourceOf(w, well).Max
	const held = 5
	inventory.Add(w, cistern, comp.ItemStack{Item: water, Count: held})
	inventory.Add(w, far, comp.ItemStack{Item: water, Count: held})

	tests := []struct {
		name     string
		from, to ecs.Entity
		count    int
		want     bool
		well     int
		cistern  int
	}{
		{"out of reach is refused", cistern, far, 1, false, full, held},
		{"more than held is refused", cistern, well, held + 1, false, full, held},
		{"into a full source is refused", cistern, well, 1, false, full, held},
		{"from a source takes its stock", well, cistern, 3, true, full - 3, held + 3},
		{"more than a source's stock is refused", well, cistern, full, false, full - 3, held + 3},
		{"into a source adds to its stock", cistern, well, 2, true, full - 1, held + 1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := inventory.Move(w, tc.from, tc.to, water, tc.count); got != tc.want {
				t.Errorf("Move = %v, want %v", got, tc.want)
			}
			if got := sourceOf(w, well).Stock; got != tc.well {
				t.Errorf("well stock = %d, want %d", got, tc.well)
			}
			if got := inventory.Count(w, cistern, water, 0); got != tc.cistern {
				t.Errorf("cistern holds %d, want %d", got, tc.cistern)
			}
		})
	}
	if got := inventory.Count(w, far, water, 0); got != held {
		t.Errorf("the far cistern holds %d, want %d", got, held)
	}
}

func TestAnEmptyContainerSpawnsWithAnEmptyInventory(t *testing.T) {
	_, w, _ := gametest.New(t, 1, func(b *content.Bundle) {
		cistern := b.Templates[cisternTemplate]
		cistern.Container, cistern.Inventory = true, nil
		b.Templates[cisternTemplate] = cistern
	})
	cistern := spawnAt(t, w, cisternTemplate, api.Cell{})
	inventories := ecs.NewMap[comp.Inventory](w.ECS)
	if !inventories.Has(cistern) {
		t.Fatal("the empty cistern spawned without an inventory")
	}
	if got := inventories.Get(cistern).Stacks; len(got) != 0 {
		t.Errorf("the empty cistern holds %v, want nothing", got)
	}
}

func TestASaveKeepsAHalfEmptySource(t *testing.T) {
	g, w, cfg := gametest.New(t, 1, nil)
	well := spawnAt(t, w, wellTemplate, api.Cell{})
	uid := w.UIDOf(well)
	want := sourceOf(w, well)
	want.Stock = want.Max / 2
	*ecs.NewMap[comp.Source](w.ECS).Get(well) = want

	_, loaded, _ := gametest.Reload(t, cfg, g)
	e, ok := loaded.ByUID(uid)
	if !ok {
		t.Fatal("the well is gone after a load")
	}
	if got := sourceOf(loaded, e); got != want {
		t.Errorf("source after load = %+v, want %+v", got, want)
	}
}
