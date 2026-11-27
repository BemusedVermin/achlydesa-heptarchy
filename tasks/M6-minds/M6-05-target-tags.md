---
id: M6-05
title: What an action may target, by tag, with corpses opt-in
status: done
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

- [x] The tests above pass.
- [x] `go tool task check` passes.

## Out of scope

- A consume action for moving clay: M6-21.

## Notes

- Shape: `targets: {tags: [...], not: [...], corpse: bool}` on an action (`content.TargetFilter`). Tags are matched against the target's template tags, as `others[crawls]` selectors already do. The loader rejects a filter on a non-body action and a tag both required and excluded.
- Corpses are listed by `Startable` and `Options` when an action sets `corpse: true`, not only accepted when named, so an agent picking from `Startable` (M6-21) can find them. No shipped action sets it, so the player sees no change.
- M3-04.F2.F1 had chosen a config-keyed action tag over a field; this task's Deliverables ask for `corpse: true` on the filter, so that is what was built. Nothing of the earlier code survived M4-01.
- Shown by `TestOnlyACorpseActionMayTargetADeadBody` and `TestTagFilterKeepsABodyOut` (`internal/core/actions/targets_test.go`) and two loader cases in `internal/content/actions_test.go`. `go tool task check-ci BASE=origin/master` passes; the game goldens did not change.
- Docs: `docs/content-guide.md` gains "What an action may target", and its "code never reads a tag" line now says code never *names* one.
- The task-reviewer agent could not be started here; I worked through its checklist myself.
- No follow-ups filed.

