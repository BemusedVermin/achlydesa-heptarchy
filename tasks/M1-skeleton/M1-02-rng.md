---
id: M1-02
title: "rng: named deterministic streams"
status: review
milestone: M1
kind: feature
depends_on: [M1-01]
discovered_by: null
size: S
---

# M1-02 · rng: named deterministic streams

## Goal

Every subsystem draws from its own random stream derived from the world seed, so adding randomness in one subsystem never shifts another, and every stream's state can be saved and restored.

## Context

- Tech design: "Simulation loop and determinism → Determinism rules" (Named streams), the Randomness row of "Library choices", and the `rng` row of "Core packages".
- `math/rand/v2`: `rand.NewPCG(seed1, seed2)` and `rand.New(src)`. `*rand.PCG` implements `MarshalBinary` and `UnmarshalBinary`; confirm with `go doc math/rand/v2 PCG`.

## Deliverables

- `internal/core/rng/rng.go`
- `internal/core/rng/rng_test.go`

## Spec

Firm signatures:

```go
// Streams hands out one PCG stream per name, all derived from the world seed.
type Streams struct{ /* unexported */ }

func New(worldSeed uint64) *Streams
func (s *Streams) Seed() uint64

// Get returns the named stream, creating it on first use. The sequence depends
// only on the world seed and the name, never on the order streams were created in.
func (s *Streams) Get(name string) *rand.Rand

// Derived returns a fresh generator for (name, salt) that is not registered or
// saved: callers recreate it from the same inputs. Procgen uses ("procgen", exportCount).
func (s *Streams) Derived(name string, salt uint64) *rand.Rand

// State and Restore cover every stream created with Get, for saves.
func (s *Streams) State() (map[string][]byte, error)
func (s *Streams) Restore(state map[string][]byte) error
```

Seeding, exactly:

- `Get(name)` uses `rand.NewPCG(worldSeed, fnv64a(name))`.
- `Derived(name, salt)` uses `rand.NewPCG(splitmix64(worldSeed ^ salt), fnv64a(name))`.
- `fnv64a` is FNV-1a 64 from `hash/fnv`. `splitmix64` is the standard finalizer: add `0x9E3779B97F4A7C15`, then xor-shift-multiply by `0xBF58476D1CE4E5B9` and `0x94D049BB133111EB` with shifts 30, 27 and 31.
- Never use `hash/maphash`: its seed changes every process.

Stream names are declared as constants in the package that owns them (`combat` declares `const rngStream = "combat"`), lowercase, starting with the package name. Document that convention in `doc.go`.

## Acceptance criteria

- [x] The same seed and name give identical first 1,000 values across two `Streams` values.
- [x] Creating "a" then "b" gives the same sequences as creating "b" then "a".
- [x] Different names give different sequences.
- [x] Draw some values, `State()`, keep drawing; a new `Streams` with `Restore()` produces the same continuation.
- [x] `Derived` with the same inputs reproduces; a different salt differs.
- [x] A `rapid` property over random seeds and names checks the first two criteria.
- [x] `go tool task check` passes.

## Out of scope

- Writing stream state into save files (M1-12).

## Notes

- Added `pgregory.net/rapid` to `go.mod`: the criteria require a property test, and the tech design names it as the property-testing library.
- `Restore` replaces the whole stream set: streams absent from the saved state restart from their seed on next `Get`.
- Files touched outside the list: `internal/core/rng/doc.go` (stream-name convention), `go.mod`, `go.sum`.
- No follow-ups filed.
