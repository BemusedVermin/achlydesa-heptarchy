package anatomy_test

import (
	"fmt"
	"maps"
	"slices"
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/clock"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game/gametest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

// rig is a game whose only entity is the sample party, with its world and party kept for tests.
type rig struct {
	cfg   game.Config
	g     *game.Game
	w     *sim.World
	party ecs.Entity
}

func newRig(t *testing.T, seed uint64, mutate func(*content.Bundle)) *rig {
	t.Helper()
	bundle := contenttest.SampleBundle(t)
	contenttest.NoStart(&bundle)
	if mutate != nil {
		mutate(&bundle)
	}
	reg := contenttest.Registry(t, bundle)
	r := &rig{}
	r.cfg = game.Config{
		Seed: seed, Content: reg,
		Spawn: func(w *sim.World, _ *clock.Clock) error {
			var err error
			r.party, err = w.Spawn("template.party.main", nil)
			return err
		},
	}
	r.g, r.w, _ = gametest.Start(t, r.cfg)
	return r
}

func (r *rig) member(t *testing.T) ecs.Entity {
	t.Helper()
	members := anatomy.Members(r.w, r.party)
	if len(members) == 0 {
		t.Fatal("the party has no members")
	}
	return members[0]
}

func (r *rig) slots(body ecs.Entity) []string {
	var slots []string
	parts := ecs.NewMap[comp.Part](r.w.ECS)
	for _, part := range anatomy.Parts(r.w, body) {
		slots = append(slots, parts.Get(part).Slot)
	}
	return slots
}

// plan is the body's part templates in order, so expectations follow the content.
func (r *rig) plan(t *testing.T, body ecs.Entity) []content.PartTemplate {
	t.Helper()
	id := ecs.NewMap[comp.Body](r.w.ECS).Get(body).Plan
	plan, ok := r.w.Content.Body(id)
	if !ok {
		t.Fatalf("unknown body plan %q", id)
	}
	defs := make([]content.PartTemplate, 0, len(plan.Parts))
	for _, entry := range plan.Parts {
		def, ok := r.w.Content.Part(entry.Part)
		if !ok {
			t.Fatalf("unknown part %q", entry.Part)
		}
		defs = append(defs, def)
	}
	return defs
}

func totals(grants ...[]comp.Grant) map[string]int {
	out := map[string]int{}
	for _, list := range grants {
		for _, grant := range list {
			out[grant.Kind] += grant.Value
		}
	}
	return out
}

func TestPartyMembersGetTheirPlansParts(t *testing.T) {
	r := newRig(t, 1, nil)
	members := anatomy.Members(r.w, r.party)
	if len(members) == 0 {
		t.Fatal("the party has no members")
	}
	parts := ecs.NewMap[comp.Part](r.w.ECS)
	identities := ecs.NewMap[comp.Identity](r.w.ECS)
	for _, member := range members {
		plan := r.plan(t, member)
		var want []string
		for _, def := range plan {
			want = append(want, def.Slot)
		}
		if !slices.Equal(r.slots(member), want) {
			t.Fatalf("slots = %v, want %v", r.slots(member), want)
		}
		got := anatomy.Parts(r.w, member)
		for i, part := range got {
			if r.w.UIDOf(part) == 0 {
				t.Error("a part has no UID")
			}
			data := parts.Get(part)
			if data.Vital != plan[i].Vital {
				t.Errorf("part %q vital = %v, want %v", data.Slot, data.Vital, plan[i].Vital)
			}
			if data.Integrity != plan[i].Integrity || data.MaxIntegrity != plan[i].Integrity {
				t.Errorf("part %q integrity = %d of %d, want %d", data.Slot, data.Integrity, data.MaxIntegrity,
					plan[i].Integrity)
			}
		}
		if ecs.NewMap[comp.Pos](r.w.ECS).Has(got[0]) {
			t.Error("an attached part has a position")
		}
		template, ok := r.w.Content.Template(identities.Get(member).Template)
		if !ok {
			t.Fatalf("unknown template %q", identities.Get(member).Template)
		}
		if protected := ecs.NewMap[comp.Protected](r.w.ECS).Has(member); protected != template.Protected {
			t.Errorf("protected = %v, want %v as the template says", protected, template.Protected)
		}
	}
}

func TestChannelsAreTheBodyThenEachPartTagged(t *testing.T) {
	r := newRig(t, 1, nil)
	member := r.member(t)
	plan := r.plan(t, member)
	channels := timeline.Channels(r.w, member)
	if len(channels) != 1+len(plan) {
		t.Fatalf("channels = %d, want the body and %d parts", len(channels), len(plan))
	}
	if channels[0].Entity != member || !slices.Equal(channels[0].Tags, []string{"body"}) {
		t.Errorf("first channel = %+v, want the body", channels[0])
	}
	for i, c := range channels[1:] {
		want := []string{"part", plan[i].Slot}
		for _, grant := range plan[i].Grants {
			want = append(want, "grant:"+grant.Kind)
		}
		if !slices.Equal(c.Tags, want) {
			t.Errorf("%s tags = %v, want %v", plan[i].Slot, c.Tags, want)
		}
	}
	ecs.NewMap[comp.PartOf](r.w.ECS).Remove(anatomy.Parts(r.w, member)[1])
	if got := len(timeline.Channels(r.w, member)); got != len(plan) {
		t.Errorf("channels after severing a part = %d, want %d", got, len(plan))
	}
}

func TestLosingTheSpeedPartsSlowsSteps(t *testing.T) {
	r := newRig(t, 1, nil)
	member := r.member(t)
	fast, err := timeline.StepTicks(r.w, member)
	if err != nil {
		t.Fatal(err)
	}
	fastSpeed, err := timeline.Speed(r.w, member)
	if err != nil {
		t.Fatal(err)
	}
	parts := ecs.NewMap[comp.Part](r.w.ECS)
	severed := 0
	for _, part := range anatomy.Parts(r.w, member) {
		if slices.ContainsFunc(parts.Get(part).Grants, func(g comp.Grant) bool { return g.Kind == "speed" }) {
			ecs.NewMap[comp.PartOf](r.w.ECS).Remove(part)
			severed++
		}
	}
	if severed == 0 {
		t.Fatal("no part grants speed, so this test checks nothing")
	}
	speed, err := timeline.Speed(r.w, member)
	if err != nil || speed >= fastSpeed {
		t.Errorf("speed without its speed parts = %d, %v; want less than %d", speed, err, fastSpeed)
	}
	slow, err := timeline.StepTicks(r.w, member)
	if err != nil {
		t.Fatal(err)
	}
	if slow <= fast {
		t.Errorf("step ticks = %d without its speed parts and %d with; want slower", slow, fast)
	}
}

func TestAFailingSpeedFormulaIsAnError(t *testing.T) {
	r := newRig(t, 1, nil)
	member := r.member(t)
	bundle := contenttest.SampleBundle(t)
	bundle.Timeline.Speed = "grant.speed % (grant.speed - grant.speed)"
	r.w.Content = contenttest.Registry(t, bundle)
	if speed, err := timeline.Speed(r.w, member); err == nil {
		t.Errorf("speed = %d with a formula that divides by zero, want an error", speed)
	}
	if ticks, err := timeline.StepTicks(r.w, member); err == nil {
		t.Errorf("step ticks = %d with a formula that divides by zero, want an error", ticks)
	}
}

func TestGrantsSumThePlansGrants(t *testing.T) {
	r := newRig(t, 1, nil)
	member := r.member(t)
	var grants [][]comp.Grant
	for _, def := range r.plan(t, member) {
		grants = append(grants, def.Grants)
	}
	want := totals(grants...)
	if got := anatomy.GrantTotals(r.w, member); !maps.Equal(got, want) {
		t.Errorf("GrantTotals = %v, want %v", got, want)
	}
	if got := anatomy.Grants(r.w, member, "nothing"); got != 0 {
		t.Errorf("Grants of a kind no part grants = %d, want 0", got)
	}
}

func TestDetachingAPartDropsItFromPartsAndGrants(t *testing.T) {
	r := newRig(t, 1, nil)
	member := r.member(t)
	all := anatomy.GrantTotals(r.w, member)
	count := len(anatomy.Parts(r.w, member))
	part := anatomy.Parts(r.w, member)[1]
	partData := ecs.NewMap[comp.Part](r.w.ECS).Get(part)
	without := maps.Clone(all)
	for _, grant := range partData.Grants {
		without[grant.Kind] -= grant.Value
	}
	maps.DeleteFunc(without, func(_ string, total int) bool { return total == 0 })
	if !anatomy.Functioning(r.w, part) {
		t.Fatal("a fresh part is not functioning")
	}
	ecs.NewMap[comp.PartOf](r.w.ECS).Remove(part)
	if anatomy.Functioning(r.w, part) {
		t.Error("a severed part is functioning")
	}
	if got := len(anatomy.Parts(r.w, member)); got != count-1 {
		t.Errorf("parts = %d, want %d", got, count-1)
	}
	if got := anatomy.GrantTotals(r.w, member); !maps.Equal(got, without) {
		t.Errorf("grants = %v, want %v", got, without)
	}
	ecs.NewMap[comp.PartOf](r.w.ECS).Add(part, &comp.PartOf{}, member)
	if got := anatomy.GrantTotals(r.w, member); !maps.Equal(got, all) {
		t.Errorf("grants after reattaching = %v, want %v", got, all)
	}
	ecs.NewMap[comp.Part](r.w.ECS).Get(part).Integrity = 0
	if got := anatomy.GrantTotals(r.w, member); !maps.Equal(got, without) {
		t.Errorf("grants with a broken part = %v, want %v", got, without)
	}
}

func TestAMaterialOverrideChangesThePartsMaterial(t *testing.T) {
	r := newRig(t, 1, func(b *content.Bundle) {
		b.Templates["template.test.crawler"] = content.Template{Body: ptr(api.ContentID("body.test.crawler"))}
	})
	crawler, err := r.w.Spawn("template.test.crawler", nil)
	if err != nil {
		t.Fatal(err)
	}
	parts := ecs.NewMap[comp.Part](r.w.ECS)
	materials := map[string]api.ContentID{}
	for _, part := range anatomy.Parts(r.w, crawler) {
		materials[parts.Get(part).Slot] = parts.Get(part).Material
	}
	if materials["torso"] != "material.clay" || materials["head"] != "material.flesh" {
		t.Errorf("materials = %v, want a clay torso and a flesh head", materials)
	}
	if ecs.NewMap[comp.Protected](r.w.ECS).Has(crawler) {
		t.Error("an unprotected template got Protected")
	}
}

func ptr[T any](v T) *T { return &v }

func TestSpawningIsDeterministic(t *testing.T) {
	describe := func(r *rig) []string {
		var out []string
		identities := ecs.NewMap[comp.Identity](r.w.ECS)
		for _, member := range anatomy.Members(r.w, r.party) {
			out = append(out, fmt.Sprintf("%s#%d", identities.Get(member).Template, r.w.UIDOf(member)))
			for _, part := range anatomy.Parts(r.w, member) {
				out = append(out, fmt.Sprintf("%s#%d", identities.Get(part).Template, r.w.UIDOf(part)))
			}
		}
		return out
	}
	first, second := describe(newRig(t, 1, nil)), describe(newRig(t, 1, nil))
	if !slices.Equal(first, second) || len(first) == 0 {
		t.Errorf("first = %v\nsecond = %v", first, second)
	}
}
