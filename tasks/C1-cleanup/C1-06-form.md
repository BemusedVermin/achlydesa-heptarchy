---
id: C1-06
title: "What are 'form', 'action', 'move' and 'type'?"
status: proposed
milestone: C1
kind: question
depends_on: []
discovered_by: null
size: S
---

# C1-06 · What are "form", "action", "move" and "type"?

## Goal

The owner picks one option, and the answer is recorded here and where the glossary or design doc needs it.

## Context

- Vocabulary audit finding 7.

## Deliverables

- The answer, written in this task's Notes and in `docs/glossary.md` (and `docs/` where the design says otherwise).

## Spec

The design's "form" (the seven combat forms) is `ActionDef` and `action.*` in content. `templates.cue` calls a list of eight IDs "the seven forms". "Move" is form × type in the design but movement in code (`api.Moved`, `move_cost`). `content.DamageType` contradicts the design, where Name, Keepsake and Relic are types but not damage. The tech design names a `Form` record that doesn't exist.

- **Glossary rule**: An action is any timed doing; a form is one of the seven combat actions, tagged `form` in `forms.cue`. `_forms` becomes `_actions`. "Move" means movement only. `DamageType` becomes `Type`. The tech design is fixed to match.
- **Real Form record**: Add a `Form` content record, as the tech design says. Larger, and a design decision.

**Recommendation:** The glossary rule.

## Acceptance criteria

- [ ] The owner has answered, through `/decide`.
- [ ] The answer is in Notes and in the glossary.

## Out of scope

- Doing the work: the cleanup task named in the milestone brief does it.

## Notes

