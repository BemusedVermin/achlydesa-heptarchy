---
id: M1-01
title: Tooling, lint guards and package stubs
status: todo
milestone: M1
kind: seam
depends_on: []
discovered_by: null
size: S
---

# M1-01 · Tooling, lint guards and package stubs

## Goal

The architecture rules in `AGENTS.md` are enforced by the linter from the first commit, one Taskfile holds every development command, and CI runs the same `check` task the owner and agents run. The prototype moves out of `internal/core` so the core starts clean.

## Context

- `docs/go-technical-design.md`: "Repository layout" (the Two layers bullet), "Simulation loop and determinism → Determinism rules", "Harness and testing → Tests" (Architecture row).
- `.golangci.yml` already has the `core-is-pure` depguard rule (golangci-lint v2 format).
- `internal/core/game/game.go` and `cmd/heptarchy/main.go` are the prototype hex walk.

## Deliverables

- `internal/prototype/game.go`: the prototype moved from `internal/core/game/game.go`, now `package prototype`. Delete the old file.
- `cmd/heptarchy/main.go`: import path updated to `internal/prototype`. No other changes.
- `.golangci.yml`: modified as below.
- `Taskfile.yml`: new.
- `go.mod`: Task as a Go tool (`go get -tool github.com/go-task/task/v3/cmd/task@latest`), so `go tool task` works with nothing else installed.
- `.github/workflows/ci.yml`: new.
- `doc.go` with only a package comment in each of: `internal/core/api`, `internal/core/sim`, `internal/core/comp`, `internal/core/clock`, `internal/core/turns`, `internal/core/rng`, `internal/core/hexmap`, `internal/core/knowledge`, `internal/core/effects`, `internal/core/explore`, `internal/core/game`, `internal/content`. Each comment says what the package owns and its layer (foundation, feature or top), using the package table in the tech design.

## Spec

**`.golangci.yml`** (golangci-lint v2):

1. `linters.default: standard`, plus `depguard` and `forbidigo` enabled.
2. Keep `core-is-pure`. Add `github.com/charmbracelet` to its deny list.
3. New depguard rule `features-stay-independent`:
   - `files`: `**/internal/core/explore/**`, `**/internal/core/procgen/**`, `**/internal/core/combat/**`, `**/internal/core/agents/**`, `**/internal/core/seats/**`, `**/internal/core/cast/**`, `**/internal/core/rumor/**`, `**/internal/core/investigation/**`, and `!$test`.
   - `deny`: each of those eight packages by full import path (`$MOD/internal/core/explore`, ...), with the description `feature packages import only foundation packages (AGENTS.md rule 2)`.
4. New depguard rule `randomness-through-rng`: `files` is `**/internal/core/**` and `!**/internal/core/rng/**`; it denies `math/rand` and `math/rand/v2` with the description `use internal/core/rng named streams (AGENTS.md rule 3)`.
5. forbidigo forbids `time.Now` and `time.Since` in `internal/core/**` only. Use `analyze-types: true` so it matches the package function, and scope it with a v2 `linters.exclusions.rules` entry (`path-except`) so code outside the core may use them.

Check every rule with a deliberate violation (see Acceptance). If a pattern doesn't behave as written here, adjust the config until the violations fail, and note the change.

**`Taskfile.yml`** (Task v3 format; docs at https://taskfile.dev). Every task has a `desc`, so `go tool task --list` documents the project's commands. Task runs commands through its own built-in shell, so write them once and they behave the same in PowerShell on Windows and on Linux. Don't add PowerShell- or bash-only syntax.

| Task | Runs |
| --- | --- |
| `fmt` | `gofmt -w .` |
| `fmt-check` | Fails if `gofmt -l .` prints anything |
| `vet` | `go vet ./...` |
| `test` | `go test ./...`, passing extra arguments through with `{{.CLI_ARGS}}`, so `go tool task test -- -run Foo ./internal/core/sim` works |
| `lint` | `golangci-lint run` |
| `content` | A placeholder that prints "no content yet"; M1-06 makes it run `go tool cue vet ./content/...` |
| `check` | `fmt-check`, `vet`, `test`, `lint` and `content`, in that order, stopping at the first failure |
| `run` | `go run ./cmd/heptarchy {{.CLI_ARGS}}` |

Run the steps of `check` in sequence (`cmds` calling each task), not as parallel `deps`, so failures are readable. golangci-lint stays a separately installed binary, as its maintainers recommend; `lint` fails with a clear message if it isn't on the path.

**`.github/workflows/ci.yml`**: on `push` and `pull_request`, runs on `ubuntu-latest`. Steps:

1. `actions/checkout`.
2. `actions/setup-go` with `go-version-file: go.mod`.
3. Install golangci-lint at a version that supports the v2 config, with the official action in install-only mode or its install script.
4. `go tool task check`.

Pin each action to a major version.

## Acceptance criteria

- [ ] `go build ./...` passes and `go run ./cmd/heptarchy` still walks the prototype map.
- [ ] `golangci-lint run` passes on the clean tree.
- [ ] Layers: temporarily create `internal/core/combat/doc.go` and a file in `internal/core/explore` that imports it. Lint fails with the layers message. Remove both.
- [ ] Randomness: a temporary import of `math/rand/v2` in `internal/core/sim` fails lint; the same import in `internal/core/rng` passes. Remove both.
- [ ] Time: a temporary `time.Now()` call in `internal/core/sim` fails lint; the same call in `cmd/heptarchy` passes. Remove both.
- [ ] Purity: a temporary Bubble Tea import in `internal/core/sim` fails lint. Remove it.
- [ ] `go tool task check` exits 0 on the clean tree and non-zero when a Go file is badly formatted.
- [ ] `go tool task --list` shows every task with its description.
- [ ] `go tool task check` works the same from PowerShell on Windows and from bash on Linux (note both runs in Notes).
- [ ] Notes record the exact output line of each deliberate violation.

## Out of scope

- Any code in the new packages beyond `doc.go`. Each later task fills its own package.
- Pre-commit hooks.
- CUE setup and the real `content` task (M1-06).

## Notes
