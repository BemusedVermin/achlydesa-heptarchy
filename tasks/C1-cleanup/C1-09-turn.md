---
id: C1-09
title: "What do 'turn', 'per-turn' and 'pending' mean?"
status: proposed
milestone: C1
kind: question
depends_on: []
discovered_by: null
size: S
---

# C1-09 · What do "turn", "per-turn" and "pending" mean?

## Goal

The owner picks one option, and the answer is recorded here and where the glossary or design doc needs it.

## Context

- Vocabulary audit finding 9.

## Deliverables

- The answer, written in this task's Notes and in `docs/glossary.md` (and `docs/` where the design says otherwise).

## Spec

`comp.Pending` is an action in flight, but `timeline.Saved.Pending` means per-turn systems not yet run. `sim.PerTurn` means "after each timeline entry", and "turn queue" means the timeline. `comp.Pending.Part` holds a channel.

- **Rename**: `timeline` flag becomes `SystemsDue`, `sim.PerTurn` becomes `PerEntry`, "turn queue" becomes "timeline", `Pending.Part` becomes `Pending.Channel`.
- **Define**: Glossary: a turn is one timeline entry. Keep the names.

**Recommendation:** Rename.

## Acceptance criteria

- [ ] The owner has answered, through `/decide`.
- [ ] The answer is in Notes and in the glossary.

## Out of scope

- Doing the work: the cleanup task named in the milestone brief does it.

## Notes

