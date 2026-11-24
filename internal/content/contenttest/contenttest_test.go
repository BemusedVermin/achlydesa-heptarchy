package contenttest

import (
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
)

func TestSampleBundleCopyMatchesLoaded(t *testing.T) {
	loaded, err := content.LoadBundle(Dir())
	if err != nil {
		t.Fatal(err)
	}
	want := Registry(t, loaded).Hash()
	if got := Registry(t, SampleBundle(t)).Hash(); got != want {
		t.Errorf("copy hash %q, loaded hash %q", got, want)
	}
}

func TestSampleBundleCopiesAreIndependent(t *testing.T) {
	first := SampleBundle(t)
	delete(first.Materials, "material.clay")
	if _, ok := SampleBundle(t).Materials["material.clay"]; !ok {
		t.Error("changing one copy changed the next")
	}
}
