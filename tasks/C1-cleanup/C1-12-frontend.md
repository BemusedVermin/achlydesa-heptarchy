---
id: C1-12
title: "Where does frontend chrome end and game text begin, and which core packages may the tui import?"
status: todo
milestone: C1
kind: question
depends_on: []
discovered_by: null
size: S
---

# C1-12 · Where does frontend chrome end and game text begin, and which core packages may the tui import?

## Goal

The owner picks one option, and the answer is recorded here and where the glossary or design doc needs it.

## Context

- Game and frontend audit V1, V3, `.claude/rules/frontend.md`.

## Deliverables

- The answer, written in this task's Notes and in `docs/glossary.md` (and `docs/` where the design says otherwise).

## Spec

`frontend.md` and AGENTS.md rule 4 say game text never lives in Go in the frontend. The tui hard-codes "ALERT: hostile in sight", "Control: ", "Looking", "graft", "locked", "free in %d", "ready", "lands at tick %d", "Day %d · %s", "Actions", "Tick %d", and builds menu group labels by capitalizing the end of a content ID. Separately, `ContentMsg` carries a `*content.Registry`, but `frontend.md` allows only `game` and `api`.

- **Core supplies all of it**: The view carries rendered text, such as a channel state label and option group labels. Only key help stays in Go. Hot reload sends an opaque `func(*game.Game) error` instead of a registry.
- **Short chrome in Go**: Labels like "Tick" and "Day" count as chrome and stay. Only game state text moves. Amend `frontend.md` to allow `content` for hot reload.

**Recommendation:** The first.

**Decided by the owner:** the core supplies all of it. The view carries rendered text, such as a channel state label and option group labels. Only key help stays in Go in the tui. Hot reload sends an opaque `func(*game.Game) error` instead of a `*content.Registry`, so the tui imports only `game` and `api`. Applied by C1-23.

## Acceptance criteria

- [x] The owner has answered, through `/decide`.
- [x] The answer is in Notes and in the glossary.

## Out of scope

- Doing the work: the cleanup task named in the milestone brief does it.

## Notes
Decided 2026-11-23: the core supplies all. In the glossary and the tech design.
