package agents_test

import (
	"strings"
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/inventory"
)

// beside puts the party's leader in a free cell next to the body.
func (f *townField) beside(body ecs.Entity) {
	f.t.Helper()
	party, _ := anatomy.ControlledParty(f.w)
	leader, _ := f.w.ByUID(ecs.NewMap[comp.Party](f.w.ECS).Get(party).Leader)
	there := *ecs.NewMap[comp.Pos](f.w.ECS).Get(body)
	m := f.w.Maps.Get(there.Map)
	for _, step := range []int{1, -1} {
		cell := there.At
		cell.Hex.Q += step
		if ok, err := m.Standable(cell); err == nil && ok {
			*ecs.NewMap[comp.Pos](f.w.ECS).Get(leader) = comp.Pos{Map: there.Map, At: cell}
			return
		}
	}
	f.t.Fatal("no free cell beside the recruit")
}

func (f *townField) hireOption(recruit ecs.Entity) (api.Option, bool) {
	f.t.Helper()
	view, err := f.g.View()
	if err != nil {
		f.t.Fatal(err)
	}
	for _, option := range view.Options {
		if option.ID == api.HireOption(f.w.UIDOf(recruit)) {
			return option, true
		}
	}
	return api.Option{}, false
}

func (f *townField) party() ecs.Entity {
	party, _ := anatomy.ControlledParty(f.w)
	return party
}

func (f *townField) genericGuards() int {
	n := 0
	query := ecs.NewFilter1[comp.Generic](f.w.ECS).Query()
	for query.Next() {
		if f.w.TemplateOf(query.Entity()) == guardTemplate {
			n++
		}
	}
	return n
}

func TestHiringAGuardTakesThePriceAndAddsThemToTheParty(t *testing.T) {
	f := newTownField(t, 1, townMap, nil)
	guard := f.living(guardTemplate)[0]
	f.beside(guard)
	option, ok := f.hireOption(guard)
	if !ok || !option.Enabled {
		t.Fatalf("hire option = %+v, found %v, want enabled", option, ok)
	}
	if want := "Hire A guard for 1 coal, 2 ration"; option.Label != want {
		t.Errorf("label = %q, want %q", option.Label, want)
	}
	party := f.party()
	rations := inventory.Count(f.w, party, "item.ration", 0)
	coal := inventory.Count(f.w, party, "item.coal", 0)
	events, err := f.g.Apply(option.Command)
	if err != nil {
		t.Fatal(err)
	}
	hired := false
	for _, event := range events {
		if h, ok := event.(api.Hired); ok {
			hired = h.Recruit == f.w.UIDOf(guard) && len(h.Paid) == 2
		}
	}
	if !hired {
		t.Errorf("events = %v, want Hired with the two items paid", events)
	}
	if got := inventory.Count(f.w, party, "item.ration", 0); got != rations-2 {
		t.Errorf("rations = %d, want %d", got, rations-2)
	}
	if got := inventory.Count(f.w, party, "item.coal", 0); got != coal-1 {
		t.Errorf("coal = %d, want %d", got, coal-1)
	}
	if got, ok := anatomy.PartyOf(f.w, guard); !ok || got != party {
		t.Error("the guard is not a member of the party")
	}
	order := ecs.NewMap[comp.Party](f.w.ECS).Get(party).Order
	if order[len(order)-1] != f.w.UIDOf(guard) {
		t.Errorf("party order = %v, want the guard last", order)
	}
	if ecs.NewMap[comp.Generic](f.w.ECS).Has(guard) {
		t.Error("the hired guard is still a generic")
	}
	for f.genericGuards() < 2 && f.w.Now < 2000 {
		f.wait()
	}
	if got := f.genericGuards(); got != 2 {
		t.Errorf("%d generic guards at tick %d, want the town's replaced", got, f.w.Now)
	}
}

func TestTheHiredRecruitActsLikeAnyMemberInTheNextFight(t *testing.T) {
	f := newTownField(t, 1, townMap, nil)
	guard := f.living(guardTemplate)[0]
	f.beside(guard)
	option, _ := f.hireOption(guard)
	if _, err := f.g.Apply(option.Command); err != nil {
		t.Fatal(err)
	}
	if _, err := f.g.Apply(api.Control{Party: f.w.UIDOf(f.party()), Mode: api.ControlIndividual}); err != nil {
		t.Fatal(err)
	}
	for f.g.Current() != f.w.UIDOf(guard) && f.w.Now < 2000 {
		f.wait()
	}
	if f.g.Current() != f.w.UIDOf(guard) {
		t.Fatalf("the recruit never decided by tick %d", f.w.Now)
	}
	view, err := f.g.View()
	if err != nil {
		t.Fatal(err)
	}
	for _, option := range view.Options {
		if strings.HasPrefix(option.ID, "strike/") {
			return
		}
	}
	t.Errorf("no action options for the recruit in %v", view.Options)
}

func TestTheHireOptionIsDisabledWithItsReason(t *testing.T) {
	tests := []struct {
		name string
		edit func(*content.Bundle)
		pre  func(*townField)
		want string
	}{
		{"a full party", func(b *content.Bundle) { b.Party.Slots = 2 }, nil, "no room"},
		{"too few items", nil, func(f *townField) {
			inventory.Take(f.w, f.party(), "item.ration", 0, inventory.Count(f.w, f.party(), "item.ration", 0))
		}, "cannot pay"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newTownField(t, 1, townMap, tt.edit)
			guard := f.living(guardTemplate)[0]
			f.beside(guard)
			if tt.pre != nil {
				tt.pre(f)
			}
			option, ok := f.hireOption(guard)
			if !ok || option.Enabled || !strings.Contains(option.Reason, tt.want) {
				t.Fatalf("option = %+v, found %v, want disabled with %q", option, ok, tt.want)
			}
			before := inventory.Count(f.w, f.party(), "item.coal", 0)
			events, err := f.g.Apply(option.Command)
			if err != nil {
				t.Fatal(err)
			}
			rejected := false
			for _, event := range events {
				_, isRejected := event.(api.Rejected)
				rejected = rejected || isRejected
			}
			if !rejected || inventory.Count(f.w, f.party(), "item.coal", 0) != before {
				t.Errorf("events = %v, want a rejection and nothing paid", events)
			}
		})
	}
}

func TestRecruitContentIsChecked(t *testing.T) {
	tests := []struct {
		name string
		edit func(*content.RecruitDef)
		want string
	}{
		{"an unknown item", func(r *content.RecruitDef) { r.Wants[0].Item = "item.nope" }, "unknown item"},
		{"no wants", func(r *content.RecruitDef) { r.Wants = nil }, "at least one"},
		{"a bad worth", func(r *content.RecruitDef) { r.Worth = "nope +" }, "worth"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bundle := contenttest.SampleBundle(t)
			template := bundle.Templates[guardTemplate]
			recruit := *template.Recruit
			recruit.Wants = append([]content.StackDef(nil), recruit.Wants...)
			tt.edit(&recruit)
			template.Recruit = &recruit
			bundle.Templates[guardTemplate] = template
			if _, err := content.New(bundle); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error = %v, want one containing %q", err, tt.want)
			}
		})
	}
}
