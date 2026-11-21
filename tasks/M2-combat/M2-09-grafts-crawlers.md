---
id: M2-09
title: Graft autonomy and crawling parts
status: todo
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

- Game design: "Combat → Grafts", "Combat → Bodies and turns" (pre-entropy things don't know how to die), "Combat → Types" (Rot also eats grafts).
- M2-04 `spawn_from_part`, M2-05 triggers, M2-08 the chooser.

## Deliverables

- `internal/core/effects/primitives.go`: `act_alone`.
- `internal/core/comp/comp.go`: `Graft{ Piece api.ContentID }` on grafted parts.
- `content/data/items.cue`: one Serilmen piece item with an autonomy trigger.
- `content/data/templates.cue`: a pre-entropy creature with a crawl trigger.
- Tests in `internal/core/combat`.

## Spec

- **`act_alone`** (params: none). The source part, which must be a graft, picks a valid move in reach for its host body using the registered chooser. It may target anyone, either side. The move is resolved without spending the host's action. Publish `MoveUsed` with `Autonomous: true`.
- **`attach_part`** from a pack piece adds `comp.Graft` and copies the piece item's triggers onto the new part's `comp.Triggers`.
- **Autonomy content:** the piece's trigger is on `turns.turn_started`, role `source`, `if: roll < 0.15`, `do: act_alone`. The number is a placeholder.
- **Crawling content:** the creature's trigger is on `combat.part_severed`, role `target`, `if: roll < 0.3`, `do: spawn_from_part(faction: "faction.crawling")`, where `faction.crawling` is hostile to everyone.
- **Rot eats grafts:** a material or type trigger, so Rot damages any `Graft` part it hits on either side.

## Acceptance criteria

- [ ] With the roll forced to 0 (a test stream), a grafted member's graft acts on its own at turn start, and the host keeps its actions.
- [ ] Severing a pre-entropy creature's arm with a forced roll turns that arm into a body at its cell, with `Pos`, allegiance `faction.crawling` and a `Turn` with the `hunt` controller, so it takes its own turns.
- [ ] The party stays on alert while a crawling part outlives its source, until it's dead too.
- [ ] Rot used on a grafted ally damages the graft.
- [ ] `go tool task check` passes.

## Out of scope

- Grafts shaping out-of-combat choices such as routes (later), and dreams from grafts (M6).

## Notes
