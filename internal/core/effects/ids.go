package effects

import (
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// KnownIDs checks that each parameter named in known that raw gives is a content ID its lookup
// finds. The error names the parameter as the kind of record ("unknown item ..."), so name
// parameters after what they refer to. A primitive calls it from VerifyContent at load and again in Apply on the
// values it read, since a text formula's result is only known then.
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

// Known adapts a content registry lookup, such as reg.Item, for [KnownIDs].
func Known[T any](lookup func(api.ContentID) (T, bool)) func(api.ContentID) bool {
	return func(id api.ContentID) bool {
		_, ok := lookup(id)
		return ok
	}
}
