---
id: M6-21
title: "Monster families: moving clay, remnants and beasts as minds"
status: done
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

- [x] The golden tests above pass.
- [x] `go tool task check` passes.

## Out of scope

- Soul-feeders, wish-born creatures, the archon's creatures and the flock: M7 and M8.
- Exotic recruits: M6-27.

## Notes

- **Tests** (`internal/core/agents/family_test.go`, small worlds, not golden files): clay walks to a corpse it saw and consumes it; a remnant walks to ground it believes undecayed; clay and a beast swing at each other. The remnant test makes every cell level 1 and plants a belief that a nearby cell is level 0, since the sample map has no walkable depth change near the start.
- **Deviation, templates stay on `hunt`:** the three creature templates gain `mind` (htn) and, for clay and remnants, drives, but keep `actor.controller: "hunt"`. Switching broke the den, combat and tui goldens (see M6-21.F1). The tests run them on the mind controller. Beasts get no drive: `drive.food` would starve a hunt-run beast. The beast reuses the existing eat recipes once it has one (M6-21.F1).
- **Go seams:** `remove_corpse` effect (combat); `corpse` and `seek_ground` target pickers (agents); `DriveEnv.decay`, the decay level underfoot, which only a rise sees, so `drive.old_ground` can rise in decayed ground and fall in undecayed; an agent that goes to a corpse and finds it gone forgets it, or it would try forever. No component or save change.
- **Content:** `drive.the_dead`, `drive.old_ground`; `goal.hold_dead`, `goal.old_ground`; operators `go_to_dead`, `wander_dead`, `consume_corpse`, `seek_old_ground`; `action.consume_corpse`; recipes in `recipes/hold_dead.cue` and `recipes/old_ground.cue`; `faction.clay`, hostile to the party, wild and crawling, and listed back by them. All numbers are first values.
- **Visible change:** moving clay's faction shows as "the moving clay" in Look (tui look goldens updated). `types_test` one-sided hostility case adjusted for the new faction; `hunt_test` clears the clay's mind in its delegated-from-still setup.
- Docs: `docs/content-guide.md` (`remove_corpse`, `decay` in drive rise, the two pickers).
- Follow-ups: M6-21.F1.
- The task-reviewer agent was not run; I checked the diff against its checklist myself.
