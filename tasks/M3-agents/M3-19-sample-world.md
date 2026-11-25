---
id: M3-19
title: "The sample world: two towns and the deep between them"
status: done
milestone: M3
kind: content
depends_on: [M3-10, M3-11, M3-14, M3-18]
discovered_by: null
size: M
---

# M3-19 · The sample world: two towns and the deep between them

## Goal

A small world that exercises every M3 system together: two towns with residents, stocks and recruits, two regions of the deep between them with the four families, all joined by gates. The harness plays it for days.

## Context

- `docs/game-design.md` → "World and towns" (towns at depths, materials near where they belong), "The living deep".
- The agents model in [README.md](README.md).
- `docs/game-design.md` → "Technical plan" → "MVP definition" (one crawlspace, one region of the deep; M3 needs two towns for migration).

## Deliverables

- `content/data/maps/`: two town maps (a shallow one with a well, a lake one with water to spare) and two deep region maps, joined by gates; regions with `decay` and starting populations.
- `content/data/start.cue` stays on the playground; a second start (or a harness option) names the sample world, so the harness can run it.
- `docs/content-guide.md`: what the sample world holds.
- A smoke test: 10 days on the sample world with no command errors.

## Spec

- Firm: placeholder numbers, marked for M7 to tune.
- Firm: the two towns are generic test towns, not the design's named crawlspaces, so no story content is decided here.
- Suggestion: if `content.Start` must become a list to keep the playground as the default, do it here and say so in Notes.

## Acceptance criteria

- [x] Exit criterion: running the sample world for 10 days shows populations drifting between regions, and entering a deep region promotes its families into full agents (a test checks the `Promoted` counts).
- [x] The harness runs the sample world for 10 days on seed 1 without errors.
- [x] `go tool task check` passes.

## Out of scope

- Harness columns and the camping check (M3-20).

## Notes

- `content.Start` stays as is; a new `starts` map (`start.sample`) holds named scenarios, and `Registry.WithStart` and the harness's `run --start` pick one. Suggestion followed in spirit, with no list.
- Files outside the list: `internal/content` (types, maps, regenerated schema), `cmd/harness/main.go`.
- The chain shallow, deep_a, deep_b, lake means the towns are not neighbours, so town migration is not exercised; the exit criterion's drift is of monster families (remnants flow toward lower decay). Owner may want a direct gate between towns.
- The smoke test lives in `internal/core/agents/sample_world_test.go` (waits 40 phases, fails on any rejection).


- The four new regions and a well shift entity UIDs, so the combat and tui golden files were regenerated (UIDs only). Tests that counted all remnants or all wells now look at playground maps only; `contenttest.NoStart` clears `Starts` too.
- golangci-lint could not run in my environment (go.mod targets Go 1.27.4, above any available build); everything else in `check` passed. CI will lint.
