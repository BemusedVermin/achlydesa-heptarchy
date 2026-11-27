---
id: M6-09
title: "The settlement stage: procgen lays out a town"
status: todo
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

- [ ] The tests above pass.
- [ ] `go tool task check` passes.

## Out of scope

- The three towns: M6-22.
- Residents: M6-23.
- Digging a new crawlspace: M8.

## Notes

