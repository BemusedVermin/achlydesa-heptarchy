---
id: C1-05
title: "How are the variables in formula environments named?"
status: proposed
milestone: C1
kind: question
depends_on: []
discovered_by: null
size: S
---

# C1-05 · How are the variables in formula environments named?

## Goal

The owner picks one option, and the answer is recorded here and where the glossary or design doc needs it.

## Context

- Vocabulary audit finding 6.
- `internal/content/env.go`, `content/data/*.cue`.

## Deliverables

- The answer, written in this task's Notes and in `docs/glossary.md` (and `docs/` where the design says otherwise).

## Spec

The target part's integrity is `integrity` in EffectEnv but `target_integrity` in ScoreEnv. Target facts are flat while source facts sit under `source.`. ActionEnv has both `choice` and `chosen`. `immune_known` means "knows the target part's material". `api.Chosen` (a trigger event's numbers, such as `cells`) shares a name with an action's choices.

- **Regular names**: Target fields under `target.` in every env (`target.integrity`, `target.locked`, `target.graft`). Drop `choice` for `chosen.<x>.id`. `immune_known` becomes `material_known`. `api.Chosen` becomes `api.Params`.
- **Leave as is**: Document the differences in the content guide.

**Recommendation:** Regular names. The loader rejects any unknown variable, so stragglers fail loudly.

## Acceptance criteria

- [ ] The owner has answered, through `/decide`.
- [ ] The answer is in Notes and in the glossary.

## Out of scope

- Doing the work: the cleanup task named in the milestone brief does it.

## Notes

