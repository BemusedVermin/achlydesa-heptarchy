package combat_test

import (
	"fmt"
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

// ring is a game on an open map built from the real content: the player's fighter, and any of an
// ally, a victim (a test dummy), a clay-cored crawler, a body of old matter, a clay-armed dummy and
// a body that cannot be grafted onto. Everyone stands on row 0; at gives each body's q.
type ring struct {
	t      *testing.T
	g      *game.Game
	w      *sim.World
	bodies map[string]ecs.Entity
}

// body is the named body. It is the zero entity when the ring has none by that name.
func (r *ring) body(name string) ecs.Entity { return r.bodies[name] }

// person is the sample's Reaper without its party, so a test can place it alone. Relic is added to
// its types because the sample leaves it out (it is innate to grafts, which offers can't see yet).
func person(t *testing.T, b *content.Bundle, controller string) content.Template {
	t.Helper()
	person := b.Templates["template.person.reaper"]
	person.Actor = &content.ActorDef{Controller: controller}
	person.Triggers = nil
	person.Lists = map[string][]string{"types": {"type.cut", "type.name", "type.press", "type.relic"}}
	return person
}

func newRing(t *testing.T, at map[string]int) *ring {
	t.Helper()
	bundle, err := content.LoadBundle(sampleDir)
	if err != nil {
		t.Fatal(err)
	}
	bundle.Start = content.Start{}
	armPart := api.ContentID("part.human_arm")
	bundle.Items["item.test_arm"] = content.Item{
		Name: "text.item.test_arm", Part: &armPart, Data: content.Data{Tags: []string{"piece"}},
	}
	bundle.Text["text.item.test_arm"] = "spare arm"
	bundle.Bodies["body.test.old"] = content.BodyPlan{Parts: []content.PlanPart{
		{Part: "part.human_arm", Material: "material.old_matter"}, {Part: "part.human_torso"},
	}}
	bundle.Bodies["body.test.clay_arm"] = content.BodyPlan{Parts: []content.PlanPart{
		{Part: "part.human_head"}, {Part: "part.human_arm", Material: "material.clay"},
		{Part: "part.human_legs"}, {Part: "part.human_torso"},
	}}
	templates := map[string]content.Template{
		"fighter": person(t, &bundle, timeline.PlayerController),
		"ally":    person(t, &bundle, "idle"),
		"victim":  bundle.Templates["template.test.dummy"],
	}
	crawler := bundle.Templates["template.test.dummy"]
	crawler.Body = ptr(api.ContentID("body.crawler"))
	templates["crawler"] = crawler
	old := bundle.Templates["template.test.dummy"]
	old.Body = ptr(api.ContentID("body.test.old"))
	templates["relic"] = old
	clay := bundle.Templates["template.test.dummy"]
	clay.Body = ptr(api.ContentID("body.test.clay_arm"))
	templates["clay"] = clay
	stone := bundle.Templates["template.test.dummy"]
	stone.Tags = nil
	templates["stone"] = stone
	for name, template := range templates {
		bundle.Templates["template.test."+name] = template
		bundle.Text["text.template.test."+name] = strings.ToUpper(name[:1]) + name[1:]
	}
	reg, err := content.New(bundle)
	if err != nil {
		t.Fatal(err)
	}
	r := &ring{t: t, bodies: map[string]ecs.Entity{}}
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
			// Spawn in a fixed order, so UIDs and the player's turn are the same every run.
			for _, name := range []string{"fighter", "ally", "victim", "crawler", "relic", "clay", "stone"} {
				q, ok := at[name]
				if !ok {
					continue
				}
				pos := &comp.Pos{Map: 1, At: api.Cell{Hex: api.Hex{Q: q}}}
				body, err := w.Spawn(api.ContentID("template.test."+name), pos)
				if err != nil {
					return err
				}
				r.bodies[name] = body
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

// do takes the enabled option whose ID starts with the segments.
func (r *ring) do(segments ...string) []api.Event {
	r.t.Helper()
	option, ok := r.option(segments...)
	if !ok || !option.Enabled {
		r.t.Fatalf("no enabled option %v among %v", segments, r.optionIDs())
	}
	return r.apply(option.Command)
}

func (r *ring) wait(ticks int64) []api.Event {
	r.t.Helper()
	return r.apply(api.Wait{Actor: r.w.UIDOf(r.body("fighter")), Until: api.WaitTicks, Ticks: ticks})
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

func (r *ring) integrityOf(part ecs.Entity) int {
	return ecs.NewMap[comp.Part](r.w.ECS).Get(part).Integrity
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

// give puts items in the fighter's pack, bound to someone or to no one.
func (r *ring) give(item api.ContentID, count int, boundTo ecs.Entity) {
	stack := comp.ItemStack{Item: item, Count: count}
	if !boundTo.IsZero() {
		stack.BoundTo = r.w.UIDOf(boundTo)
	}
	inventory.Add(r.w, r.body("fighter"), stack)
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
	victim := r.body("victim")
	arm := r.part(victim, "arm")
	events := r.do("strike", "type.press", r.uid(victim), r.uid(arm))
	if !has(events, func(e api.ActionStarted) bool { return e.Lands == r.w.Now+3 }) {
		t.Fatalf("a Strike did not start with a wind-up of 3: %v", events)
	}
	if got := r.busyChannels(r.body("fighter")); got != 1 {
		t.Errorf("a Strike occupies %d channels, want one arm", got)
	}
	events = r.wait(3)
	if !has(events, always[api.PartSevered]) || r.attached(victim, "arm") != 1 {
		t.Errorf("one Strike left the arm on: %v", events)
	}
	torso := r.part(victim, "torso")
	r.do("strike", "type.press", r.uid(victim), r.uid(torso))
	r.wait(7)
	if got := r.integrityOf(torso); got != 2 {
		t.Fatalf("one Strike left the torso at %d, want 2 of 5", got)
	}
	r.do("strike", "type.press", r.uid(victim), r.uid(torso))
	r.wait(7)
	if r.attached(victim, "torso") != 0 {
		t.Error("two Strikes left the torso on")
	}
}

func TestSweepTakesTwoArmsAndWearsEveryPartAndRevealsMaterials(t *testing.T) {
	r := newRing(t, map[string]int{"fighter": 0, "victim": 1})
	victim := r.body("victim")
	events := r.do("sweep", "type.cut", r.uid(victim))
	if got := r.busyChannels(r.body("fighter")); got != 2 {
		t.Errorf("a Sweep occupies %d channels, want two arms", got)
	}
	events = append(events, r.wait(4)...)
	parts := anatomy.Parts(r.w, victim)
	if got := count[api.PartDamaged](events); got != len(parts) {
		t.Errorf("a Sweep damaged %d parts, want all %d", got, len(parts))
	}
	for _, part := range parts {
		fact, ok := knowledge.Best(r.w, r.body("fighter"), r.w.UIDOf(part), "claim.material")
		if !ok || fact.Value != string(ecs.NewMap[comp.Part](r.w.ECS).Get(part).Material) {
			t.Errorf("fact about a part = %+v, %v; want its material", fact, ok)
		}
	}
}

func TestSweepIsNotOfferedWithOnlyOneArmFree(t *testing.T) {
	r := newRing(t, map[string]int{"fighter": 0, "victim": 1})
	arm := r.part(r.body("fighter"), "arm")
	if err := timeline.Occupy(r.w, timeline.Channel{Entity: arm, Body: r.body("fighter")}, 20); err != nil {
		t.Fatal(err)
	}
	for _, id := range r.optionIDs() {
		if strings.HasPrefix(id, "sweep") || strings.HasPrefix(id, "graft") {
			t.Errorf("%q is offered with one arm busy", id)
		}
	}
	if _, ok := r.option("strike", "type.cut", r.uid(r.body("victim"))); !ok {
		t.Errorf("a Strike with the free arm is not offered: %v", r.optionIDs())
	}
}

func TestSoakStrikeSeversClayAndSpendsLakeWater(t *testing.T) {
	r := newRing(t, map[string]int{"fighter": 0, "crawler": 1})
	fighter, crawler := r.body("fighter"), r.body("crawler")
	core := r.part(crawler, "torso")
	if _, ok := r.option("strike", "type.soak", r.uid(crawler), r.uid(core)); ok {
		t.Fatal("Soak is offered with no lake water")
	}
	r.give("item.lake_water", 1, ecs.Entity{})
	events := r.do("strike", "type.soak", r.uid(crawler), r.uid(core))
	if got := inventory.Count(r.w, fighter, "item.lake_water", 0); got != 0 {
		t.Errorf("lake water left = %d, want it spent when the swing starts", got)
	}
	events = append(events, r.wait(3)...)
	if !has(events, func(e api.PartSevered) bool { return e.Part == r.w.UIDOf(core) }) {
		t.Errorf("a Soak Strike did not sever the clay core: %v", events)
	}
	// Recovery is 4 plus Soak's 2: the arm is busy for 3 + 6 ticks in all.
	r.wait(5)
	if got := r.busyChannels(fighter); got != 1 {
		t.Errorf("channels busy after 8 ticks = %d, want the Soaked arm still recovering (9 in all)", got)
	}
	r.wait(1)
	if got := r.busyChannels(fighter); got != 0 {
		t.Errorf("channels busy after 9 ticks = %d, want none", got)
	}
}

func TestSoakAddsTwoToTheRecoveryOfEveryForm(t *testing.T) {
	r := newRing(t, map[string]int{"fighter": 0, "crawler": 1})
	r.give("item.lake_water", 1, ecs.Entity{})
	crawler := r.body("crawler")
	core := r.part(crawler, "torso")
	for _, form := range []struct {
		key      string
		path     []string
		windup   int
		recovery int
	}{
		{"strike", []string{r.uid(crawler), r.uid(core)}, 3, 4},
		{"sweep", []string{r.uid(crawler)}, 4, 5},
		{"sever", []string{r.uid(crawler), r.uid(core)}, 2, 5},
	} {
		plain, ok := r.option(append([]string{form.key, "type.cut"}, form.path...)...)
		if !ok {
			t.Fatalf("no %s with Cut among %v", form.key, r.optionIDs())
		}
		soak, ok := r.option(append([]string{form.key, "type.soak"}, form.path...)...)
		if !ok {
			t.Fatalf("no %s with Soak among %v", form.key, r.optionIDs())
		}
		wantPlain := fmt.Sprintf("lands in %d, busy %d", form.windup, form.windup+form.recovery)
		wantSoak := fmt.Sprintf("lands in %d, busy %d", form.windup, form.windup+form.recovery+2)
		if !strings.Contains(plain.Label, wantPlain) || !strings.Contains(soak.Label, wantSoak) {
			t.Errorf("%s labels = %q and %q, want %q and %q", form.key, plain.Label, soak.Label, wantPlain, wantSoak)
		}
	}
}

func TestSteppingOutOfReachMakesTheStrikeMissButTheWaterIsSpent(t *testing.T) {
	r := newRing(t, map[string]int{"fighter": 0, "crawler": 1})
	r.give("item.lake_water", 1, ecs.Entity{})
	crawler := r.body("crawler")
	r.do("strike", "type.soak", r.uid(crawler), r.uid(r.part(crawler, "torso")))
	ecs.NewMap[comp.Pos](r.w.ECS).Get(crawler).At.Hex = api.Hex{Q: 5}
	events := r.wait(3)
	if !has(events, always[api.ActionMissed]) || has(events, always[api.PartDamaged]) {
		t.Errorf("a Strike on a body that stepped away did not simply miss: %v", events)
	}
	if got := inventory.Count(r.w, r.body("fighter"), "item.lake_water", 0); got != 0 {
		t.Errorf("lake water left = %d, want it spent though the swing missed", got)
	}
}

func TestSweepHurtsEachPartByItsOwnWeaknessAndTeachesIt(t *testing.T) {
	r := newRing(t, map[string]int{"fighter": 0, "clay": 1})
	r.give("item.lake_water", 1, ecs.Entity{})
	clay := r.body("clay")
	events := r.do("sweep", "type.soak", r.uid(clay))
	events = append(events, r.wait(4)...)
	lost := map[api.UID]int{}
	for _, e := range events {
		if damaged, ok := e.(api.PartDamaged); ok {
			lost[damaged.Part] = damaged.Amount
		}
	}
	armUID := r.w.UIDOf(r.part(clay, "arm"))
	for _, part := range anatomy.Parts(r.w, clay) {
		want := 1
		if r.w.UIDOf(part) == armUID {
			want = 2
		}
		if got := lost[r.w.UIDOf(part)]; got != want {
			t.Errorf("part %d lost %d, want %d", r.w.UIDOf(part), got, want)
		}
	}
	fact, ok := knowledge.Best(r.w, r.body("fighter"), armUID, "claim.material")
	if !ok || fact.Value != "material.clay" {
		t.Errorf("fact about the clay arm = %+v, %v; want material.clay", fact, ok)
	}
}

func TestRelicOnOldMatterLandsAndDoesNothing(t *testing.T) {
	r := newRing(t, map[string]int{"fighter": 0, "relic": 1})
	relic := r.body("relic")
	arm := r.part(relic, "arm")
	events := r.do("strike", "type.relic", r.uid(relic), r.uid(arm))
	events = append(events, r.wait(3)...)
	if !has(events, always[api.ActionLanded]) {
		t.Errorf("the swing did not land: %v", events)
	}
	if has(events, always[api.PartDamaged]) || r.integrityOf(arm) != 3 {
		t.Errorf("Relic hurt old matter: events %v, integrity %d", events, r.integrityOf(arm))
	}
}

func TestNameIsDisabledUnlessANameIsBoundToTheTarget(t *testing.T) {
	r := newRing(t, map[string]int{"fighter": 0, "victim": 1, "ally": -1})
	victim := r.body("victim")
	arm := r.part(victim, "arm")
	r.give("item.name", 1, r.body("ally"))
	option, ok := r.option("strike", "type.name", r.uid(victim), r.uid(arm))
	want := r.w.Content.Render(api.Text{ID: "text.type.name_refusal"})
	if !ok || option.Enabled || want == "" || option.Reason != want {
		t.Fatalf("Name on a body whose name is not held = %+v, %v; want disabled with %q", option, ok, want)
	}
	r.give("item.name", 1, victim)
	if option, _ = r.option("strike", "type.name", r.uid(victim), r.uid(arm)); !option.Enabled {
		t.Errorf("Name on the body whose name is held is disabled: %+v", option)
	}
}

func TestPressCancelsTheSwingOfABodyWindingUp(t *testing.T) {
	for _, tt := range []struct {
		kind         string
		wantCanceled bool
	}{{"type.cut", false}, {"type.press", true}} {
		t.Run(tt.kind, func(t *testing.T) {
			r := newRing(t, map[string]int{"fighter": 0, "ally": 1})
			fighter, ally := r.body("fighter"), r.body("ally")
			err := actions.Start(r.w, actions.Intent{
				Actor: ally, Action: "action.sweep", Target: fighter,
				Choices: map[string]string{"type": "type.cut"},
			})
			if err != nil {
				t.Fatal(err)
			}
			head := r.part(ally, "head")
			events := r.do("strike", tt.kind, r.uid(ally), r.uid(head))
			events = append(events, r.wait(4)...)
			canceled := has(events, always[api.ActionCanceled])
			hurt := has(events, func(e api.PartDamaged) bool { return e.Body == r.w.UIDOf(fighter) })
			if canceled != tt.wantCanceled || hurt == tt.wantCanceled {
				t.Errorf("canceled = %v, hurt = %v; want canceled %v", canceled, hurt, tt.wantCanceled)
			}
		})
	}
}

func TestStandBracesDelaysTheOthersAndLeavesTheStanderFree(t *testing.T) {
	r := newRing(t, map[string]int{"fighter": 0, "victim": 1})
	fighter, victim := r.body("fighter"), r.body("victim")
	arm := timeline.Channel{Entity: r.part(fighter, "arm"), Body: fighter}
	if err := timeline.Occupy(r.w, arm, 30); err != nil {
		t.Fatal(err)
	}
	r.do("stand")
	if !timeline.Free(r.w, arm) {
		t.Error("Stand did not ready the arm")
	}
	if got := r.busyChannels(victim); got == 0 {
		t.Error("Stand did not delay the other body")
	}
	held := ecs.NewMap[comp.Statuses](r.w.ECS)
	if !held.Has(fighter) || held.Get(fighter).Active[0].Status != "status.standing" {
		t.Error("Stand did not brace the body")
	}
}

func TestGraftNeedsAGraftableAllyAndAPieceFromThePack(t *testing.T) {
	r := newRing(t, map[string]int{"fighter": 0, "victim": 1, "stone": -1})
	fighter, victim := r.body("fighter"), r.body("victim")
	for _, id := range r.optionIDs() {
		if strings.HasPrefix(id, "graft") {
			t.Fatalf("Graft is offered with no piece in the pack: %q", id)
		}
	}
	r.give("item.test_arm", 1, ecs.Entity{})
	if stone, ok := r.option("graft", "item.test_arm", r.uid(r.body("stone"))); ok && stone.Enabled {
		t.Errorf("Graft is offered onto the body that cannot take one: %+v", stone)
	}
	before := len(timeline.Channels(r.w, victim))
	events := r.do("graft", "item.test_arm", r.uid(victim))
	if got := r.busyChannels(fighter); got != 2 {
		t.Errorf("a Graft occupies %d channels, want both arms", got)
	}
	events = append(events, r.wait(6)...)
	if !has(events, func(e api.PartAttached) bool { return e.From == api.AttachedFromPack }) ||
		len(timeline.Channels(r.w, victim)) != before+1 {
		t.Errorf("the graft did not fit: %v", events)
	}
	if got := inventory.Count(r.w, fighter, "item.test_arm", 0); got != 0 {
		t.Errorf("pieces left = %d, want it fitted", got)
	}
}

func TestSeverCutsOnAWeaknessAndOnlyWearsOtherwise(t *testing.T) {
	r := newRing(t, map[string]int{"fighter": 0, "victim": 1})
	victim := r.body("victim")
	arm := r.part(victim, "arm")
	r.do("sever", "type.press", r.uid(victim), r.uid(arm)) // flesh is not weak to Press
	r.wait(7)
	if got := r.integrityOf(arm); got != 2 || r.attached(victim, "arm") != 2 {
		t.Fatalf("a Sever off its weakness left the arm at %d, attached %d; want 2 and still on",
			got, r.attached(victim, "arm"))
	}
	r.do("sever", "type.cut", r.uid(victim), r.uid(arm)) // flesh is weak to Cut
	events := r.wait(7)
	if !has(events, func(e api.PartSevered) bool { return e.Cause == api.SeveredCut }) {
		t.Errorf("a Sever on a weakness did not cut: %v", events)
	}
}

func TestPinLocksThePartAndGiveReadiesAnAlly(t *testing.T) {
	r := newRing(t, map[string]int{"fighter": 0, "victim": 1, "ally": -1})
	victim, ally := r.body("victim"), r.body("ally")
	arm := r.part(victim, "arm")
	r.do("pin", r.uid(victim), r.uid(arm))
	if got := r.busyChannels(r.body("fighter")); got != 1 {
		t.Errorf("a Pin occupies %d channels, want one arm", got)
	}
	r.wait(1)
	if !timeline.Locked(r.w, timeline.Channel{Entity: arm, Body: victim}) {
		t.Error("the pinned arm is not locked")
	}
	for _, c := range timeline.Channels(r.w, ally) {
		if err := timeline.Occupy(r.w, c, 30); err != nil {
			t.Fatal(err)
		}
	}
	busy := r.busyChannels(ally)
	r.wait(3)
	r.do("give", r.uid(ally))
	if got := r.busyChannels(ally); got != busy-1 {
		t.Errorf("busy channels of the ally = %d, want %d: Give readies one", got, busy-1)
	}
}

func TestInASoopOnlyTheFrontMemberReaches(t *testing.T) {
	// Single file: the player's fighter at the back, the ally in front, the victim beyond.
	r := newRing(t, map[string]int{"fighter": -1, "ally": 0, "victim": 1})
	victim := r.body("victim")
	for _, id := range r.optionIDs() {
		for _, form := range []string{"strike", "sweep", "sever", "pin"} {
			if strings.HasPrefix(id, form+"/") && strings.Contains(id, "/"+r.uid(victim)) {
				t.Errorf("the member behind has %q against the creature beyond", id)
			}
		}
	}
	reach := actions.Startable(r.w, r.body("ally"))
	if !slices.ContainsFunc(reach, func(in actions.Intent) bool { return in.Target == victim }) {
		t.Error("the front member has no attack against the creature")
	}
}
