# M5 · Procgen

**Status:** outline. Break it into task files once M4 is done.

**Goal:** generate the deep between M4's fixed places. It is an infinite 3D deep built in chunks at human scale, running right up to every town's edge with no seam, with caves that wander in every direction the way Minecraft's do. Exports reseed it, keeping the party's pocket and a route home.

**Design sources:** `docs/game-design.md` → "World and towns" (One world, no seams), "The living deep" (Procgen and exports, The hex grid, An export with the party below). Tech design → "Designing for pivots" (Procgen stages, The world, Plan view), "Level of detail".

## Planned tasks

- **Chunk sources.** A procgen `hexmap.Source` that builds each chunk of the one world on demand from (world seed, export count, chunk key), deterministic and independent of the order chunks are visited. It serves M4's placed pieces where they sit and generates everything else.
- **3D noise caves.** Threshold 3D noise (opensimplex-go's 3D evaluation) into wide chambers, long winding tunnels and thin crawl soops: the cheese, spaghetti and noodle approach Minecraft uses. Sizes are in 1.5 m cells, so a tunnel is a few hexes wide and a chamber tens across. The plan view's projection makes it read as one continuous space.
- **Stage pipeline.** Named stages per geography: noise, carving, joining the fixed places, set pieces, populating, verifying. Joining carves the generated deep into each fixed place's openings so nothing shows where one ends and the other begins.
- **Sinks.** Funnels that spiral down around a shaft: a helicoid ramp, gentle far from the center so it reads as level ground in the plan view, steepening into the shaft at the axis. Generate them with depth-periodic noise in the sink's angular coordinate so each lap resembles the last without a fault line at the seam. Jumping the shaft is the fast way down; risers (opposite-handed sinks) are the natural way back up. A `Topology` with impossible geometry (descending on truly flat ground) stays possible later through the M1-05 seam.
- **Regions** of the generated deep, as areas of the one world, and a **regions graph** with gonum for connectivity and reachability guarantees.
- **Exports.** An irregular schedule from content. Reseeding regenerates chunks with the new export count, except the party's pocket and the fixed places. Seats' regions land somewhere new, and seen cells on regenerated chunks are forgotten (the map goes stale).
- **The deep changing character** over exports (decay spreading downward).
- **Scaling memory.** `Seen` as per-chunk bitsets, unloading chunks under a memory budget, and saving generated chunks or regenerating them (the question left open in M1-12).
- **Properties:** every town connects to the deep, every active seat is reachable, the pocket survives an export.

## Exit criteria

- Walk from a town into a generated deep that winds over and under itself, and back, with no seam anywhere: nothing shows where the town ends and the generated deep begins.
- Spiral down a sink, jump its shaft, and find a way back up.
- An export reshapes the deep while the party is below, and a route to at least one town exists.
- The guarantees hold over 500 seeds in property tests.

## Questions to settle before breaking it down

- The useful depth range: how deep the oldest layer sits, and whether depth goes on past it with age topping out.
- The chunk memory budget (open in the tech design).
- Which set pieces are fixed, and where.
