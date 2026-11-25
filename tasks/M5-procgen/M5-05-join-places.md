---
id: M5-05
title: Join the generated deep to the fixed places
status: todo
milestone: M5
kind: feature
depends_on: [M5-02]
discovered_by: null
size: M
---

# M5-05 · Join the generated deep to the fixed places

## Goal

A `join` stage carves the generated deep into each fixed place's openings, so nothing shows where a place ends and the deep begins: the party walks out of a place's opening into a tunnel or chamber with no seam.

## Context

- `docs/game-design.md` → "Procgen and exports" (grown right up to the edges of the fixed places), "World and towns" (towns are open to the deep).
- The brief's "Planned tasks" → Stage pipeline (joining).
- M5-01's `Fixed`; M4-06's places; the place openings are the open cells on a place's boundary that touch fill.

## Deliverables

- `internal/core/procgen/join.go`: the `join` stage and its registration.
- `content/data/geography.cue`: `join` listed after `caves`.
- `internal/core/procgen/join_test.go`: tests (below).

## Spec

- **Openings.** For each fixed place, the open cells of its boundary adjacent to fill outside the place. `Fixed` exposes the place footprints (suggestion: add `Places() []Footprint` to `Fixed`).
- **Carving.** From each opening a tunnel is carved into the surrounding ground until it meets generated open ground (a cave). It wanders: the path follows a seeded random walk biased toward the nearest cave, never a straight line. Tunnel width and the maximum reach are content.
- **Order independence.** Which tunnels reach a chunk is computed from the seed and the place's openings alone, so a chunk can find every tunnel that crosses it without reading its neighbors.
- A tunnel never alters a fixed place's cells, and never leaves the bounds.
- No guarantee of reaching a cave when none is near: M5-07 repairs that.

## Acceptance criteria

- [ ] On the three test places and several seeds, every opening of every place is the end of a carved tunnel (a test walks outward with the real stepping rules).
- [ ] Order-independence holds for chunks tunnels cross.
- [ ] `go tool task check` passes.

## Out of scope

- Guaranteeing every place reaches the deep: M5-07.
- Using the real large world: M5-06.

## Notes

Filled in by the implementer: assumptions, deviations, files touched outside Deliverables, follow-ups filed.
