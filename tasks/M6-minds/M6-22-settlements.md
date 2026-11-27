---
id: M6-22
title: "The three towns as settlements"
status: todo
milestone: M6
kind: content
depends_on: [M6-09, M6-13, M6-20]
discovered_by: null
size: M
---

# M6-22 · The three towns as settlements

## Goal

Three towns exist as generated settlements at different depths, each with the landmarks its people need: the Reaper's crawlspace (shallow), the salken village (surface) and the lake crawlspace (middle). A remote water source lies outside the first.

## Context

- `docs/game-design.md` → "World and towns" (the depth table; Towns are open to the deep); "What to do in town".
- M6-09 (the stage), M6-08 (pieces), M6-02 (sources), M6-20 (population).

## Deliverables

- `content/data/settlements/`: `settlement.reaper`, `settlement.salken_village`, `settlement.lake` with position, depth, radius, piece mix, exits and `landmark_min`.
- `content/data/pieces/`: the pieces these need beyond M6-08's first set: `piece.field`, `piece.cistern`, `piece.lake`, `piece.coal_pile`.
- A water source for the Reaper's crawlspace outside its walls: a `piece.lake` placed by the geography as a landmark in the deep nearby, so haulers must leave to fetch it.
- `content/data/start.cue`: the game starts in the Reaper's crawlspace.
- Tests: all three generate; each contains its required landmarks; the Reaper's crawlspace holds no water source inside it; each connects to the deep by `verify`.

## Spec

- **Firm:** a settlement is data. Adding a fourth town later is a record, not code.
- **Firm:** towns are open to the deep; nothing scripted guards them (design).
- **Firm:** a residents-free town is valid until M6-23 fills it.

## Acceptance criteria

- [ ] The tests above pass.
- [ ] `go tool task check` passes.

## Out of scope

- Residents and their jobs: M6-23.
- The thrukk, still town and archon's crawlspaces: later milestones.

## Notes

