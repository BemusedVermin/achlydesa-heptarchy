package explore_test

import (
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/actions"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game/gametest"
)

var dirNames = [...]string{"e", "ne", "nw", "w", "sw", "se"}

// walkThroughGateTo steps the leader along the shortest route to the gate of its map that leads to the named map.
func (fx *fixture) walkThroughGateTo(to string) {
	fx.t.Helper()
	leader := fx.partyUIDs()[0]
	for range 100 {
		pos := fx.pos(leader)
		m := fx.w.Maps.Get(pos.Map)
		var goal api.Cell
		found := false
		for _, gate := range m.Gates() {
			if gate.To == fx.mapNamed(to) {
				goal, found = gate.At, true
			}
		}
		if !found {
			fx.t.Fatalf("map %q has no gate to %q", m.Name(), to)
		}
		if pos.At == goal {
			return
		}
		route, ok, err := m.Path(pos.At, goal, actions.StepRules(fx.w), nil, 10000)
		if err != nil || !ok {
			fx.t.Fatalf("no route to the gate to %q: %v", to, err)
		}
		dir, ok, err := m.DirTo(pos.At, route[1], actions.StepRules(fx.w))
		if err != nil || !ok {
			fx.t.Fatalf("no step to %v: %v", route[1], err)
		}
		fx.stepLeader(dirNames[dir])
		if fx.pos(leader).Map != pos.Map {
			return
		}
	}
	fx.t.Fatalf("never reached the gate to %q", to)
}

func (fx *fixture) templatesOn(mapID api.MapID) map[api.ContentID]bool {
	found := map[api.ContentID]bool{}
	query := ecs.NewFilter2[comp.Pos, comp.Identity](fx.w.ECS).Query()
	for query.Next() {
		pos, identity := query.Get()
		if pos.Map == mapID {
			found[identity.Template] = true
		}
	}
	return found
}

func TestThePlaygroundGatesReachEveryCreature(t *testing.T) {
	shipped, err := content.LoadBundle(contenttest.Dir())
	if err != nil {
		t.Fatal(err)
	}
	if shipped.Start.Map != "map.play.hub" {
		t.Fatalf("the shipped start map is %q, want the playground hub", shipped.Start.Map)
	}
	bundle := contenttest.SampleBundle(t)
	contenttest.StepOnlyParty(&bundle)
	contenttest.Edit(bundle.Needs, "need.undecayed", func(n *content.NeedDef) { n.Rise = "0" })
	for _, id := range []string{"faction.party", "faction.wild", "faction.archon"} {
		contenttest.Edit(bundle.Factions, id, func(f *content.Faction) { f.HostileTo = nil })
	}
	bundle.Start.Map = shipped.Start.Map
	fx := &fixture{t: t}
	fx.g, fx.w, _ = gametest.Start(t, game.Config{Seed: 1, Content: contenttest.Registry(t, bundle)})
	hub := fx.mapNamed("map.play.hub")
	if got := fx.pos(fx.partyUIDs()[0]).Map; got != hub {
		t.Fatalf("party starts on map %d, want the hub %d", got, hub)
	}
	routes := []struct {
		to   string
		want []api.ContentID
	}{
		{"map.play.hall", []api.ContentID{"template.creature.moving_clay"}},
		{"map.play.den", []api.ContentID{"template.creature.remnant", "template.creature.beast"}},
		{"map.play.descent", []api.ContentID{"template.creature.moving_clay"}},
		{"map.test.town", []api.ContentID{"template.town.well"}},
		{"map.play.feeders", []api.ContentID{"template.creature.soul_feeder"}},
		{"map.play.shrine", []api.ContentID{"template.resident.believer", "template.creature.archon_warden"}},
	}
	for _, route := range routes {
		t.Run(route.to, func(t *testing.T) {
			fx.t = t
			fx.walkThroughGateTo(route.to)
			target := fx.mapNamed(route.to)
			if got := fx.pos(fx.partyUIDs()[0]).Map; got != target {
				t.Fatalf("leader on map %d after the gate, want %d", got, target)
			}
			present := fx.templatesOn(target)
			for _, id := range route.want {
				if !present[id] {
					t.Errorf("%s has no %s", route.to, id)
				}
			}
			fx.walkThroughGateTo("map.play.hub")
			if got := fx.pos(fx.partyUIDs()[0]).Map; got != hub {
				t.Errorf("leader on map %d after the way back, want the hub %d", got, hub)
			}
		})
	}
}
