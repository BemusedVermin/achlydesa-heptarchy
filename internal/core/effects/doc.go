// Package effects runs the effects content names. A primitive is a named effect registered from Go;
// content refers to it with parameters that are formulas, text or selectors, and Apply and Check run
// a list of references against a Context: the acting body, the channels it uses, its target and its
// choices. Body and channel selectors pick what a primitive acts on. The package registers the
// timing primitives (lock, interrupt, delay, ready) itself, and runs content triggers: effects that
// templates, materials, statuses and items fire when an event happens to whatever carries them.
package effects
