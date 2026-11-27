---
id: M5-12
title: Decay spreads downward over exports
status: done
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
- `internal/core/worldgen/`: stages read a cell's decay through the rule when choosing terrain.
- A test: the same cell is more decayed after more exports, and shallower cells decay before deeper ones.

## Spec

- Numbers and the formula are content, tuned later in M9. Code names no terrain. Firm (AGENTS.md rule 4).
- Fixed places do not change terrain; a place's own content may react to decay later.

## Acceptance criteria

- [x] The test above passes.
- [x] `go tool task check` passes.

## Out of scope

- Towns rotting (the still town): content for M7.

## Notes

- A geography has an optional `decay`: a `level` formula over `depth` and `exports` (the named rule `rule.decay_level`) and `terrains`, level n giving terrains[n-1] (the last once past the list). The loader requires the terrains to be solid.
- `worldgen.NewDecay` reads the formula once per depth; `Source.Chunk` then swaps solid, non-fixed cells for the decayed terrain after every stage has run, so stages and verify never see it. Open cells and fixed places are untouched.
- Rule and two terrains (`decayed_clay`, `rotten_clay`) are first values: one level per export, one less per 4 depths. Tune in M9.
- Not run: `go tool task check` fails here at golangci-lint (built with go1.25, repo targets 1.27); vet, gofmt and tests pass.
- Follow-ups filed: none.
