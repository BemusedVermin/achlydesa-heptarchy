---
id: M3-11
title: "Residents: needs, roles and schedules in town"
status: todo
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

- [ ] Exit criterion: a test runs a day in the test town; the hauler walks from the well to the cistern during the day phases and not at night, and the town's water stock rises.
- [ ] Exit criterion: with the well emptied, the stock runs out, residents walk to the gate and leave, and the neighbouring region's count of residents rises.
- [ ] In the playground, the party can walk into the town and watch the hauler work.
- [ ] `go tool task check` passes.

## Out of scope

- Generic NPCs and their replacement (M3-12).
- Trade, worship, town events and calling the Reaper (later milestones: M4 for reaping calls).
- Relationships and grief from a death (M4's cast model).

## Notes


