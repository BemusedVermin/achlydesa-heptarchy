---
id: M6-19
title: "Flee, terror and engaging: handing a fight over and back"
status: done
milestone: M6
kind: feature
depends_on: [M6-01, M6-12, M6-14]
discovered_by: null
size: M
---

# M6-19 · Flee, terror and engaging: handing a fight over and back

## Goal

An agent whose goal is to fight hands the body to the `hunt` controller; one whose terror outweighs it flees instead. When the fight ends, the mind decides again.

## Context

- `docs/game-design.md` → "Party" → "Recruits" (terror makes them flee mid-fight, an ordinary behavior); "Combat sides".
- `internal/core/combat/controller.go`: `hunt`, `ScoreChooser`. Combat AI stays separate (owner decision).
- M6-01 provides `DelegateTo` and `Resume`.

## Deliverables

- `content/data/drives.cue`: `drive.terror` (a need-like drive raised by triggers when a body is hit or a neighbour dies, not by a timer).
- `content/data/goals.cue` and `operators.cue`: `goal.fight` (when a hostile is believed near) and `goal.flee` (when terror is high), `operator.engage` (calls `DelegateTo(hunt)`) and `operator.flee` (steps to the believed cell farthest from known hostiles).
- `hunt` calls `Resume` when its prey is gone or the body's goal changed.
- Tests: a creature in sight of a hostile engages and fights; a terrified one runs instead; after the fight the mind resumes with its old goal; hunting behavior and M2 goldens are unchanged for creatures not under a mind.

## Spec

- **Firm:** terror is raised by an event trigger in content (`damage` taken, a neighbour's death) and decays by its own rise formula.
- **Firm:** a body that would flee with nowhere to go fights.
- **Firm:** nothing here forces a desertion or a rout; those are agents choosing.

## Acceptance criteria

- [x] The tests above pass.
- [x] `go tool task check` passes.

## Out of scope

- Dodging by choosing a chooser: left to combat AI work later.

## Notes


- Terror is `drive.safety`, not a new `drive.terror`: `docs/game-design.md` says safety "is the same drive as terror". Its `rise` is now `-1`, so it fades by one a phase; being hit adds 2 through a `combat.part_damaged` trigger (`_terrorOnHit` in `templates.cue`). Owner: say if you want a separate drive.
- Hostiles are "believed near" from the at-claims an agent holds: seen within the new `agents.hostile_ticks` (12), within sight radius of where seen, not seen dead, not out of the fight. `BeliefEnv` gains `hostile`.
- `goal.fight` weighs 0.5 when a hostile is believed near; `goal.flee` is safety's `urge * weight` and applies only with a hostile near. First values, for tuning.
- `operator.engage` has the `engage` picker, which calls `DelegateTo(hunt)` and returns no intent, so the plan is dropped. Its action (`action.strike`) is only there to satisfy content. `operator.flee` steps to the neighbouring cell farthest from the nearest believed hostile; with none farther, the `flee` picker delegates to `hunt`, so a body with nowhere to go fights.
- Deviation: `hunt` already resumes the mind when prey is gone. A changed goal is noticed by a new per-entry system, `agents.rethink`, which re-chooses a delegated mind's goal every `think_ticks` and resumes it on a change; `combat` cannot see goals.
- `think` now calls `sense` first, so a mind learns what it sees when it chooses.
- Triggers are not attached to any template (it would change saved trigger IDs and goldens); no template has a mind yet. The two tests build their own.
- Follow-ups: M6-19.F1 (a neighbour's death cannot be reached by a trigger: `body_died` has no source), M6-19.F2 (attach drive.safety and the trigger to mind templates).
- Schemas regenerated for `hostile_ticks`.
