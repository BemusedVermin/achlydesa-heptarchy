# Lens: vocabulary

Goal: one clear, concise name for each concept of the game and of its code, used the same way in the docs, the content, the Go code and the text players read. The owner decides the names; you find the problems and propose.

1. **Collect the names in use** from:
   - `docs/game-design.md`, `docs/go-technical-design.md` and `docs/glossary.md` (if it exists);
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
