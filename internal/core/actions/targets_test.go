package actions_test

import (
	"slices"
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/actions"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
)

// withFilteredActions adds body actions that differ only in their target filter, and tags the
// dummy's template with "ward".
func withFilteredActions(filters map[string]content.TargetFilter) func(*content.Bundle) {
	return func(b *content.Bundle) {
		for key, filter := range filters {
			id := "action.test." + key
			b.Actions[id] = content.ActionDef{
				Name: api.ContentID("text." + id), Key: key, Uses: use("arm", 1), Target: "body", Targets: filter,
				Range: 2, Windup: "0", Recovery: "1", OnLand: hit,
			}
			b.Text["text."+id] = key
		}
		dummy := b.Templates["template.test.dummy"]
		dummy.Tags = []string{"ward"}
		b.Templates["template.test.dummy"] = dummy
	}
}

// targetsOf lists the actions the fighter may start now against the dummy.
func targetsOf(fx *fixture) []api.ContentID {
	var out []api.ContentID
	for _, in := range actions.Startable(fx.w, fx.fighter) {
		if in.Target == fx.dummy && !slices.Contains(out, in.Action) {
			out = append(out, in.Action)
		}
	}
	slices.Sort(out)
	return out
}

func TestOnlyACorpseActionMayTargetADeadBody(t *testing.T) {
	known := []api.ContentID{"action.test.hit", "action.test.gnaw"}
	fx := newFixture(t, 1, opts{actions: known, mutate: withFilteredActions(map[string]content.TargetFilter{
		"hit":  {},
		"gnaw": {Corpse: true},
	})})
	if got, want := targetsOf(fx), known; !slices.Equal(got, slices.Sorted(slices.Values(want))) {
		t.Fatalf("against a living dummy, startable = %v, want %v", got, want)
	}
	ecs.NewMap[comp.Actor](fx.w.ECS).Remove(fx.dummy)
	ecs.NewMap[comp.Dead](fx.w.ECS).Add(fx.dummy, &comp.Dead{Tick: fx.w.Now})
	if got := targetsOf(fx); !slices.Equal(got, []api.ContentID{"action.test.gnaw"}) {
		t.Errorf("against a corpse, startable = %v, want only gnaw", got)
	}
	ok, reason := actions.Validate(fx.w, actions.Intent{Actor: fx.fighter, Action: "action.test.hit", Target: fx.dummy})
	if ok || reason.ID != "text.actions.bad_target" {
		t.Errorf("hit on a corpse: ok = %v, reason = %q; want refused as bad_target", ok, reason.ID)
	}
	if ok, reason := actions.Validate(
		fx.w, actions.Intent{Actor: fx.fighter, Action: "action.test.gnaw", Target: fx.dummy},
	); !ok {
		t.Errorf("gnaw on a corpse refused: %q", reason.ID)
	}
}

func TestTagFilterKeepsABodyOut(t *testing.T) {
	filters := map[string]content.TargetFilter{
		"any":    {},
		"warded": {Tags: []string{"ward"}},
		"unward": {Not: []string{"ward"}},
		"rooted": {Tags: []string{"root"}},
		"mixed":  {Tags: []string{"ward"}, Not: []string{"root"}},
	}
	var known []api.ContentID
	for key := range filters {
		known = append(known, api.ContentID("action.test."+key))
	}
	slices.Sort(known)
	fx := newFixture(t, 1, opts{actions: known, mutate: withFilteredActions(filters)})
	want := []api.ContentID{"action.test.any", "action.test.mixed", "action.test.warded"}
	if got := targetsOf(fx); !slices.Equal(got, want) {
		t.Errorf("startable against a warded dummy = %v, want %v", got, want)
	}
	for _, refused := range []api.ContentID{"action.test.unward", "action.test.rooted"} {
		ok, reason := actions.Validate(fx.w, actions.Intent{Actor: fx.fighter, Action: refused, Target: fx.dummy})
		if ok || reason.ID != "text.actions.bad_target" {
			t.Errorf("%s: ok = %v, reason = %q; want refused as bad_target", refused, ok, reason.ID)
		}
	}
}
