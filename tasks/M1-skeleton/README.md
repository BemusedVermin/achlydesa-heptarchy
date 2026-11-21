# M1 · Skeleton

**Goal:** the engine spine with every seam M1 touches from "Designing for pivots", and the prototype hex walk rebuilt on top of it: a map loaded from CUE, a party entity, a clock, deterministic replays and saves.

**Exit criteria:**

- `go run ./cmd/heptarchy play --seed 1` walks a hand-made 3D map loaded from `content/data`: down a slope, off a ledge, past a soop crossing underneath, all in one seamless plan view with the clock and a message log on screen. The party moves as a leader with a follower trailing single file, one keypress per turn, on the same timeline every other body will use.
- A recorded command log replays to the same hash in `go run ./cmd/harness replay`.
- Save, load and continue gives the same hash as never saving.
- Every lint guard is on: core purity, layers, randomness only through `rng`, no `time.Now` in the core.

## Tasks

| ID | Title | Depends on |
| --- | --- | --- |
| [M1-01](M1-01-tooling.md) | Tooling, lint guards and package stubs | none |
| [M1-02](M1-02-rng.md) | rng: named deterministic streams | M1-01 |
| [M1-03](M1-03-api.md) | api: the plain-data boundary and its codec | M1-01 |
| [M1-04](M1-04-sim.md) | sim: world, event bus, scheduler, command registry | M1-02, M1-03 |
| [M1-05](M1-05-comp-hexmap.md) | comp basics and hexmap: 3D cells, chunks, topology | M1-03 |
| [M1-14](M1-14-sight-pathing.md) | hexmap: sight lines and pathing in 3D | M1-05 |
| [M1-06](M1-06-content.md) | content: CUE loader, IDs, text, formulas | M1-03 |
| [M1-07](M1-07-templates-spawn.md) | Templates as component bags, and spawning | M1-04, M1-05, M1-06 |
| [M1-08](M1-08-clock.md) | clock: ticks, phases and days | M1-04, M1-06 |
| [M1-15](M1-15-turns.md) | turns: one shared timeline for every body | M1-04, M1-07, M1-08 |
| [M1-09](M1-09-game.md) | game: the public API | M1-15 |
| [M1-10](M1-10-explore.md) | explore: maps from content, movement and the seamless plan view | M1-09, M1-14 |
| [M1-16](M1-16-lead-follow.md) | Lead and follow | M1-10 |
| [M1-11](M1-11-tui.md) | Terminal frontend on the new API, with a CLI | M1-16 |
| [M1-12](M1-12-save.md) | save: versioned snapshots and migrations | M1-16 |
| [M1-13](M1-13-harness.md) | harness: seeded runs and replays | M1-12 |

M1-02, M1-03, M1-05 and M1-06 can run in parallel once M1-01 is done. M1-14, M1-15 and M1-16 are listed where they fall in the dependency order; their numbers are just the next free ones.

## Shared conventions for M1 and later

- **Module path:** `github.com/BemusedVermin/achlydesa-heptarchy`, written `$MOD` below.
- **Content IDs:** lowercase, dotted, kind first: `material.clay`, `terrain.air`, `text.explore.bump`, `template.party.main`, `map.test.hall`. Pattern: `^[a-z][a-z0-9_]*(\.[a-z0-9_]+)+$`.
- **Command and event kinds:** `<package>.<verb>` in lowercase: `explore.step`, `explore.moved`, `turns.wait`.
- **Depth** grows downward from 0 at the salk's surface, one cell per step. It never appears in the view: the core projects it away.
- **UIDs** start at 1, are never reused, and 0 means "none".
- **Ticks** are `int64`, counted from the start of the game.
