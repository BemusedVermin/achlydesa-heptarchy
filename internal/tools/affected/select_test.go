package main

import (
	"slices"
	"testing"
)

// module: c imports b imports a; d reads content and e imports d; f stands alone.
var module = []pkg{
	{ImportPath: "m/a", Dir: "internal/a"},
	{ImportPath: "m/b", Dir: "internal/b", Imports: []string{"m/a"}},
	{ImportPath: "m/c", Dir: "internal/c", Imports: []string{"m/b"}},
	{ImportPath: "m/d", Dir: "internal/d", ReadsContent: true},
	{ImportPath: "m/e", Dir: "internal/e", Imports: []string{"m/d"}},
	{ImportPath: "m/f", Dir: "cmd/f"},
}

func TestSelectPackages(t *testing.T) {
	everything := []string{"m/a", "m/b", "m/c", "m/d", "m/e", "m/f"}
	tests := []struct {
		name    string
		changed []string
		want    []string
	}{
		{"nothing changed", nil, []string{}},
		{"docs only", []string{"docs/game-design.md", "tasks/INDEX.md"}, []string{}},
		{"a leaf's dependents follow", []string{"internal/a/a.go"}, []string{"m/a", "m/b", "m/c"}},
		{"a top package alone", []string{"internal/c/c.go"}, []string{"m/c"}},
		{"testdata belongs to its package", []string{"internal/b/testdata/golden/x.txt"}, []string{"m/b", "m/c"}},
		{"content runs readers only", []string{"content/data/items.cue"}, []string{"m/d"}},
		{"content plus code", []string{"content/data/items.cue", "cmd/f/main.go"}, []string{"m/d", "m/f"}},
		{"go.mod runs everything", []string{"go.mod"}, everything},
		{"workflow runs everything", []string{".github/workflows/ci.yml"}, everything},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := selectPackages(test.changed, module)
			if !slices.Equal(got, test.want) {
				t.Errorf("selectPackages(%v) = %v, want %v", test.changed, got, test.want)
			}
		})
	}
}
