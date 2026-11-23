---
id: C1-10
title: "What is the hostile faction called?"
status: done
milestone: C1
kind: question
depends_on: []
discovered_by: null
size: S
---

# C1-10 · What is the hostile faction called?

## Goal

The owner picks one option, and the answer is recorded here and where the glossary or design doc needs it.

## Context

- Vocabulary audit finding 12.

## Deliverables

- The answer, written in this task's Notes and in `docs/glossary.md` (and `docs/` where the design says otherwise).

## Spec

`faction.deep` has the player text "the deep", the same as the place.

- **faction.wild**: Hostile creatures as a side.
- **faction.hostile**: Plain, but names the relation not the side.
- **Keep**: The place and the faction share a name on purpose.

**Recommendation:** `faction.wild`. The player-facing name is the owner's.

**Decided by the owner:** `faction.wild`. `faction.deep` is renamed, so the hostile creatures as a side no longer share a name with the place. The player-facing text for it is chosen by the applying task (a plain default, noted) and can be changed by the owner. Applied by C1-17.

## Acceptance criteria

- [x] The owner has answered, through `/decide`.
- [x] The answer is in Notes and in the glossary.

## Out of scope

- Doing the work: the cleanup task named in the milestone brief does it.

## Notes
Decided 2026-11-23: `faction.wild`. In the glossary.
