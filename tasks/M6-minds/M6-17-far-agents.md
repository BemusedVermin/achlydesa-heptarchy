---
id: M6-17
title: "Far agents: goals only, over the chunk graph"
status: done
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

- [x] The tests and the benchmark above pass.
- [x] A chunk whose two openings lie in separate areas is not crossed between them, and M6-06's tests still pass.
- [x] `Route` with a node limit returns instead of searching on past it when the target is unreachable.
- [x] `go tool task check` passes.

## Out of scope

- Population at start: M6-20.

## Notes
- Folded in M6-06.F1 (triaged 2026-11-27): area-aware graph nodes and the `Route` node limit. Size raised from M to L; the owner may split the graph change into its own task.
- **Budget for M6-28:** `BenchmarkPhaseWithFarAgents` (`internal/core/agents/lod_bench_test.go`: 100 near and 900 far minds, half of them thirsty, one phase per op) measures about 900 ms per phase on a Ryzen 9 9950X3D. The far agents are about 3% of that (around 30 ms, mostly first summaries of the graph). The rest is near agents' `goToPicker` cell paths (about 40%, M6-17.F1) and the per-entry combat alert (about 27%, M6-17.F2). Budget to enforce on the CI runner: **2,500 ms per phase**, until F1 and F2 lower it. I could not measure on the CI runner; the number assumes it is about 2.5 times slower than this machine.
- **Off the timeline:** new `timeline.Park` and `timeline.Unpark`, with a `comp.Parked` marker so a save keeps it. A parked body has no entry and is skipped by `schedule`. Not in the Deliverables, but the timeline had no way to take a body off it.
- **Far state:** new `comp.Far` (next far decision tick, and the stops ahead with their arrival ticks). Save version 15.
- **Cadence:** the far system is a per-entry system that does nothing until the party changes chunk or a far decision is due; each far body keeps its own decision tick in `comp.Far`, so a save loses nothing. The cached party chunks and earliest due tick live in a resource that a loaded game rebuilds to the same result (the save/reload test checks the hash).
- **Demotion** happens at a near mind's own decision, once its channels are free; **promotion** when the party's chunks change or a far stop lands in a near chunk. A body with no position counts as near.
- **Coarse operators:** a step whose picker is `go_to` routes over the graph (stops at chunk centers, the last at the target; arrival ticks are the route cost times the body's step ticks, spread evenly over the stops). `wander` hops to a random neighboring chunk from the agent's derived stream. Any other step applies its action's `on_start` and `on_land` effects at once to the step's target or the agent, after `effects.Check`, skipping the action's `requires` and range. Take, put and consume content (M6-13) is not done yet, so the test drinks through a test action with `satisfy`; M6-17.F4 covers far targets for take and put.
- **Planning** runs the body's own registered planner on other goroutines from `Snapshot`s; adoption, claims and routing run on the world's goroutine in UID order. Routing stays on the world goroutine because `Graph` caches summaries and is not safe for concurrent use; its summaries are still computed on other goroutines.
- **Graph:** `ChunkOpening` gains `FromArea` and `ToArea`; `OpeningSummary` gains `Areas`, `Joins` (one-way steps between areas, such as drops) and `Border` (areas of border cells only, not all 4,096 slots). `Summarize` floods the whole chunk, so it does more work than before. `ToArea` is filled in by `Graph.Summary` from the neighbor's summary. `Route(from, to, maxNodes)`: 0 means no limit. Link costs still count from chunk centers, so M6-06's cost test is unchanged. `Route` may start from any area of `from` (M6-17.F3). New exported `worldgen.Center`.
- **Files outside the Deliverables:** `internal/core/timeline/timeline.go`, `internal/core/comp/{comp,mind}.go` and their schemas, `internal/content/{types,minds}.go` (`agents.far_ticks`, required when goals exist), `content/data/goals.cue` (`far_ticks: 6`, a tenth of a 60-tick phase on average), `internal/core/agents/mind.go` (plan making split so far agents reuse it; demotion), `internal/core/save` (version 15), `internal/core/explore/graph_test.go` (compares openings without their areas), `docs/go-technical-design.md` (Far agents paragraph). `go tool task schema` also regenerated two stale `api` schema files from master (`Weight`, `Unreachable`, `Confirm`).
- Follow-ups: M6-17.F1, M6-17.F2, M6-17.F3, M6-17.F4.
