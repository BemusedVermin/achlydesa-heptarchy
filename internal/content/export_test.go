package content

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"
)

const dataPath = "../../content/data"

// sampleRegistry and sampleBundle twin contenttest, which this package cannot import without a cycle.
// Both load content/data once per test binary.
var (
	sample     = sync.OnceValues(func() (*Registry, error) { return Load(dataPath) })
	sampleJSON = sync.OnceValues(func() ([]byte, error) { return loadBundleJSON(dataPath) })
)

// sampleBundle returns a copy of the sample bundle that the test may change.
func sampleBundle(t testing.TB) Bundle {
	t.Helper()
	return copyBundle(t, sampleJSON)
}

func sampleRegistry(t testing.TB) *Registry {
	t.Helper()
	reg, err := sample()
	if err != nil {
		t.Fatal(err)
	}
	return reg
}

func loadBundleJSON(dir string) ([]byte, error) {
	bundle, err := LoadBundle(dir)
	if err != nil {
		return nil, err
	}
	return json.Marshal(bundle)
}

// copyBundle decodes a fresh bundle from its encoding, so a test may change every part of it.
func copyBundle(t testing.TB, encoded func() ([]byte, error)) Bundle {
	t.Helper()
	data, err := encoded()
	if err != nil {
		t.Fatal(err)
	}
	var bundle Bundle
	if err := json.Unmarshal(data, &bundle); err != nil {
		t.Fatal(err)
	}
	return bundle
}

// requireErrContains fails the test unless err is an error whose text holds every fragment.
func requireErrContains(t testing.TB, err error, fragments ...string) {
	t.Helper()
	if err == nil {
		t.Fatalf("got no error, want one containing %q", fragments)
	}
	for _, fragment := range fragments {
		if !strings.Contains(err.Error(), fragment) {
			t.Errorf("error %q lacks %q", err, fragment)
		}
	}
}
