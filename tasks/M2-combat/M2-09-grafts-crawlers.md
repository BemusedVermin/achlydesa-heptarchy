---
id: M2-09
title: Graft autonomy and crawling parts
status: done
milestone: M2
kind: feature
depends_on: [M2-05, M2-08]
discovered_by: null
size: S
---

# M2-09 · Graft autonomy and crawling parts

## Goal

Grafted parts sometimes act on their own, and parts severed from pre-entropy things sometimes crawl off and fight. Both are content triggers on the M2-05 runner, plus one new primitive.

## Context

- Game design: "Combat → Grafts", "Combat → Bodies and timing" (pre-entropy things don't know how to die), "Combat → Types" (Rot also eats grafts).
- M2-04 `spawn_from_part`, M2-05 triggers, M2-08 the chooser.

## Deliverables

- `internal/core/effects/primitives.go`: `act_alone`.
- `internal/core/comp/comp.go`: `Graft{ Piece api.ContentID }` on grafted parts.
- `content/data/items.cue`: one Serilmen piece item with an autonomy trigger.
- `content/data/templates.cue`: a pre-entropy creature with a crawl trigger.
- Tests in `internal/core/combat`.

## Spec

- **`act_alone`** (params: none). The source part, which must be a graft, picks a startable intent that uses only its own channel, through the registered chooser. It may target anyone, either side. It starts that intent for its host with the usual wind-up, and `ActionStarted` carries `Autonomous: true`. The host's other parts stay free.
- **`attach_part`** from a pack piece adds `comp.Graft` and copies the piece item's triggers onto the new part's `comp.Triggers`.
- **Autonomy content:** the piece's trigger is on `timeline.ready`, role `source`, `if: roll < 0.15`, `do: act_alone`. The number is a placeholder.
- **Crawling content:** the creature's trigger is on `combat.part_severed`, role `target`, `if: roll < 0.3`, `do: spawn_from_part(faction: "faction.crawling", controller: "hunt")`, where `faction.crawling` is hostile to everyone.
- **Rot eats grafts:** a material or type trigger, so Rot damages any `Graft` part it hits on either side.

## Acceptance criteria

- [x] With the roll forced to 0 (a test stream), when its host decides, a grafted member's graft starts a swing on its own, and the host's other arms stay free.
- [x] Severing a pre-entropy creature's arm with a forced roll turns that arm into a body at its cell, with `Pos`, allegiance `faction.crawling`, and an `Actor` with the `hunt` controller, so it acts on its own.
- [x] The party stays on alert while a crawling part outlives its source, until it's dead too.
- [x] Rot used on a grafted ally damages the graft.
- [x] `go tool task check` passes.

## Out of scope

- Grafts shaping out-of-combat choices such as routes (later), and dreams from grafts (M6).

## Notes

- **`act_alone` lives in `internal/core/combat/graft.go`**, not `effects/primitives.go`: it needs `actions` and the chooser, and `actions` imports `effects`. It restricts the choice to attack-tagged intents on the graft's own channel (a swing, not Give or Graft); drop the tag filter to allow any.
- **Forced roll:** the `effects` stream cannot be forced, so tests rewrite the trigger's `if` to `roll < 2` (always true), plus `integrity > 0` on the remnant so a worn-out core does not crawl.
- **Trigger plumbing (outside the list):** `timeline.Ready` is now `Subjected` (source = the deciding body); a trigger with role `source` or `target` also fires for the body's parts that own it in `comp.Triggers`, with `effects.Context.SourcePart` set to that part. `attach_part` from a pack copies the item's trigger IDs onto the part (`content.ItemTriggerIDs`) and records `Graft.Piece`. `actions.StartAutonomous` and `ActionStarted.Autonomous` added. Schemas regenerated.
- A pack still holding the piece also fires its trigger for the holder (item owner), where `act_alone` does nothing but the roll is drawn.
- **Rot** is the `on_land` of `type.rot`: `damage target_part, graft ? 2 : 0`. The 2, 0.15 and 0.3 are placeholders.
- Content added: `item.serilmen_arm`, `part.serilmen_arm`, `template.creature.remnant` (`body.remnant`, old matter), `faction.crawling`. Not in any start scenario or pack.
- Follow-up filed: M2-09.F1 (crawling parts know no actions, so they cannot yet attack).
