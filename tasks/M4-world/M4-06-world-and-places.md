---
id: M4-06
title: "Content: the world and its places"
status: in-progress
milestone: M4
kind: seam
depends_on: [M4-04]
discovered_by: null
size: M
---

# M4-06 · Content: the world and its places

## Goal

Content describes one world: its size, the terrain of every cell nothing covers, and the places (towns, set pieces, hand-built stretches) set into it at fixed positions. A loader builds one chunk source from them all.

## Context

- `docs/game-design.md` → "World and towns" (One world, no seams; The world is bounded; Small at first), "Procgen and exports" (Fixed).
- `docs/go-technical-design.md` → "Designing for pivots" → "The world"; "Content pipeline".
- `docs/content-guide.md` → "A map", "IDs".
- What exists: `content.MapDef` (layers, legend, marks, spawns), `internal/content/maps.go` (`checkMap`), `internal/core/explore/source.go` (`NewSource` builds one map's chunk source), and M4-04's `Bounds`.

## Deliverables

- `internal/content/types.go`: `WorldDef` and `PlaceDef` (below), and `Bundle.World` and `Bundle.Places`; regenerate the schemas.
- `internal/content/places.go`: the checks (below), and `Registry.World()`, `Place`, `PlaceIDs`.
- `internal/core/explore/world_source.go`: `NewWorldSource(reg) (*WorldSource, error)`, a `hexmap.Source` that places every place's layers at its offset in the world's fill, and reports marks by `place.<id>:<mark>` (or a `(place, mark)` pair, see Spec) and spawns in place ID order, then row order.
- One test on a small fixture world of two places (below).
- `content/data/world.cue`: a `world` record and the `places` schema line, with no places yet (M4-07 moves the maps in). Not wired into a game.
- `docs/content-guide.md`: the layout table and IDs gain `world.cue`, `places/` and the `place.` prefix.

## Spec

Suggested shapes:

```go
// WorldDef is the size of the one world and what fills the cells no place covers.
type WorldDef struct {
	// Cols and Rows are hexes across, in odd-r offset.
	Cols, Rows         int
	MinDepth, MaxDepth int
	// Fill is every cell no place covers: solid clay until procgen generates the deep.
	Fill api.ContentID
	// Edge is beyond the bounds.
	Edge api.ContentID
	// Open is the terrain of a marked cell.
	Open api.ContentID
}

// PlaceDef is a piece of hand-made world set at a fixed position.
type PlaceDef struct {
	Name api.ContentID
	// At is the world cell of the place's column 0, row 0 and depth 0; layer depths are relative to it.
	At     PlaceAt
	Layers []LayerDef
	Legend map[string]api.ContentID
	Marks  map[string]string
	Spawns map[string]api.ContentID
}
```

- **Offsets keep the odd-r layout honest.** A place's `At.Row` must be even, so its odd rows stay odd in the world. The loader rejects an odd row. Firm.
- **Checks:** every layer cell falls inside the world's bounds; no two places author the same cell; legend, marks and spawns follow today's map rules (`checkMap`'s checks are reused, not copied).
- **Marks are per place.** A mark name is unique within its place; the world source answers `Mark(place, name)`. Suggestion: `Start` becomes `{place, mark, party}` in M4-07.
- **Space characters** leave the world's fill, as they do today.
- This task adds the new shapes beside `maps`; M4-07 deletes `maps`.

## Acceptance criteria

- [ ] A test: a world of two places at different offsets and depths gives the right terrain at a cell of each, the fill between them, and the edge outside the bounds.
- [ ] The loader rejects a place out of bounds or overlapping another, with an error naming it.
- [ ] `go tool task content` and `go tool task check` pass.

## Out of scope

- Starting a game in the world, and deleting maps: M4-07.
- Regions, and joining the places: M5.

## Notes

Filled in by the implementer.
