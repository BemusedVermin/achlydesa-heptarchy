---
id: M3-20
title: Harness columns and the camping check
status: todo
milestone: M3
kind: test
depends_on: [M3-16, M3-19]
discovered_by: null
size: M
---

# M3-20 · Harness columns and the camping check

## Goal

The harness measures the living deep. Its CSV gains deaths, migrations and towns emptied, a policy can let days pass, and a 100-seed test shows no degenerate camping: no family settles into one town.

## Context

- `docs/game-design.md` → "Keeping it affordable" (harness logging).
- `docs/go-technical-design.md` → "The harness".
- `internal/harness` (`run.go`, `combat.go`, the lethality test as the pattern), `cmd/harness`.
- M3-10's events, M3-19's sample world.

## Deliverables

- `internal/harness/emergence.go`: counting events per run; the new CSV columns.
- A policy that rests and waits so days pass (through the view's options, as `RandomPolicy` does).
- `internal/harness/emergence_test.go`: the 100-seed check.
- `cmd/harness`: a flag to run the sample world.

## Spec

- Firm: columns `deaths`, `migrations`, `towns_emptied`, counted from events, added after the existing columns.
- Firm: **degenerate camping** is any family with more than `camping_share` of its living members in one town region for more than `camping_days` days in a row; both numbers are content (or harness flags), placeholders for M7.
- The 100-seed test may be slow: mark it to run outside `-short`, as the lethality test is run, and say in Notes how long it takes.

## Acceptance criteria

- [ ] Exit criterion: 100 seeds of 30 days on the sample world show no degenerate camping.
- [ ] `go run ./cmd/harness` on the sample world writes the three new columns with nonzero values on at least one seed.
- [ ] Exit criterion: a golden test replays the brief's scenario: a resident fetches water, a town short of water loses people.
- [ ] `go tool task check` passes.

## Out of scope

- Tuning the numbers (M8).

## Notes


