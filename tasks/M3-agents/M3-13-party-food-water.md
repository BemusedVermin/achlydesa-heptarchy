---
id: M3-13
title: The party eats and drinks
status: done
milestone: M3
kind: feature
depends_on: [M3-02]
discovered_by: null
size: M
---

# M3-13 · The party eats and drinks

## Goal

Party members get hungry and thirsty. The player feeds them rations and lake water from the inventory; a member who goes without is slowed, then loses integrity, so supplies limit how long the party can stay below.

## Context

- `docs/game-design.md` → "Party" → "Food and water" (the owner's decision: lake water is the drink; slowed first, then weakened) and "Time pressures".
- The agents model in [README.md](README.md), "Needs".
- M3-02's needs, thresholds and `satisfy`; `content/data/items.cue` (`item.lake_water`); `content/data/statuses.cue`; the `timeline.speed` rule and `rules.cue`.

## Deliverables

- `content/data/templates.cue`: `need.food` and `need.water` on the party's member templates.
- `content/data/items.cue`: `item.ration`; the starting inventory carries some.
- `content/data/actions.cue`: `action.eat` and `action.drink`, each taking one item from the inventory and meeting its need; lake water keeps its Soak use.
- `content/data/needs.cue`: thresholds: a status that slows (speed and wind-up through rules) at one level, and damage to vital parts per phase at a higher one.
- `content/data/statuses.cue`, `content/data/rules.cue`.
- `internal/core/api`, the party view: each member's needs, as text the core renders.
- `internal/frontend/tui`: draw it in the party panel.
- Tests in `internal/core/game` or `combat` for the thresholds.

## Spec

**Firm:**

- Everything here is content except the party view's needs line. If a rule cannot be written with M3-02's thresholds and the existing statuses, file a follow-up rather than adding Go that names a need.
- Eating and drinking are ordinary actions in the action menu, available while not in a fight's reach, as the player's choice for each member.
- Placeholder numbers: a member goes from fed to slowed in about two days and to losing integrity in about four; tuning is M7.

## Acceptance criteria

- [x] A test lets four days pass without eating: the member is slowed (steps take longer) after the first threshold and loses vital integrity after the second.
- [x] Drinking lake water lowers thirst and leaves one fewer lake water for Soak.
- [x] In the playground, the party panel shows hunger and thirst, and eating from the action menu lowers hunger.
- [x] `go tool task check` passes.

## Out of scope

- Recruits' food upkeep (M3-16).
- Buying food in town (trade is not in M3; the starting inventory and recruits' upkeep use what the party carries).

## Notes

- **Content only, apart from the view.** Slowing uses rules: `rule.slowed` (the `status.slowed` status, put on by the need's lower threshold) lengthens `action.step`'s recovery and the forms' wind-ups (`rule.lag`). The upper threshold damages `self:torso`. Pace: both needs rise 0.0625 per phase, so slowed at 0.5 (two days) and damage from 0.9 (about four days), a point a phase. The sample `need.food` and `need.water` rises changed to match; `status.slowed` lasts 120 ticks, longer than any phase, so each boundary renews it. Amounts: a ration or a lake water lowers its need by 0.5. All placeholders for M7.
- **Deviations:** "not in a fight's reach" is not enforced: no formula can see the alert (M3-13.F1). The playground (M3-07) does not exist yet, so the party panel criterion was checked against the sample game and a render test. The starting inventory has 6 rations and 6 lake water (was 3). `api.MemberView.Needs` and `text.explore.need` are new.
- **Test changes outside the list:** `internal/core/game/game_test.go` waits shorter (the long waits now starve the party), `internal/core/agents/needs_test.go` test action `drink` renamed `sip` (key clash), four tui menu goldens regenerated (new options).
- **Environment:** golangci-lint could not run here (the toolchain targets Go 1.27.4); format, vet, tests, content and comment checks passed.
- **Follow-ups:** M3-13.F1.
