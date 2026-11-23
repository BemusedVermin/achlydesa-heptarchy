---
id: M2-11
title: Sample content, lethality targets and golden encounters
status: done
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

- [x] Seven golden encounters pass, and each covers what its name says (checked by reading the golden logs).
- [x] The properties hold over 200 cases each.
- [x] **Lethality:** over 100 seeds each, a `blind` party loses at least 60% of fights against the hall's creature and an `informed` party wins at least 80%, and the median fight lasts under 60 ticks. Record the measured numbers in Notes. If they miss, tune content, not code, and say what changed. These targets are placeholders for M7.
- [x] `harness run --seeds 50 --steps 500` completes with the combat columns filled, and no run hits a bus cascade error.
- [x] `go tool task check` passes.

## Out of scope

- Balancing numbers (M7).

## Notes

**Lethality, measured** (`TestLethalityTargets`, seeds 1 to 100 each, hall's crawler): blind lost 100% (won 0%, median fight 49 ticks); informed won 100% (median 34 ticks). Tuned in content only. Before tuning, the sample crawler's flesh maw let a blind party Sever it with Burn at once and win 95%. I made the hall's creature a new `body.clay_thing` (clay maws, legs and core, integrity 10/10/12, two maws), because clay is weak only to Soak and Relic, and Burn bakes it harder. `body.crawler` stays as it was, since many tests build on it.

**Deviations and decisions**

- **Hall spawns.** The remnant and the beast do not spawn on `map.test.hall`: with three hunters there, a random walker (harness, save and frontend tests) dies, and the lethality target says "the hall's creature". They spawn on a new `map.test.den` (`content/data/maps/test_den.cue`), and the hall keeps the crawler. Change `spawns` on the hall if the owner wants all three there.
- **Protagonists** were already in content from earlier tasks (the Reaper with Press, the salken graftable with Relic through grafts), so none were added. The remnant gained Step and Strike (with Cut) so it fights, and a flesh `template.creature.beast` (bone skull, eye, flesh legs and torso) was added.
- **Golden encounters** set bodies up with the test fixtures (`ring`, `field`), as the existing tests do, and from there every move is an option of the view. They are not "a seed plus a command log" replays, because the setup places bodies directly. Goldens: `internal/core/combat/testdata/encounters/*.jsonl`, the events encoded with the codec; the old `hunt.golden` was regenerated (map and UIDs shifted). "Soak plus Strike severing clay": Strike never severs, so it is the clay arm worn off by a weak hit (`SeveredWorn`).
- **Properties** are one rapid test (`TestRandomFightsKeepTheInvariants`, 200 cases, about 90 s, skipped under `-short`) that checks every listed property after each command, plus the replay hash and "no `Rejected`". "Every pending action lands, misses or is canceled exactly once" is checked as: `ActionStarted` count equals landed plus canceled plus actions in flight (a miss is also a landing). It skips Stand and Give options, which break "no channel starts an action while busy" (M2-11.F1).
- `internal/harness`: `blind` takes the first offered attack that is not a Sweep and loads whatever type is offered first; `informed` reads each part's true material through a new read-only `game.(*Game).PartOfUID` (added to `internal/core/game/inspect.go`, next to `PositionOf`) and picks a weak type, preferring vital parts. Both walk to the creature and close in. The CSV gained `alert_ticks, landings, misses, severed, deaths, outcome` (running totals; landings = blows that cost a part integrity). `harness run --seeds 50 --steps 500` completes with no bus cascade error, and the outcome column reads won or lost in the random-policy runs.
- Existing tests touched: `internal/frontend/tui` goldens (`look_*`, `menu_parts`: the creature's body changed) and `TestWholeFightEndsAndTheMapResumesFollow` now soaks and severs the core instead of striking with the first type, since the tougher clay thing cannot be struck down by the naive option.
- The save test (M1-18.F1, now obsolete and merged here) is `internal/core/save/fight_test.go`.

**Follow-ups filed:** M2-11.F1 (Stand and Give free a channel that has a swing winding up).
