---
id: M6-26
title: "Recruit personality: robbery, desertion and loyalty"
status: todo
milestone: M6
kind: feature
depends_on: [M6-25]
discovered_by: null
size: M
---

# M6-26 · Recruit personality: robbery, desertion and loyalty

## Goal

Two recruits with different weights, both unpaid, leave in different ways: one asks for its pay and quits, another robs the party on the way out. Loyalty keeps some around longer.

## Context

- Folds in follow-up `tasks/followups/M4-03.F1-recruit-personality.md` (marked obsolete by this plan).
- `docs/game-design.md` → "Recruits" (Paying them: personality decides), "How minds work" (personality is the weights).
- M6-03 weights, M6-25 pay and quitting.

## Deliverables

- `content/data/drives.cue`: `drive.greed` and `drive.loyalty` (a desire that grows with time served and is relieved by being paid and by fighting beside the party).
- `content/data/operators.cue`: `operator.rob` (take items from the party's inventory when adjacent to a member and not hostile), with a recipe `recipe.leave_robbing`.
- Recruit templates give typical weights on greed and loyalty; each individual draws its own (M6-03).
- Tests (golden, same seed): a high-greed recruit robs and leaves; a low-greed one quits without taking anything; a high-loyalty one stays several days longer than a low-loyalty one at the same unmet pay.

## Spec

- **Firm:** each outcome is the planner choosing among drives; no personality code, no desertion system.
- **Firm:** a robbed party notices through its inventory and the message log, not an event that announces it.
- **Firm:** a deserter stays neutral to the party; a grudge is M8's.

## Acceptance criteria

- [ ] The golden tests above pass.
- [ ] `go tool task check` passes.

## Out of scope

- Grudges and revenge: M8.

## Notes

