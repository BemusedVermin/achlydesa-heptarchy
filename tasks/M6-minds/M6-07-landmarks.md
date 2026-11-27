---
id: M6-07
title: "Landmarks: things that advertise what they give"
status: done
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

- [x] The tests above pass.
- [x] `go tool task check` passes, and the save version is bumped.

## Out of scope

- Planning with landmarks: M6-10.
- Placing landmarks in towns: M6-08 and M6-09.

## Notes

- **Tests:** `TestACorpseGivesTheDeadFromTheTickItDies` (`agents/corpse_test.go`, a real game), `TestAnAgentKnowsAWellOnlyOnceItHasSeenIt`, `TestKnownListsTheNearestLandmarkFirstByWhereItWasSeen`, `TestAMarkWithNoBodyIsSeenAsALandmark`, `TestSenseRecordsTheGroundUnderfootAndTheNearestOfEachOtherDecay` and a rapid ring check (`agents/landmarks_test.go`), `TestEachCellOfALandmarkMarkGivesItsAffordances` (`explore/landmarks_test.go`) and `TestLandmarkMarksAndAffordancesAreChecked` (`content/places_test.go`).
- **Shape:** `comp.Affordance.Tags` is `[]api.ContentID` rather than `[]string`, kept in ID order, since the tags are `affordance.` record IDs. Templates list `affordances`; a place maps a mark character to affordances in `PlaceDef.Affordances`, and each such cell becomes an entity with only `Identity` (its template is the place ID), `Pos` and `Affordance` when the game starts. Such a mark may repeat, like a spawn, and may not be a spawn or an opening.
- **Death:** a new `afford` effect (params `bodies`, `affordance`) adds an affordance; `api.BodyDied` is now `Subjected` (the body is the target), so a template's `combat.body_died` trigger can use it. Every creature and person template carries `_corpseGivesDead` in `templates.cue`. On the remnant it goes after its existing trigger, because a trigger's place in the list is its ID and moving the sever trigger changed its rolls.
- **Facts:** `sense` records `claim.gives` (value: the affordance IDs joined by commas) and `claim.at` for each landmark seen, bodies or not. `Known(w, body, tag)` returns `[]Landmark{Subject, At}`, nearest the body's own cell first, by UID on ties, and skips a landmark whose place it does not remember.
- **Ground:** facts about a cell needed a new key, so `api.Fact` gained `Cell *api.Cell` (nil for facts about a thing); the knowledge store keys by (Subject, Cell, Claim, Origin). `sense` learns `claim.ground` (value: the world geography's decay level) for the cell underfoot and for the nearest open cell seen at each other level, not for every open cell, which would fill the 64-fact memory in one look: M6-07.F1 asks the owner to confirm. `KnownGround(w, body)` lists remembered ground, least decayed first. The lookup is in `agents`, through `content.Registry.DecayLevel`.
- **Config:** `agents.gives_claim` and `agents.ground_claim` (`knowledge.cue`), new claims `claim.gives` and `claim.ground` with their log lines, and `affordances.cue` with `affordance.water`, `.rest`, `.worship` and `.the_dead`. The well template gives water. No sample place has landmark marks yet (M6-08, M6-09).
- **Benchmark:** `BenchmarkThousandAgentsSenseOnce` now runs with a geography and two exports, so it measures ground too: about 55 ms, inside the 150 ms budget. Ground cells are walked in rings by distance, skipping depths whose level is already found.
- **Save:** version 13 (`comp.Affordance`, `api.Fact.Cell`), with a no-op step in `save_test.go`.
- **Outside the Deliverables:** `internal/core/api/{knowledge,combat}.go`, `internal/core/knowledge/knowledge.go`, `internal/core/explore/{source,world_source,explore}.go`, `internal/core/save/`, `internal/content/{places,load,records,affordances}.go`, the regenerated schemas, `docs/content-guide.md`, `docs/glossary.md` (affordance) and `docs/go-technical-design.md` (the `knowledge` row).
- **Follow-ups:** M6-07.F1.

