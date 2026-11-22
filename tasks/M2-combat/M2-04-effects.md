---
id: M2-04
title: Combat effect primitives
status: done
milestone: M2
kind: feature
depends_on: [M2-01, M2-03]
discovered_by: null
size: M
---

# M2-04 · Combat effect primitives

## Goal

The primitives combat content composes, on top of M1-19's registry and selectors: damage and severing, statuses, grafting, revealing materials, consuming items, and turning a severed part into a body. Each works outside a fight too, and each takes selectors, so one effect can hit one part, every part with a tag, or several bodies.

## Context

- Game design: "Combat → Forms", "Combat → Types", "Combat → Timing".
- This milestone's README, "The combat model".
- M1-19 the effect core and timing primitives (`lock`, `interrupt`, `delay`, `ready`), M2-01 knowledge, M2-02 bodies, M2-03 materials, types and the pack.

## Deliverables

- `internal/content/env.go`: the `EffectEnv` fields below; content `statuses` (`status.`); regenerate the schema.
- `internal/core/comp/comp.go`: `Statuses`.
- `internal/core/combat/effects.go` (registered by `combat.Install`), tests.
- `internal/core/api/combat.go`: the events these publish.
- From M1-19.F1: extend `Registry.EffectSites` for each place content now holds effects, and add a `game` test (fixture under `testdata/`; the sample content stays valid) where `game.New` fails on a deliberately bad action effect. Acceptance: an unknown primitive, a missing parameter and a malformed selector each fail, and the error lists each with its path.

## Spec

**`EffectEnv` gains** (each tagged with a snake_case `expr` name):

- `Weak`, `Immune`: the target part's material against the action's chosen type.
- `Integrity`, `MaxIntegrity`: the target part's.
- `Height`: target depth minus source depth, so it's positive when the source is above.
- `Statuses`: the target body's status IDs.
- `Locked`: whether the target part is busy because of a lock.
- `Graft`: whether the target part is a graft.

They're computed from the context once per `Apply`.

**Primitives** (behavior firm, parameter names suggested). Parameters that select take M1-19's selector grammar.

| Name | Params | Does |
| --- | --- | --- |
| `damage` | `parts`, `amount` | Lowers each selected part's integrity; at zero it's severed (worn down). Publishes `PartDamaged` and maybe `PartSevered{Cause: "worn"}` |
| `sever` | `parts`, `when` | If `when` holds, removes each selected part's `PartOf`; the part stays as an entity at the body's position, and its own pending actions are canceled. Publishes `PartSevered{Cause: "cut"}` |
| `add_status` | `bodies`, `status`, `ticks` | Adds a content status to each selected body until a tick. Publishes `StatusAdded` |
| `attach_part` | `from` (`"pack"` or `"severed"`), `item` | Attaches a carried piece, or a severed part at the target's position, to the target body, which gains its channel. Publishes `PartAttached` |
| `reveal` | `claim` | Writes a fact about each part of the target body for the source's party (Sweep). Uses `knowledge.Learn` |
| `consume` | `item` | Takes one of the item from the source's party pack; its `Check` fails, with a reason, when the pack is short |
| `spawn_from_part` | `faction`, `controller` | Turns a severed part into a one-part body with that allegiance and an `Actor` with that controller (crawling parts) |

- Statuses expire by tick: a `PerTurn` system (it runs after every timeline entry) removes expired ones. Status effects live in formulas that read `Statuses`, such as Stand's double damage taken.
- Severing an arm that's winding up cancels that swing: `sever` cancels pending actions whose channel is the severed part, and M1-18 cancels any whose part is gone.
- Every primitive must behave the same whether or not anything hostile is around: there's no fight state to depend on.

## Acceptance criteria

- [x] One table test per primitive, including a no-op case (a `sever` whose `when` is false).
- [x] `damage` with `parts: "target:arm"` hits every arm of the target; past zero it severs, publishing both events in order.
- [x] Severing an arm mid-wind-up cancels its pending swing; `ActionCanceled` names it.
- [x] `reveal` writes one fact per attached part, and `knowledge.Best` returns each part's material.
- [x] `consume` disables an action whose item the pack lacks, with a reason.
- [x] `go tool task check` passes.

## Out of scope

- The forms that use these (M2-06), triggers (M2-05), and death (M2-07).

## Notes

**Assumptions and deviations**

- `parts` takes M1-19's channel selector (`target:arm`, `target_part`, ...); the body's own channel in a selection is skipped, since it is not a part.
- `sever`'s `when` is optional (absent means always). `attach_part`'s `item` is required for `from: pack` and refused for `from: severed`; `reveal` takes an optional `source` (default `source.sweep`) and, for now, only reads `claim.material`.
- `Locked` is "the part is busy", because `comp.Busy` does not record why. Filed as M2-04.F2.
- A part from `attach_part` (either source) gets a new marker, `comp.Graft`; that is what `graft` reads. Reattaching one's own severed part also marks it, since the part cannot say where it came from.
- `attach_part from: pack` needs to know what a piece becomes, so `content.Item` gains an optional `part` (a part template ID; only a `piece` may have one, checked at load).
- A severed part keeps its integrity and gets the body's `Pos`; `attach_part from: severed` takes the lowest-UID severed part at the target's position and removes that `Pos`.
- `spawn_from_part` makes a new body entity (`comp.Body`, `Actor`, `Allegiance`) at the part's position and attaches the part to it. The target part is used when it is severed, otherwise the severed part at the target's position. No `Appearance` yet; M2-09 owns what a crawler looks like.
- The packs `consume`, `attach_part` and `reveal` use belong to the source's party when the source is a member, otherwise to the source. A bound item (`consume`) is bound to the target's UID.
- Apply for `attach_part`, `consume` and `spawn_from_part` is a quiet no-op when there is nothing to use (a Check says why first); a missing status, item or faction ID is an error. `env` is computed once per effect reference and kept for its formulas; `roll` is still drawn fresh each time.
- `EffectSites` needed no extension: actions are still the only place content holds effects (statuses hold formulas read through `statuses`, not effects). The `game` test uses `internal/core/game/testdata/bad_actions.json`.

**Files outside the Deliverables**

`internal/content/{types.go,types_check.go,load.go}` (statuses, `Item.Part`, checks), `content/data/statuses.cue` (schema constraint and three text lines), `internal/core/comp/comp.go` (`Graft`, `Allegiance` beside `Statuses`; `Allegiance` is the one M2-07 expects), `internal/core/effects/{effects.go,timing.go}` (env, exported `WindingUp`), `internal/core/game/install.go`, generated `*_go_gen.cue`.

**Follow-ups filed:** [M2-04.F1](../followups/M2-04.F1-validate-content-ids-in-effects.md), [M2-04.F2](../followups/M2-04.F2-lock-reason-on-busy.md).
