---
id: M6-12
title: "Movement actions: go to, and wander"
status: todo
milestone: M6
kind: feature
depends_on: [M6-11]
discovered_by: null
size: M
---

# M6-12 · Movement actions: go to, and wander

## Goal

An agent can go to a place it believes in, and with nothing known can pick a path at random. Both are operators over the existing step action.

## Context

- `docs/game-design.md` → "How minds work" (Beliefs: nothing known means a random path); "People with needs" (no migrate).
- `internal/core/actions` (`action.step`), `internal/core/hexmap` (pathing), `internal/core/combat/controller.go` (`stepToward`, which `hunt` uses).
- A reverted M3 fix to remember: path around standing bodies (`M3-04.F1`).

## Deliverables

- `content/data/operators.cue`: `operator.go_to` (target a believed landmark or cell) and `operator.wander` (a step in a random open direction, biased away from cells it just left).
- `internal/core/agents/`: the target pickers these operators name; a path step that avoids standing bodies; `wander` draws from the `agents` stream derived per UID and tick.
- Tests: with a believed well, an agent walks to it; with none, it moves and never stands still forever; the same seed gives the same wander.

## Spec

- **Firm:** `go_to` uses A* over loaded cells for near agents only. Far agents travel the chunk graph (M6-17).
- **Firm:** a cell the agent believes is open but is not makes the step fail, which drops the plan and the belief is replaced by what is seen.
- **Suggestion:** wander prefers unvisited cells, remembered as `claim.visited` with a short cap.

## Acceptance criteria

- [ ] The tests above pass.
- [ ] `go tool task check` passes.

## Out of scope

- Fleeing: M6-19. Far travel: M6-17.

## Notes

