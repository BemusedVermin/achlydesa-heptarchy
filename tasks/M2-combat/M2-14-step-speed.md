---
id: M2-14
title: "Make a game step fast again after the seven forms"
status: todo
milestone: M2
kind: bug
depends_on: [M2-06]
discovered_by: null
size: M
---

# M2-14 · Make a game step fast again after the seven forms

## Goal

A game step (build the view, pick an option, apply it) costs about what it did before M2-06, without changing a single outcome. Every test that plays games, and every agent and CI run, gets back to its old speed.

## Context

- M2-06 (#44) changed only content: the seven forms in `content/data/forms.cue`, innate types read as a choice, the `piece` choice, and more channels per body. No Go outside tests changed, yet every package that plays games slowed down. CI times before (d000852) and after (5d50cfd):

  | Package | Before | After |
  | --- | --- | --- |
  | `internal/harness` | 5.3s | 72.3s |
  | `cmd/harness` | 0.3s | 6.5s |
  | `internal/core/combat` | 0.3s | 8.0s |
  | `internal/core/save` | 10.7s | 21.5s |
  | `internal/core/explore` | 7.3s | 12.7s |

- The suspect, not yet measured: the timeline's `usable` hook in `internal/core/actions/options.go`. For each channel the timeline asks about (`internal/core/timeline/timeline.go`), it makes a fresh evaluator and builds the body's full candidate list. That list is every known action × every combination of its choices × every target × every channel set. It then evaluates rules on the candidates until one passes. M2-06 multiplied every one of those factors.
- `Startable` and the view's option list build the same candidates again.
- `docs/go-technical-design.md`: the determinism rules and how the timeline and actions fit together.
- M2-08 also changes `internal/core/actions`, so don't run the two side by side.

## Deliverables

- A benchmark of one harness step on the sample content: view, a random choice, apply. Suggested place: `internal/harness/bench_test.go`.
- The fix, wherever the profile points. Likely `internal/core/actions/options.go`, and maybe `internal/core/timeline/timeline.go`.

## Spec

1. **Measure first.** Run the benchmark with a CPU profile on master: `go test -run '^$' -bench . -cpuprofile cpu.out ./internal/harness`, then `go tool pprof -top cpu.out`. Name the top costs in Notes before changing code. If the suspect above is wrong, follow the profile.
2. **Fix what the profile shows.** Ideas, all suggestions:
   - Answer `usable` for all of a body's channels in one pass, instead of rebuilding the candidates once per channel.
   - Build a body's candidates once per decision and reuse them for the usable check, `Startable` and the view.
   - Run cheap tests (channel free, target in reach) before evaluating rules.
3. **Nothing observable changes (firm).** Same options, in the same order, with the same option IDs. Same RNG draws, same events, same game hash.
4. **Determinism rules hold (firm).** Any cache lives on the world (a resource) or inside one call, never in a global. It never outlives a change it depends on, and it is rebuilt after a load and after a content hot reload. Only the world's goroutine writes it.
5. **Don't touch the content.** Don't fix the speed by removing forms, choices or channels from `content/data`.

## Acceptance criteria

- [ ] Notes give the benchmark's time per step and the profile's top costs, before and after.
- [ ] `go run ./cmd/harness run --seeds 3 --steps 200 --out after.csv` writes a file byte-identical to the one master writes (`--out before.csv`).
- [ ] `go test -count=1 ./internal/harness` runs at least five times faster than on master. Notes give both times.
- [ ] `go tool task check` passes.

## Out of scope

- Shrinking the harness tests' seeds and steps. If they are still slow after the fix, file a follow-up.
- Caching plan-view reachability (M1-10.F1.F1.F1). If the profile puts it among the top costs, say so in Notes and leave it to that follow-up.
- Changing how many options the view offers (`MaxOptions`) or what they are.

## Notes

Filled in by the implementer: assumptions, deviations, files touched outside Deliverables, follow-ups filed.
