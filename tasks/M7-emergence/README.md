# M7 · Emergence

**Status:** outline. Break it into task files once M6 is done.

**Goal:** what makes runs tell stories, all through agents with needs: people born and arriving, people leaving and founding crawlspaces, rumors that drift, grudges and memory, dreams, generated candidates, and the flock.

**Design sources:** `docs/game-design.md` → "World and towns" (crawlspaces authored and emergent, the bounded world), "The living deep" (Outcomes emerge, Needs change, What happened is found), "People with needs" (Rumors travel and drift), "Three tiers of people" (new people), "Grudges and memory", "Core loop and time management → Dreams", "Candidates and seats → How wishes curdle", "Time pressures" (The flock is weather).

## Planned tasks

- **Births and growing up.** Children are born, grow, and take the test at ten, which sorts them into ibben or salken.
- **Immigration and emigration points.** Places at the world's edges where people arrive from beyond it, and leave for beyond it, when that meets their needs. Nobody respawns.
- **New crawlspaces.** People who can't meet their needs anywhere dig and settle a new crawlspace; one nobody needs is abandoned.
- **Needs that change.** Content rules add, remove and reweight an agent's needs as its circumstances change: losses, grudges, growing up, grafts, faith, where it lives.
- **Knowing what happened.** The party learns outcomes from evidence and from people, as facts in the knowledge store, never as events shown to the player:
  - a crawlspace abandoned a while ago looks it (empty cisterns, belongings left behind, nobody answering);
  - people who knew someone who left remember it, and say where they went and roughly when.
- **Rumors.** Facts carried between people and towns, drifting with each retelling (`text/template` for the words); drifted rumors as a source of false positives.
- **Memory and grudges.** Survivors remember who maimed them; revenge becomes a need; bodies come back with replaced parts; deserters live on like anyone else.
- **The deepest nemesis.** An abandoned recruit can take a passed seat, through the existing wound scoring.
- **Dreams.** Sleeping in town picks one image by content conditions over world state; clue dreams write low-confidence facts; graft count deepens Serilmen's dreams.
- **The wish generator.** Wound, then wish, then curdle mode, then price, then form, over content tables, for generated candidates.
- **The flock.** Hibblejaunts as agents with needs on the salk; where they roam closes routes and kills anyone caught. Nothing schedules it.

## Exit criteria

- Arriving at a crawlspace abandoned weeks ago, the party can tell it was abandoned and roughly when, and someone elsewhere can say who left and where they went.
- An agent's needs at the end of a long run differ from its needs at the start, because of what happened to it.
- A creature that lost an arm to Cut returns with a clay arm, weak to Soak, and hunts the party.
- A rumor about the party changes as it crosses two towns.
- A dream gives a clue that investigation later confirms.

## Questions to settle before breaking it down

- How far rumors may drift before they stop being useful.
- Whether the slave doctrine's challenge against a hibblejaunt appears in play (open in the game design).
