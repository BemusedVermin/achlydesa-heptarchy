---
id: M2-06
title: The seven forms as actions, with types and timing
status: done
milestone: M2
kind: content
depends_on: [M2-04]
discovered_by: null
size: M
---

# M2-06 · The seven forms as actions, with types and timing

## Goal

The seven forms are content actions (M1-20), so changing how combat plays means editing CUE. Code adds only what's specific to types: the `type` choice, the checks that a type can be used on a target, and the way a type changes any form it's loaded into. The sample numbers are tuned to be lethal from the start.

## Context

- Game design: "Combat → Moves = form × type", "Combat → Forms", "Combat → Types", "Combat → Timing", "Combat → Lethality".
- This milestone's README, "The combat model".
- M1-20 actions and their hooks, M2-03 types and the pack, M2-04 the combat primitives.

## Deliverables

- `internal/content/types.go`: `DamageType` gains `Windup`, `Recovery` (Expr ticks added) and `OnLand []EffectRef`; `Template.Combatant`; regenerate the schema.
- `content/data/forms.cue`: the seven forms as `action.*` entries; `content/data/types.cue`: type modifiers.
- `internal/core/comp/comp.go`: `Combatant`, `Allegiance`, `Graftable`.
- `internal/core/combat/types.go` (the choice provider, validator and modifier), tests.

## Spec

**Combatants:** `Template.Combatant{ Faction; Types []api.ContentID (innate); Graftable bool }`. Its spawner adds `comp.Combatant`, `comp.Allegiance` and `comp.Graftable`. Which forms a body knows is `Template.Actions` (M1-20).

**The `type` choice provider** offers the body's innate types, plus each consumable type whose item is in its party's pack. Forms with `choices: ["type"]` get one option per type.

**The validator** rejects an intent when:

- the target part is immune to the chosen type;
- the type's condition (M2-03's `TypeEnv`, such as Name only working on its owner) fails;
- the form needs a graftable target and doesn't have one.

**The modifier** adds the chosen type's `Windup` and `Recovery` ticks and appends its `OnLand` effects. It adds `consume` of the type's item to `OnStart`, so the item is spent when the swing starts, hit or miss.

**The forms, provisionally.** Numbers are placeholders for M7, but they must already feel lethal: with human parts at 3 integrity (5 for the vital torso), an ordinary Strike takes off an arm in one hit and a torso in two.

| Form | Uses | Target, range | Wind-up / recovery | On land |
| --- | --- | --- | --- | --- |
| Strike | 1 `grant:action` | part, 1 | 3 / 4 | `damage(parts: "target_part", amount: weak ? 99 : 3)` |
| Sweep | 2 `grant:action` | body, 1 | 4 / 5 | `damage(parts: "target:all", amount: weak ? 2 : 1)`, then `reveal(claim: "claim.material")` |
| Sever | 1 `grant:action` | part, 1 | 2 / 5 | `damage(parts: "target_part", amount: 1)`, `sever(parts: "target_part", when: weak)` |
| Pin | 1 `grant:action` | part, 1 | 1 / 3 | `lock(channels: "target_part", ticks: 8)` |
| Stand | 1 `body` | self | 0 / 0 | `add_status(bodies: "self", status: "status.standing", ticks: 10)`, `delay(channels: "crawling:all", ticks: 8)` (M2-06.F2; first `others:all`), `ready(channels: "self:all", count: 99)` |
| Give | 1 `grant:action` | ally, 1 | 0 / 6 | `ready(channels: "target:all", count: 1)` |
| Graft | 2 `grant:action` | ally, 1 | 6 / 6 | `attach_part(from: "pack", item: choice.piece)`; needs a graftable target and a `piece` choice from the pack |

Every attack is `sight: true`, and Connects is the default (still in range and sight). Ranged forms are content for later.

Example type modifiers, to prove the hooks:

- **Soak** adds 2 to recovery, because it's heavy.
- **Press** adds `interrupt(bodies: "target", when: winding_up)`, so a Press landing on a body mid-wind-up staggers it.
- **Light** has `consume` and nothing more; it's gone once opened.

Body-wide effects apply in UID order. Strike's 99 stands in for "comes off in one blow". File a `question` follow-up on whether Strike and Sever should differ on a weakness the way these placeholders do. Stand freezes everyone who crawls; the game design's canon says "In the deep, anyone who stands freezes everyone who crawls." Until crawling is modeled, `others:all` stands in; file a follow-up to add a `crawls` tag to bodies and narrow the selector.

## Acceptance criteria

- [x] Each form starts and lands on an adjacent test dummy with the expected events, occupying the right channels: Strike one arm, Sweep and Graft two.
- [x] Sweep isn't offered while only one arm is free.
- [x] Strike loaded with Soak on a clay part severs it on landing and spends one lake water at the start. With no lake water, the option is disabled with a reason.
- [x] The dummy stepping out of reach before a Strike lands makes it miss, and the lake water is still spent.
- [x] Name on the wrong target, and Relic against old matter, are rejected by the validator with reasons.
- [x] A Press landing on a body that's winding up cancels that swing.
- [x] In a one-cell soop, the member behind the front one has no attack options against the creature beyond it.
- [x] `go tool task check` passes.

## Out of scope

- Alert and death (M2-07). The bodies view and creature behavior (M2-08).

## Notes

**Assumptions and deviations**

- **Human part integrity** is now 3 (head, arms, legs) and 5 (torso), as the spec's lethality sentence needs. Flesh is weak to Cut, so a Cut Strike takes off any human part in one blow; the "ordinary" hits in the tests use Press. Crawler parts are unchanged (M2-11 tunes). Three numbers in `combat/effects_test.go` followed.
- **Stand has recovery 1, not 0.** `actions` refuses a form that takes no time (`too_quick`), and Stand's `ready self:all` frees the channel at once anyway. Stand also `requires` that the body is not already `status.standing`, or it could freeze everyone for ever; for that `ActionEnv` gained `Statuses` (the actor's status IDs), and `status.standing` is new content.
- **Name now consumes `item.name`** (new, bound, kind `name`). Without an item the type's condition `target_uid == item_bound_to` had nothing to compare to, so Name could never be used. The pack holds one stack per learned name, tied to its owner; the swing spends it, which is "speaking it uses it up". The condition's `item_bound_to` is the target's UID when the pack holds an item bound to it, else the first bound stack's owner, else 0. Keepsake works the same. M2-05's trigger idea for Name is no longer needed for the cost.
- **`choice.<name>` in text parameters.** Graft's `attach_part(item: choice.piece)` needed a text parameter that reads one of the action's choices; `effects` reads `choice.` prefixed text parameters from `Context.Choices` (`effects.ChoicePrefix`), and `attach_part`'s content check skips them. Choice `piece` offers the pieces in the pack.
- **Provider and the disabled option.** The `type` provider follows the spec (innate plus consumable types whose item is in the pack), and M2-06.F4 then widened it to every consumable type, so with no lake water Soak is listed disabled with the pack-short reason.
- **Validator**: a body-aimed move (Sweep) is refused as immune only when every attached part is immune. A form tagged `graft` needs a `Graftable` target. Type timing formulas run over `ActionEnv`; `TypeTicks` returns both. A formula error at play time adds nothing (formulas are compiled at load).
- **Sample content**: forms live in `content/data/forms.cue` but no sample template knows them yet, so existing menus and goldens are unchanged; M2-08/M2-11 hand them out. The two person templates gain `combatant` (faction party, Cut/Press/Relic/Name, graftable). The soop criterion is tested as a line (back member, front member, creature) rather than a authored one-cell map; range 1 is what shuts the back member out.
- **Shared test names**: the `actions` tests' stand-in choice is now called `flavor`, because `combat.Install` registers the real `type`. The `effects_test` rig adds its status instead of replacing the sample ones. The content test of `EffectSites` now checks for the sites it needs rather than an exact list.
- `internal/core/timeline/testdata/rapid/...fail` appeared once during a full run (a flaky property test under load, unrelated: it replays green). I deleted it.

**Files outside the Deliverables**

`internal/content/{actions.go,env.go,registry.go,load.go,types_check.go}` (type timing, `EffectSites` for types, `Statuses`, combatant checks), `internal/core/actions/actions.go` (`Statuses` in the env), `internal/core/effects/effects.go` (`choice.` parameters), `internal/core/combat/effects.go` (install, content check), `content/data/{items,statuses,templates,bodies}.cue`, `internal/core/comp/comp.go`, generated `*_go_gen.cue`, and the tests named above.

**Placeholder numbers.** Every number in the forms, types and part integrity is provisional. The combat tests no longer read them: `combat/fixture_test.go` (`pinNumbers`) sets the numbers they are written against, and `internal/content/types_test.go` sets its own type timings. Tests elsewhere that read sample content and may still assume its numbers: `anatomy_test.go` (channel and grant counts follow the human body plan), `timeline_test.go` (step timing), `internal/content/bodies_test.go`.

**Follow-ups filed (F2-F4 were later started with the owner's approval and done in this PR; F1 stays proposed):** [M2-06.F1](../followups/M2-06.F1-strike-sever-weakness.md) (question: Strike vs Sever on a weakness), [M2-06.F2](../followups/M2-06.F2-crawls-tag.md) (`crawls` tag for Stand), [M2-06.F3](../followups/M2-06.F3-per-part-weakness-in-sweeps.md) (per-part weakness in Sweep), [M2-06.F4](../followups/M2-06.F4-show-unavailable-types.md) (spent types as disabled options).
