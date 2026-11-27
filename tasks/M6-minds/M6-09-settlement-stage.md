---
id: M6-09
title: "The settlement stage: procgen lays out a town"
status: done
milestone: M6
kind: feature
depends_on: [M6-08]
discovered_by: null
size: M
---

# M6-09 · The settlement stage: procgen lays out a town

## Goal

A town is generated, not drawn. Content says where a settlement sits, how deep, how big, and which pieces it may use. The `settlement` stage lays the pieces out along streets, deterministically from the world seed.

## Context

- `docs/game-design.md` → "World and towns" ("Towns keep their layout but not their condition"; Crawlspaces are generated), "Procgen and exports" ("Ways out": a fixed place opens only where its author marks an exit).
- `docs/go-technical-design.md` → "Procgen stages". `internal/core/worldgen/`: stages declared per geography, `join` links places to the deep.
- Owner decision, 2026-11-27: use procgen for towns and landmarks, as Minecraft does; hand-authoring them is out.

## Deliverables

- `internal/content/`: `SettlementDef` (`settlement.` IDs): `at` (column, row, depth), `radius`, `pieces` with weights and counts, `street` piece, `exits` (how many ways out to the deep), and `landmark_min` (the landmarks it must contain).
- `internal/core/worldgen/settlement.go`: the stage. It uses the `worldgen` stream seeded from the world seed only, never the export count, so an export never rearranges a town.
- It plugs into the world the way places do (stamped when the world is built) and is joined to the deep by `join` at its marked exits only.
- Tests: the same seed gives the same layout; pieces never overlap; every door meets a street or another door; an export leaves the layout unchanged; `landmark_min` holds.

## Spec

- **Firm:** a settlement is laid out as a street graph grown from a centre, with pieces attached at their doors, as in Minecraft's jigsaw. Placement draws from a stream and places in a fixed order, so the result is a pure function of the seed.
- **Firm:** it fails with an error, never a default, when the pieces cannot satisfy `landmark_min`.
- **Firm:** a settlement is fixed ground. Its condition (decay, stocks, residents) changes; its layout does not.

## Acceptance criteria

- [x] The tests above pass.
- [x] `go tool task check` passes.

## Out of scope

- The three towns: M6-22.
- Residents: M6-23.
- Digging a new crawlspace: M8.

## Notes

- The stage is `worldgen.LaySettlement`, not a chunk `Stage`: a town is fixed ground, so `explore.NewWorldSource` calls it and stamps the pieces, as it does places. `NewWorldSource` now takes the world seed (`w.RNG.Seed()`); callers in tests pass 0. Exits join the footprint as openings, so `join` reaches them like a place's.
- Draws use `Derived("worldgen", hash of settlement ID)` from the seed only.
- Reading of the spec: `pieces` maps piece ID to `weight` and `count` (the most uses); `street` must be one of them and is the only piece exits leave from; `landmark_min` maps an affordance ID to the fewest landmark cells giving it. A lacking affordance makes the layout pick only pieces that give it; if none can, or none fit, it errors.
- Two pieces join where their doors are adjacent. "Every door meets a street or another door" is tested as: every piece after the first has a door beside an earlier piece's door. Other doors may stay unmet.
- `piece.street` in `content/data/pieces/pieces.cue` had its only side door inside its walkway, so nothing could join its sides. It now has doors in its walls, as its comment says.
- Content check: a settlement's disc (radius plus one hex) must lie in the world and clear every place and other settlement at the depths its pieces use.
- No settlement is in `content/data` yet (M6-22); `settlements/settlements.cue` holds the schema only. `docs/content-guide.md` gains a settlements section.
- Follow-ups filed: none.
