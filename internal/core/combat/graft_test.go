package combat_test

import (
	"slices"
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/effects"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/inventory"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

// always makes the first trigger of a record fire on every event it listens for: the stream cannot
// be forced, so a condition that every roll meets stands in for a roll of 0.
const rollAlways = "roll < 2"

// apply runs effects for a source and target, and delivers the events they publish.
func (f *field) apply(source, target, targetPart ecs.Entity, refs ...content.EffectRef) {
	f.t.Helper()
	ctx := &effects.Context{World: f.w, Source: source, Target: target, TargetPart: targetPart}
	if err := effects.Apply(ctx, refs); err != nil {
		f.t.Fatal(err)
	}
	if _, err := f.w.Bus.Drain(f.w); err != nil {
		f.t.Fatal(err)
	}
}

func TestAGraftActsOnItsOwnAndLeavesTheOtherLimbsFree(t *testing.T) {
	f := huntField(t, func(b *content.Bundle) {
		item := b.Items["item.serilmen_arm"]
		item.Triggers[0].If = rollAlways
		b.Items["item.serilmen_arm"] = item
	})
	creature := f.creature()
	party, _ := f.w.ByUID(f.view().Party.UID)
	inventory.Add(f.w, party, comp.ItemStack{Item: "item.serilmen_arm", Count: 1})
	f.apply(f.members()[0], creature, ecs.Entity{}, do("attach_part", "from", "inventory", "item", "item.serilmen_arm"))
	graft := f.partOf(creature, "arm")
	maw := f.partOf(creature, "head")
	if got := ecs.NewMap[comp.Graft](f.w.ECS).Get(graft).Piece; got != "item.serilmen_arm" {
		t.Fatalf("graft piece = %q", got)
	}
	if !ecs.NewMap[comp.Triggers](f.w.ECS).Has(graft) {
		t.Fatal("the piece's triggers did not move onto the part")
	}
	f.moveParty(east(f.at(creature), -1))

	var started []api.ActionStarted
	var mawFree, graftFree []bool
	f.w.Bus.SubscribeAll("test.auto", func(w *sim.World, e api.Event) {
		if s, ok := e.(api.ActionStarted); ok && s.Autonomous {
			started = append(started, s)
			mawFree = append(mawFree, timeline.Free(w, timeline.Channel{Entity: maw, Body: creature}))
			graftFree = append(graftFree, timeline.Free(w, timeline.Channel{Entity: graft, Body: creature}))
		}
	})
	uid := f.w.UIDOf(creature)
	if !f.until(300, func() bool { return len(started) > 0 }) {
		t.Fatal("the graft never acted on its own")
	}
	if started[0].Source != uid || started[0].Action != "action.strike" {
		t.Errorf("started = %+v, want the host's strike", started[0])
	}
	if graftFree[0] {
		t.Error("the graft's channel is free after it started a swing")
	}
	if !mawFree[0] {
		t.Error("the host's other limb was occupied by the graft's swing")
	}
}

func TestActAloneIsRefusedForAPartThatIsNotAGraft(t *testing.T) {
	f := huntField(t, nil)
	creature := f.creature()
	f.moveParty(east(f.at(creature), -1))
	maw := f.partOf(creature, "head")
	ctx := &effects.Context{World: f.w, Source: creature, SourcePart: maw}
	if err := effects.Apply(ctx, []content.EffectRef{do("act_alone")}); err != nil {
		t.Fatal(err)
	}
	if _, err := f.w.Bus.Drain(f.w); err != nil {
		t.Fatal(err)
	}
	if !timeline.Free(f.w, timeline.Channel{Entity: maw, Body: creature}) {
		t.Error("a part that is no graft acted on its own")
	}
}

// remnantField is the hall with a remnant standing where the hall's creature does, and the party
// three cells off. Every severed part crawls, and a worn-out one does not.
func remnantField(t *testing.T) (f *field, remnant ecs.Entity) {
	t.Helper()
	f = huntField(t, func(b *content.Bundle) {
		template := b.Templates["template.creature.remnant"]
		template.Triggers[0].If = rollAlways + " &&target.integrity > 0"
		b.Templates["template.creature.remnant"] = template
	})
	pos := *ecs.NewMap[comp.Pos](f.w.ECS).Get(f.creature())
	var err error
	if remnant, err = f.w.Spawn("template.creature.remnant", &pos); err != nil {
		t.Fatal(err)
	}
	f.moveParty(east(pos.At, -3))
	return f, remnant
}

func TestASeveredPartOfAPreEntropyThingCrawlsOff(t *testing.T) {
	f, remnant := remnantField(t)
	arm := f.partOf(remnant, "arm")
	at := *ecs.NewMap[comp.Pos](f.w.ECS).Get(remnant)
	f.apply(f.members()[0], remnant, arm, do("sever", "parts", "target_part"))

	partOf := ecs.NewMap[comp.PartOf](f.w.ECS)
	if !partOf.Has(arm) {
		t.Fatal("the severed arm did not become a body")
	}
	body := partOf.GetRelation(arm)
	if body == remnant {
		t.Fatal("the arm is still on the remnant")
	}
	if got := *ecs.NewMap[comp.Pos](f.w.ECS).Get(body); got != at {
		t.Errorf("the crawling arm is at %+v, want %+v", got, at)
	}
	if got := ecs.NewMap[comp.Allegiance](f.w.ECS).Get(body).Faction; got != "faction.crawling" {
		t.Errorf("faction = %q", got)
	}
	if got := ecs.NewMap[comp.Actor](f.w.ECS).Get(body).Controller; got != "hunt" {
		t.Errorf("controller = %q", got)
	}
}

func TestACrawlingArmSwingsAtAPartyMemberInReach(t *testing.T) {
	f, remnant := remnantField(t)
	reaper := f.members()[0]
	arm := f.partOf(remnant, "arm")
	f.apply(reaper, remnant, arm, do("sever", "parts", "target_part"))
	crawling := ecs.NewMap[comp.PartOf](f.w.ECS).GetRelation(arm)
	// The source and the hall's own creature fall, and the party stands next to the arm: the only
	// hostiles in its reach are the party.
	f.damage(reaper, remnant, "torso", "99")
	f.damage(reaper, f.creature(), "torso", "99")
	f.moveParty(east(f.at(crawling), -1))

	uid := f.w.UIDOf(crawling)
	members := f.memberUIDs()
	var swing *api.ActionStarted
	f.w.Bus.SubscribeAll("test.crawling_swing", func(_ *sim.World, e api.Event) {
		if s, ok := e.(api.ActionStarted); ok && s.Source == uid && swing == nil {
			swing = &s
		}
	})
	if !f.until(300, func() bool { return swing != nil }) {
		t.Fatal("the crawling arm never started an action")
	}
	if !slices.Contains([]api.ContentID{"action.strike", "action.sever", "action.pin"}, swing.Action) {
		t.Errorf("the crawling arm started %q, want a form its action grant pays for", swing.Action)
	}
	if !slices.Contains(members, swing.Target) {
		t.Errorf("the crawling arm swung at %v, want one of the party %v", swing.Target, members)
	}
}

func TestACrawlingPartTakesItsCrawlsAsIdentity(t *testing.T) {
	tests := []struct {
		name       string
		crawlsAs   api.ContentID
		want       api.ContentID
		appearance bool
	}{
		{"with crawls_as", "template.crawling.remnant_arm", "template.crawling.remnant_arm", true},
		{"without crawls_as", "", "part.remnant_arm", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := huntField(t, func(b *content.Bundle) {
				template := b.Templates["template.creature.remnant"]
				template.Triggers[0].If = rollAlways
				b.Templates["template.creature.remnant"] = template
				part := b.Parts["part.remnant_arm"]
				part.CrawlsAs = tt.crawlsAs
				b.Parts["part.remnant_arm"] = part
			})
			pos := *ecs.NewMap[comp.Pos](f.w.ECS).Get(f.creature())
			remnant, err := f.w.Spawn("template.creature.remnant", &pos)
			if err != nil {
				t.Fatal(err)
			}
			arm := f.partOf(remnant, "arm")
			f.apply(f.members()[0], remnant, arm, do("sever", "parts", "target_part"))
			crawling := ecs.NewMap[comp.PartOf](f.w.ECS).GetRelation(arm)
			if got := ecs.NewMap[comp.Identity](f.w.ECS).Get(crawling).Template; got != tt.want {
				t.Errorf("identity = %q, want %q", got, tt.want)
			}
			if got := ecs.NewMap[comp.Appearance](f.w.ECS).Has(crawling); got != tt.appearance {
				t.Errorf("has appearance = %v, want %v", got, tt.appearance)
			}
		})
	}
}

func TestAPartOfAThingThatDiesNormallyDoesNotCrawl(t *testing.T) {
	f, _ := remnantField(t)
	creature := f.creature()
	arm := f.partOf(creature, "head")
	f.apply(f.members()[0], creature, arm, do("sever", "parts", "target_part"))
	if got := ecs.NewMap[comp.PartOf](f.w.ECS).Has(arm); got {
		t.Error("a part of an ordinary creature crawled off")
	}
}

func TestTheAlertHoldsWhileACrawlingPartOutlivesItsSource(t *testing.T) {
	f, remnant := remnantField(t)
	reaper := f.members()[0]
	arm := f.partOf(remnant, "arm")
	f.apply(reaper, remnant, arm, do("sever", "parts", "target_part"))
	crawling := ecs.NewMap[comp.PartOf](f.w.ECS).GetRelation(arm)
	// The source and the hall's own creature go down: only the crawling arm is left to see.
	f.damage(reaper, remnant, "torso", "99")
	f.damage(reaper, f.creature(), "torso", "99")
	if !ecs.NewMap[comp.Dead](f.w.ECS).Has(remnant) {
		t.Fatal("the remnant did not go down")
	}
	for range 6 {
		f.option("wait/event")
	}
	if !f.view().Party.Alert {
		t.Fatal("the alert dropped while the crawling arm still lived")
	}
	f.damage(reaper, crawling, "arm", "99")
	if !ecs.NewMap[comp.Dead](f.w.ECS).Has(crawling) {
		t.Fatal("the crawling arm did not fall")
	}
	for range 400 {
		if !f.view().Party.Alert {
			return
		}
		f.option("wait/event")
	}
	t.Error("the alert never ended after the crawling arm died")
}

func TestRotEatsAGraftOnEitherSide(t *testing.T) {
	r := newRig(t)
	rot, ok := r.w.Content.Type("type.rot")
	if !ok {
		t.Fatal("no rot")
	}
	parts := ecs.NewMap[comp.Part](r.w.ECS)
	grafts := ecs.NewMap[comp.Graft](r.w.ECS)
	for name, body := range map[string]ecs.Entity{"ally": anatomy.Members(r.w, r.party)[1], "enemy": r.crawler} {
		inventory.Add(r.w, r.party, comp.ItemStack{Item: "item.test_arm", Count: 1})
		r.apply(t, r.ctx(body), do("attach_part", "from", "inventory", "item", "item.test_arm"))
		seen := false
		for _, part := range anatomy.Parts(r.w, body) {
			before := parts.Get(part).Integrity
			ctx := r.ctx(body)
			ctx.TargetPart = part
			r.apply(t, ctx, rot.OnLand...)
			want := before
			if grafts.Has(part) {
				want -= 2
				seen = true
			}
			if got := parts.Get(part).Integrity; got != want {
				t.Errorf("%s: graft %v: integrity = %d, want %d", name, grafts.Has(part), got, want)
			}
		}
		if !seen {
			t.Errorf("%s: no graft to eat", name)
		}
	}
}

func TestAReattachedLimbIsNotAGraftForRelicOrRot(t *testing.T) {
	r := newRig(t)
	rot, ok := r.w.Content.Type("type.rot")
	if !ok {
		t.Fatal("no rot")
	}
	arm := r.partsOf(r.reaper, "arm")[0]
	ctx := r.ctx(r.reaper)
	ctx.TargetPart = arm
	r.apply(t, ctx, do("sever", "parts", "target_part"))
	r.apply(t, r.ctx(r.reaper), do("attach_part", "from", "severed"))
	if got := effects.BodyEnvOf(r.w, r.reaper, arm); got.Grafts != 0 || got.Graft {
		t.Errorf("a reattached limb counts as a graft: grafts = %d, graft = %v", got.Grafts, got.Graft)
	}
	before := ecs.NewMap[comp.Part](r.w.ECS).Get(arm).Integrity
	ctx = r.ctx(r.reaper)
	ctx.TargetPart = arm
	r.apply(t, ctx, rot.OnLand...)
	if got := ecs.NewMap[comp.Part](r.w.ECS).Get(arm).Integrity; got != before {
		t.Errorf("Rot ate a reattached limb: integrity %d, want %d", got, before)
	}
}
