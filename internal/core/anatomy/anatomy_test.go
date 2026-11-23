package anatomy_test

import (
	"bytes"
	"fmt"
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
	bundle.Start = content.Start{} // the rig spawns its own party
	if mutate != nil {
		mutate(&bundle)
	}
	reg, err := content.New(bundle)
	if err != nil {
		t.Fatal(err)
	}
	r := &rig{}
	r.cfg = game.Config{
		Seed: seed, Content: reg,
		Install: func(w *sim.World, _ *clock.Clock) error {
			r.w = w
			return nil
		},
		Spawn: func(w *sim.World, _ *clock.Clock) error {
			var err error
			r.party, err = w.Spawn("template.party.main", nil)
			return err
		},
	}
	if r.g, err = game.New(r.cfg); err != nil {
		t.Fatal(err)
	}
	return r
}

func (r *rig) members() []ecs.Entity { return anatomy.Members(r.w, r.party) }

func (r *rig) slots(body ecs.Entity) []string {
	var slots []string
	parts := ecs.NewMap[comp.Part](r.w.ECS)
	for _, part := range anatomy.Parts(r.w, body) {
		slots = append(slots, parts.Get(part).Slot)
	}
	return slots
}

func TestPartyMembersHaveFiveHumanParts(t *testing.T) {
	r := newRig(t, 1, nil)
	members := r.members()
	if len(members) != 2 {
		t.Fatalf("members = %d, want 2", len(members))
	}
	parts := ecs.NewMap[comp.Part](r.w.ECS)
	for _, member := range members {
		got := anatomy.Parts(r.w, member)
		if want := []string{"head", "arm", "arm", "legs", "torso"}; !slices.Equal(r.slots(member), want) {
			t.Fatalf("slots = %v, want %v", r.slots(member), want)
		}
		for _, part := range got {
			if r.w.UIDOf(part) == 0 {
				t.Error("a part has no UID")
			}
			data := parts.Get(part)
			if data.Vital != (data.Slot == "torso") {
				t.Errorf("part %q vital = %v", data.Slot, data.Vital)
			}
			if data.Integrity != data.MaxIntegrity || data.Integrity < 1 {
				t.Errorf("part %q integrity = %d of %d", data.Slot, data.Integrity, data.MaxIntegrity)
			}
		}
		if ecs.NewMap[comp.Pos](r.w.ECS).Has(got[0]) {
			t.Error("an attached part has a position")
		}
		if !ecs.NewMap[comp.Protected](r.w.ECS).Has(member) {
			t.Error("a protagonist is not protected")
		}
	}
}

func TestAHumanHasSixTaggedChannels(t *testing.T) {
	r := newRig(t, 1, nil)
	member := r.members()[0]
	channels := timeline.Channels(r.w, member)
	if len(channels) != 6 {
		t.Fatalf("channels = %d, want 6", len(channels))
	}
	if channels[0].Entity != member || !slices.Equal(channels[0].Tags, []string{"body"}) {
		t.Errorf("first channel = %+v, want the body", channels[0])
	}
	want := map[string][]string{
		"head":  {"part", "head", "grant:targeting"},
		"arm":   {"part", "arm", "grant:action"},
		"legs":  {"part", "legs", "grant:speed"},
		"torso": {"part", "torso"},
	}
	parts := ecs.NewMap[comp.Part](r.w.ECS)
	for _, c := range channels[1:] {
		if slot := parts.Get(c.Entity).Slot; !slices.Equal(c.Tags, want[slot]) {
			t.Errorf("%s tags = %v, want %v", slot, c.Tags, want[slot])
		}
	}
	// Severing an arm removes its channel.
	arm := anatomy.Parts(r.w, member)[1]
	ecs.NewMap[comp.PartOf](r.w.ECS).Remove(arm)
	if got := len(timeline.Channels(r.w, member)); got != 5 {
		t.Errorf("channels after severing an arm = %d, want 5", got)
	}
}

func TestLosingTheLegsSlowsSteps(t *testing.T) {
	r := newRig(t, 1, nil)
	member := r.members()[0]
	fast, err := timeline.StepTicks(r.w, member)
	if err != nil {
		t.Fatal(err)
	}
	legs := anatomy.Parts(r.w, member)[3]
	ecs.NewMap[comp.PartOf](r.w.ECS).Remove(legs)
	if speed := timeline.Speed(r.w, member); speed != 0 {
		t.Errorf("speed without legs = %d, want 0", speed)
	}
	slow, err := timeline.StepTicks(r.w, member)
	if err != nil {
		t.Fatal(err)
	}
	if slow <= fast {
		t.Errorf("step ticks = %d without legs and %d with; want slower", slow, fast)
	}
}

func TestGrantsCountPartsAndSumSpeed(t *testing.T) {
	r := newRig(t, 1, nil)
	member := r.members()[0]
	for kind, want := range map[string]int{"action": 2, "targeting": 1, "speed": 2, "nothing": 0} {
		if got := anatomy.Grants(r.w, member, kind); got != want {
			t.Errorf("Grants(%q) = %d, want %d", kind, got, want)
		}
	}
}

func TestDetachingAPartDropsItFromPartsAndGrants(t *testing.T) {
	r := newRig(t, 1, nil)
	member := r.members()[0]
	arm := anatomy.Parts(r.w, member)[1]
	if !anatomy.Functioning(r.w, arm) {
		t.Fatal("a fresh arm is not functioning")
	}
	ecs.NewMap[comp.PartOf](r.w.ECS).Remove(arm)
	if anatomy.Functioning(r.w, arm) {
		t.Error("a severed arm is functioning")
	}
	if got := len(anatomy.Parts(r.w, member)); got != 4 {
		t.Errorf("parts = %d, want 4", got)
	}
	if got := anatomy.Grants(r.w, member, "action"); got != 1 {
		t.Errorf("action grants = %d, want 1", got)
	}
	// Reattaching brings it back.
	ecs.NewMap[comp.PartOf](r.w.ECS).Add(arm, &comp.PartOf{}, member)
	if got := anatomy.Grants(r.w, member, "action"); got != 2 {
		t.Errorf("action grants after reattaching = %d, want 2", got)
	}
	// A part with no integrity left does not function, so it grants nothing.
	ecs.NewMap[comp.Part](r.w.ECS).Get(arm).Integrity = 0
	if got := anatomy.Grants(r.w, member, "action"); got != 1 {
		t.Errorf("action grants with a broken arm = %d, want 1", got)
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
		for _, member := range r.members() {
			out = append(out, fmt.Sprintf("%s#%d", identities.Get(member).Template, r.w.UIDOf(member)))
			for _, part := range anatomy.Parts(r.w, member) {
				out = append(out, fmt.Sprintf("%s#%d", identities.Get(part).Template, r.w.UIDOf(part)))
			}
		}
		return out
	}
	first, second := describe(newRig(t, 1, nil)), describe(newRig(t, 1, nil))
	if !slices.Equal(first, second) || len(first) != 12 {
		t.Errorf("first = %v\nsecond = %v", first, second)
	}
}

func TestBodiesAndPartsSurviveSaveAndLoad(t *testing.T) {
	r := newRig(t, 1, nil)
	var saved bytes.Buffer
	if err := r.g.Save(&saved); err != nil {
		t.Fatal(err)
	}
	loaded := &rig{}
	cfg := r.cfg
	cfg.Install = func(w *sim.World, _ *clock.Clock) error {
		loaded.w = w
		return nil
	}
	if _, _, err := game.Load(cfg, &saved); err != nil {
		t.Fatal(err)
	}
	parties := ecs.NewFilter1[comp.Party](loaded.w.ECS).Query()
	for parties.Next() {
		loaded.party = parties.Entity()
		parties.Close()
		break
	}
	before, after := r.members(), loaded.members()
	if len(after) != len(before) {
		t.Fatalf("members = %d, want %d", len(after), len(before))
	}
	for i := range before {
		if r.w.UIDOf(before[i]) != loaded.w.UIDOf(after[i]) {
			t.Errorf("member %d UID changed", i)
		}
		if !slices.Equal(r.slots(before[i]), loaded.slots(after[i])) {
			t.Errorf("member %d slots = %v, want %v", i, loaded.slots(after[i]), r.slots(before[i]))
		}
		for kind, want := range map[string]int{"action": 2, "targeting": 1, "speed": 2} {
			if got := anatomy.Grants(loaded.w, after[i], kind); got != want {
				t.Errorf("member %d Grants(%q) = %d, want %d", i, kind, got, want)
			}
		}
		if got := len(timeline.Channels(loaded.w, after[i])); got != 6 {
			t.Errorf("member %d channels = %d, want 6", i, got)
		}
	}
}
