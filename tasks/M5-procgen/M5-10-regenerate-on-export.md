---
id: M5-10
title: Regenerate the deep on an export
status: done
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

- [x] The test above passes on a handful of seeds.
- [x] `go tool task check` passes.

## Out of scope

- Seats' regions landing somewhere new: M7.
- The deep changing character: M5-12.

## Notes

- `explore` subscribes to `ExportBegan` (`export.go`). With no geography or no world it does nothing. Otherwise it snapshots the pocket chunks, builds a new generated source whose fixed layer is the places plus the pocket (`pocketFixed`, with a transient `pocket` footprint), swaps it into the map, forgets seen cells, and moves bodies.
- Pocket: the chunks within `world.pocket_chunks` of any party member (new content, first value 1). Its openings are standable cells the party can walk to inside the pocket that border new ground, at most `world.pocket_exits` (new content, first value 4, required at least 1 when a geography is set), spread evenly along the sorted list.
- Bodies outside the places' authored cells and the pocket go to the nearest free standable cell outside the pocket (rings of growing radius, UID order). A body already on open ground stays. Seen is cleared for every chunk except the pocket and the chunks places cover, so some stale cells remain in place chunks.
- `hexmap.Map.Replace` swaps the source and drops chunks and edits outside the fixed layer. `WorldSource.ChunkKeys` added. `cellSource` now takes a `worldgen.Fixed` and bounds.
- The test runs 4 exports on a 100 by 40 by 8 world (about 5 s), with `pocket_chunks` 0 since the world has few chunks. It does not check that bodies were moved, only that every body ends on open ground.
- The real world uses `geography.large`, which has no verify, so the route guarantee holds only for geographies with verify until M5-07.F1.
- golangci-lint cannot run here (built with an older Go), so lint is left to CI; the second criterion was checked with gofmt, vet, content, comments and the affected tests.
- Follow-ups filed: M5-10.F1 (pocket not saved), M5-10.F2 (sealed pocket).
