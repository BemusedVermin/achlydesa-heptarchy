# M5 · Procgen

**Goal:** generate the deep between M4's fixed places. It is a 3D deep, built in chunks at human scale within the world's bounds, running right up to every town's edge with no seam, with caves that wander in every direction the way Minecraft's do. Exports reseed it, keeping the party's pocket and a route home.

**Design sources:** `docs/game-design.md` → "World and towns" (One world, no seams), "The living deep" (Procgen and exports, The hex grid, An export with the party below). Tech design → "Designing for pivots" (Procgen stages, The world, Plan view), "Level of detail".

## Tasks

| ID | Title | Depends on |
| --- | --- | --- |
| [M5-01](M5-01-procgen-source.md) | Geography content and the procgen chunk source | M4-07 |
| [M5-02](M5-02-cave-noise.md) | Cave noise stage | M5-01 |
| [M5-03](M5-03-sinks.md) | Sinks stage | M5-01 |
| [M5-04](M5-04-regions.md) | ~~Regions as areas and the region graph~~ (cut, see Decisions) | M5-01 |
| [M5-05](M5-05-join-places.md) | Join the generated deep to the fixed places | M5-02 |
| [M5-06](M5-06-the-large-world.md) | The large world: procgen in the game | M5-05 |
| [M5-07](M5-07-verify-guarantees.md) | Verify stage and the guarantees | M5-05 |
| [M5-08](M5-08-unload-chunks.md) | Unload far chunks and keep Seen as bitsets | M5-01 |
| [M5-09](M5-09-export-schedule.md) | The export schedule and count | M4-07 |
| [M5-10](M5-10-regenerate-on-export.md) | Regenerate the deep on an export | M5-06, M5-07, M5-08, M5-09 |
| [M5-11](M5-11-populate.md) | Populate generated ground | M5-06, M5-08 |
| [M5-12](M5-12-decay-spreads.md) | Decay spreads downward over exports | M5-10 |
| [M5-13](M5-13-play-by-play.md) | A play-by-play of every fight in the message log | none |
| [M5-14](M5-14-wipe-screen.md) | A wipe screen that waits for the player before the party wakes | M5-13 |
| [M5-15](M5-15-replay-trace.md) | Record as you play, and replay with a readable trace | M5-13 |
| [M5-16](M5-16-readable-elevation.md) | Make elevation readable on the terminal map | none |

M5-13 to M5-16 were added by the owner on 2026-11-26 after a playtest: a fight resolved with no visible beats, then the party woke the next day, and elevation was hard to read. M5-14 and M5-15 both touch the TUI app, so they run one after the other. M5-16 touches the map rendering and side panel, so keep it apart from M5-13 if their files overlap.

M5-01 and M5-09 can start at once. After M5-01, M5-02, M5-03 and M5-08 run side by side. The chain M5-02 → M5-05 → M5-06 is the critical path.

## Exit criteria

| Exit criterion | Covered by |
| --- | --- |
| Walk from a town into a generated deep that winds over and under itself, and back, with no seam anywhere. | M5-02, M5-05, M5-06 |
| Spiral down a sink, jump its shaft, and find a way back up. | M5-03 |
| An export reshapes the deep while the party is below, and a route to at least one town exists. | M5-09, M5-10 |
| The guarantees hold on a handful of seeds. | M5-07, M5-10 |

## Decisions

Settled with the owner on 2026-11-25:

- **Size:** tests use small fixture worlds; the real game's world is large (first values about 400 by 400 hexes, 64 depth steps), tuned later. Recorded in `docs/game-design.md` → "World and towns" → "Small at first".
- **Saves and memory:** generated chunks are regenerated from the seed and export count, never saved; chunks far from the party are unloaded under a content-set budget. Recorded in `docs/go-technical-design.md` → "Level of detail" and `docs/game-design.md` → "Procgen and exports".
- **Exports:** M5 adds the schedule and the regeneration; seats hook into the same event later (M7). Recorded in `docs/game-design.md` → "Procgen and exports".
- **Bodies in regenerated deep:** moved to the nearest open ground, never lost. Recorded in `docs/game-design.md` → "An export with the party below".
- **No regions in procgen** (M5-04 cut). Depth already decides what the deep is like, so decay, materials and spawns read depth through content, and the guarantees check chunks. What the piece reports as a seat's region is M7's to define. Recorded in `docs/go-technical-design.md` → "Level of detail".

Provisional (the owner didn't settle them, so they are the most reversible choice):

- **Set pieces are places.** M4's places already hold fixed hand-built pieces, so no stage is needed to place them. Which set pieces exist and where is content, added as the story needs them.
- **Populating** happens once per chunk per export count and is recorded in the save (M5-11). M6 may change it.
- **One geography** for the whole world; the ground changes with depth. A geography per area, for ground that differs side by side at the same depth, is a later seam.

Left open: the useful depth range beyond the first values, and the real chunk budget (both tuned with the large world).
