package content

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// writeData writes src as the only CUE file of a fresh data package and returns its directory.
func writeData(tb testing.TB, src string) string {
	tb.Helper()
	dir := tb.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "data.cue"), []byte("package data\n\n"+src), 0o600); err != nil {
		tb.Fatal(err)
	}
	return dir
}

const validData = `
materials: "material.clay": name: "text.material.clay"
terrains: "terrain.wall": {
	name:      "text.terrain.wall"
	solid:     true
	opaque:    true
	move_cost: "1 + encumbrance"
}
text: {
	"text.material.clay": "clay"
	"text.terrain.wall":  "wall"
	"text.hello":         "Hello, {{.who}}!"
	"text.phase.dawn":    "dawn"
	"text.action.step":   "step"
}
actions: "action.step": {
	name: "text.action.step", key: "step", uses: [{tag: "body", count: 1}]
	target: "direction", windup: "0", recovery: "step_ticks"
}
rules: agents_score: "1.0"
agents: {score_rule: "agents_score", dying_below: 0.5, promote_radius: 8}
party: {slots: 4, sight_radius: 6, follow_action: "action.step"}
combat: {ai_score: "1.0", aims: "true", attack_tag: "attack", corpse_tag: "corpse", step_action: "action.step"}
timeline: {step_ticks: "speed >= 2 ? 10 : 20", speed: "grant.speed", default_speed: 2, window_ticks: 30}
calendar: {
	phases: [{id: "phase.dawn", name: "text.phase.dawn", ticks: 60}]
}
`

func TestLoadSample(t *testing.T) {
	registry := sampleRegistry(t)
	if _, err := registry.MoveCost("terrain.packed_clay", MoveEnv{}); err != nil {
		t.Fatalf("MoveCost: %v", err)
	}
	if _, ok := registry.Material("material.clay"); !ok {
		t.Error("material.clay missing")
	}
	ids := registry.TerrainIDs()
	if !slices.IsSorted(ids) || !slices.Contains(ids, "terrain.air") || !slices.Contains(ids, "terrain.packed_clay") {
		t.Errorf("TerrainIDs = %v, want air and packed clay in order", ids)
	}
	if got := registry.Render(api.Text{ID: "text.explore.bump"}); got == "" || strings.Contains(got, "[missing text:") {
		t.Errorf("Render = %q", got)
	}
}

func TestKnowledgeContentIsChecked(t *testing.T) {
	data := validData + `
claims: "claim.material": name: "text.missing_name"
origins: "origin.sweep": name: "text.also_missing"
`
	_, err := Load(writeData(t, data))
	requireErrContains(t, err,
		`claims["claim.material"].name: text reference "text.missing_name"`,
		`claims["claim.material"] log text: text reference "text.knowledge.material"`,
		`origins["origin.sweep"].name: text reference "text.also_missing"`,
	)
}

func TestADanglingTextReferenceReportsOnce(t *testing.T) {
	bundle, err := LoadBundle(writeData(t, validData))
	if err != nil {
		t.Fatal(err)
	}
	delete(bundle.Text, "text.material.clay")
	_, err = New(bundle)
	requireErrContains(t, err, "text.material.clay")
	if n := strings.Count(err.Error(), `text reference "text.material.clay" does not exist`); n != 1 {
		t.Errorf("the missing text is reported %d times, want once: %v", n, err)
	}
}

func TestStepTicks(t *testing.T) {
	registry, err := Load(writeData(t, validData))
	if err != nil {
		t.Fatal(err)
	}
	if got := registry.Timeline().DefaultSpeed; got != 2 {
		t.Errorf("default speed = %d, want 2", got)
	}
	for speed, want := range map[int]int64{0: 20, 1: 20, 2: 10, 3: 10} {
		if got, err := registry.StepTicks(SpeedEnv{Speed: speed}); err != nil || got != want {
			t.Errorf("StepTicks(speed %d) = %d, %v; want %d", speed, got, err, want)
		}
	}
}

func TestLoadFormulaUsesEnv(t *testing.T) {
	registry, err := Load(writeData(t, validData))
	if err != nil {
		t.Fatal(err)
	}
	cost, err := registry.MoveCost("terrain.wall", MoveEnv{Encumbrance: 2})
	if err != nil || cost != 3 {
		t.Fatalf("MoveCost = %d, %v; want 3, nil", cost, err)
	}
	if _, err := registry.MoveCost("terrain.nope", MoveEnv{}); err == nil {
		t.Error("unknown terrain should error")
	}
}

func TestLoadRejects(t *testing.T) {
	tests := []struct {
		name          string
		src           string
		wantFragments []string
	}{
		{
			name: "solid terrain without move_cost",
			src: `
terrains: "terrain.wall": {name: "text.terrain.wall", solid: true, opaque: true}
text: {"text.a": "a", "text.terrain.wall": "a"}
`,
			wantFragments: []string{`terrains["terrain.wall"].move_cost`, "required"},
		},
		{
			name: "party with unknown member",
			src: `
templates: "template.party.x": party: members: ["template.nope"]
`,
			wantFragments: []string{`templates["template.party.x"].party.members[0]`, "template.nope"},
		},
		{
			name: "party nested in a party",
			src: `
templates: {
	"template.party.a": party: members: ["template.party.b"]
	"template.party.b": party: members: ["template.party.a"]
}
`,
			wantFragments: []string{"is itself a party"},
		},
		{
			name: "party with no members",
			src: `
templates: "template.party.x": party: members: []
`,
			wantFragments: []string{"at least one member"},
		},
		{
			name: "party numbers too small",
			src: `
party: {slots: 0, sight_radius: 0}
`,
			wantFragments: []string{"party.slots", "party.sight_radius"},
		},
		{
			name: "step_ticks that does not compile",
			src: `
timeline: step_ticks: "speed +"
`,
			wantFragments: []string{"timeline.step_ticks"},
		},
		{
			name: "step_ticks with an unknown variable",
			src: `
timeline: step_ticks: "swiftness"
`,
			wantFragments: []string{"timeline.step_ticks", "swiftness"},
		},
		{
			name: "negative default speed",
			src: `
timeline: default_speed: -1
`,
			wantFragments: []string{"timeline.default_speed"},
		},
		{
			name: "actor without a controller",
			src: `
templates: "template.x": actor: controller: ""
`,
			wantFragments: []string{`templates["template.x"].actor.controller`},
		},
		{
			name: "bad template id",
			src: `
templates: "person.x": {}
`,
			wantFragments: []string{`templates["person.x"]`, `"template."`},
		},
		{
			name: "bad id",
			src: `
materials: "Bad Id": name: "text.a"
text: "text.a": "a"
`,
			wantFragments: []string{`materials["Bad Id"]`},
		},
		{
			name: "wrong prefix",
			src: `
materials: "terrain.clay": name: "text.a"
text: "text.a": "a"
`,
			wantFragments: []string{`materials["terrain.clay"]`, `"material."`},
		},
		{
			name: "dangling text reference",
			src: `
materials: "material.clay": name: "text.missing"
text: "text.a": "a"
`,
			wantFragments: []string{`materials["material.clay"].name`, "text.missing"},
		},
		{
			name: "formula does not compile",
			src: `
terrains: "terrain.wall": {name: "text.terrain.wall", solid: true, opaque: true, move_cost: "1 +"}
text: {"text.a": "a", "text.terrain.wall": "a"}
`,
			wantFragments: []string{`terrains["terrain.wall"].move_cost`},
		},
		{
			name: "formula uses unknown variable",
			src: `
terrains: "terrain.wall": {name: "text.terrain.wall", solid: true, opaque: true, move_cost: "weight"}
text: {"text.a": "a", "text.terrain.wall": "a"}
`,
			wantFragments: []string{`terrains["terrain.wall"].move_cost`, "weight"},
		},
		{
			name: "formula returns the wrong type",
			src: `
terrains: "terrain.wall": {name: "text.terrain.wall", solid: true, opaque: true, move_cost: "'heavy'"}
text: {"text.a": "a", "text.terrain.wall": "a"}
`,
			wantFragments: []string{`terrains["terrain.wall"].move_cost`},
		},
		{
			name: "bad text template",
			src: `
text: "text.a": "{{.who"
`,
			wantFragments: []string{`text["text.a"]`},
		},
		{
			name: "bad calendar",
			src: `
text: {"text.a": "a", "text.terrain.wall": "a"}
calendar: {phases: [{id: "day.dawn", name: "text.missing", ticks: 0}]}
`,
			wantFragments: []string{"calendar.phases[0].ticks", "calendar.phases[0]", "text.missing"},
		},
		{
			name: "all problems are reported",
			src: `
materials: "material.clay": name: "text.missing"
terrains: "terrain.wall": {name: "text.terrain.wall", solid: true, opaque: true}
text: {"text.a": "a", "text.terrain.wall": "a"}
`,
			wantFragments: []string{`materials["material.clay"].name`, `terrains["terrain.wall"].move_cost`},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Load(writeData(t, test.src))
			requireErrContains(t, err, test.wantFragments...)
		})
	}
}

func TestHash(t *testing.T) {
	first, err := Load(writeData(t, validData))
	if err != nil {
		t.Fatal(err)
	}
	second, err := Load(writeData(t, validData))
	if err != nil {
		t.Fatal(err)
	}
	if first.Hash() != second.Hash() || first.Hash() == "" {
		t.Errorf("same data gave hashes %q and %q", first.Hash(), second.Hash())
	}
	changed, err := Load(writeData(t, strings.Replace(validData, `"clay"`, `"mud"`, 1)))
	if err != nil {
		t.Fatal(err)
	}
	if changed.Hash() == first.Hash() {
		t.Error("changing a value should change the hash")
	}
}

func TestRender(t *testing.T) {
	registry, err := Load(writeData(t, validData))
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name string
		in   api.Text
		want string
	}{
		{"argument", api.Text{ID: "text.hello", Args: map[string]string{"who": "Ada"}}, "Hello, Ada!"},
		{"missing argument", api.Text{ID: "text.hello"}, "Hello, !"},
		{"missing id", api.Text{ID: "text.nope"}, "[missing text: text.nope]"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := registry.Render(test.in); got != test.want {
				t.Errorf("Render = %q, want %q", got, test.want)
			}
		})
	}
}

func FuzzLoad(f *testing.F) {
	f.Add(validData)
	f.Add("")
	f.Add(`text: "text.a": "{{"`)
	f.Add(`terrains: x: {solid: true, move_cost: "1 +"}`)
	f.Fuzz(func(t *testing.T, src string) {
		_, _ = Load(writeData(t, src))
	})
}
