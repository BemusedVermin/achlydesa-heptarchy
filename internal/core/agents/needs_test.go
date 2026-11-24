package agents_test

import (
	"math"
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

const leader = "template.person.reaper"

// fixture is a game on the sample content whose leader carries needs and knows how to drink.
type fixture struct {
	t   *testing.T
	g   *game.Game
	w   *sim.World
	cfg game.Config
}

// newFixture starts the leader with the given levels and lets mutate adjust the needs first.
func newFixture(t *testing.T, levels map[string]float64, mutate func(*content.Bundle)) *fixture {
	t.Helper()
	f := &fixture{t: t}
	f.g, f.w, f.cfg = gametest.New(t, 1, func(bundle *content.Bundle) {
		contenttest.StandingCrawler(bundle)
		contenttest.Edit(bundle.Needs, "need.water", func(need *content.NeedDef) { need.Rise = "0.1" })
		contenttest.Edit(bundle.Needs, "need.food", func(need *content.NeedDef) { need.Rise = "0" })
		for key, amount := range map[string]string{"sip": "0.25", "salt": "-5"} {
			id := "action.test." + key
			bundle.Actions[id] = content.ActionDef{
				Name: api.ContentID("text." + id), Key: key, Target: content.TargetSelf,
				Uses:   []content.Requirement{{Tag: "body", Count: 1}},
				Windup: "0", Recovery: "1",
				OnLand: []content.EffectRef{contenttest.Do("satisfy", "need", "need.water", "amount", amount)},
			}
			bundle.Text["text."+id] = key
		}
		contenttest.Edit(bundle.Templates, leader, func(person *content.Template) {
			person.Needs = levels
			person.Actions = append(person.Actions, "action.test.sip", "action.test.salt")
		})
		if mutate != nil {
			mutate(bundle)
		}
	})
	return f
}

func (f *fixture) leader() ecs.Entity {
	f.t.Helper()
	var found []ecs.Entity
	query := ecs.NewFilter1[comp.Identity](f.w.ECS).Query()
	for query.Next() {
		if query.Get().Template == leader {
			found = append(found, query.Entity())
		}
	}
	if len(found) != 1 {
		f.t.Fatalf("found %d leaders, want 1", len(found))
	}
	return found[0]
}

func (f *fixture) level(need api.ContentID) float64 {
	f.t.Helper()
	needs := ecs.NewMap[comp.Needs](f.w.ECS)
	body := f.leader()
	if !needs.Has(body) {
		f.t.Fatal("the leader carries no needs")
	}
	for _, level := range needs.Get(body).Levels {
		if level.Need == need {
			return level.Level
		}
	}
	f.t.Fatalf("the leader does not carry %s", need)
	return 0
}

func (f *fixture) wantLevel(need api.ContentID, want float64) {
	f.t.Helper()
	if got := f.level(need); math.Abs(got-want) > 1e-9 {
		f.t.Errorf("%s at tick %d = %v, want %v", need, f.w.Now, got, want)
	}
}

func (f *fixture) apply(cmd api.Command) {
	f.t.Helper()
	events, err := f.g.Apply(cmd)
	if err != nil {
		f.t.Fatal(err)
	}
	for _, e := range events {
		if rejected, ok := e.(api.Rejected); ok {
			f.t.Fatalf("%s rejected: %+v", cmd.CommandKind(), rejected)
		}
	}
}

// nextPhase waits until the next phase boundary has been crossed.
func (f *fixture) nextPhase() {
	f.t.Helper()
	start := f.w.Now
	f.apply(api.Wait{Source: f.g.Current(), Until: api.WaitPhase})
	if f.w.Now == start {
		f.t.Fatalf("time did not move from tick %d", start)
	}
}

// do takes the leader's option that starts the action.
func (f *fixture) do(action api.ContentID) {
	f.t.Helper()
	view, err := f.g.View()
	if err != nil {
		f.t.Fatal(err)
	}
	for _, option := range view.Options {
		if do, ok := option.Command.(api.Do); ok && do.Action == action && option.Enabled {
			f.apply(option.Command)
			return
		}
	}
	f.t.Fatalf("no enabled option for %s", action)
}

func TestNeedsRiseEachPhase(t *testing.T) {
	f := newFixture(t, map[string]float64{"need.water": 0.2}, nil)
	f.wantLevel("need.water", 0.2)
	f.nextPhase()
	f.wantLevel("need.water", 0.3)
	f.nextPhase()
	f.wantLevel("need.water", 0.4)
}

func TestNeedsRiseNoHigherThanOne(t *testing.T) {
	f := newFixture(t, map[string]float64{"need.water": 0.95}, nil)
	f.nextPhase()
	f.wantLevel("need.water", 1)
}

func TestSatisfyKeepsTheLevelInRange(t *testing.T) {
	f := newFixture(t, map[string]float64{"need.water": 0.2}, nil)
	f.do("action.test.sip")
	f.wantLevel("need.water", 0)
	f.nextPhase()
	f.wantLevel("need.water", 0.1)
	f.do("action.test.salt")
	f.wantLevel("need.water", 1)
}

func TestSatisfyIgnoresABodyWithoutTheNeed(t *testing.T) {
	f := newFixture(t, map[string]float64{"need.food": 0.5}, nil)
	f.do("action.test.sip")
	f.wantLevel("need.food", 0.5)
}

func TestThresholdAppliesWhileItHolds(t *testing.T) {
	f := newFixture(t, map[string]float64{"need.water": 0.4, "need.food": 0}, func(bundle *content.Bundle) {
		contenttest.Edit(bundle.Needs, "need.water", func(need *content.NeedDef) {
			need.Thresholds = []content.Threshold{{
				At: 0.5,
				Do: []content.EffectRef{contenttest.Do("satisfy", "need", "need.food", "amount", "-0.1")},
			}}
		})
	})
	f.nextPhase()
	f.wantLevel("need.water", 0.5)
	f.wantLevel("need.food", 0.1)
	f.nextPhase()
	f.wantLevel("need.water", 0.6)
	f.wantLevel("need.food", 0.2)
	f.do("action.test.sip")
	f.nextPhase()
	f.wantLevel("need.water", 0.45)
	f.wantLevel("need.food", 0.2)
}

func TestFormulasReadNeedsAndPhase(t *testing.T) {
	f := newFixture(t, map[string]float64{"need.water": 0.5, "need.food": 0}, func(bundle *content.Bundle) {
		contenttest.Edit(bundle.Needs, "need.food", func(need *content.NeedDef) {
			need.Rise = `phase == "phase.day" ? source.needs.water / 10 : 0`
		})
	})
	f.nextPhase()
	f.wantLevel("need.water", 0.6)
	f.wantLevel("need.food", 0.05)
}

func TestSwappingToContentWithoutANeedLetsTheGamePlayOn(t *testing.T) {
	f := newFixture(t, map[string]float64{"need.water": 0.2, "need.food": 0.5}, nil)
	bundle := contenttest.SampleBundle(t)
	delete(bundle.Needs, "need.water")
	for key, template := range bundle.Templates {
		delete(template.Needs, "need.water")
		bundle.Templates[key] = template
	}
	if err := f.g.SwapContent(contenttest.Registry(t, bundle)); err != nil {
		t.Fatal(err)
	}
	f.nextPhase()
	f.nextPhase()
	f.wantLevel("need.water", 0.2)
}

func TestNeedsSurviveASave(t *testing.T) {
	f := newFixture(t, map[string]float64{"need.water": 0.2, "need.food": 0.7}, nil)
	f.nextPhase()
	want, err := f.g.Hash()
	if err != nil {
		t.Fatal(err)
	}
	loaded, world, _ := gametest.Reload(t, f.cfg, f.g)
	got, err := loaded.Hash()
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Error("the loaded game hashes differently")
	}
	f.g, f.w = loaded, world
	f.wantLevel("need.water", 0.3)
	f.wantLevel("need.food", 0.7)
}

func TestContentErrorsFailAtLoad(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*content.Bundle)
		want   string
	}{
		{"a template names an unknown need", func(bundle *content.Bundle) {
			contenttest.Edit(bundle.Templates, leader, func(person *content.Template) {
				person.Needs = map[string]float64{"need.air": 0.1}
			})
		}, `unknown need "need.air"`},
		{"a starting level above one", func(bundle *content.Bundle) {
			contenttest.Edit(bundle.Templates, leader, func(person *content.Template) {
				person.Needs = map[string]float64{"need.water": 1.5}
			})
		}, "from 0 to 1"},
		{"thresholds out of order", func(bundle *content.Bundle) {
			contenttest.Edit(bundle.Needs, "need.water", func(need *content.NeedDef) {
				need.Thresholds = []content.Threshold{{At: 0.8}, {At: 0.5}}
			})
		}, "ascending"},
		{"a rise that names an unknown variable", func(bundle *content.Bundle) {
			contenttest.Edit(bundle.Needs, "need.water", func(need *content.NeedDef) { need.Rise = "thirst" })
		}, "rise"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bundle := contenttest.SampleBundle(t)
			tt.mutate(&bundle)
			_, err := content.New(bundle)
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error = %v, want one that mentions %q", err, tt.want)
			}
		})
	}
}

func TestThresholdNamingAnUnknownEffectFailsWhenAGameStarts(t *testing.T) {
	bundle := contenttest.SampleBundle(t)
	contenttest.Edit(bundle.Needs, "need.water", func(need *content.NeedDef) {
		need.Thresholds = []content.Threshold{{At: 0.5, Do: []content.EffectRef{contenttest.Do("parch")}}}
	})
	_, err := game.New(game.Config{Seed: 1, Content: contenttest.Registry(t, bundle)})
	if err == nil || !strings.Contains(err.Error(), `unknown primitive "parch"`) {
		t.Errorf("error = %v, want an unknown primitive", err)
	}
}

func TestSatisfyNamingAnUnknownNeedFailsWhenAGameStarts(t *testing.T) {
	bundle := contenttest.SampleBundle(t)
	contenttest.Edit(bundle.Needs, "need.water", func(need *content.NeedDef) {
		need.Thresholds = []content.Threshold{{
			At: 0.5, Do: []content.EffectRef{contenttest.Do("satisfy", "need", "need.air", "amount", "1")},
		}}
	})
	_, err := game.New(game.Config{Seed: 1, Content: contenttest.Registry(t, bundle)})
	if err == nil || !strings.Contains(err.Error(), `unknown need "need.air"`) {
		t.Errorf("error = %v, want an unknown need", err)
	}
}
