# M6 · Emergence

**Status:** outline. Break it into task files once M5 is done.

**Goal:** the systems that make runs tell stories: rumors that drift, grudges and memory, dreams, generated candidates, and the flock.

**Design sources:** `docs/game-design.md` → "People with needs" (Rumors travel and drift), "Grudges and memory", "Core loop and time management → Dreams", "Candidates and seats → How wishes curdle", "Time pressures" (The flock is weather).

## Planned tasks

- **Rumors.** Facts carried between people and towns, drifting with each retelling (`text/template` for the words); drifted rumors as a source of false positives.
- **Memory and grudges.** Survivors remember who maimed them; revenge becomes a need; bodies come back with replaced parts; deserters persist as `Individual`s.
- **The deepest nemesis.** An abandoned recruit can take a passed seat, through the existing wound scoring.
- **Dreams.** Sleeping in town picks one image by content conditions over world state; clue dreams write low-confidence facts; graft count deepens Serilmen's dreams.
- **The wish generator.** Wound, then wish, then curdle mode, then price, then form, over content tables, for generated candidates.
- **The flock.** A scheduled surface movement that closes routes and kills anyone caught.

## Exit criteria

- A creature that lost an arm to Cut returns with a clay arm, weak to Soak, and hunts the party.
- A rumor about the party changes as it crosses two towns.
- A dream gives a clue that investigation later confirms.

## Questions to settle before breaking it down

- How far rumors may drift before they stop being useful.
- Whether the slave doctrine's challenge against a hibblejaunt appears in play (open in the game design).
