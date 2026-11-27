---
id: M5-15
title: Record as you play, and replay with a readable trace
status: todo
milestone: M5
kind: feature
depends_on: [M5-13]
discovered_by: null
size: S
---

# M5-15 · Record as you play, and replay with a readable trace

## Goal

A player who hits a bug can hand over a seed and a command log, and anyone can replay it and read, command by command, what happened. Today the log is only written on a clean exit, and a replay prints only a hash.

## Context

- Filed by the owner on 2026-11-26, while asking how to capture logs to recreate bugs.
- `heptarchy play --record FILE` (`cmd/heptarchy/main.go`) writes `Game.History` through `tui.WriteReplay` only after `program.Run` returns. A crash or a killed terminal leaves nothing.
- `harness replay FILE --seed N` (`cmd/harness/main.go`, `internal/harness/replay.go`) replays the log and prints the final hash. `--expect` checks it.
- The message log (`api.View.Log`), with M5-13's fight lines, already says what happened in words.

## Deliverables

- `cmd/heptarchy/main.go` and `internal/frontend/tui/app.go`: rewrite the record file after every applied command, not only at exit.
- `internal/harness/replay.go` and `cmd/harness/main.go`: a `--trace` flag that prints, for each command in order, its index, its kind and the log lines it added, then the hash.
- `docs/go-technical-design.md`: the harness section names `--trace` and says the record is kept as you play.

## Spec

- **Recording (suggestion).** The app gets an optional record hook that runs after each successful `Apply`. `play` passes one that rewrites the file whole through `WriteReplay`. Logs are small, so rewriting is simpler than appending, and a half-written line can't happen.
- **Trace.** `harness replay --trace` applies the commands one at a time through `Game.Apply`, not `game.Replay`. After each one, it prints the new lines of `View.Log` as `#<index> <kind>` followed by the indented lines with their ticks. It prints the same final hash as without `--trace`.

## Acceptance criteria

- [ ] With `--record`, the file holds every command applied so far while the game is still running (a test of the record hook).
- [ ] `harness replay --trace` prints each command's index, kind and log lines, and the same hash as without `--trace` (a golden file under `testdata/`).
- [ ] `go tool task check` passes.

## Out of scope

- Stepping through a replay interactively in the TUI: a later idea.
- Recording content changes from `--dev` reloads, which a replay already warns about.

## Notes

Filled in by the implementer: assumptions, deviations, files touched outside Deliverables, follow-ups filed.
