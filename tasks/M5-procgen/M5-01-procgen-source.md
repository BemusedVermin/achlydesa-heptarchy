---
id: M5-01
title: Geography content and the procgen chunk source
status: done
milestone: M5
kind: seam
depends_on: [M4-07]
discovered_by: null
size: M
---

# M5-01 · Geography content and the procgen chunk source

## Goal

A `procgen.Source` builds any chunk of the world on demand from the world seed, the export count and the chunk key, running the named stages of the world's geography. Given the same three inputs it returns the same chunk, whatever order chunks are visited in. With no stages yet, it serves the fixed places and fills the rest with the world's fill.

## Context

- `docs/game-design.md` → "Procgen and exports" (Fixed, Generated, Reseeding).
- `docs/go-technical-design.md` → "Designing for pivots" → "The world" and "Procgen stages"; "Core packages" → `procgen`; "Determinism rules".
- What exists: `hexmap.Source` and `FuncSource` (`internal/core/hexmap/chunk.go`), `explore.WorldSource` (the fixed places set in the world's fill), `rng.Streams.Derived`, whose doc comment already names `("procgen", exportCount)`.
- Layers: `procgen` is a feature package, so it may not import `explore`. The fixed places reach it through an interface declared in `procgen` and implemented by `explore.WorldSource`.

## Deliverables

- `internal/content/types.go`: `GeographyDef` and `Bundle.Geographies`; `WorldDef.Geography`; regenerate the schemas. `content/data/geography.cue` with one geography (`geography.deep`) listing no stages yet, and `world.cue` naming it.
- `internal/content/geography.go`: checks (below), `Registry.Geography`.
- `internal/core/procgen/`: `doc.go`, `source.go` (`Source`, `Config`), `stage.go` (`Stage`, the stage registry). The registry hangs off a value, not a package variable.
- `internal/core/explore/world_source.go`: `WorldSource` implements the fixed-layer interface (below).
- One test: the same chunk key from two sources built in different visiting orders is identical; a different export count or world seed changes a chunk that has a stage that uses randomness (a test-only stage registered in the test).
- `docs/content-guide.md`: `geography.cue` and the `geography.` prefix.

## Spec

Suggested shapes (names firm where marked):

```go
// GeographyDef lists the stages that build chunks of one kind of ground, in order.
type GeographyDef struct {
	Stages []StageDef
}

// StageDef names a registered stage and its numbers, which are content.
type StageDef struct {
	// Stage is the registered stage's name, such as "caves".
	Stage  string
	Params map[string]float64
}

// Fixed is what procgen needs of the hand-made layer. (firm: declared in procgen)
type Fixed interface {
	hexmap.Source
	// Covers reports whether a cell belongs to a place.
	Covers(api.Cell) bool
}

// Stage changes one chunk. It reads only its Context and its own random stream, so a chunk's result depends
// on nothing else. (firm)
type Stage interface {
	Run(ctx *Context, chunk *hexmap.Chunk) error
}
```

- `Context` carries: the chunk key, the world bounds, an `*rng.Generator` derived from `(world seed, export count, chunk key)`, the stage's params, and read-only access to `Fixed`. Suggestion: derive the generator by mixing the key into `rng.Streams.Derived("procgen", exportCount)`'s salt; it never touches a registered stream, so world play cannot change generation.
- **Order independence is the property that matters.** A stage that needs a neighbor chunk's contents must compute them from the seed (as noise does), never read them from a loaded chunk.
- **Fixed cells win.** After the stages run, cells `Fixed.Covers` are copied over from the fixed layer, so a stage cannot damage a place.
- **Checks:** every stage name is registered; a geography's stages and params are valid for that stage (each stage validates its own params through the registry).
- Cells outside the world's bounds are never generated.

## Acceptance criteria

- [x] `go tool task test -- ./internal/core/procgen ./internal/core/explore ./internal/content` passes.
- [x] `play --seed 1` starts and the plan view looks as it did (no stages yet).
- [x] `procgen` imports no feature package (`go tool task lint`).
- [x] `go tool task check` passes.

## Out of scope

- Any terrain-shaping stage: M5-02 (caves), M5-03 (sinks), M5-05 (joining).
- Using the source in the real game's world: M5-06. This task may wire it in the test and `contenttest` worlds only.
- Regions: M5-04. Populating: M5-11.

## Notes

- **Where the stage checks live.** Content only checks that a stage is named and that `world.geography` exists; whether a stage is registered and takes its params is checked by procgen (`Stages.Validate` for all of content, and `NewSource` for its own geography), as effect primitives are checked by `effects`. Nothing calls `Stages.Validate` yet: M5-06 wires the source into the game.
- **`world.geography` is optional** (empty means every unplaced cell is the fill), so hand-built test worlds need no geography. `geographies` is optional in the bundle for the same reason. The geography collection gets an ID prefix check but no `Collection` constant, so choices cannot list it.
- **`StageDef`** sits in `internal/content/types.go` beside `GeographyDef`; `content.GeographyPath` formats a stage's path for errors on both sides.
- **Randomness.** Each stage of each chunk draws from `Derived("procgen", salt)`, the salt an FNV hash of the export count, the chunk key, the stage's index and its name. `Source` keeps its own `rng.New(seed)` and never touches a registered stream.
- **Fixed cells win, and so do cells outside the bounds:** after the stages run, every such cell is reset to the fixed layer's chunk. A chunk wholly outside the bounds skips the stages.
- **Files outside the Deliverables:** `internal/core/rng/rng.go` gains `Generator`, the type `Derived` returns, with only `Uint64`, `IntN` and `Float64`, so procgen draws randomness without importing `math/rand` or reaching the standard generator. `internal/content/records.go` and `load.go` register the collection and its check.
- The test builds the source over `explore.WorldSource` from the sample content (tests may import across features) and checks order independence, fixed and out-of-bounds cells, export count and seed, and stage rejection.
- **Follow-up filed:** M5-01.F1, so that only `rng` imports `math/rand`; `Streams.Get` and `internal/harness` still use it.
