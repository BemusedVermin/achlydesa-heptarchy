---
id: M6-21
title: "Monster families: moving clay, remnants and beasts as minds"
status: todo
milestone: M6
kind: content
depends_on: [M6-05, M6-07, M6-15, M6-16, M6-19, M6-20]
discovered_by: null
size: M
---

# M6-21 · Monster families: moving clay, remnants and beasts as minds

## Goal

Each creature family has the drives and recipes that make its sketch in the design true by itself, with no code for any family.

## Context

- `docs/game-design.md` → "Monsters with needs" (the sketches table: moving clay holds the dead, remnants retreat from decay, families fight each other).
- Added back from the reverted M3: `tasks/M3-agents/M3-18-monster-families.md` (read for what to reuse).
- M6-05 (corpse targets), M6-07 (the dead and undecayed ground as landmarks), M6-15, M6-16, M6-19, M6-20.

## Deliverables

- `content/data/drives.cue`: `drive.the_dead` (moving clay), `drive.old_ground` (remnants), `drive.hunger` for beasts reuse `drive.food`.
- `content/data/recipes.cue` and `operators.cue`: `operator.consume_corpse` (an action that takes a corpse as its target and removes it, giving `drive.the_dead` relief), recipes for each family.
- `content/data/templates.cue`: the creature templates gain `mind` (planner, drives with typical weights); the existing `hunt` creatures keep fighting through M6-19.
- Hostility between families through `factions.cue` (moving clay vs beasts), so families fight each other when they meet.
- Tests (golden, small world): moving clay walks to a corpse it saw and consumes it; a remnant moves toward ground it believes undecayed; two hostile families fight on meeting.

## Spec

- **Firm:** no family has a system. Each is drives, operators, recipes and hostility.
- **Firm:** a creature that believes nothing wanders, so the first sighting teaches it (M6-12).
- **Suggestion:** one recipe per family and per drive, small.

## Acceptance criteria

- [ ] The golden tests above pass.
- [ ] `go tool task check` passes.

## Out of scope

- Soul-feeders, wish-born creatures, the archon's creatures and the flock: M7 and M8.
- Exotic recruits: M6-27.

## Notes

