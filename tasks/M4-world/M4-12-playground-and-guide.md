---
id: M4-12
title: Playground and test places at human scale; the content guide
status: todo
milestone: M4
kind: content
depends_on: [M4-07, M4-08]
discovered_by: null
size: M
---

# M4-12 · Playground and test places at human scale; the content guide

## Goal

The playground is one place-built world to walk around at 1.5 m a cell, with the test rooms joined by tunnels rather than gates. The content guide says how to build in the one world.

## Context

- `docs/content-guide.md` → "A place" (after M4-07), "The playground", "Regions" (after M4-08).
- `docs/game-design.md` → "The hex grid" (Human scale; Headroom).
- What exists after M4-07: `content/data/places/play_*.cue` and `test_*.cue`, unjoined; `start` begins in the hall place.

## Deliverables

- `content/data/places/play_*.cue`: the playground's rooms rebuilt as one connected area: the hall, the den, the descent (soops down to the deep) and the feeders' room, joined by tunnels, with a small town room holding containers. `start` begins in it.
- `content/data/places/test_*.cue`: test fixtures stay small and keep their IDs; join or move them only where a test needs it. Tests that start on a fixture start at its place.
- `docs/content-guide.md`:
  - State the scale: a cell is 1.5 m across and 1.5 m deep, and a body fills one cell.
  - A recipe for placing a piece in the world (offset, even row, depth), for regions, and for placing bodies.
  - A recipe for a structure: walls, a door, a room, a shaft, a container.
  - "The playground" table rewritten for the connected playground.
  - Delete every remaining word about maps, gates, stock marks and populations.

## Spec

- The playground keeps what each old map showed (the moving clay in the hall, the remnant and the beast in the den, the soops and the clay at depth in the descent, the feeders), so nothing a later task demos disappears.
- Test places are fixtures (`test.` IDs): change them only together with the tests that use them.

## Acceptance criteria

- [ ] `go tool task run -- play --seed 1` starts in the playground, and a test walks from the hall to the den and to the bottom of the descent with no jump in position.
- [ ] `git grep -i -E "\bmaps?/|gate|stock mark|populations" -- docs/content-guide.md` finds nothing.
- [ ] `go tool task check` passes.

## Out of scope

- The sample world: M4-11.

## Notes

Filled in by the implementer.
