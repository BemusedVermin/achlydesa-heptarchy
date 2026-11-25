---
id: M3-18
title: Monster families as content
status: done
milestone: M3
kind: content
depends_on: [M3-05, M3-10]
discovered_by: null
size: M
---

# M3-18 · Monster families as content

## Goal

The design's monster sketches live on their needs: moving clay gathers at bodies, remnants retreat from decayed ground, soul-feeders hunt the living and prefer the dying, and the archon's creatures guard believers and spare them. Each family also drifts off screen.

## Context

- `docs/game-design.md` → "Monsters with needs" (the sketches table and what falls out of it), "Players can use needs".
- The agents model in [README.md](README.md), all sections.
- M3-03 to M3-05 (behaviors, primitives, chooser), M3-09 (`FamilyDef`), M3-10 (drift and attrition).

## Deliverables

- `content/data/templates.cue`: moving clay, remnant and a new soul-feeder on the `agents` controller with needs and behaviors; a believer resident and an archon's creature.
- `content/data/families.cue`, `behaviors.cue`, `needs.cue`, `factions.cue` (an archon faction not hostile to believers).
- `content/data/bodies.cue`, `materials.cue` as the soul-feeder needs.
- The playground: one room per family, reachable by gates.
- Tests; M2's golden encounters updated only where a creature's controller changed, with the reason in Notes.

## Spec

- Firm: all content. A family needs no Go of its own; a gap in the primitives becomes a follow-up.
- Moving clay: need `need.hold_dead`, behavior `consume` toward dead bodies; drift toward regions with dead bodies.
- Remnants: need `need.undecayed`, urge from the region's `decay`; behavior `leave` toward the less-decayed neighbour; drift toward low decay.
- Soul-feeders: hunt the living with the agents chooser preferring `target.dying`.
- The archon's creatures: `guard` around believers; hostility by faction spares believers.
- M2's lethality targets (`internal/harness/lethality_test.go`) still hold, or Notes says which number moved and why.

## Acceptance criteria

- [x] Exit criterion: a test kills a creature near moving clay; the clay walks to the corpse.
- [x] A remnant in a decayed region moves toward its less-decayed neighbour, on screen and off.
- [x] An archon's creature attacks a party member who comes near a believer and leaves the believer alone.
- [x] In the playground, each family's room shows its behavior.
- [x] `go tool task check` passes.

## Out of scope

- Wish-born creatures and the flock (they need seats (M4) and the surface's schedule, later).
- Grudges (M6).

## Notes

- Content: `need.hold_dead`, `need.undecayed`, `need.soul`; `action.hold`; behaviors `behavior.clay.hold`, `behavior.remnant.leave`, `behavior.feeder.hunt`, `behavior.archon.guard`, `behavior.monster.engage`; templates for clay, remnant (both moved from `hunt` to `agents`), a soul-feeder (`body.soul_feeder`, `material.soul_husk` weak to Lees), an archon's warden and a believer; `faction.archon` (hostile to the party, which now lists it back) and `faction.believers`; families `family.moving_clay`, `family.remnant`, `family.soul_feeder`, `family.archon_wardens`, `family.believers` with drift. All numbers are placeholders.
- Playground: new rooms `map.play.feeders` and `map.play.shrine`, gates 5 and 6 on the hub (now 15 wide). Decay is region data: den 0.8, hub 0.2. The hall is the clay room, the den the remnant room.
- Deviation, one Go seam: `region` (the region's record, so `region.props.decay`) is now visible to `consider` and to need `rise` and `urge`. Without it a remnant on screen cannot read decay. Documented in `docs/content-guide.md`. No new component, so no save bump.
- Creatures wait unless a hostile is within 8 hexes (`behavior.monster.engage`), so clay holds a corpse while the party is farther off and fights when it is near. Clay and soul-feeders do not wander: wandering agents flood every golden with moves.
- Goldens updated (combat encounters, tui menus): UIDs and map IDs shifted by the two new rooms, and the den remnant now leaves for the hub after a phase, which adds a move or two to encounters that run past one. No creature's fight changed. Lethality targets still hold.
- Tests that needed edits: `residents_test.go` (clear the den spawn so a leaving remnant does not kill the party; ignore remnant migration), `regions_test.go` and `offscreen_test.go` (hub has two more neighbours), `explore/playground_test.go` (peaceful factions, remnant need frozen, two new rooms).
- Soul-feeders' `need.soul` is now met by wounding the living, and `leave` ranks gates by content: follow-ups M3-18.F2 and M3-18.F1, done in this pull request.
- The believer has no behaviors (it stands); an archon's warden guards the nearest believer within 3 hexes.


