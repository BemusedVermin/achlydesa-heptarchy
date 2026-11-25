package contenttest

import (
	"encoding/json"
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
// value tests may hold.
var sample = sync.OnceValues(func() (*content.Registry, error) {
	bundle, err := content.LoadBundle(Dir())
	if err != nil {
		return nil, err
	}
	pinStart(&bundle)
	return content.New(bundle)
})

// testStartMap is the map the sample content starts on in tests; the shipped start is the playground.
const testStartMap = "map.test.hall"

func pinStart(bundle *content.Bundle) {
	if bundle.Start.Map != "" {
		bundle.Start.Map = testStartMap
	}
}

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

var sampleBundle = sync.OnceValues(func() ([]byte, error) {
	bundle, err := content.LoadBundle(Dir())
	if err != nil {
		return nil, err
	}
	pinStart(&bundle)
	return json.Marshal(bundle)
})

// SampleBundle returns a copy of the sample content that the test may change. The CUE is parsed once
// per binary; each call decodes its own copy.
func SampleBundle(t T) content.Bundle {
	t.Helper()
	encoded, err := sampleBundle()
	if err != nil {
		t.Fatal(err)
	}
	var bundle content.Bundle
	if err := json.Unmarshal(encoded, &bundle); err != nil {
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
func NoStart(bundle *content.Bundle) { bundle.Start, bundle.Starts = content.Start{}, nil }

// Do builds an effect reference from the primitive's name and its parameters as name, value pairs.
func Do(name string, params ...string) content.EffectRef {
	ref := content.EffectRef{Do: name, Params: map[string]string{}}
	for i := 0; i+1 < len(params); i += 2 {
		ref.Params[params[i]] = params[i+1]
	}
	return ref
}
