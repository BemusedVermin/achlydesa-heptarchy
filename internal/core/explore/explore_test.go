package explore_test

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/clock"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/rng"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/prototype"
)

const sampleDir = "../../../content/data"

// fixture is a running game on one of the test maps.
type fixture struct {
	t   *testing.T
	cfg game.Config
	g   *game.Game
	w   *sim.World
}

// newFixture starts the sample content's party on the named map, or on its own start map if empty.
func newFixture(t *testing.T, mapID api.ContentID, mutate func(*content.Bundle)) *fixture {
	t.Helper()
	bundle, err := content.LoadBundle(sampleDir)
	if err != nil {
		t.Fatal(err)
	}
	if mapID != "" {
		bundle.Start.Map = mapID
	}
	if mutate != nil {
		mutate(&bundle)
	}
	reg, err := content.New(bundle)
	if err != nil {
		t.Fatal(err)
	}
	fx := &fixture{t: t}
	fx.cfg = game.Config{Seed: 1, Content: reg, Extra: func(w *sim.World, _ *clock.Clock) error {
		fx.w = w
		return nil
	}}
	if fx.g, err = game.New(fx.cfg); err != nil {
		t.Fatal(err)
	}
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

// toLeader makes the other members wait for an event until the leader decides. Followers arrive
// with M1-16; until then every member is player-controlled and takes its own turn.
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

// stepLeader steps the leader and returns the events.
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
		// A party of one, so the view's sight is the leader's alone, as in the prototype.
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
	if view.Map == nil || view.Map.Name != "The test hall" {
		t.Fatalf("map view = %+v, want the test hall", view.Map)
	}
	// The visible cells are exactly the prototype's field of view, around the same center.
	want := map[api.Hex]bool{}
	for h, ok := range prototype.New().View().Visible {
		if ok {
			want[api.Hex{Q: h.Q, R: h.R}] = true
		}
	}
	got := map[api.Hex]bool{}
	for _, c := range view.Map.Cells {
		if c.Visible {
			got[api.Hex{Q: c.Hex.Q - at.At.Hex.Q, R: c.Hex.R - at.At.Hex.R}] = true
		}
	}
	// M1-14's sight is line based and the prototype's is shadowcasting, so a few hexes behind the
	// pillars differ; everything else agrees.
	differ := 0
	for h := range want {
		if !got[h] {
			differ++
		}
	}
	for h := range got {
		if !want[h] {
			differ++
		}
	}
	if differ > len(want)/10 {
		t.Errorf("%d of %d hexes differ from the prototype's field of view", differ, len(want))
	}
}

func TestWalkingTheSlopeGoesDownFourDepthsWithNoGap(t *testing.T) {
	fx := newFixture(t, "map.test.descent", nil)
	leader := fx.view().Party.Members[0].UID
	start := fx.pos(leader)
	for range 6 { // to the opening in the hall's east wall
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
		for q := start.At.Hex.Q + 6; q <= start.At.Hex.Q+10; q++ { // the opening and the four slope hexes
			h := api.Hex{Q: q, R: start.At.Hex.R}
			if _, drawn := cellAt(view.Map, h); !drawn {
				t.Errorf("after %d steps down: no cell drawn at slope hex %v", depth, h)
			}
		}
		if c, _ := cellAt(view.Map, center); c.Below != 0 || c.Glyph != "." {
			t.Errorf("under the leader: %+v, want the floor at Below 0", c)
		}
	}
}

func TestTheSoopUnderTheHallIsNeverDrawnAndNoStepReachesIt(t *testing.T) {
	fx := newFixture(t, "map.test.descent", nil)
	leader := fx.view().Party.Members[0].UID
	start := fx.pos(leader)
	// Walk every hex of the hall's floor that the leader can reach with plain steps in the hall.
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
	fx := newFixture(t, "map.test.descent", nil)
	leader := fx.view().Party.Members[0].UID
	for range 7 { // the opening, then the first slope cell at depth 1
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
	fx := newFixture(t, "map.test.descent", nil)
	leader := fx.view().Party.Members[0].UID
	for range 10 { // down the whole slope to the chamber
		fx.stepLeader("e")
	}
	if fx.pos(leader).At.Depth != 4 {
		t.Fatalf("leader at %+v, want the chamber's depth", fx.pos(leader))
	}
	busy := func() int64 {
		e, _ := fx.w.ByUID(leader)
		return ecs.NewMap[comp.Busy](fx.w.ECS).Get(e).Until - fx.w.Now
	}
	events := fx.stepLeader("w") // up the slope: its first step is a step up
	if moved, _ := find[api.Moved](events); moved.Kind != api.MoveUp {
		t.Fatalf("moved = %+v, want a step up", moved)
	}
	up := busy()
	events = fx.stepLeader("e") // and back down
	if moved, _ := find[api.Moved](events); moved.Kind != api.MoveDown {
		t.Fatalf("moved = %+v, want a step down", moved)
	}
	if down := busy(); up <= down {
		t.Errorf("a step up kept the channel busy %d ticks, a step down %d; want up to cost more", up, down)
	}
}

func TestASlowerBodysStepsTakeLonger(t *testing.T) {
	ticks := func(speed int64) int64 {
		fx := newFixture(t, "", func(b *content.Bundle) {
			b.Timeline.StepTicks = "speed >= 2 ? 10 : 25"
			b.Timeline.DefaultSpeed = int(speed)
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
	fx := newFixture(t, "map.test.descent", nil)
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
	fx := newFixture(t, "map.test.descent", nil)
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

func TestReplayOfRandomStepsGivesTheSameHash(t *testing.T) {
	fx := newFixture(t, "map.test.descent", nil)
	rnd := rng.New(7).Get("test.steps")
	moved := 0
	for range 200 {
		var steps []api.Option
		for _, o := range fx.view().Options {
			if strings.HasPrefix(o.ID, "step/") && o.Enabled {
				steps = append(steps, o)
			}
		}
		if len(steps) == 0 {
			t.Fatal("no enabled step")
		}
		events := fx.apply(steps[rnd.IntN(len(steps))].Command)
		if _, ok := find[api.Moved](events); ok {
			moved++
		}
	}
	if moved == 0 {
		t.Fatal("no random step moved anyone")
	}
	want, err := fx.g.Hash()
	if err != nil {
		t.Fatal(err)
	}
	replayed, err := game.Replay(fx.cfg, fx.g.History())
	if err != nil {
		t.Fatal(err)
	}
	if got, err := replayed.Hash(); err != nil || got != want {
		t.Errorf("replay hash = %x, %v; want %x", got, err, want)
	}
}

func TestBlockedStepsAreDisabledWithTheBumpReason(t *testing.T) {
	fx := newFixture(t, "", nil)
	for range 12 { // walk west until the ring wall stops the leader
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
