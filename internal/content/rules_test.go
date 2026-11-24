package content

import (
	"strings"
	"testing"

	"github.com/expr-lang/expr"
	"github.com/expr-lang/expr/vm"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// loadWith loads validData plus extra CUE.
func loadWith(t *testing.T, extra string) (*Registry, error) {
	t.Helper()
	return Load(writeData(t, validData+extra))
}

func TestRulesAreInlinedAndCheckedWhereUsed(t *testing.T) {
	reg, err := loadWith(t, `
rules: {
	weak:    "chosen.type.id in target.material.lists.weak_to"
	either:  "rule.weak || rule.brittle"
	brittle: "\"brittle\" in target.part.tags"
}
types: "type.cut": name: "text.type.cut"
text: "text.type.cut": "a"
`)
	if err != nil {
		t.Fatal(err)
	}
	program, err := reg.Compile("rule.either ? 1 : 0", EffectEnv{}, expr.AsInt64())
	if err != nil {
		t.Fatal(err)
	}
	run := func(program *vm.Program, env EffectEnv) any {
		t.Helper()
		out, err := expr.Run(program, env)
		if err != nil {
			t.Fatal(err)
		}
		return out
	}
	part := recordWith("part.x", "brittle")
	env := EffectEnv{Target: BodyEnv{Part: part, Material: EmptyRecord(), Template: EmptyRecord()}}
	if got := run(program, env); got != int64(1) {
		t.Errorf("brittle part: rule.either = %v, want 1", got)
	}
	if got := run(program, EffectEnv{}); got != int64(0) {
		t.Errorf("no part: rule.either = %v, want 0", got)
	}
}

func recordWith(id string, tags ...string) Record {
	record := EmptyRecord()
	record.ID, record.Tags = id, tags
	return record
}

func TestMissingDataReadsAsEmpty(t *testing.T) {
	reg, err := loadWith(t, "")
	if err != nil {
		t.Fatal(err)
	}
	for _, src := range []string{
		`chosen.type.props.nope == 0`,
		`!("x" in target.material.lists.nope)`,
		`chosen.type.id == ""`,
		`len(target.part.tags) == 0`,
	} {
		program, err := reg.Compile(src, EffectEnv{}, expr.AsBool())
		if err != nil {
			t.Fatalf("%s: %v", src, err)
		}
		out, err := expr.Run(program, EffectEnv{})
		if err != nil || out != true {
			t.Errorf("%s = %v, %v; want true, nil", src, out, err)
		}
	}
	got, err := reg.Speed(GrantEnv{Grant: map[string]int{}})
	if err != nil || got != 0 {
		t.Errorf("speed with no grants = %d, %v; want 0, nil", got, err)
	}
}

func TestRuleErrorsNameTheFormula(t *testing.T) {
	tests := []struct {
		name  string
		extra string
		want  []string
	}{
		{
			name: "unknown rule in a formula",
			extra: `
terrains: "terrain.pit": {name: "text.terrain.pit", solid: true, opaque: true, move_cost: "rule.nope"}
text: "text.terrain.pit": "pit"
`,
			want: []string{`terrains["terrain.pit"].move_cost`, `unknown rule "nope"`},
		},
		{
			name: "unknown rule inside a rule",
			extra: `
rules: a: "rule.nope"
`,
			want: []string{`rules["a"]`, `unknown rule "nope"`},
		},
		{
			name: "rules that use each other",
			extra: `
rules: {
	a: "rule.b"
	b: "rule.a"
}
`,
			want: []string{"rules[", "cycle"},
		},
		{
			name:  "a rule name that is not lower-case words",
			extra: "rules: \"Bad Name\": \"true\"\n",
			want:  []string{`rules["Bad Name"]`, "lower-case"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := loadWith(t, test.extra)
			requireErrContains(t, err, test.want...)
		})
	}
}

func TestUnusedRuleIsAWarning(t *testing.T) {
	reg, err := loadWith(t, `
rules: {
	used:   "true"
	unused: "false"
}
terrains: "terrain.pit": {name: "text.terrain.pit", solid: true, opaque: true, move_cost: "rule.used ? 1 : 2"}
text: "text.terrain.pit": "pit"
`)
	if err != nil {
		t.Fatal(err)
	}
	warnings := strings.Join(reg.Warnings(), "\n")
	if !strings.Contains(warnings, `rules["unused"]`) || strings.Contains(warnings, `rules["used"]`) {
		t.Errorf("warnings = %q, want one about the unused rule only", warnings)
	}
}

func TestVocabulary(t *testing.T) {
	const closed = `
vocabulary: {
	tags: items: ["consumable"]
	grants: ["action"]
	lists: "materials.weak_to": "types"
}
types: "type.cut": name: "text.type.cut"
text: {"text.type.cut": "a", "text.item.x": "a"}
`
	tests := []struct {
		name    string
		extra   string
		wantErr string
	}{
		{"nothing declared", `
items: "item.x": {name: "text.item.x", tags: ["whatever"]}
parts: "part.x": {name: "text.a", slot: "arm", material: "material.clay", integrity: 1, grants: [{kind: "anything", value: 1}]}
materials: "material.clay": lists: weak_to: ["nope"]
text: {"text.a": "a", "text.item.x": "a"}
`, ""},
		{"undeclared item tag", closed + `
items: "item.x": {name: "text.item.x", tags: ["piece"]}
`, `items["item.x"].tags[0]`},
		{"declared item tag", closed + `
items: "item.x": {name: "text.item.x", tags: ["consumable"]}
`, ""},
		{"undeclared grant kind", closed + `
parts: "part.x": {name: "text.a", slot: "arm", material: "material.clay", integrity: 1, grants: [{kind: "flight", value: 1}]}
`, `parts["part.x"].grants[0].kind`},
		{"grant without a value", closed + `
parts: "part.x": {name: "text.a", slot: "arm", material: "material.clay", integrity: 1, grants: [{kind: "action"}]}
`, `parts["part.x"].grants[0].value`},
		{"a weakness that names no type", closed + `
materials: "material.clay": lists: weak_to: ["type.nope"]
`, `materials["material.clay"].lists["weak_to"][0]`},
		{"a weakness that names a type", closed + `
materials: "material.clay": lists: weak_to: ["type.cut"]
`, ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := loadWith(t, test.extra)
			switch {
			case test.wantErr == "" && err != nil:
				t.Errorf("Load failed: %v", err)
			case test.wantErr != "" && (err == nil || !strings.Contains(err.Error(), test.wantErr)):
				t.Errorf("Load error = %v, want it to contain %q", err, test.wantErr)
			}
		})
	}
}

func TestChoicesAndModsAreChecked(t *testing.T) {
	tests := []struct {
		name  string
		extra string
		want  string
	}{
		{"choice from a bad collection", `
choices: type: from: "terrains"
`, `choices["type"].from`},
		{"choice offer that does not compile", `
choices: type: {from: "types", offer: "nope +"}
`, `choices["type"].offer`},
		{"mods requires that does not compile", `
types: "type.cut": {name: "text.type.cut", requires: "nope +"}
text: "text.type.cut": "a"
`, `types["type.cut"].requires`},
		{"mods refusal without text", `
types: "type.cut": {name: "text.type.cut", requires: "false", refusal: "text.missing"}
text: "text.type.cut": "a"
`, `types["type.cut"].refusal`},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := loadWith(t, test.extra)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Errorf("Load error = %v, want it to contain %q", err, test.want)
			}
		})
	}
}

func TestRecordLookup(t *testing.T) {
	reg := sampleRegistry(t)
	soak, ok := reg.Record("type.soak")
	if !ok || soak.ID != "type.soak" || len(soak.Lists["consumes"]) == 0 {
		t.Errorf("Record(type.soak) = %+v, %v", soak, ok)
	}
	unknown, ok := reg.Record("type.nope")
	if ok || unknown.ID != "type.nope" {
		t.Errorf("Record(type.nope) = %+v, %v; want just the ID, false", unknown, ok)
	}
	clay, _ := reg.Record("material.clay")
	if clay.Props == nil || clay.Props["nope"] != 0 {
		t.Errorf("clay props = %v", clay.Props)
	}
}

func TestFollowActionIsChecked(t *testing.T) {
	bundle, err := LoadBundle(writeData(t, validData))
	if err != nil {
		t.Fatal(err)
	}
	look := bundle.Actions["action.step"]
	look.Key, look.Target = "look", TargetNone
	bundle.Actions["action.look"] = look
	for follow, want := range map[string]string{
		"action.look": "must target a direction",
		"action.nope": `unknown action "action.nope"`,
	} {
		bundle.Party.FollowAction = api.ContentID(follow)
		if _, err := New(bundle); err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("follow_action %q: error = %v, want it to contain %q", follow, err, want)
		}
	}
}

func TestARecordsNameIsItsOwnIDAsText(t *testing.T) {
	tests := []struct {
		name    string
		extra   string
		wantErr string
	}{
		{"beside its record", `
types: "type.cut": name: "text.type.cut"
text: "text.type.cut": "Cut"
`, ""},
		{"another record's text", `
types: "type.cut": name: "text.a"
text: "text.a": "a"
`, `types["type.cut"].name: text ID is "text.a", want "text.type.cut"`},
		{"a part may share a name", `
parts: "part.x": {name: "text.a", slot: "arm", material: "material.clay", integrity: 1, grants: []}
text: "text.a": "arm"
`, ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := loadWith(t, test.extra)
			switch {
			case test.wantErr == "" && err != nil:
				t.Errorf("Load failed: %v", err)
			case test.wantErr != "" && (err == nil || !strings.Contains(err.Error(), test.wantErr)):
				t.Errorf("Load = %v, want an error containing %q", err, test.wantErr)
			}
		})
	}
}
