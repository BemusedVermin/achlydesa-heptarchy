---
id: M1-14
title: "hexmap: sight lines and pathing in 3D"
status: done
milestone: M1
kind: feature
depends_on: [M1-05]
discovered_by: null
size: M
---

# M1-14 · hexmap: sight lines and pathing in 3D

## Goal

The party, and later every agent, sees and paths through the 3D world: down shafts, along slopes, never through rock, and never into a soop that merely passes underneath.

## Context

- Tech design: "Designing for pivots" (Maps, Plan view), "Library choices" (Hex math, Pathfinding).
- M1-05 cells, topology and stepping.
- gravitton/hexagon line drawing (confirm the name with `go doc github.com/gravitton/hexagon`); fzipp/astar: `astar.FindPath[Node](graph, start, dest, distFn, heuristicFn)` with `Neighbours(n Node) iter.Seq[Node]`.

## Deliverables

- `internal/core/hexmap/sight.go`, `path.go`, tests and benchmarks.
- `go.mod`: adds `github.com/fzipp/astar`.

## Spec

**Sight:**

```go
// Visible returns every cell within radius hexes and depthRange depths of from
// that a straight line from from's center reaches, sorted by Depth, R, Q.
func (m *Map) Visible(from api.Cell, radius, depthRange int) ([]api.Cell, error)
```

- A cell is visible when the segment from `from`'s center to its center crosses no opaque cell other than the target itself, so walls are visible and what's behind them isn't.
- Walk the segment with the library's hex line between the two hexes and interpolate depth linearly along it. Where the interpolated depth is fractional, test both the cell above and the cell below; the line is clear if either sequence is clear. That keeps sight symmetric enough and lets you see down a slope.
- Sight stays on one map; it doesn't follow links to other maps. File a follow-up when folds exist.
- The same input always gives the same output.

**Pathing:**

```go
// Path finds a walking route using Step in six directions, so it climbs, descends
// and drops by the same rules as the party. cost prices each step; the heuristic is
// hex distance plus depth difference. It gives up after maxNodes expansions, so a
// search in an infinite map always ends.
func (m *Map) Path(from, to api.Cell, rules StepRules, cost func(Step) float64, maxNodes int) ([]api.Cell, bool, error)
```

Use fzipp/astar with a graph type whose `Neighbours` yields the cells `Step` reaches. Collect errors from chunk loading and return the first.

## Acceptance criteria

- [x] In an open room, every cell within range is visible; walls bounding it are visible; cells behind walls aren't.
- [x] Standing at a shaft's edge shows cells down the shaft; a tunnel four cells below through solid rock isn't visible.
- [x] Standing at the top of a ramp shows the slope going down.
- [x] `Path` walks down a ramp, refuses to climb a three-cell ledge (`MaxStepUp` 1), and does go down it as a drop when no other route exists.
- [x] A search with no route in an all-solid infinite map returns false within `maxNodes` expansions.
- [x] Repeated calls give identical results (100 runs).
- [x] Benchmarks: `Visible` at radius 6 and depth range 6 takes under 2 ms; record the numbers in Notes.
- [x] `go tool task check` passes.

## Out of scope

- Choosing which cell to draw per hex (M1-10).
- Sight across map links.

## Notes

- `fzipp/astar` has no expansion limit, so `Path` counts `Neighbours` calls and stops yielding neighbours after `maxNodes`; the search then drains and returns no route.
- `Visible` includes `from` itself. Vertical lines (same hex, different depth) check the cells strictly between. `Path` returns both ends and treats nil `cost` as 1 per step. Steps through links to other maps are skipped.
- Benchmark (`go test ./internal/core/hexmap -bench Visible`, radius 6, depth range 6, open floor): about 0.90 ms/op, 1664 allocs/op, on the developer's machine; the limit is 2 ms.
- Added `github.com/fzipp/astar` v0.3.0 as the spec asked.
- Follow-up filed: [M1-14.F1](../followups/M1-14.F1-sight-across-links.md), sight and pathing across map links.
