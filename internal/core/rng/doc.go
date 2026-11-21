// Package rng owns the named deterministic random streams, one per subsystem. Foundation layer.
//
// Stream names are declared as constants in the package that owns them (for
// example combat declares const rngStream = "combat"). They are lowercase and
// start with the package name, so two packages never share a stream by accident.
package rng
