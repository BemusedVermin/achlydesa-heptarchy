// Package combat holds the effect primitives fights are made of: damage and severing, statuses,
// grafting, revealing materials, using up items, and turning a severed part into a body. Each is
// registered with the effects package by Install and works the same with nothing hostile around.
// It also holds what makes a type change a form: the "type" and "piece" choices, the checks that a
// type can be used on a target, and the modifier that adds the type's timing, landing effects and
// item cost to whatever form it is loaded into. The forms themselves are content. Triggers and
// death build on this in later tasks. Feature layer: it imports only foundation packages.
package combat
