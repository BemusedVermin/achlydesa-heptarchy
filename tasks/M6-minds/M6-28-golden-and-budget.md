---
id: M6-28
title: "Golden scenarios for the exit criteria, and the step budget"
status: todo
milestone: M6
kind: test
depends_on: [M6-17, M6-21, M6-23, M6-26]
discovered_by: null
size: M
---

# M6-28 · Golden scenarios for the exit criteria, and the step budget

## Goal

Each exit criterion has a seeded scenario that proves it, and a world of about 1,000 individuals steps within a stated budget.

## Context

- This milestone's brief → "Exit criteria". `docs/go-technical-design.md` → "Harness and testing" (tests for rules that can really break; golden files under `testdata/`).
- Metrics beyond this are M9's; the harness is not extended here.

## Deliverables

- `internal/core/agents/golden_test.go` (and `testdata/`): the scenarios in "Exit criteria" on small fixture worlds, each asserting from state (positions, inventories, deaths), never from an announcing event.
- `internal/core/agents/budget_test.go`: a benchmark of the sample world's populated start and one phase of stepping, failing above the budget the M6-17 Notes record (loosened by a stated margin for CI).
- A determinism test: the same seed gives the same hash after N phases, and a save in the middle gives the same hash as a run without one.

## Spec

- **Firm:** a scenario that needs a stand-in is a bug in the task that owns the feature, not here.
- **Firm:** no near-duplicate cases; a table for the scenarios.

## Acceptance criteria

- [ ] `go tool task test -- ./internal/core/agents` passes.
- [ ] The benchmark runs under the budget and its numbers are recorded in Notes.
- [ ] `go tool task check` passes.

## Out of scope

- Harness columns and metrics: M9.

## Notes

