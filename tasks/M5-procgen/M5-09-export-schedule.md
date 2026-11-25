---
id: M5-09
title: The export schedule and count
status: todo
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

- [ ] The test above passes.
- [ ] `go tool task check` passes.

## Out of scope

- What an export does to the world: M5-10. Seats' regions and stage effects: M7.

## Notes

Filled in by the implementer: assumptions, deviations, files touched outside Deliverables, follow-ups filed.
