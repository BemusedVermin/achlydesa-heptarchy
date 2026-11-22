---
id: M2-11
title: Sample content, lethality targets and golden encounters
status: todo
milestone: M2
kind: test
depends_on: [M2-09]
discovered_by: null
size: M
---

# M2-11 · Sample content, lethality targets and golden encounters

## Goal

A small, complete content set that exercises every form, type and primitive, tuned so that combat is lethal and a blind party loses, plus the tests that pin combat's behavior down so later work can change rules without silently breaking them.

## Context

- Game design: "Combat" (especially "Lethality"), "Party → Protagonists in combat", "Technical plan → MVP definition" (minimal content).
- Tech design: "Harness and testing → Tests".

## Deliverables

- `content/data/`: the protagonists (the Reaper: Press, Pin, Stand; the elevated salken: Relic, Sever, Graft, graftable), three enemy templates (a moving-clay thing, a pre-entropy remnant, a flesh creature with bone and eye parts), and their spawns on `map.test.hall`.
- `internal/core/combat/testdata/encounters/*.jsonl`: scripted encounters with golden event logs.
- `internal/core/combat/golden_test.go`, `property_test.go`
- `internal/harness`: two policies and CSV columns for combat (ticks on alert, landings, misses, parts severed, deaths, outcome). `blind` takes the first offered option, never Sweeps first, and carries only the starting pack. `informed` knows every material and loads matching types into the right forms.

## Spec

**Golden encounters:** each is a seed plus a command log. The test replays it and compares the delivered events, encoded with the codec, to a golden file. Regenerate with `-update`. Cover:

1. Soak plus Strike severing clay.
2. A Sweep revealing materials, then a targeted Sever.
3. Pin locking an arm that's winding up, which cancels its swing, then Stand freezing everyone in sight.
4. Give and Graft between members.
5. A crawling part outliving its body.
6. A dodge: stepping out of reach before a swing lands.
7. Outpacing a creature whose legs were severed, until the alert ends.

From M1-18.F1, also: a test in `internal/core/save` that saves mid-fight with actions in flight (include a kind-specific component on a pending entity), loads, runs on, and gets the same `TimelineView` `lands` entries and the same hash as continuing without a save.

**Properties** (`rapid`, over random legal option choices for up to 300 ticks):

- No part has `PartOf` two bodies.
- No channel starts an action while it's busy.
- Only bodies in reach are ever hit.
- Every pending action lands, misses or is canceled exactly once.
- No two bodies ever share a cell.
- A severed part never acts for its former body.
- Dead bodies never decide.
- Replaying the same choices gives the same hash.
- Every option offered resolves without `Rejected`.

## Acceptance criteria

- [ ] Seven golden encounters pass, and each covers what its name says (checked by reading the golden logs).
- [ ] The properties hold over 200 cases each.
- [ ] **Lethality:** over 100 seeds each, a `blind` party loses at least 60% of fights against the hall's creature and an `informed` party wins at least 80%, and the median fight lasts under 60 ticks. Record the measured numbers in Notes. If they miss, tune content, not code, and say what changed. These targets are placeholders for M7.
- [ ] `harness run --seeds 50 --steps 500` completes with the combat columns filled, and no run hits a bus cascade error.
- [ ] `go tool task check` passes.

## Out of scope

- Balancing numbers (M7).

## Notes
