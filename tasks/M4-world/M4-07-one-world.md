---
id: M4-07
title: "One world: the maps become places"
status: todo
milestone: M4
kind: seam
depends_on: [M4-06]
discovered_by: null
size: M
---

# M4-07 · One world: the maps become places

## Goal

A game runs in one world built from M4-06's places. A position is a cell. Map IDs, the maps registry and every "which map" check are deleted, and today's maps become places of that one world, moved over as they are.

## Context

- `docs/game-design.md` → "World and towns" (One world, no seams), "The hex grid".
- `docs/go-technical-design.md` → "Designing for pivots" → "The world", "Plan view"; "Saves and replays".
- What exists: `api.MapID`, `comp.Pos.Map`, `comp.RestPlace`, `comp.Seen`, `hexmap.Maps`, `content.MapDef`, `content.Start`, and `explore.NewSource`.

## Deliverables

- `internal/core/api`: delete `MapID`, and drop the map from every event and view that carries one.
- `internal/core/comp`: `Pos` is `{At api.Cell}`; `RestPlace` holds a cell; `Seen` holds one sorted cell list.
- `internal/core/hexmap`: delete `Maps` and `maps.go`; `sim.World` holds one `*hexmap.Map`.
- `internal/core/explore`: build the one world from `NewWorldSource` and `content.World()` with M4-04's bounds, spawn every place's spawns, and stop checking maps.
- `internal/core/game`, `internal/core/save`: one list of edits; bump `CurrentVersion`.
- `internal/content`: delete `MapDef`, `maps.go` and the `map.` prefix; `Start` names a place and a mark; regenerate the schemas.
- Content: every file in `content/data/maps/` becomes a place in `content/data/places/` at an offset of its own, apart from the others, as it is (no rescaling). `start` begins in the hall place.
- `internal/frontend/tui`: nothing names a map; update its goldens.
- Tests that named a map start at a place instead; tests that only checked map bookkeeping go.
- `docs/content-guide.md`: "A map" becomes "A place". `docs/glossary.md`: no **map** as a unit of the world.

## Spec

- **Interim, on purpose.** The places sit side by side in clay with nothing joining them until M5's procgen joins them. Say so in the PR.
- The plan view behaves as before: it projects the world around the party.
- Determinism: spawns run in place ID order, then the order each place's layers give. Firm.

## Acceptance criteria

- [ ] `git grep -w -E "MapID|ByMap|hexmap.Maps" -- internal content docs` finds nothing.
- [ ] `go tool task run -- play --seed 1` opens in the hall place, and the plan view looks as it did.
- [ ] `go tool task check` passes.

## Out of scope

- Regions as areas, and joining the places: M5.
- Agents of any kind beyond M2's combat: M6.

## Notes

Filled in by the implementer.
