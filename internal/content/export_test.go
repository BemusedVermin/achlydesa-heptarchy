package content

import "testing"

const dataPath = "../../content/data"

// sampleRegistry and sampleBundle twin contenttest, which this package cannot import without a cycle.
func sampleRegistry(t testing.TB) *Registry {
	t.Helper()
	reg, err := Load(dataPath)
	if err != nil {
		t.Fatal(err)
	}
	return reg
}

func sampleBundle(t testing.TB) Bundle {
	t.Helper()
	bundle, err := LoadBundle(dataPath)
	if err != nil {
		t.Fatal(err)
	}
	return bundle
}
