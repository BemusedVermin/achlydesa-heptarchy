---
id: M4-11
title: The human-scale sample world
status: todo
milestone: M4
kind: content
depends_on: [M4-02, M4-03, M4-08]
discovered_by: null
size: M
---

# M4-11 · The human-scale sample world

## Goal

The sample world is one small, hand-built continuous world at 1.5 m a cell: a shallow town and a lake town, built of walls, rooms, doors, wells and cisterns, joined by a stretch of deep that takes real time to cross. Every creature and person in it is a placed body.

## Context

- `docs/game-design.md` → "The hex grid" (Human scale; Headroom), "World and towns" (Small at first), "People with needs", "Monsters with needs".
- The brief's "Decisions" (size, headroom, containers).
- What exists after M4-07 and M4-08: `content/data/places/sample_*.cue` (the old four maps, side by side and unjoined), regions per place, containers from M4-02, recruits from M4-03.

## Deliverables

- `content/data/world.cue`: the world's size, small: about 250×250 hexes and 40 depth steps, enough for the places below. Provisional.
- `content/data/places/sample_*.cue`, rebuilt:
  - **The shallow town:** about 100 hexes across at a shallow depth. Houses with walls, rooms and doors, a well as a shaft down to water, cisterns and larders as containers, residents with homes and jobs, generics and hirable recruits.
  - **The lake town:** deeper, beside a lake of water cells, with the same kinds of building.
  - **The deep between them:** hand-built tunnels and chambers that take at least a few in-game hours to walk, with moving clay, remnants, soul-feeders and the archon warden placed as bodies. It is a stand-in that M5's procgen replaces.
- `content/data/regions.cue`: the two towns as `crawlspace` regions, and the deep as one or two regions with decay rising with depth.
- If water as a terrain doesn't exist, a `terrain.water` (not solid, not opaque) and what drinking from it takes. If a behavior can't drink from a terrain cell, keep the water in the well's container and say so in Notes (a follow-up for drinking from terrain).
- Tests in `internal/core/agents` or `internal/core/game` that start on `start.sample`: the counts of bodies by template, and a path exists from the shallow town to the lake town.
- `docs/content-guide.md` → "The sample world": the new layout.

## Spec

- **No cell stands for a structure.** A well is a shaft of open cells with water at the bottom and a container or water cells to draw from; a house is walls and a door. Firm.
- Bodies stand in one cell, standing or crawling (headroom decision).
- Placeholder numbers stay placeholders for M9.
- Authoring large places: a helper in CUE (a comprehension that draws a rectangle of walls with a door) is welcome if it keeps the files readable; it lives in `places/` as a hidden definition.

## Acceptance criteria

- [ ] `go run ./cmd/harness run --start start.sample --steps 200` runs, and a test shows the party can walk from the shallow town's start to the lake town (a path exists).
- [ ] A test: residents of the shallow town haul water from the well to a cistern and drink, on screen.
- [ ] `go tool task content` and `go tool task check` pass.

## Out of scope

- The playground and test places: M4-12.
- The full walk and the exit checks: M4-14.
- Generating the deep: M5.

## Notes

Filled in by the implementer.
