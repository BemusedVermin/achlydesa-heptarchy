---
id: C1-16
title: "What should a speed formula that fails at run time do?"
status: todo
milestone: C1
kind: question
depends_on: []
discovered_by: null
size: S
---

# C1-16 · What should a speed formula that fails at run time do?

## Goal

The owner picks one option, and the answer is recorded here and where the glossary or design doc needs it.

## Context

- Foundation audit C1.

## Deliverables

- The answer, written in this task's Notes and in `docs/glossary.md` (and `docs/` where the design says otherwise).

## Spec

`anatomy/install.go` `speedOf` turns an error into speed 0, so the body never gets to step and nothing reports it. AGENTS.md says content bugs return errors.

- **Return the error**: `timeline.SetSpeed` hooks return `(int, error)` and the timeline propagates it.
- **Guarantee at load**: The loader proves the formula cannot fail, and run time keeps a 0 fallback with a logged warning.

**Recommendation:** Return the error.

**Decided by the owner:** return the error. `timeline.SetSpeed` hooks return `(int, error)` and the timeline propagates it, instead of `speedOf` turning it into speed 0. A content bug in a speed formula now stops the step and is reported. Applied by C1-22.

## Acceptance criteria

- [x] The owner has answered, through `/decide`.
- [x] The answer is in Notes and in the glossary.

## Out of scope

- Doing the work: the cleanup task named in the milestone brief does it.

## Notes
Decided 2026-11-23: return the error.
