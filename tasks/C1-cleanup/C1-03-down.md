---
id: C1-03
title: "Which words for being out of the fight, for dropping, and for a wipe?"
status: proposed
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

## Acceptance criteria

- [ ] The owner has answered, through `/decide`.
- [ ] The answer is in Notes and in the glossary.

## Out of scope

- Doing the work: the cleanup task named in the milestone brief does it.

## Notes

