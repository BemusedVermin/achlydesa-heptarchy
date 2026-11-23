---
id: C1-11
title: "Where does player-facing text live, and how are its IDs shaped?"
status: done
milestone: C1
kind: question
depends_on: []
discovered_by: null
size: S
---

# C1-11 · Where does player-facing text live, and how are its IDs shaped?

## Goal

The owner picks one option, and the answer is recorded here and where the glossary or design doc needs it.

## Context

- Content audit V2, vocabulary audit finding 12.

## Deliverables

- The answer, written in this task's Notes and in `docs/glossary.md` (and `docs/` where the design says otherwise).

## Spec

`.claude/rules/content.md` says player text is in `text.cue`, but 10 other files declare `text:`. Record names are split between `text.cue` and the record's file (`text.material.clay` vs the other materials). Record names live under `text.action.*` while package messages live under `text.actions.*`. Map text IDs are `text.map.test_hall` while every other record follows `text.<id>`. `text.cue` has two `text:` blocks.

- **Beside the record**: A record's name sits beside the record and is `text.<record id>`. Engine messages go in one file per Go package under `text/` as `text.<package>.<event>`. Map IDs become `text.map.test.hall`. A loader check enforces the name pattern, with parts exempt.
- **All in text.cue**: Everything in `text.cue`, split into sections.

**Recommendation:** Beside the record.

**Decided by the owner:** beside the record. A record's name sits beside the record and is `text.<record id>`. Engine messages go in one file per Go package under `text/` as `text.<package>.<event>`. Map IDs become `text.map.test.hall`. A loader check enforces the name pattern, with parts exempt, and `.claude/rules/content.md` is corrected to say so. Applied by C1-17.

## Acceptance criteria

- [x] The owner has answered, through `/decide`.
- [x] The answer is in Notes and in the glossary.

## Out of scope

- Doing the work: the cleanup task named in the milestone brief does it.

## Notes
Decided 2026-11-23: beside the record. In the glossary.
