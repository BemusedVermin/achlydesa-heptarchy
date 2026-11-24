package agents_test

import (
	"strings"
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game"
)

// floorBehavior is an idle behavior that goes for the nearest packed clay and scores 1 only when it
// is right below the agent.
func floorBehavior(bundle *content.Bundle) {
	bundle.Behaviors["behavior.test.floor"] = content.BehaviorDef{
		Name: "text.behavior.test.floor", Do: "idle", Consider: []string{"distance == 1 ? 1 : 0"},
		Target: &content.TargetDef{Kind: content.BehaviorTargetCell, Terrain: "terrain.packed_clay"},
	}
	bundle.Text["text.behavior.test.floor"] = "eyeing the floor"
	contenttest.Edit(bundle.Templates, drinker, func(template *content.Template) {
		template.Behaviors = []api.ContentID{"behavior.test.floor", idle}
	})
}

func TestACellTargetIsTheNearestCellOfItsTerrain(t *testing.T) {
	f := newAgentField(t, 1, floorBehavior)
	f.waitUntil(30)
	if got := f.doing(f.drinker()); got != "behavior.test.floor" {
		t.Errorf("doing %q, want the floor below found", got)
	}
}

func TestACellTargetIsLookedForOnlyWithinTheRadius(t *testing.T) {
	f := newAgentField(t, 1, func(bundle *content.Bundle) {
		floorBehavior(bundle)
		bundle.Agents.CellRadius = 0
	})
	f.waitUntil(30)
	if got := f.doing(f.drinker()); got != idle {
		t.Errorf("doing %q, want %q with no clay in reach", got, idle)
	}
}

func TestConsiderationsReadTheirEnv(t *testing.T) {
	f := newAgentField(t, 1, func(bundle *content.Bundle) {
		bundle.Behaviors["behavior.test.env"] = content.BehaviorDef{
			Name: "text.behavior.test.env", Do: "idle",
			Consider: []string{
				`phase == "phase.dawn" ? 1 : 0`,
				`now >= 0 ? 1 : 0`,
				`doing == "" || doing == "behavior.test.env" ? 1 : 0`,
				`target.template.id == "" ? 1 : 0`,
				`source.template.id == "template.test.drinker" ? 1 : 0`,
			},
		}
		bundle.Text["text.behavior.test.env"] = "reading"
		contenttest.Edit(bundle.Templates, drinker, func(template *content.Template) {
			template.Behaviors = []api.ContentID{"behavior.test.env", idle}
		})
	})
	f.waitUntil(15)
	if got := f.doing(f.drinker()); got != "behavior.test.env" {
		t.Errorf("doing %q, want the behavior whose considerations all hold", got)
	}
}

func TestBehaviorContentErrorsFailAtLoad(t *testing.T) {
	behavior := func(edit func(*content.BehaviorDef)) func(*content.Bundle) {
		return func(bundle *content.Bundle) { contenttest.Edit(bundle.Behaviors, toWater, edit) }
	}
	tests := []struct {
		name   string
		mutate func(*content.Bundle)
		want   string
	}{
		{"a template names an unknown behavior", func(bundle *content.Bundle) {
			contenttest.Edit(bundle.Templates, drinker, func(template *content.Template) {
				template.Behaviors = []api.ContentID{"behavior.test.fly"}
			})
		}, `unknown behavior "behavior.test.fly"`},
		{"a consideration names an unknown variable", behavior(func(def *content.BehaviorDef) {
			def.Consider = []string{"thirst"}
		}), "consider[0]"},
		{"an unknown target kind", behavior(func(def *content.BehaviorDef) {
			def.Target = &content.TargetDef{Kind: "smell"}
		}), "target.kind"},
		{"a mark target without a mark", behavior(func(def *content.BehaviorDef) {
			def.Target = &content.TargetDef{Kind: content.BehaviorTargetMark}
		}), "target.mark: required"},
		{"a cell target of an unknown terrain", behavior(func(def *content.BehaviorDef) {
			def.Target = &content.TargetDef{Kind: content.BehaviorTargetCell, Terrain: "terrain.lava"}
		}), `unknown terrain "terrain.lava"`},
		{"a body filter on a mark target", behavior(func(def *content.BehaviorDef) {
			def.Target.Tags = []string{"graftable"}
		}), "target.tags"},
		{"a negative keep bonus", func(bundle *content.Bundle) { bundle.Agents.KeepBonus = -1 }, "keep_bonus"},
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

func TestBehaviorPrimitiveErrorsFailWhenAGameStarts(t *testing.T) {
	tests := []struct {
		name string
		edit func(*content.BehaviorDef)
		want string
	}{
		{"an unknown primitive", func(def *content.BehaviorDef) { def.Do = "fly" }, `unknown behavior primitive "fly"`},
		{"a param the primitive does not take", func(def *content.BehaviorDef) {
			def.Params = map[string]string{"speed": "2"}
		}, `takes no "speed"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bundle := contenttest.SampleBundle(t)
			contenttest.Edit(bundle.Behaviors, wander, tt.edit)
			_, err := game.New(game.Config{Seed: 1, Content: contenttest.Registry(t, bundle)})
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("error = %v, want one that mentions %q", err, tt.want)
			}
		})
	}
}
