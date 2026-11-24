package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestFix(t *testing.T) {
	tests := []struct {
		name, text, want string
	}{
		{
			"module path",
			"import \"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api\"\n",
			"import \"github.com/bemusedvermin/achlydesa-heptarchy/internal/core/api\"\n",
		},
		{
			"embedded field",
			"#Item: {\n\tData: #Data\n\tweight: int\n}\n",
			"#Item: {\n\t#Data\n\tweight: int\n}\n",
		},
		{
			"field of a struct type that isn't embedded",
			"#Item: {\n\tParent: #Data\n}\n",
			"#Item: {\n\tParent: #Data\n}\n",
		},
		{
			"named type with constants",
			"#Dir: uint8 // #enumDir\n\n#enumDir:\n\t#DirE |\n\t#DirW\n",
			"#Dir: #enumDir\n\n#enumDir:\n\t#DirE |\n\t#DirW\n",
		},
		{
			"named type whose set is not in the file",
			"#Dir: uint8 // #enumDir\n",
			"#Dir: uint8 // #enumDir\n",
		},
		{
			"a field on the top level",
			"Data: #Data\n",
			"Data: #Data\n",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := fix(test.text); got != test.want {
				t.Errorf("fix(%q) = %q, want %q", test.text, got, test.want)
			}
		})
	}
}

func TestUnenforced(t *testing.T) {
	tests := []struct {
		name, text string
		want       []string
	}{
		{"enforced", "#Dir: #enumDir\n\n#enumDir:\n\t#DirE |\n\t#DirW\n", nil},
		{"named only in a comment", "#Dir: uint8 // #enumDir\n\n#enumDir:\n\t#DirE |\n\t#DirW\n", []string{"Dir"}},
		{"another type's set", "#Dir: #enumSide\n\n#enumDir:\n\t#DirE\n", []string{"Dir"}},
		{"no sets", "#Cell: {\n\tdepth: int\n}\n", nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := unenforced(test.text); !slices.Equal(got, test.want) {
				t.Errorf("unenforced = %v, want %v", got, test.want)
			}
		})
	}
}

func TestCommittedSchemasEnforceEverySet(t *testing.T) {
	root := filepath.Join("..", "..")
	found := 0
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() || !strings.HasSuffix(path, "_go_gen.cue") {
			return err
		}
		found++
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if types := unenforced(string(data)); len(types) > 0 {
			t.Errorf("%s: %v accept values outside their constants; run go tool task schema", path, types)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if found == 0 {
		t.Fatal("no generated schemas found under internal")
	}
}
