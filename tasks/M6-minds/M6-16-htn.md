---
id: M6-16
title: "The HTN planner and recipes for everyone else"
status: done
milestone: M6
kind: feature
depends_on: [M6-10, M6-12, M6-13, M6-14]
discovered_by: null
size: M
---

# M6-16 · The HTN planner and recipes for everyone else

## Goal

Most individuals plan by recipe: a goal expands through content-written recipes into operators, cheaply and predictably.

## Context

- `docs/game-design.md` → "How minds work".
- Prior art: HTN planning used for thousands of agents in large simulations, with behavior caching and levels of detail.
- M6-10's `Recipe` and `Operator`; M6-11's `Planner` interface.

## Deliverables

- `internal/core/agents/htn.go`: a planner registered as `htn`. A recipe is `goal`, optional `when`, and `steps` (operators or sub-recipes); it picks the first recipe, in content order, whose `when` and preconditions hold in the agent's beliefs.
- Recipes are checked at load: every operator exists, no recipe cycles, every sub-recipe terminates.
- `content/data/recipes.cue`: `recipe.drink` (go to the nearest believed water, take, drink), `recipe.eat`, `recipe.sleep`, `recipe.wander`.
- Tests: the same goal gives a recipe plan; a recipe whose precondition fails falls to the next; a cycle in content fails at load; no recipe means no plan (the agent wanders).

## Spec

- **Firm:** plans share the `Plan` type with GOAP, so execution, claims and invalidation are one path.
- **Firm:** HTN does no search beyond choosing a recipe, so its cost is flat.
- **Firm:** many recipes are expected, so keep `recipes.cue` small files by role.

## Acceptance criteria

- [x] The tests above pass.
- [x] `go tool task check` passes.

## Out of scope

- Resident roles (hauler, miner): M6-23. Monster recipes: M6-21.

## Notes

- `Recipe` gains an optional `when` (a bool formula over `BeliefEnv`), which the spec names but M6-10 left out. Recipes are tried in ID order; the shipped ones use exclusive `when`s so order does not matter. Cycles are rejected at load.
- `Snapshot` gains `Nearest` (the nearest believed landmark per remembered affordance), so a step that brings the agent near a landmark gets its `go_to` target. Built in `planInput`, which calls `Known` a second time per known affordance.
- New content the recipes need: operators `go_to_food`, `wander_food`, `go_to_rest`, `wander_rest`, `sleep`, `roam` (its only effect, leaving water, is a harmless stand-in because an operator needs an effect) and `action.sleep` (satisfies `drive.rest`; first values). Recipes sit in `recipes.cue` (wander) and `recipes/` (drink, eat, sleep), one file per need.
- No claims are made by HTN plans. Docs: `docs/content-guide.md` recipe bullet.
- Follow-up: M6-16.F1 (Planner.Plan cannot return an error, so a failing recipe formula is skipped).
