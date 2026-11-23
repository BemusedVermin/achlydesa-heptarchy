package contenttest

import (
	"path/filepath"
	"runtime"
	"sync"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// T is the part of testing.TB the helpers use, which rapid.T shares.
type T interface {
	Helper()
	Fatal(args ...any)
}

// sample loads content/data once. The registry is never written to, which makes it the one shared
// value tests may hold (C1-13).
var sample = sync.OnceValues(func() (*content.Registry, error) { return content.Load(Dir()) })

// Dir is the path of the sample content directory, whatever package the test runs in.
func Dir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "..", "content", "data")
}

// Sample returns the registry of the sample content, shared by every test in the binary.
// A test that changes content takes SampleBundle instead.
func Sample(t T) *content.Registry {
	t.Helper()
	reg, err := sample()
	if err != nil {
		t.Fatal(err)
	}
	return reg
}

// SampleBundle loads the sample content again, so the test may change the bundle it gets.
func SampleBundle(t T) content.Bundle {
	t.Helper()
	bundle, err := content.LoadBundle(Dir())
	if err != nil {
		t.Fatal(err)
	}
	return bundle
}

// Registry builds a registry from the bundle, failing the test if the content is invalid.
func Registry(t T, bundle content.Bundle) *content.Registry {
	t.Helper()
	reg, err := content.New(bundle)
	if err != nil {
		t.Fatal(err)
	}
	return reg
}

// Edit applies f to the record at key and stores the result back, because map values aren't addressable.
func Edit[V any](m map[string]V, key string, f func(*V)) {
	v := m[key]
	f(&v)
	m[key] = v
}

// StandingCrawler keeps the sample moving clay where it is, so random walkers aren't killed.
func StandingCrawler(bundle *content.Bundle) {
	Edit(bundle.Templates, "template.creature.moving_clay", func(clay *content.Template) {
		clay.Actor = &content.ActorDef{Controller: "still"}
		clay.Actions = nil
		clay.Lists = nil
	})
}

// StepOnlyParty teaches the sample party members only the step. A member who can also fight is still
// ready while its legs recover, which tests that walk don't want.
func StepOnlyParty(bundle *content.Bundle) {
	for _, id := range []string{"template.person.reaper", "template.person.salken"} {
		Edit(bundle.Templates, id, func(person *content.Template) {
			person.Actions = []api.ContentID{"action.step"}
		})
	}
}

// NoStart removes the start scenario, for tests that spawn their own party.
func NoStart(bundle *content.Bundle) { bundle.Start = content.Start{} }
