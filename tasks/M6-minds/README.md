# M6 · Minds

**Goal:** the living world. Creatures and people decide what to do for reasons of their own, in towns the generator builds and across a deep that is populated from the first tick, so what happens emerges from them (AGENTS.md rule 10).

**Design sources:** `docs/game-design.md` → "How minds work" (new, from this plan), "World and towns", "The living deep" (Outcomes emerge, Needs change, What happened is found), "Monsters with needs", "People with needs", "Keeping it affordable", "Party" → "Food and water", "Recruits". `docs/go-technical-design.md` → "Designing for pivots" (Behaviors), "Level of detail".

## Tasks

| ID | Title | Depends on |
| --- | --- | --- |
| [M6-01](M6-01-hostility-handoff.md) | Hostility in the foundation, and a two-way controller hand-off | none |
| [M6-02](M6-02-sources.md) | Sources and containers | none |
| [M6-03](M6-03-drives.md) | Drives: needs and desires with a weight per individual | none |
| [M6-04](M6-04-memory-sight.md) | A capped memory per agent, learned by sight | none |
| [M6-05](M6-05-target-tags.md) | What an action may target, by tag, with corpses opt-in | none |
| [M6-06](M6-06-chunk-graph.md) | The chunk graph: the openings between chunks | none |
| [M6-07](M6-07-landmarks.md) | Landmarks: things that advertise what they give | M6-02, M6-04 |
| [M6-08](M6-08-pieces.md) | Pieces: building blocks a town is laid out from | M6-07 |
| [M6-09](M6-09-settlement-stage.md) | The settlement stage: procgen lays out a town | M6-08 |
| [M6-10](M6-10-operators-beliefs.md) | Operators and beliefs: the state both planners plan over | M6-03, M6-04, M6-07 |
| [M6-11](M6-11-mind-controller.md) | The mind controller: run a plan, claim what it needs | M6-01, M6-10 |
| [M6-12](M6-12-movement-actions.md) | Movement actions: go to, and wander | M6-11 |
| [M6-13](M6-13-thing-actions.md) | Thing actions: take, put and consume | M6-02, M6-11 |
| [M6-14](M6-14-goal-choice.md) | Goal choice: drives score into a goal | M6-03, M6-11 |
| [M6-15](M6-15-goap.md) | The GOAP planner for important individuals | M6-10, M6-12, M6-13, M6-14 |
| [M6-16](M6-16-htn.md) | The HTN planner and recipes for everyone else | M6-10, M6-12, M6-13, M6-14 |
| [M6-17](M6-17-far-agents.md) | Far agents: goals only, over the chunk graph | M6-06, M6-12, M6-14 |
| [M6-18](M6-18-hearsay.md) | Hearsay: agents tell each other what they know | M6-04, M6-11 |
| [M6-19](M6-19-flee-engage.md) | Flee, terror and engaging: handing a fight over and back | M6-01, M6-12, M6-14 |
| [M6-20](M6-20-populate-world.md) | Populate the whole world at start, with minds | M6-03, M6-06, M6-14 |
| [M6-21](M6-21-monster-families.md) | Monster families: moving clay, remnants and beasts as minds | M6-05, M6-07, M6-15, M6-16, M6-19, M6-20 |
| [M6-22](M6-22-settlements.md) | The three towns as settlements | M6-09, M6-13, M6-20 |
| [M6-23](M6-23-residents.md) | Residents: roles, jobs and a town that can run dry | M6-15, M6-16, M6-18, M6-22 |
| [M6-24](M6-24-party-needs.md) | The party eats, drinks and rests | M6-03, M6-07, M6-13 |
| [M6-25](M6-25-hiring.md) | Hiring recruits, pay and quitting | M6-19, M6-23, M6-24 |
| [M6-26](M6-26-recruit-personality.md) | Recruit personality: robbery, desertion and loyalty | M6-25 |
| [M6-27](M6-27-exotic-recruits.md) | Graftable and exotic recruits | M6-21, M6-26 |
| [M6-28](M6-28-golden-and-budget.md) | Golden scenarios for the exit criteria, and the step budget | M6-17, M6-21, M6-23, M6-26 |

M6-01 to M6-06 can all start at once. Lanes after that:

- **World:** M6-07 → M6-08 → M6-09, then M6-22 with M6-13 and M6-20.
- **Mind core:** M6-10 → M6-11, then M6-12, M6-13, M6-14, M6-18 side by side, then the planners M6-15 and M6-16.
- **Scale:** M6-06 and M6-12 and M6-14 → M6-17; M6-20 beside it.
- **Party:** M6-24 needs only M6-03, M6-07 and M6-13; recruits follow M6-23.

About ten tasks touch `internal/core/agents`, so run the mind-core lane in order and pull master between them to avoid merge conflicts. No new foundation package is needed: `.golangci.yml` is protected on task branches, and everything lands in `agents` or in the existing foundation packages (`knowledge`, `inventory`, `comp`, `worldgen`, `actions`, `anatomy`, `timeline`).

## Exit criteria

| Exit criterion | Covered by |
| --- | --- |
| A creature that knows nothing wanders, and once it has seen something that meets a need, goes to it. | M6-04, M6-07, M6-12, M6-14, M6-28 |
| Moving clay walks to a corpse the party left behind and takes it. | M6-05, M6-21, M6-28 |
| A resident hauls water from a source outside the town into its cistern, and when the haulers die and the cistern runs dry, residents go looking for water, with nothing announcing it. | M6-02, M6-13, M6-22, M6-23, M6-28 |
| A hired recruit fights beside the party and then quits when pay goes unmet, and two recruits with different weights leave differently (one robs). | M6-25, M6-26, M6-28 |
| The party goes hungry, thirsty and tired, eats, drinks and rests to heal, and wakes at its rest place after a wipe. | M6-24 |
| Towns are generated, with the same layout after an export, each with the landmarks its people need. | M6-08, M6-09, M6-22 |
| The whole bounded world is populated at the start, about 1,000 individuals step within the stated budget, and the same seed gives the same hash, including across a save in the middle of a plan. | M6-11, M6-17, M6-20, M6-28 |

## Decisions

Settled with the owner on 2026-11-27, recorded in `docs/game-design.md` → "How minds work" unless noted:

- **Decision model:** drives score into a goal, then a planner finds the steps. Important individuals use GOAP, picked when they are generated; everyone else uses HTN recipes written in content. Combat AI stays separate (the `hunt` controller and its chooser); a mind hands a fight over and takes the body back.
- **Personality:** a weight per drive on each individual, drawn around the template's typical weights at spawn.
- **Beliefs:** a capped memory per agent, learned by sight and by hearsay from agents met. Nothing known means a random path. How stories drift is M8.
- **Scale:** about 1,000 individuals at once, first target. The tech design's "Level of detail" records it.
- **Supplies:** sources refill slowly at a content rate and people haul to containers. Recorded in "How minds work".
- **A wish** is one more drive with a large weight, set by M7.
- **The whole world is populated at start,** as individuals; an export creates nobody. Recorded in the tech design's "Level of detail".
- **Towns are generated** by procgen from content-defined pieces, with layouts fixed by the world seed alone, not hand-authored (this changes "World and towns" from "authored and emergent"). Recorded there.
- **No migrate behavior:** going elsewhere is just going for what you need (the design's behavior list no longer names it).

Settled while planning, and provisional (not asked; marked so in the task specs):

- **Planning model:** one symbolic state of beliefs and one set of content operators serve both planners (M6-10). Plans are invalidated on a failed step, a lost claim, a contradicted belief or an export, and are saved.
- **Reservations** stop two agents planning for the same item or bed (M6-11).
- **Upkeep is an act,** the player's or the recruit's own, never a daily automatic deduction (M6-25).
- **Exit criteria are proved by golden scenarios,** not harness metrics, which are M9's (M6-28).
- **Everything lands in `agents`** and the existing foundation packages.

Follow-ups folded in or kept:

- `M4-03.F1` (recruit personality) is folded into M6-26 and is `obsolete`.
- `M3-16.F2` (deserters' hostility) is done: deserters stay neutral until M8's grudges.
- The corpse-target and target-tag follow-ups from the reverted M3 are added back in M6-05.

Left open: the first values of every drive, goal weight, population and cistern size (tuned in M9); whether scanning about 2,500 chunks at start is fast enough (M6-20 records the time and may populate far chunks lazily from the same stream); the thrukk, still town and archon's crawlspaces (later milestones).

## Prior art

From the research of 2026-11-25 and 2026-11-27:

- Orkin, "Three States and a Plan: The A.I. of F.E.A.R." (GDC 2006), and "GOAP: Ten Years Old and No Fear!" on GOAP.
- Ubisoft's postmortem on GOAP in Assassin's Creed Odyssey and Immortals Fenyx Rising (Game Developer): modularity and debugging gains, the production costs.
- HTN planning for thousands of agents, with levels of detail and behavior caching (large multiplayer simulations).
- RimWorld's think trees, job givers and reservations; the Sims' needs and advertising objects.
- Dörner's Psi theory, Reiss's 16 basic desires, Dwarf Fortress's facets, values and dreams.
- Talk of the Town (Ryan et al.) on knowledge and hearsay; Generative Agents (Park et al., 2023) for the ideas, not the language models.
- Minecraft's village generation from structure pieces, for the settlement stage.
