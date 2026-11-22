package content

import (
	"strings"
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

func TestSampleActions(t *testing.T) {
	reg, err := Load(sampleDir)
	if err != nil {
		t.Fatal(err)
	}
	def, ok := reg.Action("action.step")
	if !ok || def.Key != "step" || def.Target != TargetDirection {
		t.Fatalf("action.step = %+v, %v", def, ok)
	}
	env := ActionEnv{Speed: 2, StepTicks: 10, Choice: map[string]string{}}
	if windup, err := reg.ActionWindup("action.step", env); err != nil || windup != 0 {
		t.Errorf("windup = %d, %v; want 0", windup, err)
	}
	if recovery, err := reg.ActionRecovery("action.step", env); err != nil || recovery != 10 {
		t.Errorf("recovery = %d, %v; want step_ticks = 10", recovery, err)
	}
	if ok, err := reg.ActionRequires("action.step", env); err != nil || !ok {
		t.Errorf("an action with no requires = %v, %v; want it to pass", ok, err)
	}
	if _, present, err := reg.ActionConnects("action.step", env); err != nil || present {
		t.Errorf("connects present = %v, %v; want the default", present, err)
	}
	sites := reg.EffectSites()
	if len(sites) != 1 || sites[0].Path != `actions["action.step"].on_land` {
		t.Errorf("EffectSites = %+v", sites)
	}
	template, _ := reg.Template("template.person.reaper")
	if len(template.Actions) != 1 || template.Actions[0] != "action.step" {
		t.Errorf("the reaper knows %v, want action.step", template.Actions)
	}
}

func TestLoadRejectsBadActions(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Bundle)
		want   []string
	}{
		{
			name:   "unknown target kind",
			mutate: func(b *Bundle) { change(b, func(d *ActionDef) { d.Target = "sideways" }) },
			want:   []string{`actions["action.step"].target`, "sideways"},
		},
		{
			name:   "no channels",
			mutate: func(b *Bundle) { change(b, func(d *ActionDef) { d.Uses = nil }) },
			want:   []string{`actions["action.step"].uses`, "at least one channel"},
		},
		{
			name:   "bad channel tag and negative count",
			mutate: func(b *Bundle) { change(b, func(d *ActionDef) { d.Uses = []Requirement{{Tag: "Arm", Count: -1}} }) },
			want:   []string{"uses[0].tag", "uses[0].count"},
		},
		{
			name:   "windup does not compile",
			mutate: func(b *Bundle) { change(b, func(d *ActionDef) { d.Windup = "nope + 1" }) },
			want:   []string{`actions["action.step"].windup`},
		},
		{
			name:   "recovery missing",
			mutate: func(b *Bundle) { change(b, func(d *ActionDef) { d.Recovery = "" }) },
			want:   []string{`actions["action.step"].recovery`, "required"},
		},
		{
			name:   "requires is not a boolean",
			mutate: func(b *Bundle) { change(b, func(d *ActionDef) { d.Requires = "speed + 1" }) },
			want:   []string{`actions["action.step"].requires`},
		},
		{
			name:   "unknown name text",
			mutate: func(b *Bundle) { change(b, func(d *ActionDef) { d.Name = "text.action.nope" }) },
			want:   []string{`actions["action.step"].name`, "text.action.nope"},
		},
		{
			name:   "bad key",
			mutate: func(b *Bundle) { change(b, func(d *ActionDef) { d.Key = "Step!" }) },
			want:   []string{`actions["action.step"].key`},
		},
		{
			name: "duplicate key",
			mutate: func(b *Bundle) {
				other := b.Actions["action.step"]
				b.Actions["action.other"] = other
			},
			want: []string{"already the key of"},
		},
		{
			name:   "repeated choice",
			mutate: func(b *Bundle) { change(b, func(d *ActionDef) { d.Choices = []string{"type", "type"} }) },
			want:   []string{"choices[1]", "listed twice"},
		},
		{
			name: "template with an unknown action",
			mutate: func(b *Bundle) {
				template := b.Templates["template.person.reaper"]
				template.Actions = []api.ContentID{"action.nope"}
				b.Templates["template.person.reaper"] = template
			},
			want: []string{`templates["template.person.reaper"].actions[0]`, "action.nope"},
		},
		{
			name: "action ID without the prefix",
			mutate: func(b *Bundle) {
				b.Actions["step.again"] = b.Actions["action.step"]
				delete(b.Actions, "action.step")
			},
			want: []string{`actions["step.again"]`, `must start with "action."`},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bundle, err := LoadBundle(sampleDir)
			if err != nil {
				t.Fatal(err)
			}
			tc.mutate(&bundle)
			_, err = New(bundle)
			if err == nil {
				t.Fatal("New accepted bad content")
			}
			for _, fragment := range tc.want {
				if !strings.Contains(err.Error(), fragment) {
					t.Errorf("error %q lacks %q", err, fragment)
				}
			}
		})
	}
}

// change edits action.step in place.
func change(b *Bundle, edit func(*ActionDef)) {
	def := b.Actions["action.step"]
	edit(&def)
	b.Actions["action.step"] = def
}
