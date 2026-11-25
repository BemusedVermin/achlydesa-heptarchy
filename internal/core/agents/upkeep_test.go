package agents_test

import (
	"slices"
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

const coalItem = "item.coal"

// coalUpkeep makes a guard's upkeep a coal a day, so that an unpaid recruit does not starve.
func coalUpkeep(bundle *content.Bundle) {
	guard := bundle.Templates[guardTemplate]
	recruit := *guard.Recruit
	recruit.Upkeep = []content.StackDef{{Item: coalItem, Count: 1}}
	guard.Recruit = &recruit
	bundle.Templates[guardTemplate] = guard
}

func (f *townField) hireGuard() ecs.Entity {
	f.t.Helper()
	guard := f.living(guardTemplate)[0]
	f.beside(guard)
	option, ok := f.hireOption(guard)
	if !ok || !option.Enabled {
		f.t.Fatalf("hire option = %+v, found %v, want enabled", option, ok)
	}
	if _, err := f.g.Apply(option.Command); err != nil {
		f.t.Fatal(err)
	}
	return guard
}

func (f *townField) waitPhase() []api.Event {
	f.t.Helper()
	events, err := f.g.Apply(api.Wait{Source: f.g.Current(), Until: api.WaitPhase})
	if err != nil {
		f.t.Fatal(err)
	}
	return events
}

func (f *townField) emptyCoal() {
	inventory.Take(f.w, f.party(), coalItem, 0, inventory.Count(f.w, f.party(), coalItem, 0))
}

func (f *townField) discontent(body ecs.Entity) int {
	discontent := ecs.NewMap[comp.Discontent](f.w.ECS)
	if !discontent.Has(body) {
		return 0
	}
	return discontent.Get(body).Days
}

func (f *townField) terror(body ecs.Entity) float64 {
	for _, held := range ecs.NewMap[comp.Needs](f.w.ECS).Get(body).Levels {
		if held.Need == "need.terror" {
			return held.Level
		}
	}
	f.t.Fatal("no terror need")
	return 0
}

func (f *townField) leaveTheMap() {
	f.t.Helper()
	here := ecs.NewMap[comp.Pos](f.w.ECS).Get(f.guardLeader()).Map
	for _, id := range f.w.Maps.IDs() {
		if id == here {
			continue
		}
		for _, gate := range f.w.Maps.Get(id).Gates() {
			*ecs.NewMap[comp.Pos](f.w.ECS).Get(f.guardLeader()) = comp.Pos{Map: id, At: gate.At}
			f.w.Bus.Publish(api.PartyCrossed{Party: f.w.UIDOf(f.party()), From: here, To: id})
			if _, err := f.w.Bus.Drain(f.w); err != nil {
				f.t.Fatal(err)
			}
			return
		}
	}
	f.t.Fatal("no map with a gate to leave for")
}

func (f *townField) guardLeader() ecs.Entity {
	leader, _ := f.w.ByUID(ecs.NewMap[comp.Party](f.w.ECS).Get(f.party()).Leader)
	return leader
}

func TestAnUnpaidRecruitGrumblesThenDesertsAndLivesOn(t *testing.T) {
	f := newTownField(t, 1, townMap, coalUpkeep)
	guard := f.hireGuard()
	f.emptyCoal()
	graftedPart := anatomy.Parts(f.w, guard)[0]
	ecs.NewMap[comp.Graft](f.w.ECS).Add(graftedPart, &comp.Graft{Piece: "piece.test"})
	var grumbled, deserted int
	for range 6 * 4 {
		if deserted > 0 {
			break
		}
		for _, event := range f.waitPhase() {
			switch event.(type) {
			case api.Grumbled:
				if deserted > 0 {
					t.Error("grumbled after deserting")
				}
				grumbled++
			case api.Deserted:
				deserted++
			}
		}
	}
	if grumbled == 0 || deserted != 1 {
		t.Fatalf("grumbled %d times and deserted %d times, want grumbling then one desertion", grumbled, deserted)
	}
	if _, member := anatomy.PartyOf(f.w, guard); member {
		t.Error("the deserter is still a member")
	}
	if slices.Contains(ecs.NewMap[comp.Party](f.w.ECS).Get(f.party()).Order, f.w.UIDOf(guard)) {
		t.Error("the deserter is still in the party order")
	}
	if got := ecs.NewMap[comp.Allegiance](f.w.ECS).Get(guard).Faction; got != "faction.deserters" {
		t.Errorf("faction = %q, want the deserters", got)
	}
	if got := ecs.NewMap[comp.Actor](f.w.ECS).Get(guard).Controller; got != "agents" {
		t.Errorf("controller = %q, want agents", got)
	}
	if !ecs.NewMap[comp.Individual](f.w.ECS).Has(guard) {
		t.Error("the deserter is not an Individual")
	}
	if !ecs.NewMap[comp.Graft](f.w.ECS).Has(graftedPart) || !f.w.Alive(graftedPart) {
		t.Error("the deserter lost its graft")
	}
	f.leaveTheMap()
	if !f.w.Alive(guard) || !ecs.NewMap[comp.Individual](f.w.ECS).Has(guard) {
		t.Error("the deserter did not outlive the party's leaving the map")
	}
}

func TestRecruitsConfigIsChecked(t *testing.T) {
	tests := []struct {
		name string
		edit func(*content.RecruitsConfig)
		want string
	}{
		{"desert before grumble", func(c *content.RecruitsConfig) { c.DesertAt = c.GrumbleAt }, "desert_at"},
		{"no flee threshold", func(c *content.RecruitsConfig) { c.FleeAt = 0 }, "flee_at"},
		{"an unknown terror need", func(c *content.RecruitsConfig) { c.TerrorNeed = "need.nope" }, "terror_need"},
		{
			"an unknown faction",
			func(c *content.RecruitsConfig) { c.DesertersFaction = "faction.nope" },
			"deserters_faction",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bundle := contenttest.SampleBundle(t)
			tt.edit(&bundle.Recruits)
			if _, err := content.New(bundle); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error = %v, want one containing %q", err, tt.want)
			}
		})
	}
}

func TestAPaidDayLowersDiscontentToZero(t *testing.T) {
	f := newTownField(t, 1, townMap, coalUpkeep)
	guard := f.hireGuard()
	f.emptyCoal()
	for f.discontent(guard) < 2 && f.w.Now < 5000 {
		f.waitPhase()
	}
	if got := f.discontent(guard); got != 2 {
		t.Fatalf("discontent = %d, want 2 after two unpaid days", got)
	}
	inventory.Add(f.w, f.party(), comp.ItemStack{Item: coalItem, Count: 10})
	for f.discontent(guard) > 0 && f.w.Now < 10000 {
		f.waitPhase()
	}
	if got := f.discontent(guard); got != 0 {
		t.Errorf("discontent = %d, want 0 after paid days", got)
	}
	if got := inventory.Count(f.w, f.party(), coalItem, 0); got >= 10 {
		t.Errorf("%d rations left, want the upkeep taken", got)
	}
}

func TestARecruitWhoSeesAnAllyMaimedFleesUntilTheTerrorEases(t *testing.T) {
	f := newTownField(t, 1, townMap, coalUpkeep)
	guard := f.hireGuard()
	leader := f.guardLeader()
	arm := anatomy.Parts(f.w, leader)[0]
	f.w.Bus.Publish(api.PartSevered{
		Body: f.w.UIDOf(leader), Part: f.w.UIDOf(arm), Cause: api.SeveredWorn, Tick: f.w.Now,
	})
	if _, err := f.w.Bus.Drain(f.w); err != nil {
		t.Fatal(err)
	}
	if got := f.terror(guard); got < 0.5 {
		t.Fatalf("terror = %v, want at least the flee threshold", got)
	}
	var fled bool
	for _, event := range f.waitPhase() {
		if _, ok := event.(api.Fled); ok {
			fled = true
		}
	}
	if !fled {
		t.Fatal("no Fled event")
	}
	if got := ecs.NewMap[comp.Actor](f.w.ECS).Get(guard).Controller; got != "agents" {
		t.Errorf("controller = %q, want agents while fleeing", got)
	}
	if _, member := anatomy.PartyOf(f.w, guard); !member {
		t.Error("fleeing ended membership")
	}
	for f.terror(guard) >= 0.5 && f.w.Now < 5000 {
		f.waitPhase()
	}
	f.waitPhase()
	if ecs.NewMap[comp.Fleeing](f.w.ECS).Has(guard) {
		t.Error("still fleeing after the terror eased")
	}
	if got := ecs.NewMap[comp.Actor](f.w.ECS).Get(guard).Controller; got != "follow" {
		t.Errorf("controller = %q, want follow once calm", got)
	}
}

func TestAFleeingRecruitHasNoPlayerOptions(t *testing.T) {
	f := newTownField(t, 1, townMap, coalUpkeep)
	guard := f.hireGuard()
	if _, err := f.g.Apply(api.Control{Party: f.w.UIDOf(f.party()), Mode: api.ControlIndividual}); err != nil {
		t.Fatal(err)
	}
	f.waitPhase()
	leader := f.guardLeader()
	arm := anatomy.Parts(f.w, leader)[0]
	f.w.Bus.Publish(api.PartSevered{
		Body: f.w.UIDOf(leader), Part: f.w.UIDOf(arm), Cause: api.SeveredCut, Source: f.w.UIDOf(leader), Tick: f.w.Now,
	})
	if _, err := f.w.Bus.Drain(f.w); err != nil {
		t.Fatal(err)
	}
	f.waitPhase()
	for range 40 {
		if f.terror(guard) < 0.5 {
			break
		}
		if f.g.Current() == f.w.UIDOf(guard) {
			t.Fatal("the player was asked to decide for a fleeing recruit")
		}
		f.wait()
	}
}
