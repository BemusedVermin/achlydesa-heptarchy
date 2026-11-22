---
id: M2-06
title: The seven forms as actions, with types and timing
status: todo
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
| Stand | 1 `body` | self | 0 / 0 | `add_status(bodies: "self", status: "status.standing", ticks: 10)`, `delay(channels: "others:all", ticks: 8)`, `ready(channels: "self:all", count: 99)` |
| Give | 1 `grant:action` | ally, 1 | 0 / 6 | `ready(channels: "target:all", count: 1)` |
| Graft | 2 `grant:action` | ally, 1 | 6 / 6 | `attach_part(from: "pack", item: choice.piece)`; needs a graftable target and a `piece` choice from the pack |

Every attack is `sight: true`, and Connects is the default (still in range and sight). Ranged forms are content for later.

Example type modifiers, to prove the hooks:

- **Soak** adds 2 to recovery, because it's heavy.
- **Press** adds `interrupt(bodies: "target", when: winding_up)`, so a Press landing on a body mid-wind-up staggers it.
- **Light** has `consume` and nothing more; it's gone once opened.

Body-wide effects apply in UID order. Strike's 99 stands in for "comes off in one blow". File a `question` follow-up on whether Strike and Sever should differ on a weakness the way these placeholders do. Stand freezes everyone who crawls; the game design's canon says "In the deep, anyone who stands freezes everyone who crawls." Until crawling is modeled, `others:all` stands in; file a follow-up to add a `crawls` tag to bodies and narrow the selector.

## Acceptance criteria

- [ ] Each form starts and lands on an adjacent test dummy with the expected events, occupying the right channels: Strike one arm, Sweep and Graft two.
- [ ] Sweep isn't offered while only one arm is free.
- [ ] Strike loaded with Soak on a clay part severs it on landing and spends one lake water at the start. With no lake water, the option is disabled with a reason.
- [ ] The dummy stepping out of reach before a Strike lands makes it miss, and the lake water is still spent.
- [ ] Name on the wrong target, and Relic against old matter, are rejected by the validator with reasons.
- [ ] A Press landing on a body that's winding up cancels that swing.
- [ ] In a one-cell soop, the member behind the front one has no attack options against the creature beyond it.
- [ ] `go tool task check` passes.

## Out of scope

- Alert and death (M2-07). The bodies view and creature behavior (M2-08).

## Notes
