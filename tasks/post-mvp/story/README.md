# Story (post-MVP)

**Status:** outline, post-MVP. Break it into task files only once the MVP (M1 to M7) is done. Task IDs: `PM-STORY-01` and on.

**Goal:** the main story, the characters' arcs and emergent stories all run on one scene system. Authored scenes have fixed casts; emergent scenes are cast from whoever in the simulation fits. Casting, strain and threads keep characters consistent, and a director paces heavy and soft scenes. Everything a writer touches is content.

**Design sources:** `docs/post-mvp.md` → "Story" (How the story runs, Story beats, Emergent stories, The cast model after the MVP, In the engine). `docs/game-design.md` → "Story" for the canon it serves, "People with needs" for requests.

**Prior art the design draws on:** storylets and quality-based narrative (Failbetter, Emily Short), most-specific line matching (Valve's dynamic dialogue, Elan Ruskin, GDC 2012), story roles cast by personality (Wildermyth), social practices (Versu), stress from acting against one's traits (Crusader Kings III), beats and a drama manager (Façade), pacing (RimWorld's storytellers), curated emergent narrative and story sifting (James Ryan; Max Kreminski's Felt and Winnow), and memory-driven rivals (Shadow of Mordor's Nemesis system).

## Planned tasks

- **The `story` package.** Add it as a feature package: AGENTS.md rule 2 and the depguard lists in `.golangci.yml`.
- **Traits and strain.** Traits are content (setting-specific ones such as devout, dutiful and proud, plus generic ones such as kind, cruel and fearful) and components on characters. Acting against one's traits adds strain; past a threshold a turn scene plays and traits shift. The cast model's wound scoring counts strain.
- **Hooks.** Content-defined facts about a character for scenes to hang on, such as "lost his betrothed at the Osten".
- **The chronicle.** Significant events with their causes and several tags each, fed from the bus. Scenes, patterns and lines read it.
- **Scenes and lines.** The scene format: roles, condition, slot, tone, lines, choices, effects and threads. Lines are picked by most-specific match over facts. A scene view and a terminal screen to play them.
- **Casting.** Role requirements and fit formulas; a minimum fit below which a scene doesn't play; declared breaks for acting against type; roles authored characters can rule themselves out of; choices made in character, weighted by traits, for anyone who isn't the player.
- **Threads.** Opened by breaks, deaths, promises, betrayals and grudges; matched to the kinds of scene that can move them forward; closed by a scene, or by a scene where someone remembers when the person is gone.
- **Story patterns.** Staged patterns over the chronicle. A partial match can play foreshadowing; a complete match casts and plays its scene.
- **The director.** One scene per slot, preferring spine scenes, then open threads, then sifted stories, then everyday scenes. Heavy and soft alternate under an intensity budget, with cooldowns per character. It stages set pieces (gathers the cast, pauses the clock) and nudges the simulation when the spine needs it, for example in who a seat's offer finds.
- **Scenes without the party.** These resolve through the characters' own choices with no text, go into the chronicle, and leave events that `rumor` spreads.
- **Requests.** Unmet needs open threads, and scenes put the ask to the party.
- **Authored content.** The spine, the protagonists' arcs and the draft beats from the game design, with fallbacks or plot armor for every spine scene. The specific beats are subject to change, so no code may depend on any particular one: they're content like everything else.
- **Harness checks.** The spine completes on every seed; no thread stays open past a tuned number of days; a measure of how varied the emergent stories are.

## Exit criteria

These use today's draft beats as test cases; if a beat is rewritten or cut, swap in whatever replaces it.

- The opening plays from the scene system, with both the spare and the reap branch.
- An Osten in a town the party never visits produces a different story depending on the winner's traits, and nobody acts against type without a declared break and its thread.
- A kind character pushed into a cruel choice gains strain, and a later scene follows it up.
- The abandoned-recruit pattern foreshadows with a rumor and completes with "You left me."
- Over 200 seeds, every spine scene fires.

## Questions to settle before breaking it down

- The bond's ending: whether the piece steered it too, or it's the one thing the protagonists keep (open in the game design).
- The starting trait list, and how many traits a character carries.
- How scenes look in the terminal.
- The intensity budget and thread deadlines, to be tuned like the clock numbers.
- Whether the slave doctrine's challenge against a hibblejaunt appears in play (open in the game design).
