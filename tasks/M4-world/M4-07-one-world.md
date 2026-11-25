---
id: M4-07
title: "One world: the maps become places"
status: done
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

- [x] `git grep -w -E "MapID|ByMap|hexmap.Maps" -- internal content docs` finds nothing.
- [x] `go tool task run -- play --seed 1` opens in the hall place, and the plan view looks as it did.
- [x] `go tool task check` passes.

## Out of scope

- Regions as areas, and joining the places: M5.
- Agents of any kind beyond M2's combat: M6.

## Notes

- **Interim, on purpose:** the three places sit side by side in the world's clay with nothing joining them; M5's procgen joins them. A party started in one place stays there.
- **Layout:** world 70 by 30 hexes, depths 0 to 7. Hall at column 4, row 4; descent at column 30, row 4; den at column 56, row 4; all at depth 0, at least five hexes of clay apart and from the edge. Places keep their rows as they were. IDs are now `place.test.*`, names `text.place.test.*`.
- **Plan view:** the TUI goldens did not change, and `play --seed 1` opens in the hall. The combat encounter goldens changed only by the hall's offset (every hex +2 q, +4 r) and the dropped `map` field; I checked that line by line.
- **Names:** `api.MapView` keeps its name (it is the plan view); it loses `Map` and `Name`, which the frontend never showed. `Start` is `{place, mark, party}`, as M4-06 suggested. `save.MapEdit`/`MapEdits` became `save.Edit`/`Edits` (`"edits"`), and `CurrentVersion` is 3. `explore.LoadMaps` became `LoadWorld`. `hexmap.Map` loses its ID and name; `hexmap.Link` keeps only `Cell`, and the step test through a link to another map went.
- **`effects.SetMaps` is gone:** effects reads `w.Map` directly, since `sim` is a foundation package. Content with no `world` leaves `w.Map` nil, as content with no maps left the registry empty.
- **The world's edge is the world's bounds:** `explore` sets them on the one `hexmap.Map` from `WorldSource.Bounds` and `Edge`.
- **Tests:** fixtures name a place instead of a map. `combat`'s field picks the moving clay nearest the party, since several places now hold one in the same world. The ledge fixture is a place at depth 3, so its negative layers stay in bounds. The actions fixture also drops the world, so its bodies stand nowhere, as before.
- **Outside the Deliverables:** `internal/core/{actions,combat,effects,sim}` (map checks dropped), `internal/harness/run.go` (the CSV's `map` column is gone), `content/data/text/explore.cue` ("The world could not be read"), `ClimbCost` moved from the deleted `maps.go` to `registry.go`, and `docs/go-technical-design.md` (`comp`, `hexmap` rows, the save, the data folder).
- **Docs:** `docs/content-guide.md`'s map recipe is now "A place"; `docs/glossary.md` gains **world** and **place**, saying there is no map as a unit of the world.
- `go tool task check` and `go tool task check-ci BASE=origin/master` pass.
- No follow-ups filed.
