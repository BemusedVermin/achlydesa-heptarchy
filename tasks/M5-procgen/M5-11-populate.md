---
id: M5-11
title: Populate generated ground
status: todo
milestone: M5
kind: feature
depends_on: [M5-06, M5-08]
discovered_by: null
size: M
---

# M5-11 · Populate generated ground

## Goal

Generated ground gets its starting creatures as bodies, chosen from content by depth. Nothing is a count: each creature is an individual body from a template.

## Context

- `docs/game-design.md` → "The living deep", "Monsters with needs".
- `docs/go-technical-design.md` → "Procgen stages" (populate); "Determinism rules" (parallel, then in order).
- What exists: place `Spawns` (`explore.Spawn`) already spawn a body from a template at a marked cell at the start of a game.

## Deliverables

- `internal/content/types.go`, `content/data/`: `GeographyDef.Spawns` (template IDs, each with a weight, a depth band and a density), so what lives where is data. They sit on the geography so that a later geography per area brings its own.
- `internal/core/procgen/populate.go`: the `populate` stage computes the spawns of a chunk as plain data (cell, template) from the seed; it does not touch the world.
- `internal/core/explore/`: the world's goroutine applies a chunk's spawns once, in chunk-key then cell order, the first time the chunk is generated for the current export count.
- `internal/core/save/`: which chunks are populated for this export count; bump the save version.
- A test: walking into a chunk spawns bodies once, reloading does not spawn again, and the same seed spawns the same bodies.

## Spec

- **Provisional:** a chunk is populated once, the first time it loads for an export count. This is world generation, not a respawn system: nothing replaces a dead creature, and a chunk unloaded and reloaded is not populated again (recorded in the save). If M6 wants something else, it changes this task's rule.
- Bodies keep a stable `UID` and carry no knowledge of the chunk; far agents are M6's.
- Spawn density and the template mix are content. The test uses only existing templates.

## Acceptance criteria

- [ ] The test above passes.
- [ ] `go tool task check` passes.

## Out of scope

- What creatures do (needs, behaviors): M6.
- Immigrants arriving from beyond the world: M6.

## Notes

Filled in by the implementer: assumptions, deviations, files touched outside Deliverables, follow-ups filed.
