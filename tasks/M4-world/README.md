# M4 · One world

**Status:** outline. Break it into task files once M3 is done.

**Goal:** the salk, the towns and the deep are one continuous world at human scale. The party walks from anywhere to anywhere with nothing loading in between. Separate maps, gates and regions are deleted from code, content and docs, not hidden behind a new layer.

**Design sources:** `docs/game-design.md` → "World and towns" (One world, no seams), "The living deep" (Procgen and exports, The hex grid), "People with needs" (Keeping it affordable). Tech design → "Designing for pivots" (The world, Plan view), "Level of detail".

## Why this milestone exists

M3 stood regions and towns up as separate hand-made maps joined by gates, with one region per map. That contradicted the design: the world is seamless, and nothing divides it into maps or regions. Everything M3 built on that model (gates, the playground hub, region and town entities per map, promotion on entering a map, migration through gates) is converted here, before procgen and seats build on it.

## Planned tasks

- **One world.** There is one `hexmap.Map`, the world. Content places each town, set piece and hand-built stretch of the deep at a fixed position in it (a hex offset and a depth). Cells nothing covers are solid clay until M5 generates them. A position is a cell. The start puts the party at a cell of the world.
- **Delete separate maps.** `api.MapID`, `Pos.Map`, `hexmap.Maps`, `content.MapDef` as a registry of maps, `Seen.ByMap`, per-map edits in saves, and every "which map" check go.
- **Delete gates.** `hexmap.Gate`, `GateDef`, `GateSite`, `explore/gates.go`, `crossGate`, `rankedGate`, gate marks, the playground hub and the gate-chained starts go.
- **Delete regions.** `api.Region`, `RegionDef`, `RegionEnv`, `RegionView`, `RegionOf`, `SpawnRegions`, `agents/regions.go`, the region panel, the `region` variable in formulas, and region records on maps go. What they carried moves:
  - what the ground is like (decay, tags) comes from content formulas over a cell's position and depth;
  - a town is a place with a footprint in the world, carrying its stock, stock marks, residents and generics;
  - formulas that read the region read the cell or the town instead.
- **Detail by distance.** Chunks within the detail radius of the party hold full agents. Beyond it, each chunk holds counts per creature family. A chunk's counts turn into full agents as it comes within the radius, before anything in it can come into sight, and fold back as it leaves. Named agents and anyone with a grudge stay individuals. This replaces promotion on entering a map.
- **Moving without gates.** On screen, the behavior that was `leave` walks toward a destination it ranks: a town, water, less-decayed ground. Off screen, counts and individuals drift from chunk to chunk toward what they need, taking time by distance. Migration, restocking and towns emptying keep working.
- **Human scale.** One hex and one depth step are 1.5 m. The test maps, the playground and the sample world are rebuilt as one continuous world at that scale. Buildings have walls, rooms and doors; wells are shafts; lakes are water; towns run hundreds of hexes across; the deep between the towns takes real time to cross. The content guide states the scale and how to build a structure.
- **Saves, the harness and the view follow.** A save holds cells in one world (bump the save version). Harness columns report by town and chunk. The plan view and the frontend never name a map or a region.
- **Docs.** The glossary loses gate, map and region, and aggregate and promote/demote are redefined per chunk. The content guide describes placing pieces in the world.

## Exit criteria

- Walk from the shallow town down through the deep to the lake town and back: no gate, no load, no jump in the view.
- Walk toward a crowd of monsters: they are full agents before they come into sight. Walk away and they fold back into counts.
- Off screen, residents restock, monsters drift and migrate, and a town can empty, as in M3.
- No cell in the sample world stands for a whole structure: every house, well and cistern is built of cells at 1.5 m.
- `git grep -i -w -E "gates?|regions?|MapID"` over `internal/`, `content/`, `docs/` and `AGENTS.md` finds nothing.
- M3's property and harness tests pass on the continuous world, rewritten where they named maps, gates or regions.

## Questions to settle before breaking it down

- **Headroom.** Does a standing body need two cells of height (3 m), with one cell enough for crawling? People crawl inside shelter, so low passages could be the norm.
- **The detail radius** in chunks, and the chunk size at 1.5 m.
- **Distances between the fixed places,** which set how long crossing the deep takes.
- **What the piece reports without regions.** The game design says it reports a seat's region, meaning the direction of the pull. Settle the reading by M6: a direction and a depth band, the nearest town, or something else.
