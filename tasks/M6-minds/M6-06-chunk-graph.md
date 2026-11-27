---
id: M6-06
title: "The chunk graph: the openings between chunks"
status: todo
milestone: M6
kind: seam
depends_on: []
discovered_by: null
size: M
---

# M6-06 · The chunk graph: the openings between chunks

## Goal

Far agents can find a route through the world without loading it. The world exposes a graph whose nodes are chunks and whose edges are the openings between neighbouring chunks.

## Context

- `docs/go-technical-design.md` → "Level of detail" (Far agents), "Determinism rules" (parallel, then in order).
- `internal/core/explore/unload.go`: far chunks are unloaded but their bodies stay, and `hexmap` sight and pathing need loaded chunks, so something else must serve far travel.
- `internal/core/worldgen`: chunks are a pure function of seed, export count and content; `verify` already summarizes chunks (`M5-07.F1`).

## Deliverables

- `internal/core/worldgen/graph.go`: `Graph` with `Neighbors(key)` and `Route(from, to)` (A*), built from per-chunk opening summaries; each summary is pure data computed from the seed and export count.
- Cache keyed by (export count, chunk key); the explore package drops it on an export.
- Tests: opening summaries match the loaded chunk's real openings on a few seeds; a route exists between two chunks joined by a tunnel; the same seed and export count give the same graph.

## Spec

- **Firm:** summaries are computed on other goroutines as pure functions and applied in chunk-key order.
- **Firm:** the graph holds fixed places' chunks exactly as the world has them, and changes with an export only where generation changes.
- **Suggestion:** an opening is a cell pair on a shared edge plus its walk cost; `Route` returns the chunk keys and a cost.

## Acceptance criteria

- [ ] The tests above pass; the same graph results with the cache cold or warm.
- [ ] `go tool task check` passes.

## Out of scope

- Agents using the graph: M6-17.

## Notes

