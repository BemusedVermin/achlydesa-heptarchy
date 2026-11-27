---
id: M6-25
title: "Hiring recruits, pay and quitting"
status: todo
milestone: M6
kind: feature
depends_on: [M6-19, M6-23, M6-24]
discovered_by: null
size: M
---

# M6-25 · Hiring recruits, pay and quitting

## Goal

The party can hire a generic person, pay them in what they want, and a recruit whose pay goes unmet grumbles and then quits, all as acts through the planner.

## Context

- `docs/game-design.md` → "Party" → "Recruits" (Price, Paying them; quitting is an act like hiring); "Three tiers of people".
- Added back from the reverted M3: `tasks/M3-agents/M3-15-hiring-recruits.md` and `M3-16-upkeep-desertion.md`.
- Review finding: upkeep must not be a daily automatic deduction, which would be a system (rule 10).

## Deliverables

- `internal/content/`: `Template.Recruit{Wants, Worth}`: item tags the person wants and a scale for what they bring.
- `content/data/actions.cue`: `action.hire` (takes a bundle of wanted items from the party's inventory, scaled by worth, and makes the person a party member under the follow controller, up to `party.slots`) and `action.pay` (hands upkeep from the list to a recruit).
- `content/data/drives.cue` / `operators.cue`: `drive.pay` rises with time since the last payment; operators `operator.grumble` and `operator.quit` (the quit act removes the person from the party, sets allegiance to `faction.deserters`, keeps grafts and pieces).
- `content/data/`: a few recruit templates (guard, miner, field hand) placed in the three towns.
- Tests: hiring takes the bundle and joins the party; an unpaid recruit grumbles and then walks off with its inventory; a paid one stays; a deserter is neutral to the party (M3-16.F2).

## Spec

- **Firm:** the recruit decides to quit through the planner like anyone, with `drive.pay`'s weight and its other drives; nothing deducts upkeep by itself.
- **Firm:** a quit changes allegiance as its own effect, not through a desertion hook.
- **Firm:** a recruit takes one of the four fight slots; a fifth slot stays a playtest question.

## Acceptance criteria

- [ ] The tests above pass.
- [ ] `go tool task check` passes.

## Out of scope

- Robbery, loyalty, differences between recruits: M6-26.
- Exotic recruits: M6-27.

## Notes

