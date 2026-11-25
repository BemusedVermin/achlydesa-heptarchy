# M4 · One world

**Goal:** the game runs in one world. Every place is set into it at a fixed position, nothing names a map, and M3's stand-ins are gone. Procgen (M5) then generates the deep between the places, so nobody hand-draws the world.

**Design sources:** `docs/game-design.md` → "World and towns" (One world, no seams; the bounded world; small at first), "The hex grid" (human scale, headroom). Tech design → "Designing for pivots" → "The world".

## Why this milestone exists

M3 put AI before the world existed, so it hard-coded behavior around stand-ins: maps as regions joined by gates, counts instead of creatures, and a system per outcome. M4 reverts M3 and makes the world one place. How agents decide comes back, designed properly, in M6.

## Tasks

| ID | Title | Depends on |
| --- | --- | --- |
| [M4-01](M4-01-revert-m3.md) | Revert M3 | none |
| [M4-04](M4-04-bounded-world-store.md) | hexmap: one bounded world store | none |
| [M4-06](M4-06-world-and-places.md) | Content: the world and its places | M4-04 |
| [M4-07](M4-07-one-world.md) | One world: the maps become places | M4-06 |
| [M4-15](M4-15-delete-weak-tests.md) | Delete weak tests | M4-01 |

M4-01 is done in the pull request that replanned M4. M4-15 runs beside the world lane (M4-04 → M4-06 → M4-07). M4-04's open pull request needs `origin/master` merged in once the revert lands.

## Exit criteria

| Exit criterion | Covered by |
| --- | --- |
| The game starts in one world holding every place, and the plan view looks as it did. | M4-06, M4-07 |
| Nothing in code, content or docs names a map as a unit of the world. | M4-07 |
| M3's code and content are gone; the decisions it recorded stay in the game design. | M4-01 |
| The tests left catch real breakage; tests of test helpers and trivial code are gone. | M4-15 |

## Decisions

Settled with the owner on 2026-11-25:

- **Revert M3** rather than delete its stand-ins one by one. What to add back is listed in M4-01.
- **Scale:** one hex and one depth step are 1.5 m, and every body fills one cell (headroom). Recorded in `docs/game-design.md` → "The hex grid".
- **Small at first:** the world is small and its places are today's maps, moved over as they are. Procgen joins them and grows the deep; human-scale towns come when there's a reason to build them.
- **Hobby-sized:** no harness work, few tests, and no task that isn't the world.

## Interim

After M4-07 the places sit side by side in clay with nothing joining them. Procgen (M5) joins them.
