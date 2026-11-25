---
id: M5-12
title: Decay spreads downward over exports
status: todo
milestone: M5
kind: feature
depends_on: [M5-10]
discovered_by: null
size: S
---

# M5-12 · Decay spreads downward over exports

## Goal

The deep changes character over the game: each export lets decay reach a little deeper, so the upper deep grows more decayed and old matter retreats downward.

## Context

- `docs/game-design.md` → "World and towns" (towns keep their layout, not their condition), "Procgen and exports" (the deep changes character).
- M5-10 regenerates the deep with the new count.

## Deliverables

- `internal/content/types.go`, `content/data/`: a rule giving a cell's decay from `(depth, export count)` as a named formula in `rules.cue`, plus the terrains decayed ground uses, on the geography.
- `internal/core/procgen/`: stages read a cell's decay through the rule when choosing terrain.
- A test: the same cell is more decayed after more exports, and shallower cells decay before deeper ones.

## Spec

- Numbers and the formula are content, tuned later in M9. Code names no terrain. Firm (AGENTS.md rule 4).
- Fixed places do not change terrain; a place's own content may react to decay later.

## Acceptance criteria

- [ ] The test above passes.
- [ ] `go tool task check` passes.

## Out of scope

- Towns rotting (the still town): content for M7.

## Notes

Filled in by the implementer: assumptions, deviations, files touched outside Deliverables, follow-ups filed.
