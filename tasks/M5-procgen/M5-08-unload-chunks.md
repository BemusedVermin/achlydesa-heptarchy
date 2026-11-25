---
id: M5-08
title: Unload far chunks and keep Seen as bitsets
status: done
milestone: M5
kind: seam
depends_on: [M5-01]
discovered_by: null
size: M
---

# M5-08 · Unload far chunks and keep Seen as bitsets

## Goal

The world holds at most a content-set number of chunks in memory by dropping those far from every party member, and the party's remembered cells cost a bitset per chunk, so a large world does not grow memory with every cell explored.

## Context

- Decision, 2026-11-25: generated chunks are regenerated from the seed and export count on load; only edits are saved; far chunks are unloaded under a memory budget. Recorded in `docs/go-technical-design.md` → "Level of detail" and `docs/game-design.md` → "Procgen and exports".
- What exists: `hexmap.Map.chunks` caches every loaded chunk; edits live in an overlay; `comp.Seen{Cells []api.Cell}` is a sorted slice whose doc says the representation will change.
- `docs/go-technical-design.md` → "Level of detail" (near radius in chunks).

## Deliverables

- `internal/core/hexmap/map.go`: `Map.Unload(keep func(ChunkKey) bool)` or an equivalent that drops chunks, never edits; unloading must also clear any cached sight derived from them.
- `internal/content/types.go`, `content/data/world.cue`: a `chunk_budget` number on the world; regenerate the schemas.
- `internal/core/explore/`: after a party moves, chunks beyond the budget, farthest from any party member first, are unloaded.
- `internal/core/comp/comp.go`, `internal/core/explore/explore.go`, `internal/core/save/`: `Seen` stored per chunk as a bitset keyed by `ChunkKey`; the explore helpers keep their signatures. Bump the save version.
- Tests: unloading and reloading a chunk gives the same terrain; an edit survives an unload; a save round-trips `Seen`.

## Spec

- The budget is in chunks (content). First value tuned with the large world. Firm: a chunk within the near radius of any party member is never unloaded.
- Unload order is by distance, ties by `ChunkKey`, so it is deterministic. Firm.
- A reloaded chunk is generated again from the source; it must equal what it was (M5-01's guarantee), plus the overlay.

## Acceptance criteria

- [x] A test with a budget of a few chunks walks far in a large fixture world and the loaded-chunk count never exceeds the budget, with identical terrain on return.
- [x] `go tool task check` passes.

## Out of scope

- Regenerating chunks on an export: M5-10 (it uses this unloading to drop chunks).

## Notes

- **Near radius is content too.** The spec keeps chunks within "the near radius" but nothing in content held it, so the world gained `near_chunks` beside `chunk_budget` (first values 2, from "Level of detail", and 256, untuned). M6's far agents may want to read the same number.
- **Chunk distance** is `hexmap.ChunkDistance`: the hex distance between two chunk keys' (Q, R), or their depth difference when larger. A chunk's distance to the party is its distance to the nearest member's chunk, over the members of every party.
- **When it runs:** after a party member's step (once its cost is read) and after a wake. A `chunk_budget` of 0 keeps every chunk. Near chunks are kept even when they alone exceed the budget.
- `Map.Unload` clears the whole sight cache whenever it drops a chunk, rather than tracking which results touched which chunks.
- `comp.Seen` is now `Chunks []SeenChunk`, each a key and 64 words of bits in `hexmap.SlotOf` order, sorted by key. `markSeen` and `seenIn` keep their signatures. Save version 3 to 4, with no migration.
- New hexmap helpers: `Map.Loaded`, `CompareKeys`, `ChunkKey.Cell`, `SlotOf`, `ChunkCells`.
- Files outside the list: `internal/core/explore/wake.go`, `internal/core/explore/install.go`, the regenerated `*_go_gen.cue`, `internal/core/save/save_test.go` (migration-chain test needed a 3→4 entry).
- Follow-ups: none.
