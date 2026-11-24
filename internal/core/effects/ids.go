package effects

import (
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// KnownIDs checks that each parameter of raw named in known is an ID its lookup finds; the error
// calls the parameter's name the kind of record.
func KnownIDs(raw map[string]string, known map[string]func(api.ContentID) bool) error {
	var problems []error
	for _, name := range slices.Sorted(maps.Keys(known)) {
		value, ok := raw[name]
		if ok && !known[name](api.ContentID(value)) {
			problems = append(problems, fmt.Errorf("unknown %s %q", name, value))
		}
	}
	return errors.Join(problems...)
}

// Known adapts a registry lookup, such as reg.Item, for [KnownIDs].
func Known[T any](lookup func(api.ContentID) (T, bool)) func(api.ContentID) bool {
	return func(id api.ContentID) bool {
		_, ok := lookup(id)
		return ok
	}
}
