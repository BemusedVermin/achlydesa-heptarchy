---
id: M4-09
title: "Far agents: two fidelities"
status: todo
milestone: M4
kind: feature
depends_on: [M4-05, M4-07]
discovered_by: null
size: M
---

# M4-09 · Far agents: two fidelities

## Goal

Agents far from the party decide less often and travel the coarse chunk graph; near the party they act through the timeline cell by cell. It is one agent model at two fidelities, and nobody stops existing.

## Context

- `docs/game-design.md` → "Keeping it affordable".
- `docs/go-technical-design.md` → "Level of detail" (the radius, `agents.far_ticks` and the graph, decided on 2026-11-25), "Tick order" step 4, "Determinism rules".
- The brief's "The one-world model" → "Two fidelities".
- What exists: the agents controller (`controller.go`, `behaviors.go`), `move_toward` and the other primitives, `timeline.Suspend`/`Resume`, M4-05's `CoarseRoute`, and `content.AgentsConfig` in `behaviors.cue`.

## Deliverables

- `internal/content/types.go`: `AgentsConfig.NearChunks` (int) and `AgentsConfig.FarTicks` (Expr over the calendar, or int; say which). Regenerate the schemas. `content/data/behaviors.cue` sets 2 and a tenth of a phase.
- `internal/core/agents/fidelity.go`: which agents are near, recomputed after each party step (per entry is fine if it's cheap; measure).
- `internal/core/agents`: a far agent's decision comes every `far_ticks`; `move_toward` (and anything that walks) moves a far agent along a `CoarseRoute`, arriving at the next opening after that leg's steps × its step ticks. Near agents are unchanged.
- Tests (below), and a benchmark of a day of the sample world with the party idle.
- `docs/go-technical-design.md` → "Level of detail": how fidelity is switched, if it differs from the Spec.

## Spec

- **Near** is within `near_chunks` chunks (Chebyshev distance on chunk keys, depth included) of any member of the controlled party, or in sight of one. Firm: anything the player could see acts cell by cell.
- **Switching** keeps the body where it is. Going far, it finishes or cancels its pending actions at the next decision; going near, a far agent mid-leg is placed at the cell its leg has reached (the nearest standable cell on the route's cell path at that fraction), so nothing jumps in view. Suggestion: compute that cell lazily only when it comes near.
- **Fights off screen.** A far agent whose behavior engages a hostile within its own chunk, or whose target does, becomes near until the fight is over (no hostile within `near_chunks`). Provisional.
- **Needs and thresholds** tick per phase for everyone, as now.
- **Determinism.** Far decisions are timeline entries like any other, at a coarser cadence, so the order rule (landings, then decisions by UID) holds. Any parallel route search returns results the world's goroutine applies in UID order. Firm.

## Acceptance criteria

- [ ] A test: a far creature with `move_toward` a cell two chunks beyond reaches it in about the time the cell path would take (within the M4-05 factor), without deciding every step.
- [ ] A test: the party walks toward a far agent mid-leg; it becomes near on a cell of its route, and the plan view never shows it jump more than one step.
- [ ] A test: two hostile far agents meeting in one chunk fight cell by cell, and one dies.
- [ ] Replay: the same seed and commands give the same `Game.Hash`, with agents switching fidelity on the way.
- [ ] `go tool task check` passes.

## Out of scope

- Leaving for a ranked destination: M4-10.
- Tuning `near_chunks` and `far_ticks` to the time budget: M4-13 measures, M8 tunes.

## Notes

Filled in by the implementer.
