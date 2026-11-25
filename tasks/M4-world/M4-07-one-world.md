---
id: M4-07
title: "One world: delete maps and gates"
status: todo
milestone: M4
kind: seam
depends_on: [M4-01, M4-02, M4-03, M4-06]
discovered_by: null
size: M
---

# M4-07 · One world: delete maps and gates

## Goal

A game runs in one world built from M4-06's places. A position is a cell. Map IDs, the maps registry, gates and every "which map" check are deleted, and today's maps become places of that one world, moved over as they are.

## Context

- `docs/game-design.md` → "World and towns" (One world, no seams), "The hex grid".
- `docs/go-technical-design.md` → "Designing for pivots" → "The world", "Plan view"; "Saves and replays".
- The brief's "The one-world model" → "One world" and "Interim".
- What exists: `api.MapID`, `comp.Pos.Map`, `comp.RestPlace.Map`, `comp.Seen.ByMap`, `hexmap.Maps`, `hexmap.Link.Map`, `hexmap.Gate`, `content.MapDef`, `GateDef`, `explore.GateSite`, `internal/core/explore/gates.go`, `agents/leave.go` (`crossGate`, `rankedGate`), `api.PartyCrossed`, `api.Migrated`, the save's `MapEdits`, `content/data/maps/`, and `start.cue`. Each of M4-01 to M4-03 has already removed the agents code that kept per-map state.

## Deliverables

- `internal/core/api`: delete `MapID`; drop the `Map` field from `Moved`, `PartyWoke`, `MapView`, `PartyView` and `RegionView`; delete `PartyCrossed` and `Migrated`.
- `internal/core/comp`: `Pos` is `{At api.Cell}`; `RestPlace` is `{At api.Cell}`; `Seen` holds one sorted cell list (keep the helpers in explore as the only readers).
- `internal/core/hexmap`: delete `Maps`, `maps.go`, `gates.go` and `Link.Map`; a `Map` has no ID (rename it `World` if that reads better, and say so in Notes). `sim.World` holds one `*hexmap.Map`.
- `internal/core/explore`: build the one world from `NewWorldSource` and `content.World()` with M4-04's bounds; spawn every place's spawns; delete `gates.go` and gate marks; follow, wake, sight and the plan view stop checking maps.
- `internal/core/agents`: delete `leave.go` and the `leave` primitive; `RegionOf`, `Neighbours` and `partyMap` go. Until M4-08, `region` in formulas is the empty record and the region section of the view is empty.
- `internal/core/game`, `internal/core/save`: one list of edits; bump `CurrentVersion`.
- `internal/content`: delete `MapDef`, `GateDef`, `RegionDef`, `maps.go` and the `map.` prefix; `Start` is `{place, party, party_at_mark}`; regenerate the schemas.
- Content: every file in `content/data/maps/` becomes a place in `content/data/places/` at an offset of its own, apart from the others, as it is (no rescaling). The gate marks go. The playground hub is deleted; `start` begins in the hall place. `start.sample` begins in the shallow town place. Remove behaviors that used `leave` from templates.
- `internal/frontend/tui`: nothing names a map; update its goldens.
- Tests: every test that crossed a gate or named a map either goes (gate tests) or starts at a place.
- `docs/content-guide.md`: "A map" becomes "A place"; "The playground" and "The sample world" say they are places of one world, pending M4-11 and M4-12. `docs/go-technical-design.md` → "Core packages": `comp`'s `Pos{At}`, and `hexmap` holds one world. `docs/glossary.md`: delete **gate**.

## Spec

- **Interim, on purpose.** The old maps sit side by side in clay with nothing joining them, and nobody travels between places, until M4-11 and M4-15. That's expected; say so in the PR.
- The plan view is unchanged in behavior: it projects the world around the party, and `MapView` loses only its map ID and name (the region panel names where the party is, from M4-08).
- Determinism: spawns run in place ID order, then the order each place's layers give. Firm.
- Name each golden and harness output this task regenerates in Notes.
- If this task grows past M, split the save and view changes into a follow-up only if the build still passes without them; otherwise keep it whole and say why in Notes.

## Acceptance criteria

- [ ] `git grep -i -w -E "gates?|MapID|ByMap|Migrated|PartyCrossed" -- internal content docs` finds nothing.
- [ ] `go tool task run -- play --seed 1` opens in the hall place, and the plan view looks as it did.
- [ ] A save of a game with an edited cell loads with the edit, and an older save fails cleanly with the version error.
- [ ] `go run ./cmd/harness run --start start.sample --steps 40` runs.
- [ ] `go tool task check-pr` and `go tool task check` pass.

## Out of scope

- Regions as areas, and the region panel: M4-08.
- Agents going anywhere in the world for their needs: M4-15.
- Rebuilding places at human scale: M4-11, M4-12.

## Notes

Filled in by the implementer.
