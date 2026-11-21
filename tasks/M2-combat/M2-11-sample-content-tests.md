---
id: M2-11
title: Sample content, golden encounters and properties
status: todo
milestone: M2
kind: test
depends_on: [M2-09]
discovered_by: null
size: M
---

# M2-11 · Sample content, golden encounters and properties

## Goal

A small, complete content set that exercises every form, type and primitive, plus the tests that pin combat's behavior down so M3 and later can change things without silently breaking combat.

## Context

- Game design: "Combat", "Party → Protagonists in combat", "Technical plan → MVP definition" (minimal content).
- Tech design: "Harness and testing → Tests".

## Deliverables

- `content/data/`: the protagonists (the Reaper: Press, Pin, Stand; the elevated salken: Relic, Sever, Graft, graftable), three enemy templates (a moving-clay thing, a pre-entropy remnant, a flesh creature with bone and eye parts), and their spawns on `map.test.hall`.
- `internal/core/combat/testdata/encounters/*.jsonl`: scripted encounters with golden event logs.
- `internal/core/combat/golden_test.go`, `property_test.go`
- `internal/harness`: a `first` policy that always picks the first offered option, and CSV columns for combat (turns on alert, parts severed, deaths).

## Spec

**Golden encounters:** each is a seed plus a command log. The test replays it and compares the delivered events, encoded with the codec, to a golden file. Regenerate with `-update`. Cover:

1. Soak plus Strike severing clay.
2. A Sweep revealing materials, then a targeted Sever.
3. Pin and Stand in one turn, with Stand's freeze delaying the others.
4. Give and Graft between members.
5. A crawling part outliving its body.
6. Outpacing a creature whose legs were severed, until the alert ends.

**Properties** (`rapid`, over random legal option choices for up to 60 turns):

- No part has `PartOf` two bodies.
- A body's actions in a turn never exceed its functioning, unlocked action parts plus extras.
- Only bodies in reach are ever hit.
- No two bodies ever share a cell.
- A severed part never acts for its former body.
- Dead bodies never take turns.
- Replaying the same choices gives the same hash.
- Every option offered resolves without `Rejected`.

## Acceptance criteria

- [ ] Six golden encounters pass, and each covers what its name says (checked by reading the golden logs).
- [ ] The six properties hold over 200 cases each.
- [ ] `harness run --seeds 50 --steps 500` completes with the combat columns filled, and no run hits a bus cascade error.
- [ ] `go tool task check` passes.

## Out of scope

- Balancing numbers (M7).

## Notes
