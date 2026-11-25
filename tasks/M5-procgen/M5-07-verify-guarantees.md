---
id: M5-07
title: Verify stage and the guarantees
status: done
milestone: M5
kind: feature
depends_on: [M5-05]
discovered_by: null
size: M
---

# M5-07 · Verify stage and the guarantees

## Goal

A `verify` stage checks the generated deep against the guarantees and repairs it where it fails: every fixed place connects to the deep, and every place can reach every other. Property tests prove it on a handful of seeds.

## Context

- `docs/game-design.md` → "Procgen and exports" (Guarantees), "An export with the party below" (a route to at least one town).
- `docs/go-technical-design.md` → "Procgen stages" (verify); fzipp/astar, which pathing already uses.
- M5-05's tunnels.
- The brief's exit criterion: "The guarantees hold on a handful of seeds."

## Deliverables

- `internal/core/worldgen/verify.go`: the `verify` stage, its registration, and the chunk graph it reads.
- `content/data/geography.cue`: `verify` last.
- `internal/core/worldgen/verify_test.go`: property tests with `pgregory.net/rapid` over seeds.

## Spec

- **The check** runs over the chunk graph of walkable connectivity (chunks as nodes, open openings as edges, per M4's stepping rules at chunk borders), built from a summary of each chunk's openings, never the whole world at once. Components come from a flood fill. It is a pure function of the seed, so every chunk can compute the same result.
- **The repair** carves the shortest corridor (A* on the chunk graph with fzipp/astar, then cells) between the nearest disconnected components, carved deterministically so order independence holds. A repair never touches a fixed place's cells.
- **Guarantees tested:** every place has a walkable route to every other place; no walkable pocket is smaller than a threshold while another larger one exists (stranded pockets are joined or filled), threshold content.
- Materials appearing where they belong (coal near dead crawlspaces) waits for the content that has those materials.

## Acceptance criteria

- [x] A rapid property: for any of many seeds, a walk using the real stepping rules from any place's start reaches every other place's start.
- [x] Order independence still holds.
- [x] `go tool task check` passes.

## Out of scope

- Guarantees across an export (the pocket, a route to a town): M5-10.
- A route home from every cell for the party: not promised.

## Notes

- Deviation: the check runs over the whole ground before verify (once per source, memoized), not over per-chunk summaries; fine for the sample world, filed as M5-07.F1 for the large one. The repair uses a BFS over tunnel steps (as join's), not A* on a chunk graph.
- Files outside the list: `source.go` gains `Context.Before` (the chunks as earlier stages leave them) and `Context.Shared` (compute once per source and stage).
- Walkability counts only steps that can be made both ways, so a drop never joins areas. An opening's area is the standable ground it or its exit stands on or steps to. A pocket under `min_pocket` cells with no place cells is filled; larger cut-off pockets are joined.
- Step limits and `min_pocket` (24) are verify params in content, duplicating the party's step limits; they must match `party.cue`.
- The stage is not registered anywhere yet; M5-06 wires the procgen stages into the game.
- Test count is lowered to 12 checks (about 1.5 s each); `-rapid.checks` overrides.
- Follow-up filed: M5-07.F1.
- Package renamed `internal/core/procgen` to `internal/core/worldgen` and made a foundation package, at the owner's request: it imports only `content`, `api`, `hexmap` and `rng`, so `explore` and `game` may import it. The feature package `procgen` is kept for reseeding on exports and the party's pocket (M5-10), not yet written. Edited `AGENTS.md`, `.claude/rules/core.md`, `docs/go-technical-design.md` and the paths in M5-11, M5-12 and the two follow-ups to match. The `procgen` rng stream name is unchanged so goldens stay valid. Open PR 183 still imports `procgen` and needs updating from this branch.
