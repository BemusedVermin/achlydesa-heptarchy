---
id: C1-07
title: "What is the controller for a creature that keeps still called?"
status: proposed
milestone: C1
kind: question
depends_on: []
discovered_by: null
size: S
---

# C1-07 · What is the controller for a creature that keeps still called?

## Goal

The owner picks one option, and the answer is recorded here and where the glossary or design doc needs it.

## Context

- Vocabulary audit finding 8.

## Deliverables

- The answer, written in this task's Notes and in `docs/glossary.md` (and `docs/` where the design says otherwise).

## Spec

`combat/rest.go` has the controller `rest` and `combat.rest_ticks`, while "rest" elsewhere means the party sleeping in town (`RestPlace`, "Rest and dream"). No shipped template uses the `rest` controller, and a built-in `idle` controller also does nothing.

- **still**: Rename the controller to `still` and the config to `still_ticks`. Keep "rest" for sleeping.
- **Fold into idle**: Delete `rest`; M2-07.F2 already discusses `idle`.

**Recommendation:** `still`.

## Acceptance criteria

- [ ] The owner has answered, through `/decide`.
- [ ] The answer is in Notes and in the glossary.

## Out of scope

- Doing the work: the cleanup task named in the milestone brief does it.

## Notes

