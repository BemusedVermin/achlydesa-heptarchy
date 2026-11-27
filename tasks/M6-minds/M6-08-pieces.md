---
id: M6-08
title: "Pieces: building blocks a town is laid out from"
status: todo
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

- [ ] The tests above pass.
- [ ] `go tool task check` passes.

## Out of scope

- Choosing where pieces go: M6-09.
- The specific towns: M6-22.

## Notes

