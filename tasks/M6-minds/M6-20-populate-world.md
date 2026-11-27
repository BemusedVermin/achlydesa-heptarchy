---
id: M6-20
title: "Populate the whole world at start, with minds"
status: done
milestone: M6
kind: feature
depends_on: [M6-03, M6-06, M6-14]
discovered_by: null
size: M
---

# M6-20 · Populate the whole world at start, with minds

## Goal

A new game creates every creature in the bounded world as an individual, with its drives, weights and planner, without the party having walked there. An export creates nobody.

## Context

- `docs/go-technical-design.md` → "Level of detail" (Populated at start); `docs/game-design.md` → "Keeping it affordable".
- Replaces M5-11's rule that a chunk populates when it first loads (`tasks/M5-procgen/M5-11-populate.md` Spec marks that provisional: "If M6 wants something else, it changes this task's rule"). `internal/core/worldgen/populate.go` and `internal/core/explore/populate.go`.
- `explore.populateLoaded` and `Snapshot.Populated`, and M5-10's clearing of the set on an export.

## Deliverables

- `internal/core/worldgen/populate.go`: `Populate(key)` is called for every chunk in the world, in parallel and as pure data, from per-chunk streams (so load order does not matter).
- `internal/core/explore/`: apply the spawns once at world creation, in chunk-key then cell order, spawning far agents with no body cell loaded as an entity with a position and a chunk key; drop `populateLoaded` on step.
- Spawns assign a drive weight set (M6-03, drawn per chunk) and a planner tier: a spawn entry may set `mind.planner: goap` with a chance, so important individuals are picked by the generator.
- An export no longer clears or re-runs population; bodies are moved to open ground by the existing pocket rule.
- Save: drop `Snapshot.Populated`; bump the save version.
- Tests: the same seed gives the same individuals; reloading a save creates nobody; an export creates nobody and loses nobody; total count for the sample world is in the band the content states.

## Spec

- **Firm:** start-up work for about 2,500 chunks is bounded, parallel and cached; the time is recorded in Notes and measured again by M6-28.
- **Firm:** populate is world generation, not respawn: a dead creature is never replaced (AGENTS.md rule 10).
- **Provisional:** if scanning every chunk at start is too slow in practice, populate far chunks lazily from the same stream, so results do not change; file a follow-up rather than changing the rule.

## Acceptance criteria

- [x] The tests above pass.
- [x] `go tool task check` passes.

## Out of scope

- Immigrants and births: M8.
- Specific families and towns: M6-21 and M6-22.

## Notes

- `worldgen.Source.PopulateAll(workers)` runs `Populate` for every chunk in the bounds on several goroutines and joins the results in chunk-key order; `explore.populateWorld` spawns them once, from `Start`. `Load` skips `Start`, so a reload creates nobody. An export no longer touches population. `explore.Populated`, `RestorePopulated` and `Snapshot.Populated` are gone; save version 15.
- Start-up on the real 400 by 400 world: about 3.9 s and about 3,900 individuals, about four times the 1,000 target (density is for M9 to tune). No cache beyond each chunk's own ground reader. M6-28 measures it again. No lazy fallback was needed.
- **Spawn mind:** `GeographySpawn.Mind {planner, chance}`. The content check requires the template to have a mind and the chance in [0, 1]. The chance is drawn in `Populate` from the chunk's stream, so the generator picks. `agents.ValidateContent` checks the planner is registered. No sample spawn uses it yet: no planner exists until M6-15.
- **Drive weights** already draw from `agents`, salted by template and spawn cell, which does not depend on load order, so no per-chunk salt was added.
- **No chunk-key component:** a far body keeps its `comp.Pos` (a standable cell by construction) and its chunk key is `hexmap.KeyOf(Pos.At)`. M6-17 can add a component if it wants one.
- Whole-world population made every body schedule itself at spawn, and deciding loaded each body's chunk (747 of 2,500 chunks at start). Interim fix until M6-17 takes far agents off the timeline: `explore`'s step check refuses a step for a body in an unloaded chunk, and `combat.seenCells` sees nothing from one. Such a body stays dormant until something reschedules it; M6-17's promotion should do that. Added `hexmap.Map.IsLoaded`.
- Files outside the list: `internal/content/{types,geography}.go` and the regenerated schema, `internal/core/agents/mind.go`, `internal/core/combat/view.go`, `internal/core/hexmap/map.go`, `internal/core/game/save.go`, `internal/core/save/`. Tests: `explore/populate_test.go` reworked, `worldgen/populate_test.go`, and a real-world count test that takes about 4 s because it builds the real world.
- No follow-ups filed.

