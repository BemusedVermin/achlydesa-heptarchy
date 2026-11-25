---
id: M4-03
title: Delete the upkeep and flight systems
status: todo
milestone: M4
kind: debt
depends_on: [M4-02]
discovered_by: null
size: S
---

# M4-03 · Delete the upkeep and flight systems

## Goal

The systems that decided for recruits are gone: upkeep, discontent, desertion, forced flight, and the events that announced them. Until M6 gives recruits desires and plans, a hired recruit follows the party and fights, and nothing else.

## Context

- `docs/game-design.md` → "Party" → "Recruits".
- `docs/go-technical-design.md` → "Designing for pivots" (Outcomes emerge). AGENTS.md rule 10.
- What exists: `internal/core/agents/upkeep.go` (`tickUpkeep`, `desert`, `tickFlight`), `recruits.go` (hiring), `comp.Discontent`, `comp.Fleeing`, `content.RecruitsConfig`, `api.Grumbled`, `api.Deserted`, `api.Fled`, the follow controller in `internal/core/explore/follow.go`, and `_severedTerror` in `templates.cue`.

## Deliverables

- Delete `upkeep.go` and its tests, the `agents.upkeep` and `agents.flight` systems, `comp.Discontent`, `comp.Fleeing`, `api.Grumbled`, `api.Deserted`, `api.Fled`, and `RecruitsConfig`'s `grumble_at`, `desert_at`, `flee_at`, `terror_need` and `flee_behavior`. Regenerate the schemas.
- Content: drop what only those systems used, such as the upkeep and flight settings, and their texts. Hiring stays as it is.
- Tests and goldens that exercised the deleted systems: delete the tests, and regenerate the goldens.
- `internal/core/save`: bump `CurrentVersion`.
- `docs/content-guide.md`: drop the upkeep and flight paragraphs. `docs/glossary.md`: delete **discontent**.

## Spec

- **Delete, don't replace.** Pay, asking, quitting and fleeing by choice are M6's, through desires and the planner. Nothing here adds a behavior, a need or an event. Firm.
- Terror stays a need if content still raises it; nothing acts on it until M6.

## Acceptance criteria

- [ ] `git grep -w -E "Grumbled|Deserted|Fled|Discontent|Fleeing|tickUpkeep|tickFlight|desert_at|grumble_at|flee_at" -- internal content docs` finds nothing.
- [ ] A test: a hired recruit, never paid, keeps following the party for days.
- [ ] Changed goldens are regenerated and named in Notes.
- [ ] `go tool task check` passes.

## Out of scope

- Recruits' pay, asking for it, quitting and fleeing: M6.
- Recruit personality (M4-03.F1): M6.

## Notes

Filled in by the implementer.
