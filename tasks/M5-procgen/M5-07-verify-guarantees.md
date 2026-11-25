---
id: M5-07
title: Verify stage and the guarantees
status: todo
milestone: M5
kind: feature
depends_on: [M5-04, M5-05]
discovered_by: null
size: M
---

# M5-07 · Verify stage and the guarantees

## Goal

A `verify` stage checks the generated deep against the guarantees and repairs it where it fails: every fixed place connects to the deep, and every place can reach every other. Property tests prove it on a handful of seeds.

## Context

- `docs/game-design.md` → "Procgen and exports" (Guarantees), "An export with the party below" (a route to at least one town).
- `docs/go-technical-design.md` → "Procgen stages" (verify); gonum for the guarantees.
- M5-04's region graph; M5-05's tunnels.
- The brief's exit criterion: "The guarantees hold on a handful of seeds."

## Deliverables

- `internal/core/procgen/verify.go`: the `verify` stage and its registration.
- `content/data/geography.cue`: `verify` last.
- `internal/core/procgen/verify_test.go`: property tests with `pgregory.net/rapid` over seeds.

## Spec

- **The check** runs over the chunk graph of walkable connectivity (chunks as nodes, open openings as edges, per M4's stepping rules at chunk borders). It is a pure function of the seed, so every chunk can compute the same result.
- **The repair** carves the shortest corridor (A* on the chunk graph, then cells) between the nearest disconnected components, carved deterministically so order independence holds. A repair never touches a fixed place's cells.
- **Guarantees tested:** every place has a walkable route to every other place; no walkable region is smaller than a threshold while another larger one exists (stranded pockets are joined or filled), threshold content.
- Materials appearing where they belong (coal near dead crawlspaces) waits for the content that has those materials; this task checks only that a material only appears in regions that list it.

## Acceptance criteria

- [ ] A rapid property: for any of many seeds, a walk using the real stepping rules from any place's start reaches every other place's start.
- [ ] Order independence still holds.
- [ ] `go tool task check` passes.

## Out of scope

- Guarantees across an export (the pocket, a route to a town): M5-10.
- A route home from every cell for the party: not promised.

## Notes

Filled in by the implementer: assumptions, deviations, files touched outside Deliverables, follow-ups filed.
