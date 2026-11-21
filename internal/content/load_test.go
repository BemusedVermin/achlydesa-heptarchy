package content

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

const sampleDir = "../../content/data"

// writeData writes src as the only CUE file of a fresh data package and returns its directory.
func writeData(t *testing.T, src string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "data.cue"), []byte("package data\n\n"+src), 0o600); err != nil {
		t.Fatal(err)
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
}
`

func TestLoadSample(t *testing.T) {
	registry, err := Load(sampleDir)
	if err != nil {
		t.Fatal(err)
	}
	cost, err := registry.MoveCost("terrain.packed_clay", MoveEnv{})
	if err != nil || cost != 1 {
		t.Fatalf("MoveCost = %d, %v; want 1, nil", cost, err)
	}
	if _, ok := registry.Material("material.clay"); !ok {
		t.Error("material.clay missing")
	}
	ids := registry.TerrainIDs()
	if len(ids) != 2 || ids[0] != "terrain.air" || ids[1] != "terrain.packed_clay" {
		t.Errorf("TerrainIDs = %v", ids)
	}
	if got := registry.Render(api.Text{ID: "text.explore.bump"}); got != "Packed clay. No way through." {
		t.Errorf("Render = %q", got)
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
		name string
		src  string
		want []string // every fragment must appear in the error
	}{
		{
			name: "solid terrain without move_cost",
			src: `
terrains: "terrain.wall": {name: "text.a", solid: true, opaque: true}
text: "text.a": "a"
`,
			want: []string{`terrains["terrain.wall"].move_cost`, "required"},
		},
		{
			name: "bad id",
			src: `
materials: "Bad Id": name: "text.a"
text: "text.a": "a"
`,
			want: []string{`materials["Bad Id"]`},
		},
		{
			name: "wrong prefix",
			src: `
materials: "terrain.clay": name: "text.a"
text: "text.a": "a"
`,
			want: []string{`materials["terrain.clay"]`, `"material."`},
		},
		{
			name: "dangling text reference",
			src: `
materials: "material.clay": name: "text.missing"
text: "text.a": "a"
`,
			want: []string{`materials["material.clay"].name`, "text.missing"},
		},
		{
			name: "formula does not compile",
			src: `
terrains: "terrain.wall": {name: "text.a", solid: true, opaque: true, move_cost: "1 +"}
text: "text.a": "a"
`,
			want: []string{`terrains["terrain.wall"].move_cost`},
		},
		{
			name: "formula uses unknown variable",
			src: `
terrains: "terrain.wall": {name: "text.a", solid: true, opaque: true, move_cost: "weight"}
text: "text.a": "a"
`,
			want: []string{`terrains["terrain.wall"].move_cost`, "weight"},
		},
		{
			name: "formula returns the wrong type",
			src: `
terrains: "terrain.wall": {name: "text.a", solid: true, opaque: true, move_cost: "'heavy'"}
text: "text.a": "a"
`,
			want: []string{`terrains["terrain.wall"].move_cost`},
		},
		{
			name: "bad text template",
			src: `
text: "text.a": "{{.who"
`,
			want: []string{`text["text.a"]`},
		},
		{
			name: "all problems are reported",
			src: `
materials: "material.clay": name: "text.missing"
terrains: "terrain.wall": {name: "text.a", solid: true, opaque: true}
text: "text.a": "a"
`,
			want: []string{`materials["material.clay"].name`, `terrains["terrain.wall"].move_cost`},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := Load(writeData(t, test.src))
			if err == nil {
				t.Fatal("expected an error")
			}
			for _, fragment := range test.want {
				if !strings.Contains(err.Error(), fragment) {
					t.Errorf("error %q lacks %q", err, fragment)
				}
			}
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
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "data.cue"), []byte("package data\n\n"+src), 0o600); err != nil {
			t.Fatal(err)
		}
		_, _ = Load(dir) // any result is fine; it must not panic
	})
}
