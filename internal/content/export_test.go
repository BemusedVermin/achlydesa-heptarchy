package content

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"
)

const dataPath = "../../content/data"

// These twin contenttest, which this package cannot import without a cycle.
var (
	sample     = sync.OnceValues(func() (*Registry, error) { return Load(dataPath) })
	sampleJSON = sync.OnceValues(func() ([]byte, error) { return loadBundleJSON(dataPath) })
)

func sampleBundle(tb testing.TB) Bundle {
	tb.Helper()
	return copyBundle(tb, sampleJSON)
}

func sampleRegistry(tb testing.TB) *Registry {
	tb.Helper()
	reg, err := sample()
	if err != nil {
		tb.Fatal(err)
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

func copyBundle(tb testing.TB, encoded func() ([]byte, error)) Bundle {
	tb.Helper()
	data, err := encoded()
	if err != nil {
		tb.Fatal(err)
	}
	var bundle Bundle
	if err := json.Unmarshal(data, &bundle); err != nil {
		tb.Fatal(err)
	}
	return bundle
}

func requireErrContains(tb testing.TB, err error, fragments ...string) {
	tb.Helper()
	if err == nil {
		tb.Fatalf("got no error, want one containing %q", fragments)
	}
	for _, fragment := range fragments {
		if !strings.Contains(err.Error(), fragment) {
			tb.Errorf("error %q lacks %q", err, fragment)
		}
	}
}
