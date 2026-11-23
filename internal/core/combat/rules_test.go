package combat_test

import (
	"strings"
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/effects"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/inventory"
)

const brittleWeak = `chosen.type.id in target.material.lists.weak_to || "brittle" in target.part.tags`

// soakAll damages every part of the crawler by 2 where it is weak to the chosen type, and 1 elsewhere.
func soakAll(t *testing.T, r *rig, kind string) map[string]int {
	t.Helper()
	ctx := r.ctx(r.crawler)
	ctx.Choices = map[string]string{"type": kind}
	before := map[ecs.Entity]int{}
	for _, slot := range []string{"head", "legs", "torso"} {
		for _, part := range r.partsOf(r.crawler, slot) {
			before[part] = r.integrity(part)
		}
	}
	r.apply(t, ctx, do("damage", "parts", "target:all", "amount", "rule.weak ? 2 : 1"))
	lost := map[string]int{}
	for part, was := range before {
		lost[partSlot(r, part)] = was - r.integrity(part)
	}
	return lost
}

func partSlot(r *rig, part ecs.Entity) string {
	for _, slot := range []string{"head", "legs", "torso"} {
		for _, found := range r.partsOf(r.crawler, slot) {
			if found == part {
				return slot
			}
		}
	}
	return ""
}

func TestDamageIsEvaluatedPerPart(t *testing.T) {
	r := newRig(t)
	// Soak: the clay core (torso) is weak to it; the maw and legs are flesh.
	got := soakAll(t, r, "type.soak")
	if got["torso"] != 2 || got["head"] != 1 || got["legs"] != 1 {
		t.Errorf("lost = %v, want 2 from the clay core and 1 from every other part", got)
	}
}

func TestChangingTheWeakRuleNeedsNoGoChange(t *testing.T) {
	r := newRigWith(t, func(b *content.Bundle) {
		b.Rules["weak"] = brittleWeak
		maw := b.Parts["part.test.crawler_maw"]
		maw.Tags = []string{"brittle"}
		b.Parts["part.test.crawler_maw"] = maw
	})
	got := soakAll(t, r, "type.soak")
	if got["torso"] != 2 || got["head"] != 2 || got["legs"] != 1 {
		t.Errorf("lost = %v, want the brittle maw to count as weak too", got)
	}
}

func TestSwappingContentRecomputesAFormulaAlreadyEvaluated(t *testing.T) {
	r := newRig(t)
	if got := soakAll(t, r, "type.soak"); got["head"] != 1 {
		t.Fatalf("before the swap the maw lost %d, want 1", got["head"])
	}
	bundle := contenttest.SampleBundle(t)
	bundle.Rules["weak"] = brittleWeak
	maw := bundle.Parts["part.test.crawler_maw"]
	maw.Tags = []string{"brittle"}
	bundle.Parts["part.test.crawler_maw"] = maw
	reg, err := content.New(bundle)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.g.SwapContent(reg); err != nil {
		t.Fatal(err)
	}
	if got := soakAll(t, r, "type.soak"); got["head"] != 2 {
		t.Errorf("after the swap the maw lost %d, want 2 by the new rule", got["head"])
	}
}

func TestTextParameterFormulas(t *testing.T) {
	r := newRig(t)
	holder := inventory.HolderOf(r.w, r.reaper)
	inventory.Add(r.w, holder, comp.ItemStack{Item: "item.lake_water", Count: 1})
	before := inventory.Count(r.w, holder, "item.lake_water", 0)
	ctx := r.ctx(r.crawler)
	ctx.Choices = map[string]string{"type": "type.soak"}
	r.apply(t, ctx, do("consume", "item", "= chosen.type.lists.consumes[0]"))
	if got := inventory.Count(r.w, holder, "item.lake_water", 0); got != before-1 {
		t.Errorf("lake water = %d, want %d: one spent by the formula's item", got, before-1)
	}
	err := effects.Apply(ctx, []content.EffectRef{do("consume", "item", `= "item.nothing"`)})
	if err == nil || !strings.Contains(err.Error(), `parameter "item"`) {
		t.Errorf("a formula naming an unknown item: error = %v, want one naming the parameter", err)
	}
}

func TestRuleReadingRollIsRefusedWhereAPrimitiveChecks(t *testing.T) {
	r := newRigWith(t, func(b *content.Bundle) { b.Rules["lucky"] = "roll < 0.5" })
	// attach_part has a Check; its parameters are text, so a formula in one is what could draw.
	err := effects.Validate(r.w, "test", []content.EffectRef{
		do("attach_part", "from", "= rule.lucky ? \"inventory\" : \"severed\"", "item", "item.test_arm"),
	})
	if err == nil || !strings.Contains(err.Error(), "cannot read roll") {
		t.Errorf("Validate = %v, want a refusal to read roll", err)
	}
	if err := effects.Validate(r.w, "test", []content.EffectRef{
		do("damage", "parts", "target:all", "amount", "rule.lucky ? 1 : 2"),
	}); err != nil {
		t.Errorf("a primitive with no Check may read roll: %v", err)
	}
}
