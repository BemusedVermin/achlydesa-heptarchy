---
id: M6-10
title: "Operators and beliefs: the state both planners plan over"
status: todo
milestone: M6
kind: seam
depends_on: [M6-03, M6-04, M6-07]
discovered_by: null
size: M
---

# M6-10 · Operators and beliefs: the state both planners plan over

## Goal

One symbolic model that GOAP (M6-15) and HTN (M6-16) both use: what an agent believes, the operators that change it, and how a plan step becomes a timeline action. Nothing here searches yet.

## Context

- `docs/game-design.md` → "How minds work" (Goal, then plan; Beliefs; Landmarks).
- Prior art: Orkin, "Three States and a Plan" (GOAP); the Ubisoft GOAP postmortem on what it cost in production; HTN used in large-crowd games.
- `internal/core/actions`: content actions with channels, wind-up and effects. A plan step is one such action.

## Deliverables

- `internal/content/`: `Operator` (`operator.` IDs): `pre` and `eff` over symbolic facts, `cost`, `does` (the action ID and how its target is picked), `kind` tag. `Goal` (`goal.` IDs): the drive it serves and the state it wants. `Recipe` (`recipe.` IDs): a goal and an ordered list of operators and sub-recipes (read by M6-16).
- `internal/core/agents/state.go`: `State`, a set of symbolic facts of three forms: `has(tag, n)`, `near(tag)`, `known(tag)`, `drive_met(id)`; `StateOf(w, body)` builds it from the body's inventory, position and memory, never from the world. `Apply(op, state)` and `Holds(op, state)`.
- `internal/core/agents/plan.go`: `Plan{Goal, Steps}`, `Step{Operator, Target}`.
- Tests: a state built from beliefs ignores a well the agent has not seen; applying an operator's effects changes the state as written; content naming an unknown operator or tag fails at load.

## Spec

- **Firm:** operators and goals are content, so a new job is content. Code names no operator or tag.
- **Firm:** the state is symbolic and small, so search is cheap. A fact about a thing carries its UID only inside a step's `Target`, never inside the state.
- **Firm:** the same operators serve GOAP and HTN, so a recipe step is an operator.
- **Suggestion:** state is a sorted slice of facts, comparable and hashable for caching.

## Acceptance criteria

- [ ] The tests above pass.
- [ ] `go tool task check` passes.

## Out of scope

- Searching: M6-15. Recipes expanding: M6-16.
- Executing a plan: M6-11.

## Notes

