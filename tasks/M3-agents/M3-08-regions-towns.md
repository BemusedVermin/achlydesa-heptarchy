---
id: M3-08
title: Regions and towns as entities
status: todo
milestone: M3
kind: seam
depends_on: [M3-06]
discovered_by: null
size: M
---

# M3-08 · Regions and towns as entities

## Goal

Each map is a region, an entity of its own that will hold off-screen populations, and a town is a region with stocks and residents. Regions know their neighbours through gates, so later tasks can move populations and people between them.

## Context

- The agents model in [README.md](README.md), "Regions, towns and gates".
- `docs/go-technical-design.md` → "Level of detail" and "Designing for pivots" → "Entity templates" (a town is a template).
- `docs/game-design.md` → "World and towns", "Procgen and exports" (decay deepens with exports; remnants want undecayed ground).
- M3-06's gates; `internal/core/sim/spawn.go` (templates as component bags).

## Deliverables

- `internal/content/types.go`: `RegionDef` on `MapDef` (or a `regions` record keyed by map; see Spec), `TownDef`; schemas regenerated.
- `internal/core/comp/comp.go`: `Region`, `Town`, `Populations`, `Individual`; registered and saved.
- `internal/core/agents/regions.go`: spawning region entities at start; `Neighbours`; the stock as a `comp.Inventory` on the town entity.
- `internal/core/api`: a `RegionView` section (the region's name, and the town's stocks when the party is in one).
- Tests; `internal/core/save/save.go`: bump `CurrentVersion`.

## Spec

**Firm:**

```go
// comp
type Region struct { Map api.MapID; Def api.ContentID }
type Town struct{}                       // a region that is a town; its stock is the entity's Inventory
type Population struct { Family api.ContentID; Count int }
type Populations struct { Of []Population } // sorted by Family
type Individual struct{}                  // never demoted into a count

// agents
func RegionOf(w *sim.World, id api.MapID) (ecs.Entity, bool)
func Neighbours(w *sim.World, region ecs.Entity) []ecs.Entity // through gates, in map ID order
```

- Every map gets a region entity at the start of a game. Content gives a region its name, its numbers (`lists`/`data` such as `decay`, for formulas) and, for a town, its starting stock and `town: true`.
- A body standing on a town's map can find the town entity, so `carry` (M3-04) can deliver into the stock: an inventory owner at a mark means the town when the mark is listed in the town's `stock_marks`.
- `Individual` is only declared here; M3-09 uses it.

**Suggestion:** put `region:` on `MapDef` rather than a separate record, so one file says everything about a map.

## Acceptance criteria

- [ ] Starting a game on the playground creates one region per map, and `Neighbours` follows the gates.
- [ ] A town's stock starts with its content amounts and changes when a test hauler delivers water to a stock mark.
- [ ] The view's region section names the region the party is in, and shows the town's stocks there.
- [ ] Saving and loading keeps regions, populations and stocks.
- [ ] `go tool task check` passes.

## Out of scope

- Promotion and demotion (M3-09); off-screen drift (M3-10); residents (M3-11).

## Notes


