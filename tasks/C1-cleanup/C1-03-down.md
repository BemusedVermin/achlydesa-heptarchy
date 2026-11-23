---
id: C1-03
title: "Which words for being out of the fight, for dropping, and for a wipe?"
status: todo
milestone: C1
kind: question
depends_on: []
discovered_by: null
size: S
---

# C1-03 · Which words for being out of the fight, for dropping, and for a wipe?

## Goal

The owner picks one option, and the answer is recorded here and where the glossary or design doc needs it.

## Context

- Vocabulary audit findings 3 and 4.

## Deliverables

- The answer, written in this task's Notes and in `docs/glossary.md` (and `docs/` where the design says otherwise).

## Spec

"Fall" means both dropping off a ledge (`api.Fell`, `_fallInjury`) and being down or dead (`fallen()`, "{{.name}} falls."). "Lost" means both a recoverable wipe (`api.PartyLost`, `combat.party_lost`) and the permanent end (`game.go:188`). The design says "party wipe" for the first.

- **Fall = drops only; wipe**: Rename `fallen()` to `outOfFight()`, reword the two texts, rename `PartyLost` to `PartyWiped` (`combat.party_wiped`, `text.combat.party_wiped`), keep "lost" for the unrecoverable state, and show `text.game.lost` only when the party cannot wake.
- **Leave as is**: Define both senses in the glossary.

**Recommendation:** The first. The player-facing wording (for example "goes down") is the owner's.

**Decided by the owner:** "fall" means dropping off a ledge only. `fallen()` becomes `outOfFight()` and the two texts that say "falls" for a downed or dead body are reworded. `api.PartyLost` becomes `api.PartyWiped` (`combat.party_wiped`, `text.combat.party_wiped`). "Lost" is kept for the unrecoverable end, and `text.game.lost` shows only when the party cannot wake. The player-facing wording of the new texts is the owner's to give; the applying task picks a plain default and notes it. Applied by C1-17.

## Acceptance criteria

- [x] The owner has answered, through `/decide`.
- [x] The answer is in Notes and in the glossary.

## Out of scope

- Doing the work: the cleanup task named in the milestone brief does it.

## Notes
Decided 2026-11-23: the recommended option. In the glossary.
