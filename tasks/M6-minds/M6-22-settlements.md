---
id: M6-22
title: "The three towns as settlements"
status: done
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

- [x] The tests above pass.
- [x] `go tool task check` passes.

## Out of scope

- Residents and their jobs: M6-23.
- The thrukk, still town and archon's crawlspaces: later milestones.

## Notes

- No stage places pieces in the deep, so the Reaper's water is `settlement.reaper_lake`: a one-piece settlement 50 hexes east of the Reaper's crawlspace, joined to the deep by its exit. Filed M6-22.F1 to make the geography place it.
- The game starts in a settlement: `start.place` may name a settlement, and its mark is the first stamped piece carrying it. `piece.street` gains a `party_start` mark. `explore.WorldSource` records settlement marks; `checkStart` accepts a settlement; `contenttest.Fixture` drops settlements (they do not fit its small world).
- `LaySettlement` change: landmark counts include the affordances of a piece's spawned templates (a well or cistern gives water through its template), and the street piece is laid as a fallback when no lacking piece fits. Before, a lacking landmark stopped the street growing at about six pieces. Radii (30) and street counts (20) are first values; F2 filed to tune.
- Sources: `template.source.lake` (water), `template.source.field` (rations), `template.container.coal_pile` (20 coal). Their numbers are placeholders.
- Depths: salken village 0, Reaper's crawlspace 6, lake crawlspace 18. Positions are first values.
- Docs: `docs/content-guide.md` updated for `start` and `landmark_min`.
- Test helper `verifiedWorldOf` split from `newVerifiedWorldWith` in `verify_test.go` to take a registry.
- Follow-ups filed: M6-22.F1, M6-22.F2.

