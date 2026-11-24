---
id: M3-18
title: Monster families as content
status: todo
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

- [ ] Exit criterion: a test kills a creature near moving clay; the clay walks to the corpse.
- [ ] A remnant in a decayed region moves toward its less-decayed neighbour, on screen and off.
- [ ] An archon's creature attacks a party member who comes near a believer and leaves the believer alone.
- [ ] In the playground, each family's room shows its behavior.
- [ ] `go tool task check` passes.

## Out of scope

- Wish-born creatures and the flock (they need seats (M4) and the surface's schedule, later).
- Grudges (M6).

## Notes


