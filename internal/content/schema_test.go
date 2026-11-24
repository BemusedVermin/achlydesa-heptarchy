package content

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// copySchemaTree copies the CUE of the repository (cue.mod, the generated schemas under internal and
// content/data) into a temporary module root, so content can be loaded there with extra files added.
// It returns the data directory.
func copySchemaTree(t *testing.T) string {
	t.Helper()
	root := filepath.Join("..", "..")
	dest := t.TempDir()
	for _, sub := range []string{"cue.mod", "internal", "content"} {
		err := filepath.WalkDir(filepath.Join(root, sub), func(path string, entry fs.DirEntry, err error) error {
			if err != nil || entry.IsDir() {
				return err
			}
			if filepath.Ext(path) != ".cue" {
				return nil
			}
			rel, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			target := filepath.Join(dest, rel)
			if err := os.MkdirAll(filepath.Dir(target), 0o750); err != nil {
				return err
			}
			return os.WriteFile(target, data, 0o600)
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	return filepath.Join(dest, "content", "data")
}

func TestCUEConstraintsHoldAgainstTheRealSchema(t *testing.T) {
	dir := copySchemaTree(t)
	if _, err := Load(dir); err != nil {
		t.Fatalf("the copied sample does not load: %v", err)
	}
	// A material both weak and immune to one type must be refused by materials.cue.
	const override = `package data

materials: "material.both": {name: "text.material.both", lists: {weak_to: ["type.cut"], immune_to: ["type.cut"]}}
`
	if err := os.WriteFile(filepath.Join(dir, "override.cue"), []byte(override), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(dir); err == nil || !strings.Contains(err.Error(), "_overlap") {
		t.Errorf("a type that is both a weakness and an immunity loaded: err = %v", err)
	}
}
