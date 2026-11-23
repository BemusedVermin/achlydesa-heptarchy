---
id: C1-02
title: "Is an own severed part that is put back a graft?"
status: todo
milestone: C1
kind: question
depends_on: []
discovered_by: null
size: S
---

# C1-02 · Is an own severed part that is put back a graft?

## Goal

The owner picks one option, and the answer is recorded here and where the glossary or design doc needs it.

## Context

- Vocabulary audit finding 2. `docs/game-design.md` on Relic and Rot.

## Deliverables

- The answer, written in this task's Notes and in `docs/glossary.md` (and `docs/` where the design says otherwise).

## Spec

`comp.Graft` with an empty `Piece` is added by `attach_part from: severed`, so `source.grafts` and the `graft` formula variable count a reattached own limb. That unlocks Relic (`choices.cue:9`) and makes Rot eat the limb (`types.cue:39`). The design says grafts are pieces of Serilmen.

- **Pieces only**: `comp.Graft` means a fitted piece. A reattached limb gets its own marker (`comp.Reattached`) or none. Relic and Rot then ignore reattached limbs. Behavior changes.
- **Both are grafts**: Keep the code, and amend the design: any part attached after the fact is a graft. No behavior change.

**Recommendation:** Pieces only, as the design reads today.

**Decided by the owner:** pieces only. `comp.Graft` means a fitted piece of Serilmen. A reattached own limb gets its own marker (`comp.Reattached`) or none, so `source.grafts`, Relic and Rot ignore it. This changes behavior: Relic no longer unlocks from, and Rot no longer eats, a reattached limb. The design's Graft row is reworded to say a reattach is not a graft. Applied by C1-17.

## Acceptance criteria

- [x] The owner has answered, through `/decide`.
- [x] The answer is in Notes and in the glossary.

## Out of scope

- Doing the work: the cleanup task named in the milestone brief does it.

## Notes
Decided 2026-11-23: pieces only. In the glossary and the design's Graft row.
