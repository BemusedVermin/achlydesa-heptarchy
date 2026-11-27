---
id: M6-11
title: "The mind controller: run a plan, claim what it needs"
status: done
milestone: M6
kind: seam
depends_on: [M6-01, M6-10]
discovered_by: null
size: M
---

# M6-11 · The mind controller: run a plan, claim what it needs

## Goal

A `mind` controller decides for an agent: it holds a plan, starts the next step as a timeline action, claims what the plan needs, and drops or remakes the plan when it stops holding.

## Context

- `docs/game-design.md` → "How minds work" (Reservations; combat is handed over); "Keeping it affordable".
- `internal/core/timeline/controllers.go` and M6-01 (hand-off); `internal/core/actions` (Start, Startable).
- RimWorld's job reservations are the model for claims.

## Deliverables

- `internal/core/agents/mind.go`: the `mind` controller (a name templates use as `actor.controller`) and a `Planner` interface, `Plan(snapshot, goal) (Plan, bool)`, registered by name; a planner is chosen per body by its template's `mind.planner`.
- `internal/core/comp/`: `Mind{Goal, Plan, Step, Planner}` and `Claims{...}` on the body; regenerate the schema.
- Claims: a claim names a target UID and what is claimed (an item count, a bed); a second agent's plan cannot claim the same. Clashes resolve when plans are applied, by UID order; the loser replans on its next decision.
- A plan is invalid when a step's action fails to start, a claim is lost, a belief the step needs is contradicted, or an export happens. An invalid plan is dropped and made again.
- A trivial planner for tests only: one operator, no search.
- Tests: a plan runs step by step; a lost claim drops it; a save in the middle of a plan loads and finishes the same way, with the same hash.

## Spec

- **Firm:** planning reads a snapshot of the world and returns a plan; it draws no shared stream. If it needs randomness it derives a stream from UID and tick, so planning may run on other goroutines and be applied in UID order.
- **Firm:** everything a mind holds (goal, plan, step, claims) is saved; bump the save version.
- **Firm:** a body delegated to another controller (M6-01) is skipped until the delegate resumes.

## Acceptance criteria

- [x] The tests above pass.
- [x] `go tool task check` passes.

## Out of scope

- Which goal: M6-14. Which planner searches: M6-15, M6-16.
- Movement and thing actions: M6-12 and M6-13.

## Notes

- **Please check, outside the Deliverables:** `timeline.Saved` now keeps the pending `Seq` counter. Without it, a game saved while an action was pending reloaded with different `Seq` values and a different hash, which failed this task's save test. The loop in `timeline.Schedule` that resumes above the highest live `Seq` is kept on purpose, as a guard for worlds built without `Import`.
- **Picker seam, not in the Deliverables:** `agents.Picker`, `RegisterPicker` and a built-in `self` picker are the minimum needed to start a step's action, since `does.target` names a picker. M6-12 and M6-13 add the real pickers. `agents.ValidateContent`, now called from `game.validateContent`, fails the game at start when an operator names an unregistered picker, a template names an unregistered planner, or a template uses the `mind` controller with no `mind` block.
- `content.Template` gains `mind: {planner}`. The planner name is copied into `comp.Mind` at spawn, so a hot reload that changes it reaches only bodies spawned later.
- `Plan`, `Step` and `Target` moved to `comp` (as `Plan`, `PlanStep`, `PlanTarget`), because `comp.Mind` holds a plan; `agents` keeps them as aliases. "No plan" is a plan with no steps, so the component holds no pointer. Claims a plan wants are `Plan.Claims`, set by the planner; the claims a body holds are `comp.Claims`.
- A claim with an item claims `n` of the thing's stock (its source plus its inventory), and others may claim the rest; a claim with no item claims the whole thing and clashes with any other claim on it. Claims count stock only when made. A claim is lost when the thing is gone or the body's `comp.Claims` no longer matches its plan; the lost-claim test despawns the well, as nothing else takes a claim away yet. How a take uses up its claim is M6-13's.
- **Clashes resolve as each body decides, not in a separate apply pass.** Planning does not run on other goroutines yet, so the first body to claim wins; bodies that decide at the same tick go in UID order, so the lower UID wins. The loser has no plan and tries again at its next decision. A batch apply for far agents belongs with M6-17.
- A plan is also dropped when `Mind.Goal` changes (so M6-14 only sets the goal) and when an export begins: `comp.Mind.Exports` keeps the export count the plan was made at.
- A step is started only once every channel of the body is free, so steps run one after another. A plan whose step cannot start is dropped and the body waits; it plans again at its next decision.
- `Snapshot.Rand` is `RNG.Derived("agents.plan", hash(UID, tick))`, so planning draws no shared stream.
- Finding clashes scans every body holding claims, once per plan made. Fine for now; M6-28's budget will show whether it needs an index.
- Save version 14.
- Docs: `docs/content-guide.md` (the `mind` block, pickers), `docs/glossary.md` (mind; claim of a plan, as against a knowledge claim).
- Files outside the Deliverables: `internal/content/{types,load}.go`, `internal/core/game/game.go`, `internal/core/timeline/save.go`, `internal/core/save/`, `internal/core/agents/{plan,drives}.go`.
- Follow-ups: none.

