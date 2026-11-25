---
id: M5-04
title: Regions as areas and the region graph
status: obsolete
milestone: M5
kind: seam
depends_on: [M5-01]
discovered_by: null
size: M
---

# M5-04 · Regions as areas and the region graph

## Goal

A region is a named area of the one world carrying data (decay, and which materials appear there). Content draws the fixed places' regions by hand, and procgen assigns the rest from depth and noise. A graph of regions and the openings between them gives the connectivity guarantees something to read.

## Context

- `docs/game-design.md` → "World and towns" (depth is the geography), "Procgen and exports" (Guarantees: materials appear where they belong), "The living deep" (the deep changes character).
- `docs/go-technical-design.md` → Level of detail → "Regions are areas of one world" (a region describes ground and holds no population).
- The brief's "Planned tasks" → Regions as areas.

## Deliverables

- `internal/content/types.go`: `RegionDef` and `Bundle.Regions`; `PlaceDef.Region` (optional, the region a place sits in); regenerate the schemas. `content/data/regions.cue` with a few regions by depth band.
- `internal/content/regions.go`: checks, `Registry.Region`, `RegionIDs`.
- `internal/core/procgen/regions.go`: `RegionAt(source config, cell) api.ContentID`, deterministic from the seed and cell (the same chunk-free evaluation as noise), and `Graph`: gonum's graph of regions joined by openings.
- `go.mod`: gonum (named in the tech design; Notes say why).
- `internal/core/procgen/regions_test.go`: tests (below).

## Spec

- `RegionDef`: `Name` (a text ID), a depth band, `Decay` (a number), and `Materials` (content IDs that may appear, each with a weight). Suggestion.
- **Assignment.** A region at a cell: the fixed place's region where `Covers`, else the band for the cell's depth, chosen by low-frequency noise among the bands that overlap there. Depends on nothing but the seed and the cell, so it is order-independent.
- **The graph.** Nodes are regions, and chunk-level connections follow open cells between chunks. Built from a chunk-and-opening summary, never the whole world at once. Firm: the graph is data a later stage and the verify stage read; nothing walks it at play time.
- No population is held by a region, and nothing here changes a region over time (M5-12).

## Acceptance criteria

- [ ] The same cell gives the same region from any chunk visiting order.
- [ ] A fixed place's cells report its region; depth bands tile the world with no gaps.
- [ ] `go tool task check` passes.

## Out of scope

- Materials changing terrain: M5-12 and M6's content. Decay changing over exports: M5-12.
- Verifying and repairing connectivity: M5-07.

## Notes

Cut on 2026-11-25 by the owner, before anyone started it. Depth already decides what the deep is like, so named regions added little on top of it. Where each piece went:

- Decay: a formula of a cell's depth and the export count (M5-12).
- Spawns: by depth band on the geography (M5-11).
- Materials by depth: wait for the content that has those materials.
- The connectivity graph: chunks, not regions (M5-07).
- Where a seat is, which the piece reports: M7 defines it.
