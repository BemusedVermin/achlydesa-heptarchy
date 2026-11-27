---
id: M6-19
title: "Flee, terror and engaging: handing a fight over and back"
status: todo
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

- [ ] The tests above pass.
- [ ] `go tool task check` passes.

## Out of scope

- Dodging by choosing a chooser: left to combat AI work later.

## Notes

