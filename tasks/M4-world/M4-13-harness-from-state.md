---
id: M4-13
title: The harness measures from state, within the time budget
status: todo
milestone: M4
kind: test
depends_on: [M4-11, M4-15]
discovered_by: null
size: M
---

# M4-13 · The harness measures from state, within the time budget

## Goal

The harness counts outcomes from the world's state, not from events announcing them, and a seeded year of the sample world runs within 60 s. Tests show the off-screen outcomes the milestone promises.

## Context

- `docs/game-design.md` → "Keeping it affordable" (The harness measures; the game doesn't announce).
- `docs/go-technical-design.md` → "The harness"; "Tests" (Performance: a seeded year in at most 60 s, decided on 2026-11-25).
- What exists: `internal/harness/emergence.go` and `combat.go` (the tracker counts `Migrated` and `TownEmptied`), `run.go`'s columns (`migrations`, `towns_emptied`), `testdata/water_scenario.golden`, and the camping check.

## Deliverables

- `internal/harness`: columns taken from the world each day: `deaths` (dead bodies), `people_<region>` for each region tagged `crawlspace` (living bodies standing in it), `empty_crawlspaces`, and the camping check by position (the most bodies of one faction within one region). `migrations` and `towns_emptied` go.
- The harness reads state through `game` (a view section or an inspect function); `game` gains only what it needs, read-only.
- `internal/harness/testdata/*.golden`: regenerated, and named in Notes.
- Emergence tests on `start.sample`, off screen (the party far away, idle):
  - residents fetch water and food from containers;
  - monsters roam and hunt (a death among their prey within N days);
  - with the shallow town's well and cisterns emptied, its people end up in the lake town and `people_<shallow>` reaches 0;
  - a recruit left unpaid walks off.
- A benchmark, `BenchmarkSampleYear`, and a test behind a flag (or `-short` skip) that fails when one seeded year takes over 60 s on the CI runner. If it does, tune `agents.near_chunks` and `agents.far_ticks` first, and file a follow-up for anything that needs code.
- `docs/go-technical-design.md` → "The harness": the new columns.

## Spec

- No count may come from an event announcing an outcome. Deaths come from `comp.Dead`, people from positions. Firm (rule 10).
- Regions come from M4-08's `RegionAt`, so a body in the fill between places counts nowhere.

## Acceptance criteria

- [ ] `go run ./cmd/harness run --start start.sample --seeds 2 --days 30` writes the new columns.
- [ ] Each emergence test above passes, and none reads an event to decide.
- [ ] One seeded year of `start.sample` runs within 60 s; the time is in Notes.
- [ ] `go tool task check-pr` and `go tool task check` pass.

## Out of scope

- Tuning numbers for good outcomes: M8.

## Notes

Filled in by the implementer.
