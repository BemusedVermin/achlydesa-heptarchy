# M4 · One world

**Status:** outline. Break it into task files once M3 is done.

**Goal:** the salk, the towns and the deep are one continuous world at human scale, and the party walks from anywhere to anywhere with nothing loading in between. M3's separate maps and gates are deleted, not wrapped; towns and regions become areas of the one world.

**Design sources:** `docs/game-design.md` → "World and towns" (One world, no seams), "The living deep" (Procgen and exports, The hex grid). Tech design → "Designing for pivots" (The world, Plan view), "Level of detail".

## Why this milestone exists

M3 stood regions and towns up as separate hand-made maps joined by gates, one region per map. That contradicts the design, where the world is seamless. M4 cleans up everything M3 built on that model before procgen and seats build on it: gates, the playground hub, regions and towns tied to maps, promotion on entering a map, and migration through gates.

## Planned tasks

- **One world.** There is one `hexmap` world. Content places each town, set piece and hand-built stretch of the deep at a fixed position in it (a hex offset and a depth). Cells nothing covers are solid clay until M5 generates them. A position is a cell, and the start puts the party at a cell of the world.
- **Delete separate maps.** `api.MapID`, `Pos.Map`, `hexmap.Maps`, the content registry of maps, `Seen.ByMap`, per-map edits in saves, and every "which map" check.
- **Delete gates.** `hexmap.Gate`, `GateDef`, `GateSite`, `explore/gates.go`, `crossGate`, `rankedGate`, gate marks, the playground hub and the gate-chained starts.
- **Regions and towns as areas.** A region is an area of the world drawn in content, no longer a map, carrying what a map's `region` carries now: tags, props such as decay, lists, populations, and for a town its stock, stock marks, residents and generics. Neighbouring regions are those that touch. Formulas, the region panel and the off-screen tick read the region a cell lies in.
- **Detail by nearness.** A region's aggregates become full agents as the party nears it, before anything can come into sight, and fold back as it leaves, keeping named agents and anyone with a grudge. This replaces promotion on entering a map.
- **Moving without gates.** On screen, `leave` walks toward the neighbouring region it ranks highest, by the same `rank` formula. Off screen, migration moves aggregates and individuals between neighbouring regions, taking time by distance. Restocking and towns emptying keep working.
- **Human scale.** One hex and one depth step are 1.5 m. Rebuild the test maps, the playground and the sample world as one continuous world at that scale:
  - buildings with walls, rooms and doors;
  - wells as shafts, and lakes as water;
  - towns hundreds of hexes across;
  - a deep between the towns that takes real time to cross.

  The content guide states the scale and how to build a structure.
- **Saves, the harness and the view follow.** A save holds cells in one world (bump the save version). The harness and the plan view never name a map.
- **Docs.** The glossary drops gate and redefines region as an area of the world. The content guide describes placing towns and regions in the world.

## Exit criteria

- Walk from the shallow town down through the deep to the lake town and back: no gate, no load, no jump in the view.
- Approach a region full of monsters: they are full agents before they come into sight. Walk away and they fold back into aggregates.
- Off screen, residents restock, monsters drift and migrate between neighbouring regions, and a town can empty, as in M3.
- No cell in the sample world stands for a whole structure: every house, well and cistern is built of cells at 1.5 m.
- `git grep -i -w -E "gates?|MapID|ByMap"` over `internal/`, `content/` and `docs/` finds nothing.
- M3's property and harness tests pass on the continuous world, rewritten where they named maps or gates.

## Questions to settle before breaking it down

- **Headroom.** Does a standing body need two cells of height (3 m), with one cell enough for crawling? People crawl inside shelter, so low passages could be the norm.
- **How content draws a region:** whole chunks, or painted cells.
- **How near is near** for promotion, and the chunk size at 1.5 m.
- **Distances between the fixed places,** which set how long crossing the deep takes.
