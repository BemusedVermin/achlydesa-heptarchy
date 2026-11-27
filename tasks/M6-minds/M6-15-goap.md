---
id: M6-15
title: "The GOAP planner for important individuals"
status: todo
milestone: M6
kind: feature
depends_on: [M6-10, M6-12, M6-13, M6-14]
discovered_by: null
size: M
---

# M6-15 · The GOAP planner for important individuals

## Goal

An important individual plans by search: from its believed state to its goal's wanted state, through content operators, at the least cost.

## Context

- `docs/game-design.md` → "How minds work" (important individuals plan with GOAP; chosen when generated).
- Orkin's F.E.A.R. GOAP, and the Ubisoft postmortem on its production cost: keep the operator vocabulary small and debug with plan traces.
- M6-10 gives state and operators; M6-11 gives the `Planner` interface.

## Deliverables

- `internal/core/agents/goap.go`: a planner registered as `goap`. Regressive or forward A* over `State`, with an operator's `cost`, a node cap `agents.goap_nodes` (content) and a depth cap.
- A plan cache keyed by (state hash, goal), cleared when beliefs change.
- A plan trace in the message log under a debug flag, so a plan can be read.
- Tests: finds `go_to well → take water → go_to larder → put` from beliefs; returns no plan within the cap rather than hanging; the same state and goal always give the same plan; an operator with unmet preconditions is never used.

## Spec

- **Firm:** search is deterministic: ties break by operator ID. It reads a snapshot and draws no stream.
- **Firm:** a plan is made again only when it is invalid (M6-11), never on every decision.
- **Firm:** content names who uses GOAP: a template's `mind.planner: goap`, or a spawn entry that sets it (M6-20). Nothing in code picks "important".

## Acceptance criteria

- [ ] The tests above pass, with a benchmark of one plan recorded in Notes.
- [ ] `go tool task check` passes.

## Out of scope

- HTN: M6-16. Combat: the `hunt` controller.

## Notes

