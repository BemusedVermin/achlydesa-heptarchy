---
id: M3-11
title: "Residents: needs, roles and schedules in town"
status: done
milestone: M3
kind: feature
depends_on: [M3-04, M3-08]
discovered_by: null
size: M
---

# M3-11 · Residents: needs, roles and schedules in town

## Goal

A town is full of people living by their needs. Residents drink from the town's cistern, haulers fetch water on a schedule, everyone sleeps at night, and when water runs out they leave through a gate. This is the first town map.

## Context

- The agents model in [README.md](README.md), "People" and "Towns".
- `docs/game-design.md` → "People with needs" (needs, behaviors, roles and customs as constraints, towns can empty) and "World and towns".
- M3-03 and M3-04's primitives, M3-08's town stock, M3-06's gates.

## Deliverables

- `content/data/maps/test_town.cue` (new): a small town with a well, a cistern (a stock mark), houses (rest marks) and a gate.
- `content/data/templates.cue`: resident templates (a water hauler and a plain resident), on the `agents` controller, in a `faction.people` that neither the party nor the wild is hostile to by default.
- `content/data/behaviors.cue`, `content/data/needs.cue`: people's needs (water, food, rest, safety, belonging, faith and duty, grief) with placeholder curves, and their behaviors.
- `internal/core/agents/primitives.go`: `leave` (walk to a gate and cross it).
- `content/data/factions.cue`; the playground gets a gate to the town.
- `internal/core/agents/residents_test.go`.

## Spec

**Firm:**

- **Roles are behavior lists, schedules are considerations.** A role is a template (or a template mixin) listing the behaviors it may do; a schedule is a consideration over `phase`, such as fetching water only by day. Customs are considerations too. No role or schedule is named in Go.
- **Water:** the hauler's `carry` takes water from the well and delivers it to the cistern stock; residents' drink behavior is `consume` with an action whose effects take one water from the stock and `satisfy` `need.water`.
- **Leaving:** `leave` walks to a gate and moves the body through it; when the destination map is off screen, the body is counted into that region's population of its family (`Migrated`) and removed. A resident whose water need is high and whose town stock is empty scores `leave` highest.
- Needs other than water, food and rest have curves but no behavior that meets them yet; their placeholder curves keep them below any threshold.

**Suggestion:** the stock mark is where the cistern is drawn, so the player sees the haulers walk to it.

## Acceptance criteria

- [x] Exit criterion: a test runs a day in the test town; the hauler walks from the well to the cistern during the day phases and not at night, and the town's water stock rises.
- [x] Exit criterion: with the well emptied, the stock runs out, residents walk to the gate and leave, and the neighbouring region's count of residents rises.
- [x] In the playground, the party can walk into the town and watch the hauler work.
- [x] `go tool task check` passes.

## Out of scope

- Generic NPCs and their replacement (M3-12).
- Trade, worship, town events and calling the Reaper (later milestones: M4 for reaping calls).
- Relationships and grief from a death (M4's cast model).

## Notes



- `leave` (`agents/leave.go`) walks to the nearest gate and crosses it. On the party's map the body appears at the gate's mark; elsewhere it is counted into the destination region's population and despawned. Its family is its `comp.Family`, or the first family whose templates list its template. No `Migrated` event: M3-10 owns it (follow-up M3-11.F1).
- Not in the spec: a `take_stock` effect (a resident's drink and eat actions take one item from the town stock, from beside a stock mark), `consume` now starts self- and no-target actions after walking next to the behavior's target, and `ConsiderEnv.Stock` (the town's stock by item, for the leave and drink considerations).
- Residents are counted in the test town's region and promoted when the party arrives; only the well is a spawn. This keeps them off the timeline, and out of the encounter goldens, until the party is there. Goldens were regenerated: only body UIDs moved, by the well.
- Roles are two templates (plain, hauler); schedules are phase considerations in `behaviors.cue`. Rest: a sleep action at the single `house` mark at night, not several houses. Safety, belonging, faith, duty and grief rise at 0.005 per phase with no thresholds and no behaviors.
- A resident that is thirsty and finds no water is slowed at need 0.5 and takes damage from 0.9, so it must leave before then; the tests raise water's rise to 0.1 to run quickly.
- Edited existing tests: the hub now neighbours the town (`regions_test.go`); `needs_test.go` drops `action.drink_stock` with `action.drink`; `controller_test.go` waits for the finished behavior to clear instead of three fixed waits, which other bodies' events made unreliable.
- The playground hub is widened by two columns for a gate to `map.test.town`.
- golangci-lint v2.15.0 had to be built with go1.27.4 (`GOTOOLCHAIN=go1.27.4`) to run here. No reviewer agent was available; no separate review was done.
- Follow-up filed: M3-11.F1.
