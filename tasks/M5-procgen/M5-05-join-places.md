---
id: M5-05
title: Join the generated deep to the fixed places
status: done
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

- [x] On the three test places and several seeds, every opening of every place is the end of a carved tunnel (a test walks outward with the real stepping rules).
- [x] Order-independence holds for chunks tunnels cross.
- [x] `go tool task check` passes.

## Out of scope

- Guaranteeing every place reaches the deep: M5-07.
- Using the real large world: M5-06.

## Notes

- Deviation at the owner's direction: the Spec says an opening is any open cell on a place's boundary touching fill. Read literally that gave 33 openings on the sample world, because the test descent's wall-less soop and ledge layers make every cell an opening. Openings are now marked in content: a place lists mark characters in `openings`, and `explore` reports each such cell (`content.PlaceDef.Openings`, checked in `internal/content/places.go`). `hexmap.Footprint` holds those marked cells. A marked cell with no neighbor outside the places is an error from `procgen.NewSource`. The sample descent marks three: both ends of its depth-2 ledge and the end of its depth-4 soop. M5-05.F1 was folded into this task at the owner's request; he will triage its file.
- Footprints: `hexmap.Footprint` (foundation layer, so `explore` can return it without importing `procgen`) and `Fixed.Places()`, built in `explore/world_source.go`. `procgen.Source` computes `Openings()` once and gives them to stages in `Context.Openings`. `Context.Siblings` gives a stage the other stages' params; join reads the caves' params from it. Caves got a `caveField` so join can ask whether a cell is cave from the seed alone.
- Tunnels walk level, one climb or one step down at a time (the real step rules), toward the nearest cave found within `reach`. Each step is drawn with weight `1 + bias * progress`. Width is added around the walk last, sparing the ground under a climb's landing, and a climb is not offered where the walk already opened that ground. Every chunk recomputes the tunnels near it, so chunk order cannot matter (rapid test).
- With no cave within `reach` an opening gets only its exit cell; M5-07 repairs that.
- Content: `join` sits between `caves` and `sinks`, so sink footprints, which set every cell, win over tunnels. `width: 2`, `reach: 24`, `bias: 1.5` are first values to tune.
- The acceptance test checks that every exit is open and has a real step outward, and that openings with a cave within a third of `reach` reach ground touching one. Reaching a cave from every opening is not guaranteed.
- `docs/content-guide.md` describes `openings`.
- The stage is not registered in the game yet (M5-06).
