---
id: M5-09
title: The export schedule and count
status: done
milestone: M5
kind: feature
depends_on: [M4-07]
discovered_by: null
size: S
---

# M5-09 · The export schedule and count

## Goal

Exports happen on an irregular schedule from content. The world counts them, and the count is saved and available to procgen as the reseeding input.

## Context

- `docs/game-design.md` → "Procgen and exports" (Reseeding), "Time pressures".
- `docs/go-technical-design.md` → Core packages → `clock` ("export scheduling").
- What exists: `content/data/calendar.cue` has day phases only; `clock.Clock.Advance`; `rng` named streams, saved with the game.

## Deliverables

- `internal/content/types.go`, `content/data/calendar.cue`: an `exports` record on the calendar (first day, minimum and maximum days between exports); regenerate the schemas. Placeholder numbers for tuning.
- `internal/core/clock/`: when the day passes an export day, the clock emits an `api.ExportBegan` event and increases the world's export count; the next export day is drawn from a named stream.
- `internal/core/api/`: `ExportBegan` event (day, count).
- `internal/core/sim/`: the export count on the world; `internal/core/save/`: saved and restored; bump the save version.
- A test: with a fixed seed the export days are identical across runs and a save made mid-schedule resumes the same next day.

## Spec

- The count starts at 0. The seed passed to procgen is `(world seed, count)`. Firm.
- The event says what happened; nothing here regenerates anything (M5-10 subscribes).
- Export days are irregular: each gap is drawn from `[min, max]`. Firm.

## Acceptance criteria

- [x] The test above passes.
- [x] `go tool task check` passes.

## Out of scope

- What an export does to the world: M5-10. Seats' regions and stage effects: M7.

## Notes

- The next export day is stored on the world (`NextExportDay`, 0 until the first is scheduled) beside `ExportCount`, and saved with it in a new `exports` block. The first export is on `first_day` and draws nothing; each later gap is drawn from the `clock.exports` stream.
- The export fires after `DayStarted` and before per-day systems.
- Placeholder numbers: first day 30, gaps 20 to 40. Content validation rejects a first day or minimum gap below 1, or a maximum below the minimum.
- Save version is now 5; the migration test chain in `save_test.go` gained a step.
- `docs/go-technical-design.md` save paragraph now lists the export fields.
- The resume test restores the RNG state and world fields by hand; the game save path itself is the two-line copy in `game/save.go`.
- Follow-up filed: M5-09.F1 (a flaky watcher test that failed the first CI run).
