---
id: C1-15
title: "Where does shared git code for the tools go, given agentkit is protected?"
status: todo
milestone: C1
kind: question
depends_on: []
discovered_by: null
size: S
---

# C1-15 · Where does shared git code for the tools go, given agentkit is protected?

## Goal

The owner picks one option, and the answer is recorded here and where the glossary or design doc needs it.

## Context

- Tools audit D2.

## Deliverables

- The answer, written in this task's Notes and in `docs/glossary.md` (and `docs/` where the design says otherwise).

## Spec

`gitOutput` in agentkit and `git` in samebehavior are the same function, and affected has its own. A shared package outside `internal/tools/agentkit/` would let an agent change the guard's behavior unless it is added to `protectedPaths`.

- **Share outside agentkit, leave agentkit alone**: `internal/tools/internal/gitcmd` for samebehavior and affected only.
- **Share and protect**: Agentkit imports it, and the path is added to `protectedPaths`.

**Recommendation:** The first.

**Decided by the owner:** share outside agentkit and leave agentkit alone. `internal/tools/internal/gitcmd` holds the git helper for `samebehavior` and `affected` only; agentkit keeps its own copy and is not changed, so an agent cannot alter the guard. Applied by C1-24.

## Acceptance criteria

- [x] The owner has answered, through `/decide`.
- [x] The answer is in Notes and in the glossary.

## Out of scope

- Doing the work: the cleanup task named in the milestone brief does it.

## Notes
Decided 2026-11-23: the recommended option.
