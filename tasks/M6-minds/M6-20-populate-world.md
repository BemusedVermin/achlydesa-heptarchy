---
id: M6-20
title: "Populate the whole world at start, with minds"
status: todo
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

- [ ] The tests above pass.
- [ ] `go tool task check` passes.

## Out of scope

- Immigrants and births: M8.
- Specific families and towns: M6-21 and M6-22.

## Notes

