---
id: C1-04
title: "How are the acting body, the body acted on, a fact's origin and a formula's subject named?"
status: done
milestone: C1
kind: question
depends_on: []
discovered_by: null
size: M
---

# C1-04 · How are the acting body, the body acted on, a fact's origin and a formula's subject named?

## Goal

The owner picks one option, and the answer is recorded here and where the glossary or design doc needs it.

## Context

- Vocabulary audit finding 5.
- `internal/content/env.go`.

## Deliverables

- The answer, written in this task's Notes and in `docs/glossary.md` (and `docs/` where the design says otherwise).

## Spec

Event fields name the acting body `Actor` (9 fields), `Source` (4) and `Body`. Formulas call it `source`. Knowledge also uses `source` for where a fact came from (`source.sweep`), and `reveal` mixes both senses in one effect. `content.Side` means "what a formula sees of one body" while `Faction` is called a side.

- **Roles name the fields**: Event fields become `Source` (the actor) and `Target`. `api.Fell.Actor` becomes `Body`. Knowledge `Source` becomes `Origin` (`origins`, `origin.sweep`). `content.Side` becomes `BodyEnv`.
- **Actor role**: Rename the formula role `source` to `actor` instead, and keep knowledge `Source`.
- **Leave as is**: Document the field-to-role mapping per event in the glossary.

**Recommendation:** The first.

**Decided by the owner:** roles name the fields. Event fields become `Source` (the acting body) and `Target` (the body or part acted on); `api.Fell.Actor` becomes `Body`. Knowledge `Source` becomes `Origin` (`origins`, `origin.sweep`), so `source` in a formula always means the acting body, and `reveal` stops mixing the two senses. `content.Side` becomes `BodyEnv`, so "side" means only a faction. Applied by C1-17.

## Acceptance criteria

- [x] The owner has answered, through `/decide`.
- [x] The answer is in Notes and in the glossary.

## Out of scope

- Doing the work: the cleanup task named in the milestone brief does it.

## Notes
Decided 2026-11-23: roles name the fields. In the glossary.
