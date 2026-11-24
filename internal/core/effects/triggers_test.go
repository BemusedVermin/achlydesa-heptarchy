package effects_test

import (
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/clock"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/effects"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/inventory"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// triggerRig is a game with two crawlers, a and b, and a test primitive, note, which records
// its text and the body it was aimed at.
type triggerRig struct {
	w     *sim.World
	a, b  ecs.Entity
	notes []string
}

func triggerBundle(t *testing.T, mutate func(*content.Bundle)) content.Bundle {
	t.Helper()
	bundle := contenttest.SampleBundle(t)
	contenttest.NoStart(&bundle)
	body := api.ContentID("body.test.crawler")
	bundle.Templates["template.test.crawler"] = content.Template{Body: &body}
	bundle.Statuses["status.braced"] = content.Status{Name: "text.status.braced"}
	bundle.Text["text.status.braced"] = "braced"
	mutate(&bundle)
	return bundle
}

func newTriggerRig(t *testing.T, mutate func(*content.Bundle)) (*triggerRig, error) {
	t.Helper()
	reg := contenttest.Registry(t, triggerBundle(t, mutate))
	r := &triggerRig{}
	at := comp.Pos{Map: 1, At: api.Cell{Hex: api.Hex{Q: 2, R: 1}, Depth: 3}}
	_, err := game.New(game.Config{
		Seed: 1, Content: reg,
		Install: func(w *sim.World, _ *clock.Clock) error {
			r.w = w
			note := effects.Primitive{
				Name: "note", Params: []string{"text"}, Kinds: map[string]effects.Kind{"text": effects.KindText},
				Apply: func(ctx *effects.Context, p effects.Params) error {
					text, err := p.String("text")
					r.notes = append(r.notes, fmt.Sprintf("%s:%d", text, w.UIDOf(ctx.Target)))
					return err
				},
			}
			return effects.Register(w, note)
		},
		Spawn: func(w *sim.World, _ *clock.Clock) error {
			if _, err := w.Spawn("template.party.main", nil); err != nil { // someone to play the first decision
				return err
			}
			var err error
			if r.a, err = w.Spawn("template.test.crawler", &at); err != nil {
				return err
			}
			r.b, err = w.Spawn("template.test.crawler", &at)
			return err
		},
	})
	return r, err
}

func mustTriggerRig(t *testing.T, mutate func(*content.Bundle)) *triggerRig {
	t.Helper()
	r, err := newTriggerRig(t, mutate)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func (r *triggerRig) apply(t *testing.T, ctx *effects.Context, refs ...content.EffectRef) error {
	t.Helper()
	if err := effects.Apply(ctx, refs); err != nil {
		t.Fatal(err)
	}
	_, err := r.w.Bus.Drain(r.w)
	return err
}

func (r *triggerRig) part(body ecs.Entity, slot string) ecs.Entity {
	query := ecs.NewFilter1[comp.Part](r.w.ECS).Query()
	defer query.Close()
	for query.Next() {
		if query.Get().Slot != slot {
			continue
		}
		if ecs.NewMap[comp.PartOf](r.w.ECS).Has(query.Entity()) &&
			ecs.NewMap[comp.PartOf](r.w.ECS).GetRelation(query.Entity()) == body {
			return query.Entity()
		}
	}
	return ecs.Entity{}
}

func trigger(on, role, ifExpr string, do ...content.EffectRef) content.Trigger {
	return content.Trigger{On: on, If: ifExpr, Do: do, Role: role}
}

func TestTemplateTriggerFiresOnlyForItsBodysParts(t *testing.T) {
	r := mustTriggerRig(t, func(b *content.Bundle) {
		tmpl := b.Templates["template.test.crawler"]
		tmpl.Triggers = []content.Trigger{
			trigger("combat.part_severed", content.RoleTarget, "", contenttest.Do("note", "text", "severed")),
		}
		b.Templates["template.test.crawler"] = tmpl
	})
	maw := r.part(r.a, "head")
	if maw.IsZero() {
		t.Fatal("the crawler has no maw")
	}
	if err := r.apply(
		t,
		&effects.Context{World: r.w, Target: r.a, TargetPart: maw},
		contenttest.Do("sever", "parts", "target_part"),
	); err != nil {
		t.Fatal(err)
	}
	want := []string{fmt.Sprintf("severed:%d", r.w.UIDOf(r.a))}
	if !slices.Equal(r.notes, want) {
		t.Errorf("notes = %v, want %v (b's trigger must not fire)", r.notes, want)
	}
}

func TestMaterialTriggerHardensClayWhenBurned(t *testing.T) {
	hardenOnBurn := func(b *content.Bundle) {
		clay := b.Materials["material.clay"]
		clay.Triggers = []content.Trigger{trigger(
			"combat.part_damaged", content.RolePart, `chosen.type.id == "type.burn"`,
			contenttest.Do("raise_max_integrity", "parts", "target_part", "amount", "2"),
		)}
		b.Materials["material.clay"] = clay
	}
	tests := []struct {
		name       string
		kind       string
		wantRaised bool
	}{
		{"burn hardens clay", "type.burn", true},
		{"cut leaves it", "type.cut", false},
		{"no type leaves it", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := mustTriggerRig(t, hardenOnBurn)
			parts := ecs.NewMap[comp.Part](r.w.ECS)
			before := map[ecs.Entity]int{}
			var core ecs.Entity
			for _, slot := range []string{"head", "legs", "torso"} {
				if e := r.part(r.a, slot); !e.IsZero() {
					before[e] = parts.Get(e).MaxIntegrity
					if parts.Get(e).Material == "material.clay" {
						core = e
					}
				}
			}
			if core.IsZero() {
				t.Fatal("the crawler has no clay part")
			}
			ctx := &effects.Context{World: r.w, Target: r.a, Choices: map[string]string{"type": tt.kind}}
			if err := r.apply(t, ctx, contenttest.Do("damage", "parts", "target:all", "amount", "1")); err != nil {
				t.Fatal(err)
			}
			for e, was := range before {
				now := parts.Get(e).MaxIntegrity
				wantNow := was
				if e == core && tt.wantRaised {
					wantNow += 2
				}
				if now != wantNow {
					t.Errorf(
						"part %s (%s): max integrity %d, want %d",
						parts.Get(e).Slot,
						parts.Get(e).Material,
						now,
						wantNow,
					)
				}
			}
		})
	}
}

func TestTriggersFireInUIDOrder(t *testing.T) {
	r := mustTriggerRig(t, func(b *content.Bundle) {
		tmpl := b.Templates["template.test.crawler"]
		tmpl.Triggers = []content.Trigger{
			trigger("combat.part_severed", content.RoleAny, "", contenttest.Do("note", "text", "body")),
		}
		b.Templates["template.test.crawler"] = tmpl
		clay := b.Materials["material.clay"]
		clay.Triggers = []content.Trigger{
			trigger("combat.part_severed", content.RoleAny, "", contenttest.Do("note", "text", "part")),
		}
		b.Materials["material.clay"] = clay
	})
	core := r.part(r.a, "torso")
	ctx := &effects.Context{World: r.w, Target: r.a, TargetPart: core}
	if err := r.apply(t, ctx, contenttest.Do("sever", "parts", "target_part")); err != nil {
		t.Fatal(err)
	}
	uid := r.w.UIDOf(r.a)
	want := []string{fmt.Sprintf("body:%d", uid), fmt.Sprintf("part:%d", uid)}
	if r.w.UIDOf(r.a) >= r.w.UIDOf(core) && !slices.Equal(r.notes, want) {
		t.Fatalf("test setup: the body should have the lower UID")
	}
	if !slices.Equal(r.notes, want) {
		t.Errorf("notes = %v, want %v (the body's UID is lower than its part's)", r.notes, want)
	}
}

func TestStatusTriggerFiresForBodiesUnderIt(t *testing.T) {
	r := mustTriggerRig(t, func(b *content.Bundle) {
		braced := b.Statuses["status.braced"]
		braced.Triggers = []content.Trigger{
			trigger("combat.part_severed", content.RoleTarget, "", contenttest.Do("note", "text", "braced")),
		}
		b.Statuses["status.braced"] = braced
	})
	maw := r.part(r.a, "head")
	ctx := &effects.Context{World: r.w, Target: r.a, TargetPart: maw}
	if err := r.apply(
		t,
		ctx,
		contenttest.Do("add_status", "bodies", "target", "status", "status.braced", "ticks", "5"),
		contenttest.Do("sever", "parts", "target_part"),
	); err != nil {
		t.Fatal(err)
	}
	if len(r.notes) != 1 {
		t.Errorf("notes = %v, want one from the braced body", r.notes)
	}
	r.notes = nil
	bCtx := &effects.Context{World: r.w, Target: r.b, TargetPart: r.part(r.b, "head")}
	if err := r.apply(t, bCtx, contenttest.Do("sever", "parts", "target_part")); err != nil {
		t.Fatal(err)
	}
	if len(r.notes) != 0 {
		t.Errorf("notes = %v, want none from a body without the status", r.notes)
	}
}

func TestItemTriggerFiresForHolders(t *testing.T) {
	r := mustTriggerRig(t, func(b *content.Bundle) {
		coal := b.Items["item.coal"]
		coal.Triggers = []content.Trigger{
			trigger("combat.part_severed", content.RoleTarget, "", contenttest.Do("note", "text", "coal")),
		}
		b.Items["item.coal"] = coal
	})
	inventory.Add(r.w, r.a, comp.ItemStack{Item: "item.coal", Count: 1})
	for _, body := range []ecs.Entity{r.a, r.b} {
		ctx := &effects.Context{World: r.w, Target: body, TargetPart: r.part(body, "head")}
		if err := r.apply(t, ctx, contenttest.Do("sever", "parts", "target_part")); err != nil {
			t.Fatal(err)
		}
	}
	want := []string{fmt.Sprintf("coal:%d", r.w.UIDOf(r.a))}
	if !slices.Equal(r.notes, want) {
		t.Errorf("notes = %v, want %v (only the holder of coal)", r.notes, want)
	}
}

func TestSelfRefiringTriggerErrors(t *testing.T) {
	r := mustTriggerRig(t, func(b *content.Bundle) {
		tmpl := b.Templates["template.test.crawler"]
		tmpl.Triggers = []content.Trigger{trigger(
			"combat.status_added", content.RoleTarget, "",
			contenttest.Do("add_status", "bodies", "target", "status", "status.braced", "ticks", "5"),
		)}
		b.Templates["template.test.crawler"] = tmpl
	})
	ctx := &effects.Context{World: r.w, Target: r.a}
	err := r.apply(t, ctx, contenttest.Do("add_status", "bodies", "target", "status", "status.braced", "ticks", "5"))
	if err == nil || !strings.Contains(err.Error(), "cascade") {
		t.Fatalf("a trigger that re-fires itself gave %v, want a cascade error", err)
	}
}

func TestTriggerEffectErrorsSurfaceFromDrain(t *testing.T) {
	r := mustTriggerRig(t, func(b *content.Bundle) {
		tmpl := b.Templates["template.test.crawler"]
		tmpl.Triggers = []content.Trigger{trigger(
			"combat.part_severed", content.RoleTarget, "",
			contenttest.Do("add_status", "bodies", "target", "status", "status.braced", "ticks", "0"),
		)}
		b.Templates["template.test.crawler"] = tmpl
	})
	ctx := &effects.Context{World: r.w, Target: r.a, TargetPart: r.part(r.a, "head")}
	err := r.apply(t, ctx, contenttest.Do("sever", "parts", "target_part"))
	if err == nil || !strings.Contains(err.Error(), "ticks must be at least 1") {
		t.Fatalf("a failing trigger effect gave %v", err)
	}
}

func TestNewRejectsBadTriggers(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*content.Bundle)
		want   string
	}{
		{
			name: "unknown event kind",
			mutate: func(b *content.Bundle) {
				clay := b.Materials["material.clay"]
				clay.Triggers = []content.Trigger{
					trigger("combat.no_such_event", content.RolePart, "", contenttest.Do("note", "text", "x")),
				}
				b.Materials["material.clay"] = clay
			},
			want: `materials["material.clay"].triggers[0].on: unknown event kind "combat.no_such_event"`,
		},
		{
			name: "unknown primitive",
			mutate: func(b *content.Bundle) {
				clay := b.Materials["material.clay"]
				clay.Triggers = []content.Trigger{
					trigger("combat.part_damaged", content.RolePart, "", contenttest.Do("no_such_primitive")),
				}
				b.Materials["material.clay"] = clay
			},
			want: `materials["material.clay"].triggers[0].do[0]: unknown primitive "no_such_primitive"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := newTriggerRig(t, tt.mutate)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("game.New gave %v, want an error naming %q", err, tt.want)
			}
		})
	}
}

func TestContentRejectsBadTriggerShapes(t *testing.T) {
	tests := []struct {
		name    string
		trigger content.Trigger
		want    string
	}{
		{"bad role", trigger("combat.part_damaged", "bystander", "", contenttest.Do("note")), "role: unknown role"},
		{"bad condition", trigger("combat.part_damaged", content.RoleAny, "nope +", contenttest.Do("note")), "if:"},
		{"no effects", trigger("combat.part_damaged", content.RoleAny, ""), "do: at least one effect"},
		{"no event", trigger("", content.RoleAny, "", contenttest.Do("note")), "on: required"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bundle := triggerBundle(t, func(b *content.Bundle) {
				clay := b.Materials["material.clay"]
				clay.Triggers = []content.Trigger{tt.trigger}
				b.Materials["material.clay"] = clay
			})
			_, err := content.New(bundle)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("content.New gave %v, want %q", err, tt.want)
			}
		})
	}
}

func TestSourceRoleTriggerFiresForTheActingBody(t *testing.T) {
	r := mustTriggerRig(t, func(b *content.Bundle) {
		tmpl := b.Templates["template.test.crawler"]
		tmpl.Triggers = []content.Trigger{
			trigger("combat.part_severed", content.RoleSource, "", contenttest.Do("note", "text", "acted")),
		}
		b.Templates["template.test.crawler"] = tmpl
	})
	// a acts on b: only a, the source, hears of it.
	ctx := &effects.Context{World: r.w, Source: r.a, Target: r.b, TargetPart: r.part(r.b, "head")}
	if err := r.apply(t, ctx, contenttest.Do("sever", "parts", "target_part")); err != nil {
		t.Fatal(err)
	}
	want := []string{fmt.Sprintf("acted:%d", r.w.UIDOf(r.b))}
	if !slices.Equal(r.notes, want) {
		t.Errorf("notes = %v, want %v", r.notes, want)
	}
}
