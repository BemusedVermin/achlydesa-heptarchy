---
id: M6-05
title: What an action may target, by tag, with corpses opt-in
status: todo
milestone: M6
kind: seam
depends_on: []
discovered_by: null
size: S
---

# M6-05 · What an action may target, by tag, with corpses opt-in

## Goal

An action says what it may be aimed at by tags, and says whether it may be aimed at a corpse. Moving clay can then consume the dead and nothing else can.

## Context

- Added back from the reverted M3: its `M3-04.F2-corpse-targets-opt-in` and `M3-04.F2.F1-action-target-tags` follow-ups (`tasks/followups/`, status done, code reverted by M4-01).
- `docs/game-design.md` → "Monsters with needs" (moving clay holds the dead). `docs/glossary.md`: action, target.
- `internal/core/actions`: `Startable` and the target filter; `content/data/actions.cue`.

## Deliverables

- `internal/content/actions.go`: `ActionDef.Targets` as a list of tags a target must carry (and optional tags it must not), with `corpse: true` to allow the dead; regenerate the schema.
- `internal/core/actions/`: the filter honours both; an action without them behaves as today.
- Tests: a corpse is not targeted unless the action says so; a tag filter keeps a body out.

## Spec

- **Firm:** the filter is generic tags, naming no content record (AGENTS.md rule 4).
- **Firm:** every existing action keeps its present targets, so M2's goldens do not change.

## Acceptance criteria

- [ ] The tests above pass.
- [ ] `go tool task check` passes.

## Out of scope

- A consume action for moving clay: M6-21.

## Notes

