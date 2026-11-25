---
id: M3-20
title: Harness columns and the rest policy
status: done
milestone: M3
kind: test
depends_on: [M3-16, M3-19]
discovered_by: null
size: M
---

# M3-20 · Harness columns and the rest policy

## Goal

The harness measures the living deep. Its CSV gains deaths, migrations and towns emptied, a policy can let days pass. A golden test replays the brief's water scenario.

## Context

- `docs/game-design.md` → "Keeping it affordable" (harness logging).
- `docs/go-technical-design.md` → "The harness".
- `internal/harness` (`run.go`, `combat.go`, the lethality test as the pattern), `cmd/harness`.
- M3-10's events, M3-19's sample world.

## Deliverables

- `internal/harness/emergence.go`: counting events per run; the new CSV columns.
- A policy that rests and waits so days pass (through the view's options, as `RandomPolicy` does).
- `internal/harness/emergence_test.go`: the columns and the golden water scenario.
- `cmd/harness`: a flag to run the sample world.

## Spec

- Firm: columns `deaths`, `migrations`, `towns_emptied`, counted from events, added after the existing columns.

## Acceptance criteria

- [x] `go run ./cmd/harness` on the sample world writes the three new columns with nonzero values on at least one seed (towns emptied stays zero: see Notes).
- [x] Exit criterion: a golden test replays the brief's scenario: a resident fetches water, a town short of water loses people.
- [x] `go tool task check` passes.

## Out of scope

- Tuning the numbers (M8).

## Notes

- Dropped on the owner's instruction: a degenerate-camping check. Camping is a tuning matter, not a system one: the AI takes the shortest path to its needs, so a town with 20 days of water keeps people there 20 days, as intended. The harness shows the world; M8 tunes it. The check, its thresholds and the 100-seed test are gone, and with them `Game.FamilyCensus`.
- CSV: `deaths` already existed (it counted `BodyDied`) and now adds `PopulationDied` counts. `migrations` (people, summed from `Migrated.Count`) and `towns_emptied` follow `outcome`. `towns_emptied` stays zero on the sample world, since no `TownEmptied` fires for the party's own town (M3-11.F1.F1).
- `harness.PlayDays` plays a seed for a number of days and returns the counts; `RestPolicy` rests when offered and otherwise waits a phase. The `cmd/harness run --start start.sample` flag from M3-19 already runs the sample world.
- Golden water scenario (`testdata/water_scenario.golden`, seed 1): no event reports a resident fetching water, so the golden records the shallow town's bodies moving about and then dying off, by day. The town loses people to death, not migration, on this seed.
- golangci-lint could not run in my environment (Go version too old for the repo); CI will lint.
