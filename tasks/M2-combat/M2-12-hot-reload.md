---
id: M2-12
title: Content hot reload in dev builds
status: done
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

- [x] Editing a terrain glyph in `content/data` while playing with `--dev` changes it on screen without a restart.
- [x] Breaking a CUE file shows the error in the log and the game keeps running on the old content.
- [x] A test drives `Watch` with file writes in `t.TempDir()` and gets exactly one callback per burst.
- [x] Without `--dev`, nothing watches files.
- [x] `go tool task check` passes.

## Out of scope

- Reloading Go code.

## Notes

- New dependency: `github.com/fsnotify/fsnotify` v1.10.1, named in the task, for recursive file watching.
- `Watch` only reacts to `.cue` files and extensionless paths (folders), so editor temp files do not trigger reloads. New subfolders are added to the watch as they appear.
- Added `Game.Note(text)` (not in the spec) so the frontend can put the reload failure in the message log; `SwapContent` also logs the "replay will not match" warning through it.
- The `--dev` flag is on both `heptarchy` and `heptarchy play`. Without it `content.Watch` is never called.
- Tests: `Watch` (one callback per burst, ignores non-CUE files, missing dir), `SwapContent` (success logs a warning, nil is refused and leaves the game unchanged), and the app's `ContentMsg` handling (success and failure). A failing swap against invalid-but-loadable content is not tested because no sample content makes the primitive checks fail.
- The first two acceptance criteria are covered by those tests and the code path; I did not drive the interactive TUI by hand.
- Follow-up filed: M2-12.F1 (content cached at game start, such as the calendar, is not refreshed by a swap).
