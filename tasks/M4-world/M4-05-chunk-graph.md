---
id: M4-05
title: "hexmap: the coarse chunk graph"
status: todo
milestone: M4
kind: seam
depends_on: [M4-04]
discovered_by: null
size: M
---

# M4-05 · hexmap: the coarse chunk graph

## Goal

A coarse graph of the world, made of chunks and the openings between them, lets an agent far from the party find a route across the world without pathing cell by cell.

## Context

- `docs/game-design.md` → "Keeping it affordable" (Off screen, less often and coarser).
- `docs/go-technical-design.md` → "Level of detail" (the graph is chunks and the openings between them); "Library choices" (gonum graph, fzipp/astar).
- What exists: `internal/core/hexmap/path.go` (A* over cells), `step.go` (stepping rules), `chunk.go`, and M4-04's bounds.

## Deliverables

- `internal/core/hexmap/graph.go`: the chunk graph, built lazily per chunk and rebuilt for the chunks an edit touches.
- `internal/core/hexmap/graph_test.go`: table tests on small hand-built worlds, and a rapid property (see Acceptance).

## Spec

Suggested shapes:

```go
// Opening is a standable cell on a chunk's face that a body can step across into the next chunk.
type Opening struct {
	From, To api.Cell
}

// CoarseRoute is a route across chunks: the openings to pass, in order, and the steps each leg costs.
type CoarseRoute struct {
	Openings []Opening
	Steps    []int
}

func (m *Map) CoarseRoute(from, to api.Cell) (CoarseRoute, bool, error)
```

- A **node** is a connected set of standable cells within one chunk (a chunk can hold several, such as two soops crossing). An **edge** joins two nodes in neighbouring chunks through at least one opening; its cost is the step count between representative openings, measured with the cell pathing inside the chunk. Firm: nodes are per connected component, not per chunk, or routes would pass through walls.
- Stepping rules are the same as `step.go` (climbs and drops included), so a coarse route never needs a step the cell pathing would refuse.
- An edit (`Set`) marks its chunk and its neighbours stale; they are rebuilt on next use. The graph is derived and never saved or hashed.
- Ties are broken by chunk key, then cell order, so a route never depends on map iteration. Firm.
- gonum's graph or fzipp/astar for the search: use what `path.go` uses unless gonum's shortest paths are simpler, and name it in Notes.

## Acceptance criteria

- [ ] A rapid property: for random small worlds, `CoarseRoute` finds a route exactly when cell pathing finds one, and its step total is within a fixed factor of the cell path's length (state the factor in the test).
- [ ] A test: an edit that walls off an opening removes the route, and opening it again restores it.
- [ ] A benchmark: a route across a 400×400 world, from a warm graph, takes under 1 ms.
- [ ] `go tool task check` passes.

## Out of scope

- Agents travelling the graph: M4-09.
- Procgen's region graph and its guarantees: M5.

## Notes

Filled in by the implementer.
