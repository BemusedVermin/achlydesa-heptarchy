package combat_test

import (
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/actions"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/clock"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/effects"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/inventory"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/knowledge"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

// allForms is what a fighter knows: a step and the seven forms.
var allForms = []api.ContentID{
	"action.step", "action.strike", "action.sweep", "action.sever", "action.pin", "action.stand",
	"action.give", "action.graft",
}

// ring is a game on an open map: the player's fighter at the origin, an ally beside it, and a
// victim (a human with 3-integrity parts) and a crawler (a clay core) standing in reach.
type ring struct {
	t       *testing.T
	g       *game.Game
	w       *sim.World
	fighter ecs.Entity
	ally    ecs.Entity
	victim  ecs.Entity
	crawler ecs.Entity
	relic   ecs.Entity // a body of old matter
}

func person(controller string, faction api.ContentID, known []api.ContentID) content.Template {
	return content.Template{
		Actor: &content.ActorDef{Controller: controller}, Actions: known, Body: ptr(api.ContentID("body.human")),
		Combatant: &content.CombatantDef{
			Faction: faction, Graftable: true,
			Types: []api.ContentID{"type.cut", "type.press", "type.relic", "type.name"},
		},
	}
}

// newRing builds the ring. Everyone stands on row 0; at gives each body's q.
func newRing(t *testing.T, at map[string]int) *ring {
	t.Helper()
	bundle, err := content.LoadBundle(sampleDir)
	if err != nil {
		t.Fatal(err)
	}
	bundle.Start = content.Start{}
	pinNumbers(&bundle)
	armPart := api.ContentID("part.human_arm")
	bundle.Items["item.test_arm"] = content.Item{Name: "text.item.test_arm", Kind: "piece", Part: &armPart}
	bundle.Text["text.item.test_arm"] = "spare arm"
	bundle.Bodies["body.test.old"] = content.BodyPlan{Parts: []content.PlanPart{
		{Part: "part.human_arm", Material: "material.old_matter"}, {Part: "part.human_torso"},
	}}
	bundle.Templates["template.test.fighter"] = person(timeline.PlayerController, "faction.party", allForms)
	bundle.Templates["template.test.ally"] = person("idle", "faction.party", allForms)
	bundle.Templates["template.test.victim"] = person("idle", "faction.deep", allForms)
	crawler := person("idle", "faction.deep", allForms)
	crawler.Body = ptr(api.ContentID("body.crawler"))
	crawler.Combatant.Graftable, crawler.Combatant.Crawls = false, true
	bundle.Templates["template.test.crawler"] = crawler
	relic := person("idle", "faction.deep", nil)
	relic.Body = ptr(api.ContentID("body.test.old"))
	bundle.Templates["template.test.relic"] = relic
	for _, name := range []string{"fighter", "ally", "victim", "crawler", "relic"} {
		bundle.Text["text.template.test."+name] = strings.ToUpper(name[:1]) + name[1:]
	}
	reg, err := content.New(bundle)
	if err != nil {
		t.Fatal(err)
	}
	r := &ring{t: t}
	place := func(name string) *comp.Pos {
		return &comp.Pos{Map: 1, At: api.Cell{Hex: api.Hex{Q: at[name]}}}
	}
	r.g, err = game.New(game.Config{
		Seed: 1, Content: reg,
		Install: func(w *sim.World, _ *clock.Clock) error {
			r.w = w
			m := hexmap.NewMap(1, "room", hexmap.FuncSource(func(api.Cell) hexmap.Terrain {
				return hexmap.Terrain{ID: "terrain.air"}
			}), hexmap.Prisms{})
			effects.SetMaps(w, func(id api.MapID) *hexmap.Map {
				if id == 1 {
					return m
				}
				return nil
			})
			return nil
		},
		Spawn: func(w *sim.World, _ *clock.Clock) error {
			for _, spawn := range []struct {
				into *ecs.Entity
				name string
			}{
				{&r.fighter, "fighter"}, {&r.ally, "ally"}, {&r.victim, "victim"}, {&r.crawler, "crawler"},
				{&r.relic, "relic"},
			} {
				if _, ok := at[spawn.name]; !ok {
					continue
				}
				var err error
				if *spawn.into, err = w.Spawn(
					api.ContentID("template.test."+spawn.name),
					place(spawn.name),
				); err != nil {
					return err
				}
			}
			return nil
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return r
}

func (r *ring) uid(e ecs.Entity) string { return strconv.FormatUint(uint64(r.w.UIDOf(e)), 10) }

func (r *ring) view() api.View {
	r.t.Helper()
	view, err := r.g.View()
	if err != nil {
		r.t.Fatal(err)
	}
	return view
}

// option finds the option whose ID starts with the segments joined by slashes.
func (r *ring) option(segments ...string) (api.Option, bool) {
	prefix := strings.Join(segments, "/")
	for _, option := range r.view().Options {
		if option.ID == prefix || strings.HasPrefix(option.ID, prefix+"/") {
			return option, true
		}
	}
	return api.Option{}, false
}

func (r *ring) do(segments ...string) []api.Event {
	r.t.Helper()
	option, ok := r.option(segments...)
	if !ok || !option.Enabled {
		r.t.Fatalf("no enabled option %v among %v", segments, r.optionIDs())
	}
	return r.apply(option.Command)
}

func (r *ring) optionIDs() []string {
	var ids []string
	for _, option := range r.view().Options {
		ids = append(ids, option.ID)
	}
	return ids
}

func (r *ring) apply(cmd api.Command) []api.Event {
	r.t.Helper()
	events, err := r.g.Apply(cmd)
	if err != nil {
		r.t.Fatal(err)
	}
	return events
}

func (r *ring) wait(ticks int64) []api.Event {
	r.t.Helper()
	return r.apply(api.Wait{Actor: r.w.UIDOf(r.fighter), Until: api.WaitTicks, Ticks: ticks})
}

// part is the body's first attached part of the slot.
func (r *ring) part(body ecs.Entity, slot string) ecs.Entity {
	r.t.Helper()
	parts := ecs.NewMap[comp.Part](r.w.ECS)
	for _, part := range anatomy.Parts(r.w, body) {
		if parts.Get(part).Slot == slot {
			return part
		}
	}
	r.t.Fatalf("body has no %s", slot)
	return ecs.Entity{}
}

func (r *ring) attached(body ecs.Entity, slot string) int {
	n := 0
	parts := ecs.NewMap[comp.Part](r.w.ECS)
	for _, part := range anatomy.Parts(r.w, body) {
		if parts.Get(part).Slot == slot {
			n++
		}
	}
	return n
}

func (r *ring) busyChannels(body ecs.Entity) int {
	n := 0
	for _, c := range timeline.Channels(r.w, body) {
		if !timeline.Free(r.w, c) {
			n++
		}
	}
	return n
}

func (r *ring) give(item api.ContentID, count int, boundTo ecs.Entity) {
	stack := comp.ItemStack{Item: item, Count: count}
	if !boundTo.IsZero() {
		stack.BoundTo = r.w.UIDOf(boundTo)
	}
	inventory.Add(r.w, r.fighter, stack)
}

func has[E api.Event](events []api.Event, match func(E) bool) bool {
	return slices.ContainsFunc(events, func(e api.Event) bool {
		typed, ok := e.(E)
		return ok && match(typed)
	})
}

func count[E api.Event](events []api.Event) int {
	n := 0
	for _, e := range events {
		if _, ok := e.(E); ok {
			n++
		}
	}
	return n
}

func always[E api.Event](E) bool { return true }

func TestStrikeTakesOffAnArmInOneHitAndATorsoInTwo(t *testing.T) {
	r := newRing(t, map[string]int{"fighter": 0, "victim": 1})
	arm := r.part(r.victim, "arm")
	events := r.do("strike", "type.press", r.uid(r.victim), r.uid(arm))
	if !has(events, func(e api.ActionStarted) bool { return e.Lands == r.w.Now+3 }) {
		t.Fatalf("a Strike did not start with a wind-up of 3: %v", events)
	}
	if got := r.busyChannels(r.fighter); got != 1 {
		t.Errorf("a Strike occupies %d channels, want one arm", got)
	}
	events = r.wait(3)
	if !has(events, always[api.PartSevered]) || r.attached(r.victim, "arm") != 1 {
		t.Errorf("one Strike left the arm on: %v", events)
	}
	torso := r.part(r.victim, "torso")
	r.do("strike", "type.press", r.uid(r.victim), r.uid(torso))
	r.wait(4)
	if got := ecs.NewMap[comp.Part](r.w.ECS).Get(torso).Integrity; got != 2 {
		t.Fatalf("one Strike left the torso at %d, want 2 of 5", got)
	}
	r.do("strike", "type.press", r.uid(r.victim), r.uid(torso))
	r.wait(7)
	if r.attached(r.victim, "torso") != 0 {
		t.Error("two Strikes left the torso on")
	}
}

func TestSweepTakesTwoArmsAndWearsEveryPartAndRevealsMaterials(t *testing.T) {
	r := newRing(t, map[string]int{"fighter": 0, "victim": 1})
	if _, ok := r.option("sweep", "type.cut", r.uid(r.victim)); !ok {
		t.Fatalf("no Sweep among %v", r.optionIDs())
	}
	events := r.do("sweep", "type.cut", r.uid(r.victim))
	if got := r.busyChannels(r.fighter); got != 2 {
		t.Errorf("a Sweep occupies %d channels, want two arms", got)
	}
	events = append(events, r.wait(4)...)
	parts := anatomy.Parts(r.w, r.victim)
	if got := count[api.PartDamaged](events); got != len(parts) {
		t.Errorf("a Sweep damaged %d parts, want all %d", got, len(parts))
	}
	for _, part := range parts {
		fact, ok := knowledge.Best(r.w, r.fighter, r.w.UIDOf(part), "claim.material")
		if !ok || fact.Value != string(ecs.NewMap[comp.Part](r.w.ECS).Get(part).Material) {
			t.Errorf("fact about a part = %+v, %v; want its material", fact, ok)
		}
	}
}

func TestSweepIsNotOfferedWithOnlyOneArmFree(t *testing.T) {
	r := newRing(t, map[string]int{"fighter": 0, "victim": 1})
	arm := r.part(r.fighter, "arm")
	if err := timeline.Occupy(r.w, timeline.Channel{Entity: arm, Body: r.fighter}, 20); err != nil {
		t.Fatal(err)
	}
	for _, id := range r.optionIDs() {
		if strings.HasPrefix(id, "sweep") || strings.HasPrefix(id, "graft") {
			t.Errorf("%q is offered with one arm busy", id)
		}
	}
	if _, ok := r.option("strike", "type.cut", r.uid(r.victim)); !ok {
		t.Errorf("a Strike with the free arm is not offered: %v", r.optionIDs())
	}
}

func TestSoakStrikeSeversClayAndSpendsLakeWater(t *testing.T) {
	r := newRing(t, map[string]int{"fighter": 0, "crawler": 1})
	core := r.part(r.crawler, "torso")
	if option, ok := r.option("strike", "type.soak", r.uid(r.crawler), r.uid(core)); !ok || option.Enabled {
		t.Fatalf("Soak with no lake water = %+v, %v; want it listed but disabled", option, ok)
	}
	intent := actions.Intent{
		Actor: r.fighter, Action: "action.strike", Target: r.crawler, TargetPart: core,
		Choices: map[string]string{"type": "type.soak"},
	}
	if ok, reason := actions.Validate(r.w, intent); ok || r.w.Content.Render(reason) == "" {
		t.Errorf("Validate with no lake water = %v, %+v; want a refusal with a reason", ok, reason)
	}
	r.give("item.lake_water", 1, ecs.Entity{})
	events := r.do("strike", "type.soak", r.uid(r.crawler), r.uid(core))
	if !has(events, func(e api.ActionStarted) bool { return e.Lands == r.w.Now+3 }) {
		t.Errorf("the Strike did not start with a wind-up of 3: %v", events)
	}
	if got := inventory.Count(r.w, r.fighter, "item.lake_water", 0); got != 0 {
		t.Errorf("lake water left = %d, want it spent when the swing starts", got)
	}
	// Soak adds 2 to the recovery: the arm is busy for 3 + 4 + 2 ticks.
	if got := r.busyChannels(r.fighter); got != 1 {
		t.Errorf("the Strike occupies %d channels, want one arm", got)
	}
	events = r.wait(3)
	if !has(events, func(e api.PartSevered) bool { return e.Part == r.w.UIDOf(core) }) {
		t.Errorf("a Soak Strike did not sever the clay core: %v", events)
	}
	r.wait(5)
	if got := r.busyChannels(r.fighter); got != 1 {
		t.Errorf("channels busy after 8 ticks = %d, want the Soaked arm still recovering (9 in all)", got)
	}
	r.wait(1)
	if got := r.busyChannels(r.fighter); got != 0 {
		t.Errorf("channels busy after 9 ticks = %d, want none", got)
	}
}

func TestSteppingOutOfReachMakesTheStrikeMissButTheWaterIsSpent(t *testing.T) {
	r := newRing(t, map[string]int{"fighter": 0, "crawler": 1})
	r.give("item.lake_water", 1, ecs.Entity{})
	core := r.part(r.crawler, "torso")
	r.do("strike", "type.soak", r.uid(r.crawler), r.uid(core))
	ecs.NewMap[comp.Pos](r.w.ECS).Get(r.crawler).At.Hex = api.Hex{Q: 5}
	events := r.wait(3)
	if !has(events, always[api.ActionMissed]) || has(events, always[api.PartDamaged]) {
		t.Errorf("a Strike on a body that stepped away did not simply miss: %v", events)
	}
	if got := inventory.Count(r.w, r.fighter, "item.lake_water", 0); got != 0 {
		t.Errorf("lake water left = %d, want it spent though the swing missed", got)
	}
}

func TestTypeChecksRefuseWithReasons(t *testing.T) {
	r := newRing(t, map[string]int{"fighter": 0, "ally": -1, "victim": 1, "relic": 2})
	r.give("item.name", 1, r.ally)
	victimArm := r.part(r.victim, "arm")
	strike := func(target, part ecs.Entity, kind string) actions.Intent {
		return actions.Intent{
			Actor: r.fighter, Action: "action.strike", Target: target, TargetPart: part,
			Choices: map[string]string{"type": kind},
		}
	}
	tests := []struct {
		name       string
		in         actions.Intent
		wantReason api.ContentID
	}{
		{"Name on someone else", strike(r.victim, victimArm, "type.name"), "text.combat.type_barred"},
		{"Name on its owner", strike(r.ally, r.part(r.ally, "arm"), "type.name"), ""},
		{"Cut on flesh", strike(r.victim, victimArm, "type.cut"), ""},
		{"Relic on old matter", strike(r.relic, r.part(r.relic, "arm"), "type.relic"), "text.combat.immune"},
		{"Relic on flesh", strike(r.victim, victimArm, "type.relic"), ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// The old-matter body stands two cells away, so bring it into reach for this check.
			ecs.NewMap[comp.Pos](r.w.ECS).Get(r.relic).At.Hex = api.Hex{Q: 1, R: 0}
			ok, reason := actions.Validate(r.w, tt.in)
			if tt.wantReason == "" {
				if !ok {
					t.Errorf("refused: %s", r.w.Content.Render(reason))
				}
				return
			}
			if ok || reason.ID != tt.wantReason || r.w.Content.Render(reason) == "" {
				t.Errorf("Validate = %v, %+v; want a refusal with %s", ok, reason, tt.wantReason)
			}
		})
	}
}

func TestGraftNeedsAGraftableTargetAndAPieceFromThePack(t *testing.T) {
	r := newRing(t, map[string]int{"fighter": 0, "ally": -1, "crawler": 1})
	if _, ok := r.option("graft"); ok {
		t.Fatal("Graft is offered with no piece in the pack")
	}
	r.give("item.test_arm", 1, ecs.Entity{})
	if _, ok := r.option("graft", "item.test_arm", r.uid(r.crawler)); !ok {
		t.Fatalf("no Graft option for the crawler: %v", r.optionIDs())
	}
	ok, reason := actions.Validate(r.w, actions.Intent{
		Actor:   r.fighter,
		Action:  "action.graft",
		Target:  r.crawler,
		Choices: map[string]string{"piece": "item.test_arm"},
	})
	if ok || reason.ID != "text.combat.not_graftable" {
		t.Errorf("Graft onto a crawler = %v, %+v; want it refused as not graftable", ok, reason)
	}
	before := len(timeline.Channels(r.w, r.ally))
	events := r.do("graft", "item.test_arm", r.uid(r.ally))
	if got := r.busyChannels(r.fighter); got != 2 {
		t.Errorf("a Graft occupies %d channels, want both arms", got)
	}
	events = append(events, r.wait(6)...)
	if !has(events, func(e api.PartAttached) bool { return e.From == api.AttachedFromPack }) ||
		len(timeline.Channels(r.w, r.ally)) != before+1 {
		t.Errorf("the graft did not fit: %v", events)
	}
	if got := inventory.Count(r.w, r.fighter, "item.test_arm", 0); got != 0 {
		t.Errorf("pieces left = %d, want it fitted", got)
	}
}

func TestSeverCutsOnAWeaknessAndOnlyWearsOtherwise(t *testing.T) {
	r := newRing(t, map[string]int{"fighter": 0, "victim": 1})
	arm := r.part(r.victim, "arm")
	r.do("sever", "type.press", r.uid(r.victim), r.uid(arm)) // flesh is not weak to Press
	r.wait(2)
	if got := ecs.NewMap[comp.Part](r.w.ECS).Get(arm).Integrity; got != 2 || r.attached(r.victim, "arm") != 2 {
		t.Fatalf(
			"a Sever off its weakness left the arm at %d, attached %d; want 2 and still on",
			got,
			r.attached(r.victim, "arm"),
		)
	}
	r.wait(5)
	r.do("sever", "type.cut", r.uid(r.victim), r.uid(arm)) // flesh is weak to Cut
	events := r.wait(2)
	if !has(events, func(e api.PartSevered) bool { return e.Cause == api.SeveredCut }) {
		t.Errorf("a Sever on a weakness did not cut: %v", events)
	}
}

func TestPinLocksThePart(t *testing.T) {
	r := newRing(t, map[string]int{"fighter": 0, "victim": 1})
	arm := r.part(r.victim, "arm")
	r.do("pin", "type.cut", r.uid(r.victim), r.uid(arm))
	if got := r.busyChannels(r.fighter); got != 1 {
		t.Errorf("a Pin occupies %d channels, want one arm", got)
	}
	r.wait(1)
	if !timeline.Locked(r.w, timeline.Channel{Entity: arm, Body: r.victim}) {
		t.Error("the pinned arm is not locked")
	}
}

func TestStandBracesFreezesCrawlersAndReadiesTheBody(t *testing.T) {
	r := newRing(t, map[string]int{"fighter": 0, "victim": 1, "crawler": -1})
	arm := timeline.Channel{Entity: r.part(r.fighter, "arm"), Body: r.fighter}
	if err := timeline.Occupy(r.w, arm, 30); err != nil {
		t.Fatal(err)
	}
	r.do("stand")
	if !timeline.Free(r.w, arm) {
		t.Error("Stand did not ready the arm")
	}
	if got := r.busyChannels(r.crawler); got == 0 {
		t.Error("Stand did not delay the crawler in sight")
	}
	if got := r.busyChannels(r.victim); got != 0 {
		t.Errorf("Stand delayed %d channels of a body that does not crawl", got)
	}
	held := ecs.NewMap[comp.Statuses](r.w.ECS)
	if !held.Has(r.fighter) || held.Get(r.fighter).Active[0].Status != "status.standing" {
		t.Error("Stand did not brace the body")
	}
	if _, ok := r.option("stand"); ok {
		t.Error("Stand is offered again while still braced")
	}
}

func TestGiveReadiesOneChannelOfAnAlly(t *testing.T) {
	r := newRing(t, map[string]int{"fighter": 0, "ally": -1})
	for _, c := range timeline.Channels(r.w, r.ally) {
		if err := timeline.Occupy(r.w, c, 30); err != nil {
			t.Fatal(err)
		}
	}
	busy := r.busyChannels(r.ally)
	r.do("give", r.uid(r.ally))
	if got := r.busyChannels(r.ally); got != busy-1 {
		t.Errorf("busy channels of the ally = %d, want %d: Give readies one", got, busy-1)
	}
}

func TestPressCancelsTheSwingOfABodyWindingUp(t *testing.T) {
	for _, tt := range []struct {
		kind         string
		wantCanceled bool
	}{{"type.cut", false}, {"type.press", true}} {
		t.Run(tt.kind, func(t *testing.T) {
			r := newRing(t, map[string]int{"fighter": 0, "victim": 1})
			err := actions.Start(r.w, actions.Intent{
				Actor:   r.victim,
				Action:  "action.sweep",
				Target:  r.fighter,
				Choices: map[string]string{"type": "type.cut"},
			})
			if err != nil {
				t.Fatal(err)
			}
			head := r.part(r.victim, "head")
			events := r.do("strike", tt.kind, r.uid(r.victim), r.uid(head))
			events = append(events, r.wait(4)...)
			pending := timeline.PendingOf(r.w, r.w.UIDOf(r.victim))
			canceled := has(events, func(e api.ActionCanceled) bool { return true })
			if canceled != tt.wantCanceled || (tt.wantCanceled && len(pending) != 0) {
				t.Errorf("canceled = %v, pending = %d; want canceled %v", canceled, len(pending), tt.wantCanceled)
			}
			hurt := has(events, func(e api.PartDamaged) bool { return e.Body == r.w.UIDOf(r.fighter) })
			if hurt == tt.wantCanceled {
				t.Errorf("the victim's Sweep hurt the fighter = %v, want %v", hurt, !tt.wantCanceled)
			}
		})
	}
}

func TestInASoopOnlyTheFrontMemberReaches(t *testing.T) {
	// Single file: the player's fighter at the back, the ally in front, the victim beyond.
	r := newRing(t, map[string]int{"fighter": -1, "ally": 0, "victim": 1})
	for _, id := range r.optionIDs() {
		for _, form := range []string{"strike", "sweep", "sever", "pin"} {
			if strings.HasPrefix(id, form+"/") && strings.Contains(id, "/"+r.uid(r.victim)) {
				t.Errorf("the member behind has %q against the creature beyond", id)
			}
		}
	}
	reach := actions.Startable(r.w, r.ally)
	if !slices.ContainsFunc(reach, func(in actions.Intent) bool { return in.Target == r.victim }) {
		t.Error("the front member has no attack against the creature")
	}
}

func TestTypesTheNeededItemIsMissingForAreListedDisabledWithTheReason(t *testing.T) {
	r := newRing(t, map[string]int{"fighter": 0, "crawler": 1})
	core := r.part(r.crawler, "torso")
	soak, ok := r.option("strike", "type.soak", r.uid(r.crawler), r.uid(core))
	if !ok || soak.Enabled {
		t.Fatalf("Soak with no lake water = %+v, %v; want it listed but disabled", soak, ok)
	}
	name := r.w.Content.Render(api.Text{ID: "text.item.lake_water"})
	want := r.w.Content.Render(api.Text{ID: "text.combat.pack_short", Args: map[string]string{"item": name}})
	if want == "" || soak.Reason != want {
		t.Errorf("reason = %q, want %q", soak.Reason, want)
	}
	r.give("item.lake_water", 1, ecs.Entity{})
	if soak, _ = r.option("strike", "type.soak", r.uid(r.crawler), r.uid(core)); !soak.Enabled {
		t.Errorf("Soak with lake water in the pack is still disabled: %+v", soak)
	}
}

func TestSweepHurtsEachPartByItsOwnWeakness(t *testing.T) {
	r := newRing(t, map[string]int{"fighter": 0, "crawler": 1})
	r.give("item.lake_water", 1, ecs.Entity{})
	parts := ecs.NewMap[comp.Part](r.w.ECS)
	events := r.do("sweep", "type.soak", r.uid(r.crawler))
	events = append(events, r.wait(4)...)
	lost := map[api.UID]int{}
	for _, e := range events {
		if damaged, ok := e.(api.PartDamaged); ok {
			lost[damaged.Part] = damaged.Amount
		}
	}
	var weakParts, otherParts int
	for _, part := range anatomy.Parts(r.w, r.crawler) {
		material, _ := r.w.Content.Material(parts.Get(part).Material)
		got := lost[r.w.UIDOf(part)]
		if slices.Contains(material.WeakTo, "type.soak") {
			weakParts++
			if got != 2 {
				t.Errorf("a part weak to Soak lost %d, want 2", got)
			}
		} else {
			otherParts++
			if got != 1 {
				t.Errorf("a part not weak to Soak lost %d, want 1", got)
			}
		}
	}
	if weakParts == 0 || otherParts == 0 {
		t.Fatalf("the crawler has %d weak and %d other parts; the test needs both", weakParts, otherParts)
	}
}
