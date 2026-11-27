---
id: M6-07
title: "Landmarks: things that advertise what they give"
status: in-progress
milestone: M6
kind: seam
depends_on: [M6-02, M6-04]
discovered_by: null
size: M
---

# M6-07 · Landmarks: things that advertise what they give

## Goal

Anything in the world can advertise what it gives: a well gives water, a bed rest, a shrine worship, a graveyard and a corpse the dead. Agents learn of landmarks by sight and plan with them.

## Context

- `docs/game-design.md` → "How minds work" → "Landmarks"; "Monsters with needs" (moving clay needs the dead, remnants need undecayed ground).
- Prior art: The Sims' objects advertising to needs.
- M6-02 gives sources and containers; M6-04 gives memory and `sense`.

## Deliverables

- `internal/core/comp/`: `Affordance{Tags []string}` (what a thing gives); regenerate the schema.
- `internal/content/`: `Template.Affordances`; a `death` effect or trigger can add one (a corpse gives `the_dead`); marks in a place can name one.
- `internal/core/agents/`: `sense` (M6-04) records a `claim.gives` fact for each landmark seen; a lookup `Known(w, body, tag)` lists remembered landmarks of a tag, nearest first by belief.
- `content/data/`: affordance tags and the first records (`affordance.water`, `.rest`, `.worship`, `.the_dead`).
- Tests: a corpse gives the dead from the tick it dies; an agent that has not seen a well does not know it; one that has, does.

## Spec

- **Firm:** an affordance is data on any entity (body, item, place mark) so a corpse, a well and a mark all work the same.
- **Firm:** undecayed ground, which remnants need, is a property of cells, not of things. The world gives `decay` at a cell from depth through content (`geography.cue`); `sense` records `claim.ground` for the cell an agent stands in and for open cells it sees, so remnants can plan toward it. Put the lookup in `agents`, not in `worldgen`.
- **Firm:** nothing here moves anything; it only labels and is learned.

## Acceptance criteria

- [ ] The tests above pass.
- [ ] `go tool task check` passes, and the save version is bumped.

## Out of scope

- Planning with landmarks: M6-10.
- Placing landmarks in towns: M6-08 and M6-09.

## Notes

