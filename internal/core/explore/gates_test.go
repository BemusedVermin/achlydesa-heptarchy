package explore_test

import (
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/explore"
)

func (fx *fixture) mapNamed(name string) api.MapID {
	fx.t.Helper()
	for _, id := range fx.w.Maps.IDs() {
		if fx.w.Maps.Get(id).Name() == name {
			return id
		}
	}
	fx.t.Fatalf("no map %q", name)
	return 0
}

func (fx *fixture) partyUIDs() []api.UID {
	fx.t.Helper()
	var uids []api.UID
	for _, m := range fx.view().Party.Members {
		uids = append(uids, m.UID)
	}
	return uids
}

func (fx *fixture) walkToGate() []api.Event {
	fx.t.Helper()
	var events []api.Event
	for range 4 {
		events = append(events, fx.stepLeader("e")...)
	}
	return events
}

func TestTheWholePartyCrossesAGate(t *testing.T) {
	fx := newFixture(t, "map.test.gate_a", nil)
	a, b := fx.mapNamed("map.test.gate_a"), fx.mapNamed("map.test.gate_b")
	events := fx.walkToGate()
	crossed, ok := find[api.PartyCrossed](events)
	if !ok || crossed.From != a || crossed.To != b {
		t.Fatalf("crossed = %+v (found %v), want from %d to %d", crossed, ok, a, b)
	}
	mark, _ := fx.w.Maps.Get(b).Mark("from_a")
	members := fx.partyUIDs()
	for i, uid := range members {
		pos := fx.pos(uid)
		if pos.Map != b {
			t.Errorf("member %d is on map %d, want %d", i, pos.Map, b)
		}
		if i == 0 && pos.At != mark {
			t.Errorf("leader at %v, want the mark %v", pos.At, mark)
		}
	}
	if view := fx.view(); view.Map == nil || view.Map.Map != b {
		t.Errorf("plan view shows %+v, want map %d", view.Map, b)
	}
}

func TestWalkingBackKeepsWhatWasSeen(t *testing.T) {
	fx := newFixture(t, "map.test.gate_a", nil)
	a := fx.mapNamed("map.test.gate_a")
	fx.walkToGate()
	party, _ := fx.w.ByUID(fx.view().Party.UID)
	before := len(ecs.NewMap[comp.Seen](fx.w.ECS).Get(party).ByMap[a])
	if before == 0 {
		t.Fatal("nothing seen on the first map")
	}
	for range 3 {
		fx.stepLeader("w")
	}
	if got := fx.pos(fx.partyUIDs()[0]).Map; got != a {
		t.Fatalf("leader is on map %d, want %d", got, a)
	}
	after := len(ecs.NewMap[comp.Seen](fx.w.ECS).Get(party).ByMap[a])
	if after < before {
		t.Errorf("seen cells on the first map fell from %d to %d", before, after)
	}
}

func TestAWipedPartyWakesOnTheMapOfItsRestPlace(t *testing.T) {
	fx := newFixture(t, "map.test.gate_a", nil)
	a := fx.mapNamed("map.test.gate_a")
	fx.walkToGate()
	downed := ecs.NewMap[comp.Downed](fx.w.ECS)
	actors := ecs.NewMap[comp.Actor](fx.w.ECS)
	for _, uid := range fx.partyUIDs() {
		e, _ := fx.w.ByUID(uid)
		actors.Remove(e)
		downed.Add(e, &comp.Downed{})
	}
	if err := explore.Wake(fx.w); err != nil {
		t.Fatal(err)
	}
	for _, uid := range fx.partyUIDs() {
		if got := fx.pos(uid).Map; got != a {
			t.Errorf("member %d woke on map %d, want %d", uid, got, a)
		}
	}
}

func TestAGateToAnUnknownMapFailsToLoad(t *testing.T) {
	bundle := contenttest.SampleBundle(t)
	def := bundle.Maps["map.test.gate_a"]
	def.Gates = map[string]content.GateDef{"g": {To: "map.test.nowhere", Mark: "from_a"}}
	bundle.Maps["map.test.gate_a"] = def
	if _, err := content.New(bundle); err == nil {
		t.Fatal("content with a gate to an unknown map loaded")
	}
}

func TestAGateToAnUnknownMarkFailsToLoad(t *testing.T) {
	bundle := contenttest.SampleBundle(t)
	def := bundle.Maps["map.test.gate_a"]
	def.Gates = map[string]content.GateDef{"g": {To: "map.test.gate_b", Mark: "nowhere"}}
	bundle.Maps["map.test.gate_a"] = def
	if _, err := content.New(bundle); err == nil {
		t.Fatal("content with a gate to an unknown mark loaded")
	}
}
