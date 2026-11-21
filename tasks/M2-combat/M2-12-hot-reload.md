---
id: M2-12
title: Content hot reload in dev builds
status: todo
milestone: M2
kind: feature
depends_on: [M1-11, M1-13]
discovered_by: null
size: S
---

# M2-12 · Content hot reload in dev builds

## Goal

While `heptarchy play --dev` runs, saving a CUE file reloads content between commands, so a fight can be tuned without restarting. Invalid content shows its errors and keeps the old content.

## Context

- Tech design: "Content pipeline" step 7, "Designing for pivots" (Identity: components hold IDs, which is what makes this safe).
- fsnotify: `github.com/fsnotify/fsnotify`.

## Deliverables

- `internal/content/watch.go`, `watch_test.go`
- `internal/core/game/game.go`: `SwapContent(*content.Registry) error`.
- `internal/frontend/tui`: a `--dev` flag wiring the watcher to the app.

## Spec

- `Watch(dir string, onChange func(*Registry, error)) (stop func(), err error)` watches `dir` recursively. It debounces bursts (150 ms), reloads, and calls back with the new registry or the error.
- `SwapContent` re-runs the checks `game.New` does against the registries (primitives, event kinds). On success it replaces `w.Content`; on failure it returns the error and keeps the old content.
- Swapping happens only between commands, on the Bubble Tea goroutine (send the result as a `tea.Msg`). Only one goroutine changes the world.
- Entities spawned from a template keep their components; new spawns use the new template. Document this in the doc comment.
- The swap clears `game.History()` reproducibility: log a warning line in the view that replays from before the swap won't match.

## Acceptance criteria

- [ ] Editing a terrain glyph in `content/data` while playing with `--dev` changes it on screen without a restart.
- [ ] Breaking a CUE file shows the error in the log and the game keeps running on the old content.
- [ ] A test drives `Watch` with file writes in `t.TempDir()` and gets exactly one callback per burst.
- [ ] Without `--dev`, nothing watches files.
- [ ] `go tool task check` passes.

## Out of scope

- Reloading Go code.

## Notes
