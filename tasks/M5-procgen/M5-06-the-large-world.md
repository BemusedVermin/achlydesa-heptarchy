---
id: M5-06
title: The large world: procgen in the game
status: done
milestone: M5
kind: feature
depends_on: [M5-05]
discovered_by: null
size: M
---

# M5-06 · The large world: procgen in the game

## Goal

The real game's world is large and procedurally generated: the places sit in it at fixed positions and the deep around and between them is generated. Tests keep small fixture worlds.

## Context

- Decision, 2026-11-25: "If this is for tests, then small. For the real game, big." Recorded in `docs/game-design.md` → "World and towns" → "Small at first".
- `docs/game-design.md` → "World and towns", "Procgen and exports".
- What exists: `explore.LoadWorld` builds `NewWorldSource`; `content/data/world.cue` is 70 by 30 hexes, 8 deep; `contenttest` fixtures.

## Deliverables

- `content/data/world.cue`: the real world, first values about 400 by 400 hexes and 64 depth steps (600 m by 600 m by 96 m), `geography.deep` with its stages.
- `content/data/places/`: places repositioned inside it (the three test places stay and sit at spread-out positions), with a start in one of them.
- `internal/core/explore/explore.go`: `LoadWorld` builds the procgen source over `WorldSource` and gives the world's `hexmap.Map` the bounds and edge.
- `internal/core/contenttest/`: a small fixture world (a few chunks) for tests that must not generate 400 by 400.
- `docs/game-design.md` and `docs/content-guide.md`: edits listed in Notes.

## Spec

- The world's size is content; nothing in code names 400 or 64. First values are for tuning.
- The fixture world is a copy of the sample bundle with a smaller `world`, per AGENTS.md ("tests that edit content take a copy").
- Generation is lazy through `hexmap.Map`: starting a game generates only the chunks near the party.
- Saves stay seed plus edits (a regenerated world is identical).

## Acceptance criteria

- [x] `go tool task run -- play --seed 1` starts quickly, and walking out of a place's opening goes into a generated tunnel with no seam.
- [x] Two games with the same seed show the same deep; a different seed shows a different one.
- [x] A save made after walking and edits loads to an identical state (existing save tests pass on the fixture).
- [x] `go tool task check` passes.

## Out of scope

- Connectivity guarantees: M5-07. Memory budget and unloading: M5-08. Exports: M5-09, M5-10.

## Notes

- explore may not import procgen (rule 2), so `explore.Generator`, `StartWith` and `LoadWorldWith` take the generator and `game/generate.go` supplies the procgen source. `Start` and `LoadWorld` stay fixed-only.
- `Load` builds the source with the save's export count, so a loaded world matches the saved one once exports regenerate (M5-10).
- `contenttest.Sample` is now the fixture: the old 70 by 30 by 8 world, old place positions, no geography, the hall as start. `contenttest.Full` loads the real content. The content package's own places test sets its world size, since it cannot import contenttest.
- Real content: world 400 by 400 by 64 depth steps; places at (150,180,6), (200,200,10), (170,230,20); start moved to the descent, the only place with an opening. The positions are first values and not checked for joins; connectivity is M5-07.
- Procgen tests now name `geography.deep` instead of the world's geography.
- Docs: `docs/content-guide.md` notes the fixture. `game-design.md` already said it.
- `go tool task check-ci` was run; the agentkit afteredit lint test fails here from a toolchain export-data mismatch, unrelated.
- No follow-ups.
