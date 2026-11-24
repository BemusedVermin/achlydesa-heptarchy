# Lens: vocabulary

Goal: one clear, concise name for each concept of the game and of its code, used the same way in the docs, the content, the Go code and the text players read. The owner decides the names; you find the problems and propose.

1. **Collect the names in use** from:
   - `docs/game-design.md`, `docs/go-technical-design.md`, and `docs/glossary.md` and `docs/content-guide.md` if they exist;
   - content IDs and field names in `content/data`;
   - exported Go identifiers (`go doc -all <package>`, or read the code);
   - player-facing text in `text.cue`;
   - task titles.
2. **Group them by concept,** and flag:
   - **Synonyms:** one concept, several names (a body called "body", "creature" and "actor").
   - **Homonyms:** one name, several concepts.
   - **Drift:** code or content naming a concept differently from the design doc.
   - **Mismatches:** a Go name that doesn't match its content ID, a vague abbreviation, or a name that says less than it could.
3. **Propose one term per concept.**
   - For game concepts, prefer the design doc's term.
   - For technical concepts, prefer the standard term (ECS, Go, CUE) unless the design doc already chose another.
   - Propose a rename only where the confusion is real, never for taste.

Report a table: concept, proposed term, one-line definition, current names with how often and where each appears, and whether it's **settled** (one name already used everywhere) or **contested** (several names, or a proposal that differs from what the design doc says). Settled terms go straight into the glossary; contested ones become questions for the owner.

**Start from `docs/glossary.md`.** It holds the terms C1 settled (for example `moving_clay`, `inventory` not "pack", `faction.wild`, `still`, `source` and `target`, `out of the fight`, `party wipe`, `form` as a tag not a record, and `move` only for movement). Treat a term there as settled: report any place that still uses another name for it, and don't reopen the choice. Ask the owner only about concepts the glossary doesn't have.
