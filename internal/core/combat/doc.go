// Package combat holds the effect primitives fights are made of: damage and severing, statuses,
// grafting, revealing materials, using up items, and turning a severed part into a body. Each is
// registered with the effects package by Install and works the same with nothing hostile around.
// Forms, triggers and death build on them in later tasks. Feature layer: it imports only
// foundation packages.
package combat
