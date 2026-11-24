---
id: M3-08
title: Regions and towns as entities
status: done
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

- [x] Starting a game on the playground creates one region per map, and `Neighbours` follows the gates.
- [x] A town's stock starts with its content amounts and changes when a test hauler delivers water to a stock mark.
- [x] The view's region section names the region the party is in, and shows the town's stocks there.
- [x] Saving and loading keeps regions, populations and stocks.
- [x] `go tool task check` passes.

## Out of scope

- Promotion and demotion (M3-09); off-screen drift (M3-10); residents (M3-11).

## Notes

- Content: `region:` sits on `MapDef`, as suggested. A town is `region: town: {stock, stock_marks}`, not `town: true` with stock fields beside it: the presence of `town` makes the region a town, so a stock can't be given to a region that isn't one. The loader checks the stock's items and counts and that each stock mark is a plain mark of the map.
- A region's name is its map's `name`; `RegionDef` has no name of its own while one map is one region. A region's `tags`/`props`/`lists` are the map's record, so formulas and `vocabulary` can read `maps` like any other record collection.
- Region entities have no `Identity` and so no UID: they are found by map ID (`RegionOf`), which kept every body's UID unchanged. Later tasks that need to point at a town can store its map ID. `Game.Hash` changes for every seed because of the new entities; no golden held a hash.
- Added `agents.TownOf` beside the spec's `RegionOf` and `Neighbours`. `agents.SpawnRegions` is called by `game.New` right after `explore.Start` (agents can't import explore); `game.Load` restores regions from the save.
- `carry` delivers into the town whenever its `to` mark is a stock mark, even with a body standing on it (stopping next to the mark then, as before).
- The view's section is `api.View.Region` (`RegionView`: map, rendered name, town flag, stock lines in item ID order), built by `agents`. The terminal frontend doesn't draw it yet: M3-08.F1.
- Playground: the hub is now a town with 10 lake water and 5 rations and a stock mark `s` in its top row (a mark cell is open, so walking is unchanged). `docs/content-guide.md` gains "Regions and towns" and a line under "The playground".
- Files outside the Deliverables: `internal/content/maps.go`, `records.go` (checks and record data), `internal/core/agents/primitives.go` (carry), `internal/core/agents/install.go`, `internal/core/game/game.go`, `content/data/maps/play_hub.cue`, `docs/content-guide.md`.
- Save version 5, no migration (saves are disposable).
- Follow-ups: M3-08.F1 (draw the region section in the TUI).
- The installed golangci-lint was built with Go 1.25 and refused the module; I installed v2.15.0 (the CI pin) built with Go 1.27.4, and `go tool task check` passes with it.


