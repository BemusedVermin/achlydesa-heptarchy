---
id: M6-11
title: "The mind controller: run a plan, claim what it needs"
status: in-progress
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

- [ ] The tests above pass.
- [ ] `go tool task check` passes.

## Out of scope

- Which goal: M6-14. Which planner searches: M6-15, M6-16.
- Movement and thing actions: M6-12 and M6-13.

## Notes

