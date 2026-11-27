package explore_test

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game/gametest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// fixture is a running game started in one of the test places.
type fixture struct {
	t   *testing.T
	cfg game.Config
	g   *game.Game
	w   *sim.World
}

// newFixture starts the sample content's party in the named place, or at its own start if empty.
func newFixture(t *testing.T, place api.ContentID, mutate func(*content.Bundle)) *fixture {
	t.Helper()
	bundle := contenttest.SampleBundle(t)
	contenttest.StepOnlyParty(&bundle)
	if place != "" {
		bundle.Start.Place = place
	}
	for id, def := range bundle.Places {
		def.Spawns = nil
		bundle.Places[id] = def
	}
	if mutate != nil {
		mutate(&bundle)
	}
	reg := contenttest.Registry(t, bundle)
	fx := &fixture{t: t}
	fx.cfg = game.Config{Seed: 1, Content: reg}
	fx.g, fx.w, _ = gametest.Start(t, fx.cfg)
	return fx
}

func (fx *fixture) view() api.View {
	fx.t.Helper()
	v, err := fx.g.View()
	if err != nil {
		fx.t.Fatal(err)
	}
	return v
}

// pos is where the body with the UID stands, read from components.
func (fx *fixture) pos(uid api.UID) comp.Pos {
	fx.t.Helper()
	e, ok := fx.w.ByUID(uid)
	if !ok {
		fx.t.Fatalf("no entity %d", uid)
	}
	return *ecs.NewMap[comp.Pos](fx.w.ECS).Get(e)
}

func (fx *fixture) option(id string) api.Option {
	fx.t.Helper()
	for _, o := range fx.view().Options {
		if o.ID == id {
			return o
		}
	}
	fx.t.Fatalf("no option %q", id)
	return api.Option{}
}

func (fx *fixture) apply(cmd api.Command) []api.Event {
	fx.t.Helper()
	events, err := fx.g.Apply(cmd)
	if err != nil {
		fx.t.Fatal(err)
	}
	return events
}

// step takes the current body one step in a direction and returns the events.
func (fx *fixture) step(dir string) []api.Event {
	fx.t.Helper()
	o := fx.option("step/" + dir)
	if !o.Enabled {
		fx.t.Fatalf("step/%s is disabled: %s", dir, o.Reason)
	}
	return fx.apply(o.Command)
}

// toLeader makes the other members wait until the leader decides.
func (fx *fixture) toLeader() {
	fx.t.Helper()
	leader := fx.view().Party.Members[0].UID
	for range 20 {
		if fx.g.Current() == leader {
			return
		}
		fx.apply(fx.option("wait/event").Command)
	}
	fx.t.Fatalf("the leader never got a turn")
}

func (fx *fixture) stepLeader(dir string) []api.Event {
	fx.t.Helper()
	fx.toLeader()
	return fx.step(dir)
}

func find[T api.Event](events []api.Event) (T, bool) {
	for _, e := range events {
		if t, ok := e.(T); ok {
			return t, true
		}
	}
	var zero T
	return zero, false
}

func cellAt(m *api.MapView, h api.Hex) (api.CellView, bool) {
	for _, c := range m.Cells {
		if c.Hex == h {
			return c, true
		}
	}
	return api.CellView{}, false
}

func TestStartOpensOnTheLeaderAtTheHallsCenter(t *testing.T) {
	fx := newFixture(t, "", func(b *content.Bundle) {
		b.Templates["template.party.main"] = content.Template{
			Party: &content.PartyDef{Members: []api.ContentID{"template.person.reaper"}},
		}
	})
	view := fx.view()
	if view.Party == nil || len(view.Party.Members) != 1 || !view.Party.Members[0].Leader {
		t.Fatalf("party = %+v, want one member, leading", view.Party)
	}
	leader := view.Party.Members[0].UID
	if fx.g.Current() != leader || !view.Party.Members[0].Current {
		t.Errorf("current = %d, want the leader %d", fx.g.Current(), leader)
	}
	at := fx.pos(leader)
	if at.At.Depth != 0 || at.At.Hex != view.Party.Hex {
		t.Errorf("leader at %+v, party hex %v; want depth 0 at the view's center", at, view.Party.Hex)
	}
	if view.Map == nil || len(view.Map.Cells) == 0 {
		t.Fatalf("map view = %+v, want the hall drawn", view.Map)
	}
}

func TestWalkingTheSlopeGoesDownFourDepthsWithNoGap(t *testing.T) {
	fx := newFixture(t, "place.test.descent", nil)
	leader := fx.view().Party.Members[0].UID
	start := fx.pos(leader)
	for range 6 {
		fx.stepLeader("e")
	}
	if got := fx.pos(leader); got.At.Hex.Q != start.At.Hex.Q+6 || got.At.Depth != 0 {
		t.Fatalf("leader at %+v, want the hall's east opening at depth 0", got)
	}
	for depth := 1; depth <= 4; depth++ {
		events := fx.stepLeader("e")
		moved, ok := find[api.Moved](events)
		if !ok || moved.Kind != api.MoveDown || moved.To.Depth != depth {
			t.Fatalf("step %d: moved = %+v, %v; want a step down to depth %d", depth, moved, ok, depth)
		}
		fx.toLeader()
		view := fx.view()
		center := fx.pos(leader).At.Hex
		for q := start.At.Hex.Q + 6; q <= start.At.Hex.Q+10; q++ {
			h := api.Hex{Q: q, R: start.At.Hex.R}
			c, drawn := cellAt(view.Map, h)
			if !drawn {
				t.Errorf("after %d steps down: no cell drawn at slope hex %v", depth, h)
				continue
			}
			if want := q - start.At.Hex.Q - 6 - depth; c.Visible && !drawnRight(c, want) {
				t.Errorf(
					"after %d steps down: hex %v = %+v, want the height glyph and weight at Below %d",
					depth,
					h,
					c,
					want,
				)
			}
		}
		if c, _ := cellAt(view.Map, center); c.Below != 0 || c.Glyph != "." {
			t.Errorf("under the leader: %+v, want the floor at Below 0", c)
		}
	}
}

// checkSlopeIsFloor fails unless every slope hex of place.test.descent is drawn as floor with Below
// equal to its depth minus the leader's. The slope's hex k, k = 1..4, is at depth k, k hexes east of
// the hall's opening.
func checkSlopeIsFloor(t *testing.T, fx *fixture, openingQ, row, leaderDepth int, when string) {
	t.Helper()
	view := fx.view()
	for k := range 5 {
		h := api.Hex{Q: openingQ + k, R: row}
		c, drawn := cellAt(view.Map, h)
		if !drawn {
			t.Errorf("%s: no cell drawn at slope hex %v", when, h)
			continue
		}
		if want := k - leaderDepth; !drawnRight(c, want) {
			t.Errorf("%s: hex %v = %+v, want the height glyph and weight at Below %d", when, h, c, want)
		}
	}
}

func TestTheSlopeIsDrawnAsFloorWalkingDownAndBackUp(t *testing.T) {
	fx := newFixture(t, "place.test.descent", nil)
	leader := fx.view().Party.Members[0].UID
	start := fx.pos(leader)
	for range 6 {
		fx.stepLeader("e")
	}
	openingQ, row := start.At.Hex.Q+6, start.At.Hex.R
	checkSlopeIsFloor(t, fx, openingQ, row, 0, "at the opening")
	for depth := 1; depth <= 4; depth++ {
		fx.stepLeader("e")
		fx.toLeader()
		checkSlopeIsFloor(t, fx, openingQ, row, depth, "going down")
	}
	for depth := 3; depth >= 0; depth-- {
		events := fx.stepLeader("w")
		if moved, ok := find[api.Moved](events); !ok || moved.Kind != api.MoveUp || moved.To.Depth != depth {
			t.Fatalf("moved = %+v, %v; want a step up to depth %d", moved, ok, depth)
		}
		fx.toLeader()
		checkSlopeIsFloor(t, fx, openingQ, row, depth, "coming back up")
	}
}

func TestTheRangeCapLeavesFarSlopeFloorToTheWalls(t *testing.T) {
	farSlope := func(planRange int) api.CellView {
		fx := newFixture(t, "place.test.descent", func(b *content.Bundle) { b.Party.PlanRange = planRange })
		leader := fx.view().Party.Members[0].UID
		start := fx.pos(leader)
		for range 6 {
			fx.stepLeader("e")
		}
		c, drawn := cellAt(fx.view().Map, api.Hex{Q: start.At.Hex.Q + 10, R: start.At.Hex.R})
		if !drawn {
			t.Fatalf("range %d: the slope's last hex is not drawn", planRange)
		}
		return c
	}
	if c := farSlope(12); c.Glyph != "v" || c.Below != 4 {
		t.Errorf("within range: %+v, want the slope's floor at Below 4", c)
	}
	if c := farSlope(2); c.Glyph != "#" {
		t.Errorf("past the range cap: %+v, want the wall of the ceiling above the slope", c)
	}
}

func TestAnUnreachableLedgeDoesNotHideTheWallBesideIt(t *testing.T) {
	fx := newFixture(t, "place.test.ledge", func(b *content.Bundle) {
		b.Text["text.place.test.ledge"] = "ledge"
		b.Places["place.test.ledge"] = content.PlaceDef{
			Name:   "text.place.test.ledge",
			At:     content.PlaceAt{Col: 30, Row: 22, Depth: 3},
			Legend: map[string]api.ContentID{"#": "terrain.packed_clay", ".": "terrain.air"},
			Marks:  map[string]string{"@": "party_start"},
			Layers: []content.LayerDef{
				{Depth: 0, Rows: []string{"@."}},
				{Depth: -1, Rows: []string{" ."}},
				{Depth: -2, Rows: []string{" .."}},
				{Depth: -3, Rows: []string{" .."}},
			},
		}
		b.Templates["template.party.main"] = content.Template{
			Party: &content.PartyDef{Members: []api.ContentID{"template.person.reaper"}},
		}
	})
	leader := fx.view().Party.Members[0].UID
	at := fx.pos(leader).At
	h := api.Hex{Q: at.Hex.Q + 2, R: at.Hex.R}
	sight, err := fx.w.Map.Visible(at, 6, 6)
	if err != nil || !slices.Contains(sight, api.Cell{Hex: h, Depth: -2}) {
		t.Fatalf("the shelf at %v is not visible (%v); the test needs it in sight", h, err)
	}
	c, drawn := cellAt(fx.view().Map, h)
	if !drawn {
		t.Fatalf("hex %v is not drawn", h)
	}
	if c.Glyph != "#" || c.Below != 0 {
		t.Errorf("hex %v = %+v, want the wall at the party's depth, not the unreachable shelf", h, c)
	}
}

func TestTheSoopUnderTheHallIsNeverDrawnAndNoStepReachesIt(t *testing.T) {
	fx := newFixture(t, "place.test.descent", nil)
	leader := fx.view().Party.Members[0].UID
	start := fx.pos(leader)
	for _, dir := range []string{"w", "w", "sw", "se", "e", "ne", "ne", "nw", "w"} {
		fx.stepLeader(dir)
		for _, c := range fx.view().Map.Cells {
			if c.Visible && c.Below != 0 {
				t.Fatalf("hall cell %+v drawn above or below the party; the soop at depth 2 must stay hidden", c)
			}
		}
		if got := fx.pos(leader); got.At.Depth != 0 {
			t.Fatalf("leader reached depth %d at %v; start %v", got.At.Depth, got.At.Hex, start.At.Hex)
		}
	}
}

func TestSteppingOffTheLedgeDropsThreeAndClimbingBackIsDisabled(t *testing.T) {
	fx := newFixture(t, "place.test.descent", nil)
	leader := fx.view().Party.Members[0].UID
	for range 7 {
		fx.stepLeader("e")
	}
	if got := fx.pos(leader).At.Depth; got != 1 {
		t.Fatalf("leader at depth %d, want 1 on the slope's first cell", got)
	}
	events := fx.stepLeader("ne")
	moved, ok := find[api.Moved](events)
	if !ok || moved.Kind != api.MoveDrop || moved.To.Depth != 4 {
		t.Fatalf("moved = %+v, %v; want a drop to depth 4", moved, ok)
	}
	if fell, ok := find[api.Fell](events); !ok || fell.Cells != 3 {
		t.Errorf("fell = %+v, %v; want 3 cells", fell, ok)
	}
	fx.toLeader()
	if o := fx.option("step/sw"); o.Enabled || o.Reason == "" {
		t.Errorf("climbing back: %+v, want a disabled option with a reason", o)
	}
}

func TestAStepUpCostsMoreThanTheSameStepDown(t *testing.T) {
	fx := newFixture(t, "place.test.descent", nil)
	leader := fx.view().Party.Members[0].UID
	for range 10 {
		fx.stepLeader("e")
	}
	if fx.pos(leader).At.Depth != 4 {
		t.Fatalf("leader at %+v, want the chamber's depth", fx.pos(leader))
	}
	stepBusy := func(dir string) ([]api.Event, int64) {
		fx.toLeader()
		start := fx.w.Now
		events := fx.step(dir)
		e, _ := fx.w.ByUID(leader)
		return events, ecs.NewMap[comp.Busy](fx.w.ECS).Get(e).Until - start
	}
	events, up := stepBusy("w")
	if moved, _ := find[api.Moved](events); moved.Kind != api.MoveUp {
		t.Fatalf("moved = %+v, want a step up", moved)
	}
	events, down := stepBusy("e")
	if moved, _ := find[api.Moved](events); moved.Kind != api.MoveDown {
		t.Fatalf("moved = %+v, want a step down", moved)
	}
	if up <= down {
		t.Errorf("a step up kept the channel busy %d ticks, a step down %d; want up to cost more", up, down)
	}
}

func TestASlowerBodysStepsTakeLonger(t *testing.T) {
	ticks := func(speed int64) int64 {
		fx := newFixture(t, "", func(b *content.Bundle) {
			b.Timeline.StepTicks = "speed >= 2 ? 10 : 25"
			legs := b.Parts["part.human_legs"]
			legs.Grants = []comp.Grant{{Kind: "speed", Value: int(speed)}}
			b.Parts["part.human_legs"] = legs
		})
		before := fx.view().Tick
		fx.stepLeader("e")
		fx.stepLeader("e")
		return fx.view().Tick - before
	}
	if fast, slow := ticks(2), ticks(1); slow <= fast {
		t.Errorf("two steps took %d ticks slowly and %d quickly; want slow to take longer", slow, fast)
	}
}

func TestRememberedCellsStayAfterTheyLeaveSight(t *testing.T) {
	fx := newFixture(t, "place.test.descent", nil)
	leader := fx.view().Party.Members[0].UID
	first := fx.view().Map
	var remembered api.CellView
	for _, c := range first.Cells {
		if c.Visible && c.Hex != fx.pos(leader).At.Hex {
			remembered = c
			break
		}
	}
	for range 10 {
		fx.stepLeader("e")
	}
	for _, c := range fx.view().Map.Cells {
		if c.Hex == remembered.Hex {
			if c.Visible {
				t.Skipf("hex %v is still in sight", c.Hex)
			}
			return
		}
	}
	t.Errorf("hex %v was seen at the start but is gone from the view", remembered.Hex)
}

func TestMapViewHasNoDepthOrLayerKeys(t *testing.T) {
	fx := newFixture(t, "place.test.descent", nil)
	for range 8 {
		fx.stepLeader("e")
	}
	view := fx.view()
	for name, section := range map[string]any{"map": view.Map, "party": view.Party} {
		data, err := json.Marshal(section)
		if err != nil {
			t.Fatal(err)
		}
		var decoded any
		if err := json.Unmarshal(data, &decoded); err != nil {
			t.Fatal(err)
		}
		var keys func(v any) []string
		keys = func(v any) []string {
			var out []string
			switch v := v.(type) {
			case map[string]any:
				for k, child := range v {
					out = append(out, k)
					out = append(out, keys(child)...)
				}
			case []any:
				for _, child := range v {
					out = append(out, keys(child)...)
				}
			}
			return out
		}
		for _, key := range keys(decoded) {
			if strings.Contains(key, "depth") || strings.Contains(key, "layer") {
				t.Errorf("%s view has a field named %q", name, key)
			}
		}
	}
}

func TestBlockedStepsAreDisabledWithTheBumpReason(t *testing.T) {
	fx := newFixture(t, "", nil)
	for range 12 {
		fx.toLeader()
		if !fx.option("step/w").Enabled {
			break
		}
		fx.stepLeader("w")
	}
	fx.toLeader()
	o := fx.option("step/w")
	if o.Enabled || o.Reason != "Packed clay. No way through." {
		t.Errorf("step/w = %+v, want it disabled by the bump text", o)
	}
	if !slices.ContainsFunc(fx.view().Options, func(o api.Option) bool { return o.Enabled }) {
		t.Error("no option is enabled")
	}
}

// legsIntegrity is the integrity of every legs part in the world, added up.
func (fx *fixture) legsIntegrity() int {
	total := 0
	query := ecs.NewFilter1[comp.Part](fx.w.ECS).Query()
	for query.Next() {
		if part := query.Get(); part.Slot == "legs" {
			total += part.Integrity
		}
	}
	return total
}

func TestAThreeCellDropInjuresTheLegsAndAOneCellStepDownDoesNot(t *testing.T) {
	fx := newFixture(t, "place.test.descent", nil)
	leader := fx.view().Party.Members[0].UID
	for range 7 {
		fx.stepLeader("e")
	}
	before := fx.legsIntegrity()
	fx.stepLeader("ne")
	if got := fx.legsIntegrity(); got != before-2 {
		t.Errorf("legs integrity after a three-cell drop = %d, want %d", got, before-2)
	}

	before = fx.legsIntegrity()
	fx.w.Bus.Publish(api.Fell{Body: leader, Cells: 1})
	if _, err := fx.w.Bus.Drain(fx.w); err != nil {
		t.Fatal(err)
	}
	if got := fx.legsIntegrity(); got != before {
		t.Errorf("legs integrity after a one-cell fall = %d, want %d", got, before)
	}
}
