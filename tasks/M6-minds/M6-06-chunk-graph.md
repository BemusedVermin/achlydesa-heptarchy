---
id: M6-06
title: "The chunk graph: the openings between chunks"
status: done
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

- [x] The tests above pass; the same graph results with the cache cold or warm.
- [x] `go tool task check` passes.

## Out of scope

- Agents using the graph: M6-17.

## Notes

- `worldgen.Graph` (`NewGraph`, `Reset`, `Prepare`, `Summary`, `Neighbors`, `Route`) and the pure `Summarize`. `Opening` was taken (a place's opening), so a step between chunks is `ChunkOpening`, a chunk's list is `OpeningSummary`, and an edge is `ChunkLink`.
- An opening is a step (`hexmap.Map.Step`, the party's step rules) from a standable cell to a cell in another chunk inside the bounds; its cost is `CellDistance` of the step. A link's cost is the cheapest center → opening → neighbor center walk, so a route's cost approximates cells walked and is never below the straight center distance (the A* heuristic).
- Nodes are whole chunks, as the spec says: a chunk is taken to join all its openings. Filed M6-06.F1 to split nodes by local area and bound `Route`'s search.
- Summaries are worked out lazily: a miss summarizes the chunk and its eight neighbors on worker goroutines, kept in chunk-key order. Nothing warms the whole world at start (about 3,500 chunks on the sample world); M6-20 can call `Prepare` if it wants that.
- The graph reads the world's cell source, not the map's edits. Today the only kept edits are the export tunnel's, which lie inside the pocket snapshot, so the graph matches the map (tested before and after an export). Edits made by future digging would not show until the next export.
- `explore` keeps the graph as an Ark resource (`ecs.NewResource[worldgen.Graph]`), so feature packages such as agents can read it without importing explore. It is built in `buildWorld` and `Reset` on an export. Saves skip resources, so nothing is saved.
- Files outside Deliverables: `internal/core/explore/explore.go`, `export.go`, `graph_test.go`; `docs/go-technical-design.md` ("Far agents") names the type and the resource.
- `TestTheGraphHasTheOpeningsOfTheLoadedChunksBeforeAndAfterAnExport` reads every cell of three small worlds twice and takes about 1.5 s.
- Follow-ups: M6-06.F1.
