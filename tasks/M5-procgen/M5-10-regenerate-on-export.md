---
id: M5-10
title: Regenerate the deep on an export
status: todo
milestone: M5
kind: feature
depends_on: [M5-06, M5-07, M5-08, M5-09]
discovered_by: null
size: M
---

# M5-10 · Regenerate the deep on an export

## Goal

An export reshapes the deep around the party: the generated ground is rebuilt from the new export count, the party's pocket and the fixed places stay, and the party still has a route to at least one town. Cells seen on regenerated chunks are forgotten.

## Context

- `docs/game-design.md` → "Procgen and exports" (Reseeding, the map goes stale), "An export with the party below".
- Decision, 2026-11-25: bodies standing in regenerated deep are moved to the nearest open ground; they are not lost. Recorded in `docs/game-design.md`.
- M5-09's `ExportBegan`; M5-08's unloading; M5-07's guarantees.

## Deliverables

- `internal/core/explore/` (or the package that owns the world): a handler for `ExportBegan` that drops every loaded generated chunk (edits to generated cells are discarded; edits inside fixed places stay) and bumps the source's export count.
- The pocket: the party's current chamber, its cells and a radius, kept as generated before the export. Suggestion: the chunks within the near radius are kept as edits-overlay copies, so they survive the new seed.
- Bodies in regenerated cells move to the nearest open cell outside the pocket, sorted by `comp.UID` for determinism.
- `comp.Seen` cleared for regenerated chunks.
- A test: export with the party below; the pocket is unchanged, a different deep surrounds it, a route to at least one place exists, bodies are on open ground, seen cells on regenerated chunks are gone, and the same seed gives the same result.

## Spec

- Which chunks keep: those within the pocket radius of any party member and every chunk any place covers.
- The new deep must pass M5-07's verify with the pocket included as fixed ground; the verify stage treats the pocket like a place (suggestion: the pocket is registered as a transient `Fixed` footprint).
- No system announces the export beyond the existing event; the player finds out from the changed map.

## Acceptance criteria

- [ ] The test above passes on a handful of seeds.
- [ ] `go tool task check` passes.

## Out of scope

- Seats' regions landing somewhere new: M7.
- The deep changing character: M5-12.

## Notes

Filled in by the implementer: assumptions, deviations, files touched outside Deliverables, follow-ups filed.
