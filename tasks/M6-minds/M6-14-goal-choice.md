---
id: M6-14
title: "Goal choice: drives score into a goal"
status: todo
milestone: M6
kind: feature
depends_on: [M6-03, M6-11]
discovered_by: null
size: M
---

# M6-14 · Goal choice: drives score into a goal

## Goal

At each decision an agent weighs its drives, picks one goal, keeps it unless another clearly beats it, and hands it to its planner.

## Context

- `docs/game-design.md` → "How minds work" (Goal, then plan); "People with needs".
- Prior art: The Sims' needs, utility AI, Dörner's Psi and Reiss's desires (the brief's prior art list).
- The reverted M3 used a keep bonus for hysteresis; reuse the idea.

## Deliverables

- `internal/content/`: `Goal` gains `drive`, and `when` (an Expr condition over beliefs), and `weight` (an Expr over the drive's level and the individual's weight).
- `internal/core/agents/choose.go`: scores every goal whose `when` holds, picks the best, keeps the current goal unless another beats it by `agents.keep_bonus`; ties break by goal ID.
- The `mind` controller (M6-11) calls it on its own cadence, `agents.think_ticks`, not on every entry, and sets `Mind.Goal`.
- `content/data/goals.cue`: `goal.drink`, `goal.eat`, `goal.sleep`, `goal.wander` (always applicable, the weakest), first values for tuning.
- Tests: a thirstier agent goes for water over rest; a goal in progress is kept against a slightly higher one; an agent with a larger weight on a drive chooses it first; the same inputs give the same goal.

## Spec

- **Firm:** a goal never exists to force an outcome (AGENTS.md rule 10): it names a drive and a state the agent wants, nothing else.
- **Firm:** `agents.keep_bonus` and `agents.think_ticks` are content, first values for tuning.
- **Firm:** an Expr that fails at run time returns an error and is reported, never a default.

## Acceptance criteria

- [ ] The tests above pass.
- [ ] `go tool task check` passes.

## Out of scope

- The planners: M6-15, M6-16.
- Desires beyond pay (revenge, grief): M8.

## Notes

