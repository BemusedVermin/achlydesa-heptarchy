package agents_test

import (
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/agents"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/clock"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game/gametest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/knowledge"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

const (
	chatPlanner = "test.chat"
	chatterTpl  = "template.test.chatter"
	listenerTpl = "template.test.listener"
	spokenClaim = api.ContentID("claim.gives")
	spokenValue = "affordance.water"
	subjectUID  = api.UID(777)
)

type noPlan struct{}

func (noPlan) Plan(agents.Snapshot, api.ContentID) (agents.Plan, bool, error) {
	return agents.Plan{}, false, nil
}

func chatContent(b *content.Bundle) {
	for id, controller := range map[string]string{chatterTpl: agents.MindController, listenerTpl: "idle"} {
		b.Templates[id] = content.Template{
			Name:    new(api.ContentID("text." + id)),
			Actor:   &content.ActorDef{Controller: controller},
			Mind:    &content.MindDef{Planner: chatPlanner},
			Actions: []api.ContentID{"action.chat"},
			Faction: new(api.ContentID("faction.party")),
		}
	}
	b.Text["text.template.test.chatter"] = "Chatter"
	b.Text["text.template.test.listener"] = "Listener"
}

type chatFixture struct {
	g      *game.Game
	w      *sim.World
	hearer api.UID
}

// newChatFixture is a speaker who knows a fact at confidence 0.5, and a hearer the given distance
// away who already knows each of hearerKnows.
func newChatFixture(t *testing.T, distance int, hearerKnows ...api.Fact) *chatFixture {
	t.Helper()
	fx := &chatFixture{}
	fx.g, fx.w, _ = gametest.New(t, 1, chatContent, func(cfg *game.Config) {
		cfg.Install = func(w *sim.World, _ *clock.Clock) error {
			return agents.RegisterPlanner(w, chatPlanner, noPlan{})
		}
		cfg.Spawn = func(w *sim.World, _ *clock.Clock) error {
			w.Map = hexmap.NewMap(hexmap.FuncSource(func(c api.Cell) hexmap.Terrain {
				if c.Depth > 0 {
					return hexmap.Terrain{ID: "terrain.packed_clay", Solid: true}
				}
				return hexmap.Terrain{ID: "terrain.air"}
			}), hexmap.Prisms{})
			speaker, err := w.Spawn(chatterTpl, &comp.Pos{At: walkStart})
			if err != nil {
				return err
			}
			far := walkStart
			far.Hex.Q += distance
			hearer, err := w.Spawn(listenerTpl, &comp.Pos{At: far})
			if err != nil {
				return err
			}
			fx.hearer = w.UIDOf(hearer)
			knowledge.Learn(w, speaker, api.Fact{
				Subject: subjectUID, Claim: spokenClaim, Value: spokenValue, Origin: "origin.seen", Confidence: 0.5,
			})
			for _, f := range hearerKnows {
				knowledge.Learn(w, hearer, f)
			}
			return nil
		}
	})
	return fx
}

func (fx *chatFixture) hearerFacts(t *testing.T) []api.Fact {
	t.Helper()
	for range 12 {
		if _, err := fx.g.Apply(api.Wait{Source: fx.g.Current(), Until: api.WaitEvent}); err != nil {
			t.Fatal(err)
		}
	}
	hearer, ok := fx.w.ByUID(fx.hearer)
	if !ok {
		t.Fatal("the hearer is gone")
	}
	return knowledge.About(fx.w, hearer, subjectUID)
}

func TestAToldFactReachesTheHearerLessSure(t *testing.T) {
	fx := newChatFixture(t, 1)
	cfg := fx.w.Content.Agents()
	facts := fx.hearerFacts(t)
	if len(facts) != 1 {
		t.Fatalf("hearer holds %d facts about the subject, want 1: %v", len(facts), facts)
	}
	got := facts[0]
	if got.Origin != cfg.ToldOrigin || got.Value != spokenValue || got.Claim != spokenClaim {
		t.Errorf("told fact = %+v, want origin %q, value %q, claim %q", got, cfg.ToldOrigin, spokenValue, spokenClaim)
	}
	if want := 0.5 * cfg.ToldConfidence; got.Confidence != want {
		t.Errorf("told confidence = %v, want %v", got.Confidence, want)
	}
}

func TestAHearerKeepsWhatItHoldsMoreSurely(t *testing.T) {
	held := api.Fact{
		Subject: subjectUID, Claim: spokenClaim, Value: "affordance.food", Origin: "origin.seen", Confidence: 1,
	}
	facts := newChatFixture(t, 1, held).hearerFacts(t)
	if len(facts) != 1 || facts[0].Origin != "origin.seen" || facts[0].Value != "affordance.food" {
		t.Errorf("hearer facts = %v, want only the one it saw", facts)
	}
}

func TestTellingNeedsReach(t *testing.T) {
	if facts := newChatFixture(t, 6).hearerFacts(t); len(facts) != 0 {
		t.Errorf("a hearer out of reach was told %v", facts)
	}
}
