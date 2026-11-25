package agents_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/agents"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/effects"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game/gametest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

const (
	feedersMap = "map.play.feeders"
	shrineMap  = "map.play.shrine"
	remnantFam = "family.remnant"
	wardenTmpl = "template.creature.archon_warden"
)

// monsterField starts the party in one of the playground rooms, with the monster families as
// shipped.
type monsterField struct {
	t      *testing.T
	g      *game.Game
	w      *sim.World
	events []api.Event
}

func newMonsterField(t *testing.T, start api.ContentID, mutate func(*content.Bundle)) *monsterField {
	t.Helper()
	f := &monsterField{t: t}
	f.g, f.w, _ = gametest.New(t, 1, func(bundle *content.Bundle) {
		bundle.Start.Map = start
		bundle.Start.PartyAtMark = "from_hub"
		if mutate != nil {
			mutate(bundle)
		}
	})
	f.w.Bus.SubscribeAll("test.record", func(_ *sim.World, e api.Event) { f.events = append(f.events, e) })
	return f
}

func peaceful(bundle *content.Bundle) {
	for _, id := range []string{"faction.party", "faction.wild", "faction.archon"} {
		contenttest.Edit(bundle.Factions, id, func(faction *content.Faction) { faction.HostileTo = nil })
	}
}

func (f *monsterField) wait(until string) {
	f.t.Helper()
	if _, err := f.g.Apply(api.Wait{Source: f.g.Current(), Until: until}); err != nil {
		f.t.Fatal(err)
	}
}

func (f *monsterField) bodies(template api.ContentID) []ecs.Entity {
	var found []ecs.Entity
	query := ecs.NewFilter1[comp.Identity](f.w.ECS).Query()
	for query.Next() {
		if query.Get().Template == template {
			found = append(found, query.Entity())
		}
	}
	return f.w.SortByUID(found)
}

func (f *monsterField) bodiesOn(on api.MapID, template api.ContentID) []ecs.Entity {
	positions := ecs.NewMap[comp.Pos](f.w.ECS)
	return slices.DeleteFunc(f.bodies(template), func(body ecs.Entity) bool { return positions.Get(body).Map != on })
}

func (f *monsterField) kill(body ecs.Entity) {
	f.t.Helper()
	f.damage(body, "torso", "99")
	if !ecs.NewMap[comp.Dead](f.w.ECS).Has(body) {
		f.t.Fatal("the body did not die")
	}
}

func (f *monsterField) damage(body ecs.Entity, slot, amount string) {
	f.t.Helper()
	ctx := &effects.Context{World: f.w, Source: body, Target: body}
	refs := []content.EffectRef{{Do: "damage", Params: map[string]string{"parts": "target:" + slot, "amount": amount}}}
	if err := effects.Apply(ctx, refs); err != nil {
		f.t.Fatal(err)
	}
	if _, err := f.w.Bus.Drain(f.w); err != nil {
		f.t.Fatal(err)
	}
}

func (f *monsterField) strikeTargets(source api.UID) []api.UID {
	var targets []api.UID
	for _, e := range f.events {
		if s, ok := e.(api.ActionStarted); ok && s.Source == source && s.Action == "action.strike" {
			targets = append(targets, s.Target)
		}
	}
	return targets
}

func (f *monsterField) migrated(family api.ContentID, from api.MapID) int {
	count := 0
	for _, e := range f.events {
		if m, ok := e.(api.Migrated); ok && m.Family == family && m.From == from {
			count += m.Count
		}
	}
	return count
}

func (f *monsterField) members() []ecs.Entity {
	party, ok := anatomy.ControlledParty(f.w)
	if !ok {
		f.t.Fatal("no controlled party")
	}
	return anatomy.Members(f.w, party)
}

func (f *monsterField) population(on api.MapID, family api.ContentID) int {
	region, ok := agents.RegionOf(f.w, on)
	if !ok {
		f.t.Fatalf("map %d has no region", on)
	}
	for _, p := range ecs.NewMap[comp.Populations](f.w.ECS).Get(region).Of {
		if p.Family == family {
			return p.Count
		}
	}
	return 0
}

func TestMovingClayWalksToACorpseTheFightLeft(t *testing.T) {
	f := newMonsterField(t, playHall, peaceful)
	corpse, err := f.w.Spawn("template.test.dummy", &comp.Pos{Map: mapNamed(t, f.w, playHall), At: cellAt(17, 13)})
	if err != nil {
		t.Fatal(err)
	}
	f.kill(corpse)
	corpseUID := f.w.UIDOf(corpse)
	held := func() bool {
		for _, e := range f.events {
			if s, ok := e.(api.ActionStarted); ok && s.Target == corpseUID && s.Action == "action.hold" {
				return true
			}
		}
		return false
	}
	for range 400 {
		if held() {
			break
		}
		f.wait(api.WaitEvent)
	}
	if !held() {
		t.Fatal("no moving clay took hold of the corpse")
	}
}

func TestMovingClayLeavesTheLivingToTheFightWhenHostilesAreNear(t *testing.T) {
	f := newMonsterField(t, playHall, nil)
	hall := mapNamed(t, f.w, playHall)
	clay := f.bodiesOn(hall, "template.creature.moving_clay")
	sim.Put(f.w, f.members()[0], comp.Pos{Map: hall, At: cellAt(13, 10)})
	var struck bool
	for range 400 {
		f.wait(api.WaitEvent)
		for _, c := range clay {
			struck = struck || len(f.strikeTargets(f.w.UIDOf(c))) > 0
		}
		if struck {
			break
		}
	}
	if !struck {
		t.Error("no moving clay struck at the party")
	}
}

func TestARemnantLeavesDecayedGroundOnScreen(t *testing.T) {
	tests := []struct {
		name  string
		decay float64
		want  bool
	}{
		{"decayed ground", 0.8, true},
		{"sound ground", 0.1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newMonsterField(t, "map.play.den", func(bundle *content.Bundle) {
				peaceful(bundle)
				contenttest.Edit(bundle.Maps, "map.play.den", func(den *content.MapDef) {
					den.Region.Props = map[string]float64{"decay": tt.decay}
				})
			})
			den := mapNamed(t, f.w, "map.play.den")
			for range 8 {
				f.wait(api.WaitPhase)
			}
			left := f.migrated(remnantFam, den) > 0
			if left != tt.want {
				t.Errorf("remnant left the den = %v, want %v", left, tt.want)
			}
			if tt.want && len(f.bodiesOn(den, "template.creature.remnant")) != 0 {
				t.Error("the remnant is still in the den")
			}
		})
	}
}

func TestRemnantsDriftTowardLessDecayedGroundOffScreen(t *testing.T) {
	f := newMonsterField(t, playHall, func(bundle *content.Bundle) {
		contenttest.Edit(bundle.Maps, playHall, func(hall *content.MapDef) { hall.Region.Populations = nil })
		contenttest.Edit(bundle.Maps, "map.play.den", func(den *content.MapDef) {
			den.Spawns = nil
			den.Region.Populations = map[string]int{remnantFam: 8}
		})
	})
	den, hub := mapNamed(t, f.w, "map.play.den"), mapNamed(t, f.w, hubMap)
	for range 3 {
		f.wait(api.WaitPhase)
	}
	if got := f.population(den, remnantFam); got >= 8 {
		t.Errorf("%d remnants still in the decayed den, want fewer than 8", got)
	}
	total := 0
	for _, id := range f.w.Maps.IDs() {
		if strings.HasPrefix(f.w.Maps.Get(id).Name(), "map.play.") {
			total += f.population(id, remnantFam)
		}
	}
	if total != 8 {
		t.Errorf("%d remnants counted in all, want 8", total)
	}
	if f.population(hub, remnantFam) == 0 {
		t.Error("no remnant reached the less decayed hub")
	}
}

func TestMovingClayDriftsTowardTheDeadOffScreen(t *testing.T) {
	f := newMonsterField(t, playHall, func(bundle *content.Bundle) {
		contenttest.Edit(bundle.Maps, playHall, func(hall *content.MapDef) { hall.Region.Populations = nil })
		contenttest.Edit(bundle.Maps, "map.play.den", func(den *content.MapDef) {
			den.Spawns = nil
			den.Region.Populations = map[string]int{"family.moving_clay": 8}
		})
	})
	den, hub := mapNamed(t, f.w, "map.play.den"), mapNamed(t, f.w, hubMap)
	corpse, err := f.w.Spawn("template.test.dummy", &comp.Pos{Map: hub, At: cellAt(2, 4)})
	if err != nil {
		t.Fatal(err)
	}
	f.kill(corpse)
	for range 2 {
		f.wait(api.WaitPhase)
	}
	if got := f.population(hub, "family.moving_clay"); got == 0 {
		t.Error("no clay drifted to the hub, where the dead lie")
	}
	if got := f.population(den, "family.moving_clay"); got >= 8 {
		t.Errorf("%d clay still in the den, want fewer than 8", got)
	}
}

func TestASoulFeederGoesForTheDyingOfTwoNearHostiles(t *testing.T) {
	f := newMonsterField(t, feedersMap, nil)
	feeders := f.bodiesOn(mapNamed(t, f.w, feedersMap), "template.creature.soul_feeder")
	if len(feeders) != 1 {
		t.Fatalf("%d soul-feeders, want 1", len(feeders))
	}
	feeder := feeders[0]
	members := f.members()
	at := ecs.NewMap[comp.Pos](f.w.ECS)
	home := *at.Get(feeder)
	sim.Put(f.w, members[0], comp.Pos{Map: home.Map, At: cellAt(9, 3)})
	sim.Put(f.w, members[1], comp.Pos{Map: home.Map, At: cellAt(11, 3)})
	f.damage(members[1], "torso", "3")
	dying := f.w.UIDOf(members[1])
	feederUID := f.w.UIDOf(feeder)
	for range 200 {
		if len(f.strikeTargets(feederUID)) > 0 {
			break
		}
		f.wait(api.WaitEvent)
	}
	targets := f.strikeTargets(feederUID)
	if len(targets) == 0 {
		t.Fatal("the soul-feeder never struck")
	}
	if targets[0] != dying {
		t.Errorf("the soul-feeder struck %d first, want the dying member %d", targets[0], dying)
	}
}

func TestAnArchonsWardenAttacksThePartyNearABelieverAndSparesTheBeliever(t *testing.T) {
	f := newMonsterField(t, shrineMap, nil)
	shrine := mapNamed(t, f.w, shrineMap)
	wardens := f.bodiesOn(shrine, wardenTmpl)
	believers := f.bodiesOn(shrine, "template.resident.believer")
	if len(wardens) != 1 || len(believers) != 1 {
		t.Fatalf("%d wardens and %d believers, want 1 of each", len(wardens), len(believers))
	}
	warden, believer := f.w.UIDOf(wardens[0]), f.w.UIDOf(believers[0])
	party := f.members()
	partyUIDs := make([]api.UID, len(party))
	for i, member := range party {
		partyUIDs[i] = f.w.UIDOf(member)
	}
	for range 300 {
		if len(f.strikeTargets(warden)) > 0 {
			break
		}
		f.wait(api.WaitEvent)
	}
	targets := f.strikeTargets(warden)
	if len(targets) == 0 {
		t.Fatal("the warden never struck")
	}
	for _, target := range targets {
		if !slices.Contains(partyUIDs, target) {
			t.Errorf("the warden struck %d, want only party members", target)
		}
	}
	for _, e := range f.events {
		if d, ok := e.(api.PartDamaged); ok && d.Body == believer {
			t.Errorf("the believer's part %d was damaged", d.Part)
		}
	}
}

func TestAWardenLeavesAPartyAloneWhenNoBelieverIsNear(t *testing.T) {
	f := newMonsterField(t, shrineMap, func(bundle *content.Bundle) {
		contenttest.Edit(bundle.Maps, shrineMap, func(shrine *content.MapDef) { delete(shrine.Spawns, "b") })
	})
	wardens := f.bodiesOn(mapNamed(t, f.w, shrineMap), wardenTmpl)
	if len(wardens) != 1 {
		t.Fatalf("%d wardens, want 1", len(wardens))
	}
	warden := f.w.UIDOf(wardens[0])
	for range 100 {
		f.wait(api.WaitEvent)
	}
	if got := f.strikeTargets(warden); len(got) != 0 {
		t.Errorf("the warden struck %d times with nobody to guard", len(got))
	}
}

func soulLevel(tb testing.TB, f *monsterField, body ecs.Entity) float64 {
	tb.Helper()
	held := ecs.NewMap[comp.Needs](f.w.ECS).Get(body)
	for _, level := range held.Levels {
		if level.Need == "need.soul" {
			return level.Level
		}
	}
	tb.Fatal("the body carries no soul need")
	return 0
}

func TestASoulFeederWoundingTheLivingEasesItsSoulNeed(t *testing.T) {
	f := newMonsterField(t, feedersMap, peaceful)
	feeder := f.bodiesOn(mapNamed(t, f.w, feedersMap), "template.creature.soul_feeder")[0]
	victim := f.members()[0]
	before := soulLevel(t, f, feeder)
	ctx := &effects.Context{World: f.w, Source: feeder, Target: victim}
	refs := []content.EffectRef{{Do: "damage", Params: map[string]string{"parts": "target:torso", "amount": "1"}}}
	if err := effects.Apply(ctx, refs); err != nil {
		t.Fatal(err)
	}
	if _, err := f.w.Bus.Drain(f.w); err != nil {
		t.Fatal(err)
	}
	if after := soulLevel(t, f, feeder); after >= before {
		t.Errorf("soul need after wounding = %v, want below %v", after, before)
	}
}

func TestARemnantLeavesByTheGateToTheLeastDecayedGround(t *testing.T) {
	f := newMonsterField(t, "map.play.den", func(bundle *content.Bundle) {
		peaceful(bundle)
		contenttest.Edit(bundle.Maps, feedersMap, func(feeders *content.MapDef) {
			feeders.Region = &content.RegionDef{Props: map[string]float64{"decay": 0}}
		})
		contenttest.Edit(bundle.Maps, "map.play.den", func(den *content.MapDef) {
			den.Marks["h"] = "gate_to_feeders"
			den.Gates["h"] = content.GateDef{To: feedersMap, Mark: "from_hub"}
			den.Layers[0].Rows[5] = "#.......h#"
		})
	})
	den := mapNamed(t, f.w, "map.play.den")
	feeders := mapNamed(t, f.w, feedersMap)
	for range 8 {
		f.wait(api.WaitPhase)
	}
	var to []api.MapID
	for _, e := range f.events {
		if m, ok := e.(api.Migrated); ok && m.Family == remnantFam && m.From == den {
			to = append(to, m.To)
		}
	}
	if len(to) != 1 || to[0] != feeders {
		t.Errorf("remnant left for maps %v, want only the least decayed, %d", to, feeders)
	}
}
