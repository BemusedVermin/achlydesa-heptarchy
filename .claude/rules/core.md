---
paths:
  - "internal/core/**"
---

# Working in internal/core

Every rule in AGENTS.md applies here. This is how they play out in the code. Names come from `docs/glossary.md`.

## Determinism (rule 3)

- Draw randomness from `rng.Streams.Get("<subsystem>")`: one stream per subsystem, so that a new draw in one system never shifts another system's sequence. Procgen uses `Derived("procgen", exportCount)`. depguard rejects `math/rand` everywhere else in the core.
- Before you act on a map's contents, collect its keys and sort them by `comp.UID` or content ID. Results from goroutines get the same treatment.
- At a single tick, landings go first, then decisions in UID order. A new system takes its place in that order through the scheduler's cadence and run-after, never through the order of calls.
- Work on another goroutine only computes and returns. Only the world's goroutine applies the results, in a fixed order.
- Time is `World.Now`, in ticks. forbidigo rejects `time.Now` and `time.Since` in the core.

## Components and content (rules 4 and 5)

- Components in `comp` are plain structs holding content IDs and UIDs. They never hold pointers to registry records or to other entities: look the record up in the content registry when you need it.
- Code names no content record, tag or rule. When code needs a particular record, a config key in content says which one.
- Numbers, names and formulas a designer would tune go in `content/data`. Compile formulas only through `content.Registry.Compile`.
- Changing a struct in `api` or `comp` changes its CUE schema too: run `go tool task schema` and commit the regenerated files with the change.

## The boundary (rules 6 and 7)

- Commands, events and views in `api` are plain, serializable data with no behavior.
- Player-facing text leaves the core as an `api.Text` (a text ID plus arguments), or in the view as a string the core rendered from one. Never write English for the player in Go.
- An invalid player command emits `api.Rejected` with a text reason. It isn't an error.
- A content bug met at run time, such as a failing speed formula, returns an error up the call chain (`timeline.SetSpeed` hooks return `(int, error)`). Never turn it into 0 or a default.
- Valid moves, menus and their order are worked out here and served in the view's `Options`.

## Layers (rule 2)

Foundation packages include `worldgen`. Feature packages (`explore`, `procgen`, `combat`, `agents`, `seats`, `cast`, `rumor`, `investigation`) import only foundation packages. To reach another feature, publish an event or write a component it reads. depguard enforces this.

## Tests

The same seed and the same commands give the same `Game.Hash`. When you touch anything whose order matters, add a test that shows it: a replay or a rapid property.
