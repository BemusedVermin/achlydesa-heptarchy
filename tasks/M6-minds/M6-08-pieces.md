---
id: M6-08
title: "Pieces: building blocks a town is laid out from"
status: done
milestone: M6
kind: seam
depends_on: [M6-07]
discovered_by: null
size: M
---

# M6-08 · Pieces: building blocks a town is laid out from

## Goal

A piece is a small content-defined sub-place (a house, a well, a larder, a shrine, a graveyard) that can be stamped into the world at an offset. The settlement stage (M6-09) lays towns out from pieces, so nobody hand-draws a town.

## Context

- `docs/game-design.md` → "World and towns" (Crawlspaces are generated and emergent; human scale), "The hex grid" (human scale).
- `internal/content/places.go`: a `PlaceDef` has layers, a legend, marks and spawns, set at a fixed `at`. A piece reuses that shape without the position.
- Prior art: Minecraft villages, assembled from structure pieces with doors that must meet.

## Deliverables

- `internal/content/`: `PieceDef` (`piece.` IDs): layers and a legend like a place, `marks` for spawns, landmarks (M6-07), containers and sources (M6-02), and `doors`: cells where it joins a street or another piece. Validation checks references.
- `internal/core/explore/`: stamping a piece at an offset and rotation into the world's edits, as a place is stamped today; a piece may not overlap anything already stamped.
- `content/data/pieces/`: a first few pieces: `piece.house`, `piece.well`, `piece.larder`, `piece.shrine`, `piece.graveyard`, `piece.street`.
- Tests: a piece stamps at an offset and a rotation, with its marks mapped; overlapping stamps are refused.

## Spec

- **Firm:** a piece is data; code names no piece. Rotation is one of six hex directions.
- **Firm:** a piece's marks may spawn bodies, place landmarks and place containers and sources, through the same mechanism places use.
- **Suggestion:** pieces are authored small (under 12 cells across); size is the author's choice.

## Acceptance criteria

- [x] The tests above pass.
- [x] `go tool task check` passes.

## Out of scope

- Choosing where pieces go: M6-09.
- The specific towns: M6-22.

## Notes

- **Tests:** `TestAPieceStampsAtAnOffsetAndARotationWithItsMarksMapped` and `TestAStampThatOverlapsOrLeavesTheWorldIsRefused` (`explore/stamp_test.go`), `TestRotateTurnsEastToEachDirectionAndBack` (`hexmap/rotate_test.go`), `TestPieceDoorsAreChecked` (`content/places_test.go`).
- **Stamping:** "the world's edits" read as the fixed layer a place is stamped into, not `hexmap.Map` edits (those are player changes and are saved). `(*WorldSource).Stamp(reg, piece, at api.Cell, rotation api.Dir)` sets a piece into the same authored cells places use, so worldgen's `Covers` sees it and its spawns (bodies, containers, sources) and landmarks start like a place's. It returns `Stamped{Marks, Doors}`; piece marks are not recorded in the source, so `WorldSource.Mark` still finds only places. A refused stamp changes nothing and wraps `ErrOverlap` or `ErrOutside` (outside the bounds was not in the spec; refused for the same reason). Stamp before a map loads chunks from the source.
- **Rotation:** `hexmap.Rotate(h, dir)` turns a hex so `DirE` points to `dir` (a sixth turn anticlockwise per step), in `hexmap` so worldgen can use it in M6-09. A piece is placed by adding axial coordinates, so unlike a place it may start on any row.
- **Content:** `PieceDef` mirrors `PlaceDef` without `at` and `openings`, plus `doors`. A piece needs at least one door (my choice; easy to loosen). Doors may repeat and may not be spawns or landmarks. A shared `content.Layout` (with `PlaceDef.Layout()` and `PieceDef.Layout()`) feeds both `checkLayout` and explore's new `draw`, which replaces `layOut`; the schema step emits an unused `#Layout` for it. A piece's landmarks take the piece ID as their identity template, as a place's take the place ID. No sample place stamps pieces yet; M6-09 does.
- **Outside the Deliverables:** `internal/core/hexmap/rotate.go`, `internal/content/{affordances,records,load}.go`, the regenerated `types_go_gen.cue`, `docs/content-guide.md` (pieces row, ID table, a piece recipe).
- **Follow-ups:** none.

