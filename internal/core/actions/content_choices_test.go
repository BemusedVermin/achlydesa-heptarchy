package actions_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/effects"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/inventory"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// withWater gives the fighter one lake water.
func withWater(b *content.Bundle) {
	fighter := b.Templates["template.test.fighter"]
	fighter.Inventory = []content.StackDef{{Item: "item.lake_water", Count: 1}}
	b.Templates["template.test.fighter"] = fighter
}

// loadIDs lists the option IDs of the load action.
func (fx *fixture) loadIDs() []string {
	var ids []string
	for _, id := range fx.optionIDs() {
		if strings.HasPrefix(id, "load/") {
			ids = append(ids, id)
		}
	}
	return ids
}

func TestTypeChoiceComesFromContent(t *testing.T) {
	fx := newFixture(t, 1, opts{
		actions: []api.ContentID{"action.test.load"}, contentChoices: true, mutate: withWater,
	})
	want := []string{"load/type.cut", "load/type.name", "load/type.press", "load/type.relic", "load/type.soak"}
	if got := fx.loadIDs(); !slices.Equal(got, want) {
		t.Fatalf("options with lake water = %v, want %v, in ID order", got, want)
	}
	if option := fx.option("load/type.name"); !option.Enabled {
		t.Error("Name is refused without a name in the inventory; it should be offered, so the menu tells nothing")
	}
	if !fx.option("load/type.soak").Enabled {
		t.Error("Soak should be enabled while the inventory holds lake water")
	}
	if !inventory.Take(fx.w, fx.fighter, "item.lake_water", 0, 1) {
		t.Fatal("the fighter had no lake water")
	}
	for _, id := range fx.loadIDs() {
		if id == "load/type.soak" {
			t.Errorf("Soak is still offered after the water is spent: %v", fx.loadIDs())
		}
	}
	if got := fx.loadIDs(); len(got) != 4 {
		t.Errorf("innate types are always offered: got %v", got)
	}
}

func TestRelicIsOfferedOnlyToABodyWithAGraft(t *testing.T) {
	fx := newFixture(t, 1, opts{
		actions: []api.ContentID{"action.test.load"}, contentChoices: true,
		mutate: func(b *content.Bundle) {
			fighter := b.Templates["template.test.fighter"]
			fighter.Lists = map[string][]string{"types": {"type.cut"}}
			b.Templates["template.test.fighter"] = fighter
		},
	})
	if got := fx.loadIDs(); slices.Contains(got, "load/type.relic") {
		t.Fatalf("a body with no graft is offered Relic: %v", got)
	}
	arm := fx.arms[0]
	ecs.NewMap[comp.PartOf](fx.w.ECS).Add(arm, &comp.PartOf{}, fx.fighter)
	ecs.NewMap[comp.Graft](fx.w.ECS).Add(arm, &comp.Graft{})
	if got := fx.loadIDs(); !slices.Contains(got, "load/type.relic") {
		t.Fatalf("a body with a graft is not offered Relic: %v", got)
	}
}

func TestLabelNamesTheChosenRecord(t *testing.T) {
	fx := newFixture(t, 1, opts{
		actions: []api.ContentID{"action.test.load"}, contentChoices: true, mutate: withWater,
	})
	label := fx.option("load/type.cut").Label
	if !strings.Contains(label, "Cut") || strings.Contains(label, "type.cut") {
		t.Errorf("label = %q, want the type's name \"Cut\", not its ID", label)
	}
}

func TestChosenRecordsChangeTheAction(t *testing.T) {
	var log []string
	fx := newFixture(t, 1, opts{
		actions: []api.ContentID{"action.test.load"}, contentChoices: true,
		install: func(w *sim.World) error {
			return effects.Register(w, effects.Primitive{
				Name: "test.log", Params: []string{"tag"}, Kinds: map[string]effects.Kind{"tag": effects.KindText},
				Apply: func(_ *effects.Context, p effects.Params) error {
					tag, err := p.String("tag")
					log = append(log, tag)
					return err
				},
			})
		},
		mutate: func(b *content.Bundle) {
			withWater(b)
			fighter := b.Templates["template.test.fighter"]
			fighter.Lists = map[string][]string{"types": {"type.cut", "type.test", "type.locked"}}
			b.Templates["template.test.fighter"] = fighter
			load := b.Actions["action.test.load"]
			load.OnLand = []content.EffectRef{{Do: "test.log", Params: map[string]string{"tag": "own"}}}
			b.Actions["action.test.load"] = load
			b.Text["text.test.refusal"] = "Locked up."
			b.Text["text.type.test"], b.Text["text.type.locked"] = "Test", "Locked"
			b.Types["type.test"] = content.Type{
				Name: "text.type.test", Tags: []string{"innate"},
				Windup: "2", Recovery: "1",
				OnStart: []content.EffectRef{{Do: "consume", Params: map[string]string{"item": "item.lake_water"}},
					{Do: "test.log", Params: map[string]string{"tag": "start"}},
				},
				OnLand: []content.EffectRef{{Do: "test.log", Params: map[string]string{"tag": "mod"}}},
			}
			b.Types["type.locked"] = content.Type{
				Name: "text.type.locked", Tags: []string{"innate"},
				Requires: "false", Refusal: "text.test.refusal",
			}
		},
	})
	plain := fx.option("load/type.cut")
	if !strings.Contains(plain.Label, "lands in 0, busy 2") {
		t.Fatalf("a plain type's label = %q, want the action's own timing", plain.Label)
	}
	test := fx.option("load/type.test")
	if !strings.Contains(test.Label, "lands in 2, busy 5") {
		t.Errorf("type.test's label = %q, want wind-up 0+2 and recovery 2+1", test.Label)
	}
	if locked := fx.option("load/type.locked"); locked.Enabled || locked.Reason != "Locked up." {
		t.Errorf(
			"type.locked = enabled %v, reason %q; want disabled with the type's refusal",
			locked.Enabled,
			locked.Reason,
		)
	}
	fx.apply(test.Command)
	if got := inventory.Count(fx.w, fx.fighter, "item.lake_water", 0); got != 0 {
		t.Errorf("lake water after the action = %d, want it spent", got)
	}
	if want := []string{"start", "own", "mod"}; !slices.Equal(log, want) {
		t.Errorf("effects = %v, want the type's start, then the landing of the action's own, then the type's", log)
	}
	if fx.w.Now != 5 {
		t.Errorf("the next decision came at tick %d, want 5 (wind-up 2, recovery 3)", fx.w.Now)
	}
}

func TestContentChoiceMayNotShadowAProvider(t *testing.T) {
	reg := contenttest.Registry(t, testBundle(t, opts{contentChoices: true}))
	if got := reg.ChoiceNames(); !slices.Equal(got, []string{"carried", "piece", "type", "ware"}) {
		t.Fatalf("choices = %v", got)
	}
}
