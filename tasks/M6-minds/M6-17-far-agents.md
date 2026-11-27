---
id: M6-17
title: "Far agents: goals only, over the chunk graph"
status: todo
milestone: M6
kind: feature
depends_on: [M6-06, M6-12, M6-14]
discovered_by: null
size: L
---

# M6-17 · Far agents: goals only, over the chunk graph

## Goal

An agent far from every party member decides on a coarser cadence at goal level and moves along the chunk graph, so 1,000 individuals stay affordable. Near agents are unchanged.

## Context

- `docs/go-technical-design.md` → "Level of detail" (Far agents; first values 2 chunks near, `agents.far_ticks` a tenth of a phase); "Determinism rules".
- `docs/game-design.md` → "Keeping it affordable".
- M6-06 gives the graph. Far chunks may not be loaded, and sight and cell pathing need loaded chunks.
- Folded in from M6-06.F1: the graph's nodes are whole chunks, so `Graph.Route` can pass through a chunk where the entry and exit openings are not joined inside it, and its cost can be well under the real walk. `Route` to an unreachable chunk also summarizes every chunk the search reaches, thousands on a 400 × 400 world. `internal/core/worldgen/verify.go`'s `reader.scan` already flood-fills a chunk into local areas; the same split can tag each opening.

## Deliverables

- `internal/core/agents/lod.go`: `Near(w, body)` by chunk distance to any party member (content radius, `world.near_chunks` reused); a far agent is off the per-entry timeline and decided by a per-`agents.far_ticks` system.
- A far decision: choose a goal (M6-14) and, if it needs a landmark the agent believes in, route to that chunk by the graph and set an arrival tick from route cost; arrival at a chunk updates its beliefs from a per-chunk summary of landmarks (never from sight); at the destination it acts on the stock it reaches with a coarse operator (take, put, consume) without cell paths.
- Promotion: when a far agent's chunk becomes near it is put on the nearest standable cell and starts deciding as a body; demotion is the reverse.
- `internal/core/worldgen/graph.go`: openings carry the local area of each end; graph nodes become (chunk, area); `Route` still returns chunk keys and a cost; `Route` takes an optional node limit, like `hexmap.Map.Path`'s `maxNodes`. Remove the `TODO(M6-06.F1)` there.
- Decisions are computed on other goroutines from snapshots and applied in UID order; far agents draw no shared stream.
- A benchmark: 1,000 agents with a 1:9 near/far split, one phase, with the budget stated in Notes (milliseconds per phase on the CI runner) as the number M6-28 enforces.
- Tests: a far agent reaches a landmark chunk and drinks; promotion places it on standable ground; the same seed gives the same arrival ticks.

## Spec

- **Firm:** far agents never call sight or cell pathing.
- **Firm:** a body's inventory, drives and beliefs are the same entity components near or far; only the cadence and the movement change. Nothing becomes a count.
- **Firm:** summaries stay pure functions of the seed and export count, computed on other goroutines and applied in chunk-key order.
- **Suggestion:** keep a per-chunk area table only for the border cells, not all 4096 slots.
- **Firm:** this adds no outcome system: a far agent that needs water goes where it believes water is, or wanders if it believes nothing.

## Acceptance criteria

- [ ] The tests and the benchmark above pass.
- [ ] A chunk whose two openings lie in separate areas is not crossed between them, and M6-06's tests still pass.
- [ ] `Route` with a node limit returns instead of searching on past it when the target is unreachable.
- [ ] `go tool task check` passes.

## Out of scope

- Population at start: M6-20.

## Notes
- Folded in M6-06.F1 (triaged 2026-11-27): area-aware graph nodes and the `Route` node limit. Size raised from M to L; the owner may split the graph change into its own task.
