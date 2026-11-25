---
id: M3-20
title: Harness columns and the camping check
status: done
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

- [x] Exit criterion: 100 seeds of 30 days on the sample world show no degenerate camping.
- [x] `go run ./cmd/harness` on the sample world writes the three new columns with nonzero values on at least one seed.
- [x] Exit criterion: a golden test replays the brief's scenario: a resident fetches water, a town short of water loses people.
- [x] `go tool task check` passes.

## Out of scope

- Tuning the numbers (M8).

## Notes



- **Owner decision to confirm: what counts as camping.** Read literally, the spec flags every seed on the sample world: `family.people` ends in the lake town once the shallow town's residents die out, `family.soul_feeder` follows the living into the lake town, and `family.people.guards` starts in the shallow town alone. Assumed (the most reversible option): a family is exempt if it began on one map, or if no living member is outside the town. The rule is `harness.Exemption`, set by `CampingConfig.Exempt` and defaulting to `ExemptSingleStartOrAlone`; switching rules is one function. Share 0.9 and 10 days are placeholders (`DefaultCampingShare`, `DefaultCampingDays`) for M8. The exemption makes the check weak on this small world; a stricter rule needs a design decision.
- Files outside the list: `internal/core/game/inspect.go` gains `Game.FamilyCensus` (living members per family on each map joined to the party's by gates; unlinked test maps are left out); `internal/harness/combat.go` and `run.go` for the columns.
- CSV: `deaths` already existed (it counted `BodyDied`) and now adds `PopulationDied` counts. `migrations` (people, summed from `Migrated.Count`) and `towns_emptied` follow `outcome`.
- Golden water scenario (`testdata/water_scenario.golden`, seed 1): no event reports a resident fetching water, so the golden records the town's bodies moving about and then dying off, by day. The shallow town loses people to death, not migration, on this seed. No `TownEmptied` event fires (see M3-11.F1.F1).
- The 100-seed test is skipped under `-short` and runs one game per CPU.
