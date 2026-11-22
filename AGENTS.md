# Achlydesa: The Dreaming Heptarchy

A turn-based dungeon crawler in Go: a pure engine core, content as CUE data, and a Bubble Tea frontend (Ebitengine later).

Read before touching code:

- `docs/game-design.md`: what the game is.
- `docs/go-technical-design.md`: how the code is laid out. Read **Designing for pivots** first.
- `tasks/README.md`: how work is tracked. `tasks/INDEX.md`: what to do next.

## Commands

Every development command is a task in `Taskfile.yml`, run through Go so nothing extra needs installing (only golangci-lint is a separate binary):

- Everything before you finish: `go tool task check` (format, vet, tests, lint, content)
- One part: `go tool task test`, `go tool task lint`, `go tool task content`
- Pass arguments after `--`: `go tool task test -- -run TestName ./internal/core/sim`
- Regenerate CUE schemas after changing content structs: `go tool task schema`
- Play: `go tool task run -- play --seed 1`
- List them all: `go tool task --list`

Until M1-01 adds the Taskfile, use `go build ./...`, `go test ./...` and `golangci-lint run` directly.

Module path: `github.com/BemusedVermin/achlydesa-heptarchy`. The owner works in PowerShell on Windows. Put any repeatable command in `Taskfile.yml` rather than a shell script; Task's built-in shell runs it the same on Windows and Linux. Commands you write for him to type must work in PowerShell.

## Rules that never bend

1. **The core is pure.** `internal/core` imports no UI package (Bubble Tea, Bubbles, Lip Gloss, Ebitengine).
2. **Layers hold.** Foundation packages (`api`, `sim`, `comp`, `anatomy`, `rng`, `clock`, `turns`, `hexmap`, `knowledge`, `inventory`, `effects`) may be imported anywhere in the core. Feature packages (`explore`, `procgen`, `combat`, `agents`, `seats`, `cast`, `rumor`, `investigation`) import only foundation packages, never each other; they talk through components and events. `game` and `save` sit on top.
3. **Determinism.** Randomness comes only from `internal/core/rng` named streams. Never call `time.Now` in the core. Never let map iteration order or thread timing decide an outcome: sort by `comp.UID` or content ID first.
   - **One writer.** Only one goroutine at a time changes the world: the ECS, `World.Now` and the RNG streams.
   - **Parallel, then in order.** Work may run on other goroutines only as a pure computation: it reads its inputs and returns a result without touching the world. The world's goroutine then applies the results in a fixed order (by chunk key, region ID or seed). Procgen chunks, harness games run side by side, and off-screen region updates all fit this pattern.
   - **Safe registries aren't a safe world.** The `sim` registries (bus, scheduler, commands, views, UID counter) are safe for concurrent use, but that doesn't make concurrent changes to the world deterministic.
4. **Content is data.** Anything a designer would tune (numbers, names, text, glyphs, formulas) lives in `content/data` as CUE, not in Go. Player-facing strings are content text IDs.
5. **IDs, not pointers.** Components hold content IDs (strings like `material.clay`) and UIDs, never pointers into the content registry or into other entities.
6. **The boundary is plain data.** Commands, events and view types live in `internal/core/api` as serializable structs with no behavior.
7. **Choices come from the core.** Frontends render the view and send commands. Valid moves, menus and their order come from the core's view.
8. **Reuse first.** Use the libraries named in the tech design. A new dependency needs a sentence in the task's Notes saying why.
9. **Don't decide the game design.** If a task needs a design decision the docs don't make, file a question (see below) and pick the most reversible option meanwhile.

## Working on a task

The full workflow is in `tasks/README.md`. In short:

1. Pick a `todo` task in `tasks/INDEX.md` whose dependencies are all `done`. Set it to `in-progress` in both the task file and the index.
2. Read the task file, then the doc sections it links.
3. Implement only what the task's Deliverables and Spec cover. Signatures in a spec are the intended shape; keep the names unless they clash with something real, and say so in Notes if you change one.
4. Verify against every acceptance criterion and run the checks. Tick the criteria in the task file.
5. Fill in the task's Notes: assumptions, deviations, files touched outside the list, follow-ups filed.
6. Set the status to `done` in the file and the index, then open a PR. Merging the PR completes the task.

## Follow-up tasks: required

Whenever you notice work outside your task's scope (a bug, a missing seam, debt, a missing test, an unclear design point), don't fix it silently and don't leave a bare TODO. File it:

1. Copy `tasks/TEMPLATE.md` to `tasks/followups/<your task ID>.F<n>-<short-slug>.md` (for example `M1-03.F1-bus-cascade-limit.md`), numbering from F1 within your task.
2. Fill in every section, with `status: proposed`, `discovered_by: <your task ID>` and the `kind`.
3. Add a row to the Follow-ups table in `tasks/INDEX.md`.
4. List it in your task's Notes.
5. In code, a `// TODO(<follow-up ID>): ...` comment is fine. A TODO without a filed task is not.

If the follow-up blocks your task, set your task to `blocked`, say why in Notes, and stop. Never start a `proposed` task: the owner triages them into milestones first.

## Style

- gofmt. A `doc.go` with a package comment in every package. Every exported identifier has a doc comment.
- Descriptive names. Single letters only for loop indices and conventional uses such as `t *testing.T`.
- Table-driven tests; `pgregory.net/rapid` for properties; golden files under `testdata/`.
- Return errors for programmer and content bugs (wrap with `%w`). An invalid player command is not an error: emit an `api.Rejected` event.
- No global mutable state. Registries hang off a world or a content registry value.
