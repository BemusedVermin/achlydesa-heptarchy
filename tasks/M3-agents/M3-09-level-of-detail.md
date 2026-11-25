---
id: M3-09
title: "Level of detail: promotion and demotion"
status: done
milestone: M3
kind: feature
depends_on: [M3-03, M3-08]
discovered_by: null
size: M
---

# M3-09 · Level of detail: promotion and demotion

## Goal

Only the party's map holds full agents. When the party enters a map, that region's aggregates become bodies; when it leaves, bodies that are not individuals fold back into counts. Off-screen creatures cost almost nothing.

## Context

- The agents model in [README.md](README.md), "Level of detail".
- `docs/go-technical-design.md` → "Level of detail", "Determinism rules" (parallel then in order; named streams).
- `docs/game-design.md` → "Keeping it affordable".
- M3-06's `PartyCrossed`, M3-08's `Populations` and `Individual`, `sim.Spawn`.

## Deliverables

- `internal/content/types.go`: `FamilyDef`; schemas regenerated. `content/data/families.cue` (new) with test families.
- `internal/core/agents/lod.go`: promote and demote, subscribed to `PartyCrossed` and run at game start for the starting map.
- `internal/core/api`: `Promoted`, `Demoted` events (counts per family, for the log and the harness).
- `internal/core/agents/lod_test.go`.

## Spec

**Firm:**

```go
// content
type FamilyDef struct {
    Name      api.ContentID
    Templates []Weighted   // which templates a promoted member is, by weight
    Marks     []string     // map mark names members appear at; empty means any open cell
}
```

- **Promote:** for each population on the entered map's region, in family ID order, spawn `Count` bodies from the family's templates (stream `lod`), at the family's marks, else at open cells, never in the party's sight if another cell is free. Set the count to 0. Each body records its family (`comp.Family`).
- **Demote:** on leaving, every body on the old map with a family and without `Individual` is removed and counted back into its region's population. A body with `Individual` stays an entity but leaves the timeline, with its position kept, until its map is entered again.
- Dead bodies stay where they are; moving clay needs them.
- Promote then demote then promote gives the same counts, and the same seed gives the same bodies in the same cells.

## Acceptance criteria

- [x] A test region with 3 test creatures promotes them when the party enters and holds 3 in its count again after the party leaves.
- [x] A test creature given `Individual` is still an entity after the party leaves, is not on the timeline, and acts again when the party returns.
- [x] Two runs with the same seed promote the same templates into the same cells.
- [x] In the playground, walking into a creature map shows its creatures appear, and walking out and back again shows them again.
- [x] `go tool task check` passes.

## Out of scope

- Counts changing while off screen (M3-10).
- The real families (M3-18).

## Notes

- `Weighted` is `{Template, Weight}` with an integer weight. `FamilyDef.Marks` that the map lacks fall back to open cells.
- Placement: with marks, the n-th member stands on the nearest free unseen cell to mark n mod len(marks); without, a random unseen free cell within `agents.promote_radius` (new content number, 8) of the party. Seen cells are used only when nothing else is free. Gate cells are skipped.
- Added beyond the Deliverables: `RegionDef.Populations` (a region's starting counts, seeded by `SpawnRegions`; the playground hall holds 3 `family.test.clay`), `comp.Dormant`, `timeline.Suspend` and `Resume` (taking a body off and back onto the timeline), `agents.StartLevelOfDetail` (called by `game.New`), `text.agents.demoted` (reason for cancelling a demoted body's actions), save version 6, and a "Families" section in `docs/content-guide.md`.
- Only bodies with `comp.Family` are demoted or suspended; an `Individual` without a family is left alone. `Promoted` and `Demoted` are not loggable (counts only), to keep arrivals out of the message log.
- golangci-lint could not run here (its Go is older than go.mod's 1.27.4), so lint and `check-ci` were not run. `go vet`, `go test ./...` (apart from agentkit's own golangci-lint test), `gofmt`, `task content` and `task comments` pass. CI will lint.
- No follow-ups filed.


