# M6 · Minds

**Status:** outline. Break it into task files once M5 is done.

**Goal:** people and creatures that want things and work out how to get them. Needs keep them alive; desires, weighted by personality, give them reasons to live. Each agent picks a goal, plans the actions that reach it from what it believes, and carries the plan out anywhere in the world, replanning when the world proves it wrong.

**Design sources:** `docs/game-design.md` → "The living deep" (Outcomes emerge, Needs change, What happened is found), "People with needs" (Needs, Desires, Behaviors, Knowing where), "Monsters with needs", "Party" → "Recruits". Tech design → "Level of detail", "Designing for pivots" (Outcomes emerge). AGENTS.md rule 10.

## The model (proposed)

From the research of 2026-11-25 (prior art below). Settle it with the owner and write it into the design docs before breaking the milestone down.

- **Needs** are what it takes to stay alive (water, food, rest, safety), on the urgency curves that exist.
- **Desires** reflect personality. One setting-specific list in content (such as standing, family, faith, vengeance, belonging, the soul). Everyone has every desire; a person's personality is its weights over the list. Wishes are concrete goals a desire produces. Memories shift the weights.
- **Choosing a goal** is a utility score: urgency × weight × how reachable the goal is, given what the agent believes.
- **Planning** is GOAP (owner, 2026-11-25) over the actions content already defines, given preconditions and effects on a small set of facts.
  - Costs include travel over the coarse graph, and danger weighted by personality, so a timid person prices danger higher.
  - Plans are short. An agent replans only when its goal changes or its plan breaks.
  - Nothing random happens inside the search. Variety comes from personality and seeded tie-breaks.
- **Beliefs** are what the agent has seen, plus what content says it starts knowing, corrected when it sees again. Hearsay joins in M8.
- **Intentions:** a plan is held until it is done or broken (`comp.Doing`).
- **Two fidelities:** near the party, agents act cell by cell; far from it, they decide less often and plan and travel over the coarse chunk graph.
- **No leave.** Going elsewhere is just going for what they need. A town empties because its people found what they need elsewhere; an agent lives where it rests.

## Planned tasks

- **Design first.** The desire list, personality, goal scoring and the planner's facts, written into the game and tech design with the owner.
- **The coarse chunk graph** in `hexmap`: chunks and the openings between them (planned as M4-05, moved here).
- **Knowing where things are:** beliefs by sight and starting knowledge, corrected by sight (planned as M4-10).
- **Desires and personality** as content, with weights per template or per person.
- **Goal choice** by utility over needs and desires.
- **The planner:** GOAP over content actions with preconditions and effects; plan traces in the harness.
- **Needs across the world:** targets come from beliefs anywhere in the world; with nothing known, a random search (planned as M4-15).
- **Far agents:** two fidelities (planned as M4-09).
- **Recruits as agents:** pay as a need, asking for it, quitting as an act like hiring, and fleeing by choice. M4-03.F1 (personality: robbery, loyalty) folds in here.
- **The time budget:** a seeded year of the sample world within 60 s.

## Exit criteria

- With the shallow town's water gone, its people go for water they know of elsewhere, some end up living in the lake town, and with none known they search.
- An unpaid recruit asks for its pay and then quits; a frightened one flees and comes back.
- Two people with the same needs and different personalities do different things in the same place, and each choice can be read from their desires.
- An agent whose plan breaks (the cistern it remembered is empty) replans from what it now knows.
- The harness shows each agent's goal and plan, and a seeded year of the sample world runs within 60 s.

## Questions to settle before breaking it down

- The desire list for this setting, and how a person's weights are made: authored, rolled from content tables, or both.
- Which needs and desires creatures have, as against people.
- How far plans reach (steps, and over what time), and how often far agents replan.
- Which M3 behavior primitives become planner actions, and which stay as the primitives actions run.
- How a wish relates to a seat (M7): whether a passed seat takes the wish its new holder already has.

## Prior art

- Orkin, "Three States and a Plan: The A.I. of F.E.A.R." (GDC 2006): GOAP.
- Ubisoft, "AI action planning on Assassin's Creed Odyssey and Immortals Fenyx Rising" (Game Developer): GOAP in production; cost, debugging, no random costs.
- Dörner and Bach, Psi theory and MicroPsi: physiological, social and cognitive drives.
- Reiss, the 16 basic desires: everyone has all of them, ranked differently.
- Dwarf Fortress: facets, values and dreams, changed by memories.
- Ryan et al., "Toward Characters Who Observe, Tell, Misremember, and Lie" (Talk of the Town): beliefs and hearsay.
- Park et al., "Generative Agents" (2023): memory, reflection and plans that needs interrupt; the ideas, not the language models.
