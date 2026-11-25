---
id: M4-14
title: "Exit check: walk the world, and the vocabulary sweep"
status: todo
milestone: M4
kind: test
depends_on: [M4-12, M4-13]
discovered_by: null
size: S
---

# M4-14 · Exit check: walk the world, and the vocabulary sweep

## Goal

A test walks the party from the shallow town down through the deep to the lake town and back with no gate, load or jump, and the words M4 deleted are gone from the code, content and docs.

## Context

- The brief's "Exit criteria".
- `docs/glossary.md`.

## Deliverables

- `internal/core/game/world_walk_test.go` (or in `gametest`): a scripted walk on `start.sample` from the start to the lake town and back, sending only move commands, which checks after every step that every party member moved at most one cell and the plan view's centre moved at most one hex.
- A sweep of `internal/`, `content/` and `docs/` for the brief's grep, rewording any doc sentence that still uses one of the words (such as "nothing loads behind a gate" in `docs/game-design.md`: say it without the word).
- `docs/glossary.md`: no **gate**, **aggregate**, **family** or **promote / demote**; **region**, **container** and **individual** are current (the last one only if it is still used, as "every agent").

## Spec

- The walk uses the core's own options (the moves the view offers), so it fails if a route needs anything else.

## Acceptance criteria

- [ ] The walk test passes.
- [ ] `git grep -i -w -E "gates?|MapID|ByMap|TownEmptied|Migrated|PopulationDied|Promoted|Demoted|Grumbled|Deserted|FamilyCount" -- internal content docs` finds nothing.
- [ ] `go tool task check` passes.

## Out of scope

- Anything the sweep finds that isn't a word swap: file a follow-up.

## Notes

Filled in by the implementer.
