---
id: M3-10
title: "The off-screen tick: drift, stocks and migration"
status: todo
milestone: M3
kind: feature
depends_on: [M3-09]
discovered_by: null
size: M
---

# M3-10 · The off-screen tick: drift, stocks and migration

## Goal

Off-screen regions live. Each day phase, populations drift between neighbouring regions by content formulas, towns spend and refill their stocks, and a town that runs short loses people to its neighbours, so the world changes while the party is elsewhere.

## Context

- The agents model in [README.md](README.md), "Level of detail" and "Towns".
- `docs/go-technical-design.md` → "Tick order" step 4, "Determinism rules" (parallel then in order, by region ID).
- `docs/game-design.md` → "Monsters with needs" (what falls out of it), "People with needs" (towns can empty), "Keeping it affordable".

## Deliverables

- `internal/content/types.go`: `FamilyDef` gains `Drift` and `Attrition`; `TownDef` gains `Use` and `Refill`; schemas regenerated.
- `internal/core/agents/offscreen.go`: the `PerPhase` system.
- `internal/core/api`: `Migrated`, `TownEmptied`, `PopulationDied` events.
- `internal/core/agents/offscreen_test.go` with rapid properties.

## Spec

**Firm:**

- Only regions the party is not in are ticked.
- **Drift:** for each family in each region, `Drift` is an Expr over a `RegionEnv` of the region and one neighbour (their `data` numbers, counts per family, stocks, dead bodies left there) giving the fraction of the count that moves to that neighbour this phase. Moves are computed from the state at the start of the phase and applied after, in region ID then family ID order, so the order of regions can't change the result.
- **Attrition:** an Expr giving how many of a family die this phase (families fighting each other, starvation), emitting `PopulationDied`.
- **Towns:** residents are a population of a resident family. Each phase a town spends `Use` (per resident, per stock item) and gains `Refill`; when a needed stock is empty, the town's residents drift to the neighbouring town with the most of it (`Migrated`). A town whose residents reach 0 emits `TownEmptied` once.
- Counts are whole numbers; fractions are rounded with the `lod` stream so totals are kept.

**Suggestion:** compute each region's moves on worker goroutines and apply them in order, as the determinism rules allow, only if a benchmark shows the tick is slow.

## Acceptance criteria

- [ ] Property (rapid): drift alone never changes the total count of a family across all regions.
- [ ] A test town with its water refill set to 0 loses residents to its neighbour over a few days, emits `Migrated`, then `TownEmptied` once.
- [ ] A test family whose drift favours regions with dead bodies gathers in a region where bodies were left.
- [ ] The same seed gives the same counts after 10 days.
- [ ] `go tool task check` passes.

## Out of scope

- Residents as full agents in town (M3-11).
- Harness columns (M3-20).

## Notes


