---
id: M2-06
title: The seven forms as actions, with types and timing
status: done
milestone: M2
kind: content
depends_on: [M2-13]
discovered_by: null
size: M
---

# M2-06 · The seven forms as actions, with types and timing

## Goal

The seven forms, what each type does to them, and the rules of weakness are all content, built from M1-20's actions and M2-13's records, rules and choices. This task changes no Go outside tests; changing how combat plays means editing CUE. The sample numbers are lethal from the start.

## Context

- Game design: "Combat → Actions = form × type", "Combat → Forms", "Combat → Types", "Combat → Timing", "Combat → Lethality".
- This milestone's README, "The combat model".
- M1-20 actions; M2-04 the combat primitives; M2-05 triggers; M2-13 records, rules, choices from content, and chosen records' mods.

## Deliverables

- `content/data/forms.cue` (new): the seven forms as `action.*` entries.
- `content/data/rules.cue`: the rules the forms use.
- `content/data/types.cue`: each type's mods. `content/data/choices.cue`: the `type` and `piece` choices.
- `content/data/statuses.cue`: `status.standing`.
- `content/data/templates.cue`: which forms and innate types each person knows, the `graftable` tag, and a test dummy. `content/data/bodies.cue`: lethal integrity.
- `internal/core/combat/forms_test.go`: the acceptance criteria, run against the real `content/data`.

## Spec

**No Go.** Every rule here is a formula, a record or an effect. If a form needs something M2-13's engine can't express, don't add Go: file a follow-up against M2-13, set this task to `blocked`, and stop.

**Rules (`rules.cue`), provisional.** M2-13's `weak` and `immune` stay. Add:

- `hit`: `rule.immune ? 0 : (rule.weak ? 99 : 3)`. An ordinary hit takes 3; a weakness takes the part off in one blow (99 stands in for "comes off"); an immunity does nothing.
- `graze`: `rule.immune ? 0 : (rule.weak ? 2 : 1)`, for Sweep.

An immune part isn't refused; the swing lands and does nothing. A refusal in the menu would tell a party what a part is made of before it has learned it, and the game design says Relic "does nothing to old matter".

**Choices (`choices.cue`).**

- `type`: from `types`, offered when `record.id in source.template.lists.types` (the body's innate types) or the pack holds the item it consumes: `any(record.lists.consumes, {pack[#] > 0})`. This replaces M2-13's `innate` tag, which goes.
- `piece`: from `items`, offered when `"piece" in record.tags && pack[record.id] > 0`.

**Innate types (`templates.cue`).** The design gives everyone Cut and Name, and the Reaper Press (the thrukk grip). Put these in each person template's `lists.types`. Relic is "innate to grafts", which a choice offer can't see yet: leave it out and file a follow-up.

**The forms, provisionally.** Numbers are placeholders for M7, but they must already feel lethal: with human parts at 3 integrity (5 for the vital torso), an ordinary Strike takes off an arm in one hit and a torso in two. Retune `bodies.cue` to match.

| Form | Uses | Target, range | Choices | Wind-up / recovery | On land |
| --- | --- | --- | --- | --- | --- |
| Strike | 1 `grant:action` | part, 1 | `type` | 3 / 4 | `damage(parts: "target_part", amount: "rule.hit")` |
| Sweep | 2 `grant:action` | body, 1 | `type` | 4 / 5 | `damage(parts: "target:all", amount: "rule.graze")`, then `reveal(parts: "target:all", claim: "claim.material", value: "= target.material.id", source: "source.sweep")` |
| Sever | 1 `grant:action` | part, 1 | `type` | 2 / 5 | `damage(parts: "target_part", amount: "rule.immune ? 0 : 1")`, `sever(parts: "target_part", when: "rule.weak")` |
| Pin | 1 `grant:action` | part, 1 | none | 1 / 3 | `lock(channels: "target_part", ticks: "8")` |
| Stand | 1 `body` | self | none | 0 / 1 | `add_status(bodies: "self", status: "status.standing", ticks: "10")`, `delay(channels: "others:all", ticks: "8")`, `ready(channels: "self:all", count: "99")` |
| Give | 1 `grant:action` | ally, 1 | none | 0 / 6 | `ready(channels: "target:all", count: "1")` |
| Graft | 2 `grant:action` | ally, 1 | `piece` | 6 / 6 | `attach_part(from: "pack", item: "= chosen.piece.id")`; `requires: "\"graftable\" in target.template.tags"` |

- Every attack has `sight: true`, and Connects is the default (still in range and sight). Ranged forms are content for later.
- Stand needs one tick in all, because an action can't take zero; its `ready` frees the stander's channels at once.
- Per-part evaluation (M2-13) makes Sweep hit each part by its own weakness and reveal each part's own material.
- Body-wide effects apply in UID order.

**Type mods (`types.cue`).**

- Every consumable type spends its item when the swing starts, hit or miss (M2-13's migration already does this).
- **Soak** adds 2 to recovery, because it's heavy.
- **Press** adds `interrupt(bodies: "target", when: "winding_up")` on landing, so a Press that lands on a body mid-wind-up staggers it.
- **Light** has its `consume` and nothing more; it's gone once opened.
- **Name** and **Keepsake** keep M2-13's `requires` (the pack holds the item bound to the target), each with its own refusal text.

**Templates.** The person templates know all seven forms. Mark them `graftable`. Add `template.test.dummy`, a human body with no controller of its own, for the tests.

**Follow-ups to file:**

- A `question`: should Strike and Sever differ on a weakness the way these placeholders do?
- Stand freezes everyone who crawls; the game design's canon says "In the deep, anyone who stands freezes everyone who crawls." Until crawling is modeled, `others:all` stands in. File a follow-up to add a `crawls` tag to bodies and narrow the selector.
- Relic as innate to grafts (see Innate types).
- A `question`: a disabled option is information. Refusing Name or Keepsake on the wrong body tells the player whom the item belongs to, which could undercut telling holders from false positives. The reversible alternative is to let it land and do nothing, as immunity does.

## Acceptance criteria

- [x] Each form starts and lands on an adjacent dummy with the expected events, occupying the right channels: Strike one arm, Sweep and Graft two.
- [x] Sweep isn't offered while only one arm is free.
- [x] Strike loaded with Soak on a clay part severs it on landing and spends one lake water at the start. With no lake water, Soak isn't offered.
- [x] Soak adds 2 ticks to the recovery of every form it's loaded into.
- [x] The dummy stepping out of reach before a Strike lands makes it miss, and the lake water is still spent.
- [x] Sweep with Soak on a dummy with a clay arm takes 2 from the arm and 1 from every other part, and the party learns each part's material.
- [x] Relic on old matter lands and does nothing: no `PartDamaged`.
- [x] Name with no name bound to the target shows as disabled, with its refusal text.
- [x] A Press landing on a body that's winding up cancels that swing.
- [x] Stand adds `status.standing`, delays the other bodies' channels, and leaves the stander's channels free.
- [x] Graft is offered only on graftable allies and only with a piece in the pack; landing attaches the piece and spends it.
- [x] In a one-cell soop, the member behind the front one has no attack options against the creature beyond it.
- [x] No Go outside `_test.go` files changes: `git diff --stat master -- '*.go' ':!*_test.go'` is empty.
- [x] `go tool task check` passes.

## Out of scope

- Factions on templates, alert and death (M2-07). The bodies view and creature behavior (M2-08).
- Any engine change; file it against M2-13 instead.

## Notes

Implemented as specified; no Go outside `_test.go` files changed.

**Deviations and assumptions**

- Both person templates know Cut, Name and Press (`lists.types`); the spec's "the Reaper Press" was read as "each person template". Relic is left out (follow-up M2-06.F3).
- The `innate` tag is gone from types and from `vocabulary.tags.types`; `choices.type` reads `source.template.lists.types`.
- Human parts were 8/6/10/12 (head/arm/legs/torso); they are now 3/3/3/5 in `bodies.cue`, as specified. The legs at 3 mean a fall of four cells or more now takes the legs off (the fall trigger deals `cells - 1`); that is placeholder tuning for M7.
- A form whose own `requires` is false (Graft onto a body that isn't `graftable`) is not listed at all, rather than listed disabled. A choice record's `requires` (Name, Keepsake) lists the option disabled with the refusal.
- `template.test.dummy` has the `idle` controller: a body needs an actor to be a target, and `idle` is the built-in "no controller of its own". It is `graftable`.
- Stand has no guard against standing twice; the spec doesn't give one.
- Added texts: `text.status.standing`, the seven form names, `text.template.test.dummy`, and one refusal text each for Name and Keepsake.
- The forms are now on the sample party, so a body is ready while its legs recover (its arms are free). Tests that walk the party (`explore`, `save`, `tui`) now give the sample party the step only; `combat`'s older fixture pins the previous integrity numbers; `actions`' fixture gives its fighter innate types; `content`'s `TestSampleActions` no longer expects the party to know only the step.

**Files touched outside the Deliverables list**: `content/data/vocabulary.cue` (the `innate` tag goes), tests in `internal/content`, `internal/core/actions`, `internal/core/combat`, `internal/core/explore`, `internal/core/save` and `internal/frontend/tui` as above.

**Tests**: `internal/core/combat/forms_test.go` runs every acceptance criterion against the real `content/data`. Its soop test uses range 1: the member behind is two cells from the creature.

**Follow-ups filed**: M2-06.F1 (question: Strike and Sever on a weakness), M2-06.F2 (crawls tag for Stand), M2-06.F3 (Relic innate to grafts), M2-06.F4 (question: a disabled option is information), M2-06.F5 (option labels show a type's ID, not its name).
