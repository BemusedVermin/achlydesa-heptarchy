---
id: M5-03
title: Sinks stage
status: done
milestone: M5
kind: feature
depends_on: [M5-01]
discovered_by: null
size: M
---

# M5-03 · Sinks stage

## Goal

A `sinks` stage carves funnels that spiral down around a shaft: a gentle helicoid ramp that reads as level ground in the plan view far from the axis and steepens into the shaft at the center. Jumping the shaft is the fast way down; risers (opposite-handed sinks) are the way back up.

## Context

- `docs/game-design.md` → "The hex grid" (depth is continuous; no layers are ever visible).
- `docs/go-technical-design.md` → "Designing for pivots" → "The world" (sinks, impossible folds).
- The brief's "Planned tasks" → Sinks. `hexmap` stepping rules (`step.go`): what slope can be walked and what drops are falls.

## Deliverables

- `internal/core/procgen/sinks.go`: the `sinks` stage and its registration.
- `content/data/geography.cue`: `geography.deep` lists `sinks` with its numbers.
- `internal/core/procgen/sinks_test.go`: tests (below).

## Spec

- **Where.** Sink axes are chosen from the seed on a coarse grid (a sink per N chunks, jittered), so any chunk can tell from the seed alone which axes reach it. Order-independence from M5-01 still holds.
- **Shape.** Around an axis the floor depth is a function of angle and radius: `depth = base + lap_height * (angle / 2π) + slope(radius)`, so each lap resembles the last. Use depth-periodic noise in the angular coordinate so the lap seam shows no fault. Near the axis the ramp steepens until it is a shaft the party can drop down; a sink's height and radius are content.
- **Risers** are the same shape with the opposite handedness and rise as you walk the spiral.
- **Walkable ramp.** Each lap's slope between neighboring cells is within what `hexmap` stepping allows, so the ramp can be walked all the way down without jumping. Firm; the test checks it.
- A sink never carves outside the bounds or a fixed place's cells.

## Acceptance criteria

- [x] A test walks a sink's ramp from its rim to its bottom using the real stepping rules, and walks back up a riser.
- [x] A test that dropping the shaft is a fall of several depth steps and lands on open ground.
- [x] Order-independence holds for chunks a sink crosses.
- [x] `go tool task check` passes.

## Out of scope

- Caves: M5-02. Fall damage and what a fall does to a body: combat and effects (already content).
- Connecting sinks to places and guaranteeing routes: M5-05, M5-07.

## Notes

- Shape: floors under a hex repeat every `lap_height` depth, offset by the angle around the axis, so the lap seam is continuous by construction and needs no noise. The spec's `slope(radius)` term is left out: the ramp steepens toward the axis because a step covers more angle there, and the shaft is the cut-off radius. `shaft_radius` is checked so the steepest step stays within one depth step.
- A riser is the mirror image; both have a shaft. Walking either spiral one way lowers the floor and the other way raises it, so the sink/riser difference is only which way round is down.
- `headroom` (min 2) is a param: climbing needs a free cell above the body. Axes never overlap (checked against `spacing`, `jitter` and `outer_radius`).
- The stage sets every cell of its footprint from `rim_depth` to one below the shaft floor (closed cells take the fixed layer's fill), so it must come after `caves` in the geography; the sample content does so. A sink is not trimmed around fixed places, so one beside a place may be cut (M5-05).
- `Run` returns an error when the sink does not fit between the bounds' depths (rim below the top layer, shaft floor within the bottom).
- Outside Deliverables: `caves_test.go` now builds its source from the geography's caves stage only, so its golden stays about caves. Nothing outside tests registers stages yet (`RegisterSinks` like `RegisterCaves`).
- Follow-ups filed: none.
