---
id: M4-04
title: "hexmap: one bounded world store"
status: in-progress
milestone: M4
kind: seam
depends_on: []
discovered_by: null
size: S
---

# M4-04 · hexmap: one bounded world store

## Goal

`hexmap` can hold one world with edges: cells outside its bounds are solid, and chunks that are all one terrain cost one value, so a world hundreds of hexes across stays small in memory.

## Context

- `docs/game-design.md` → "World and towns" (The world is bounded; Small at first), "The hex grid" (Human scale).
- `docs/go-technical-design.md` → "Designing for pivots" → "The world"; "Core packages" → `hexmap`.
- What exists: `internal/core/hexmap/map.go` (`Map`, `At`, `Set`, `Standable`), `chunk.go` (`Chunk` stores `ChunkSize³` `uint16` slots), `topology.go`.

## Deliverables

- `internal/core/hexmap/bounds.go`: `Bounds` and its methods.
- `internal/core/hexmap/chunk.go`: a chunk filled with one terrain stores no slot array until it is set.
- `internal/core/hexmap/map.go`: a map may have bounds; `At` outside them returns the map's edge terrain (solid, opaque), and `Set` outside them is an error.
- Tests in `internal/core/hexmap`, including a rapid property: for any cell, `At` inside bounds equals the source's terrain, and outside equals the edge.

## Spec

Suggested shapes:

```go
// Bounds is the box of cells a world holds: hexes in offset rows and columns, and depths, all inclusive.
type Bounds struct {
	MinCol, MaxCol, MinRow, MaxRow int
	MinDepth, MaxDepth int
}

func (b Bounds) Contains(c api.Cell) bool
func (m *Map) SetBounds(b Bounds, edge Terrain)
```

- Bounds are in odd-r offset coordinates, the same as content layers, so "400 hexes across" is a column count. Firm.
- A map with no bounds behaves as today (endless), so nothing else changes in this task.
- Sight and pathing need no change: an out-of-bounds cell is solid and opaque.
- `Chunk` keeps its API; `Set` allocates the slot array on the first cell that differs from the fill.

## Acceptance criteria

- [ ] A benchmark or test shows a 400×400×60 world of uniform fill, all chunks touched, holds under 5 MB of chunk memory.
- [ ] Stepping off the edge of a bounded map is refused by `Standable` and by pathing.
- [ ] `go tool task check` passes.

## Out of scope

- Removing map IDs and the maps registry: M4-07.
- Content for the world's size: M4-06.
- The chunk graph: M4-05.

## Notes

Filled in by the implementer.
