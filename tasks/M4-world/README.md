# M4 · One world

**Goal:** the salk, the towns and the deep are one bounded, continuous world at human scale, and everyone in it is an individual agent all the time. The party walks from anywhere to anywhere with nothing loading in between. What happens in the world comes from agents meeting their needs. M3's separate maps, gates, counts and outcome systems are deleted, not wrapped.

**Design sources:** `docs/game-design.md`:

- "World and towns": One world, no seams; authored and emergent crawlspaces; the bounded world; small at first.
- "The living deep": Outcomes emerge; Procgen and exports; The hex grid; People with needs; Three tiers of people; Keeping it affordable.
- "Party" → "Recruits" → "Paying them".

Tech design → "Designing for pivots" (The world, Outcomes emerge), "Level of detail". AGENTS.md rule 10.

## Why this milestone exists

M3 built stand-ins that contradict the design:

- **Separate maps.** Regions and towns were hand-made maps joined by gates.
- **Counts instead of creatures.** Off-screen populations were per-region counts, turned into bodies only when the party arrived.
- **A system per outcome.** Migration, towns emptying, NPC respawns, town stock and recruit desertion each had a system and events of their own.

M4 deletes all of that before procgen and seats build on it. What's left is one world and one agent model.

## Tasks

| ID | Title | Depends on |
| --- | --- | --- |
| [M4-01](M4-01-delete-level-of-detail.md) | Everyone is an individual: delete level of detail and counts | none |
| [M4-02](M4-02-containers-replace-stock.md) | Containers replace the town stock; delete generics and emptying | M4-01 |
| [M4-03](M4-03-delete-upkeep-flight.md) | Delete the upkeep and flight systems | M4-02 |
| [M4-04](M4-04-bounded-world-store.md) | hexmap: one bounded world store | none |
| [M4-06](M4-06-world-and-places.md) | Content: the world and its places | M4-04 |
| [M4-07](M4-07-one-world.md) | One world: delete maps and gates | M4-01, M4-02, M4-03, M4-06 |
| [M4-08](M4-08-regions-as-areas.md) | Regions drawn by hand as areas | M4-07 |
| [M4-11](M4-11-sample-world.md) | The human-scale sample world | M4-02, M4-03, M4-08 |
| [M4-12](M4-12-playground-and-guide.md) | Playground and test places at human scale; the content guide | M4-07, M4-08 |
| [M4-13](M4-13-harness-from-state.md) | The harness measures from state | M4-11 |
| [M4-14](M4-14-exit-check.md) | Exit check: walk the world, and the vocabulary sweep | M4-12, M4-13 |

M4-01 and M4-04 start in parallel. Two lanes run side by side: the deletions (M4-01 → M4-02 → M4-03), which all touch `agents`, and the world foundation (M4-04 → M4-06). They meet at M4-07. After it, regions (M4-08) and the content tasks (M4-11, M4-12) follow. M4-13 and M4-14 close the milestone.

## Exit criteria

| Exit criterion | Covered by |
| --- | --- |
| Walk from the shallow town down through the deep to the lake town and back: no gate, no load, no jump in the view. | M4-07 (no gates), M4-11 (the world), M4-14 (the walk test) |
| Every creature and person in the sample world is an entity from the first tick to its death; nothing is ever a count. | M4-01, M4-11 |
| Off screen, residents fetch water and food from containers and monsters roam and hunt, through M3's behaviors; nothing refills, replaces or announces anything. | M4-02 (containers), M4-03 (deletions), M4-13 (tests) |
| No cell in the sample world stands for a whole structure: every house, well and cistern is built of cells at 1.5 m. | M4-11, M4-12 |
| `git grep -i -w -E "gates?\|MapID\|ByMap\|TownEmptied\|Migrated\|PopulationDied\|Promoted\|Demoted\|Grumbled\|Deserted\|FamilyCount"` over `internal/`, `content/` and `docs/` finds nothing. | M4-01, M4-02, M4-03, M4-07, M4-14 |

## Decisions

Settled with the owner on 2026-11-25:

- **Headroom:** every body fills one cell, standing or crawling; low passages are the norm. Recorded in `docs/game-design.md` → "The hex grid" → "Headroom". Used by M4-11, M4-12.
- **World size:** small at first and built by hand (the towns and a short stretch of deep); procgen grows the rest of the deep later. Recorded in `docs/game-design.md` → "World and towns" → "Small at first". The owner chose a hand-built stand-in deep in M4-11 that M5 replaces. Used by M4-06, M4-11.
- **How far is far:** near is within 2 chunks of any party member or in sight; far agents decide every `agents.far_ticks` (first value a tenth of a phase); the graph is chunks and the openings between them. All content, tuned with the harness. Recorded in `docs/go-technical-design.md` → "Level of detail". Used by M6.
- **Time budget:** a seeded year of the sample world costs at most 60 s in the harness. Recorded in `docs/go-technical-design.md` → "Tests" (Performance) and "Open questions". Used by M6 (M4-13 reports the time without a limit).
- **Regions:** a named area of the world, drawn by hand in content for now; procgen generates regions later. Recorded in `docs/go-technical-design.md` → "Level of detail". Used by M4-08.
- **Recruits:** pay is a need; as it goes unmet, grumbling and then walking off score higher. Personality (robbery), and loyalty keeping people around, come later (M4-03.F1). Recorded in `docs/game-design.md` → "Recruits" → "Paying them". Used by M6; M4-03 only deletes the old systems.
- **Things in places:** water and food lie in containers, entities with a position and an inventory. Recorded in `docs/game-design.md` → "The living deep" → "Outcomes emerge". Used by M4-02.

No questions are left open.

## The one-world model

Every M4 task builds on this. Where the design docs are silent, the choice here is provisional and marked.

**One world.** There is one `hexmap` world with bounds (M4-04); its size is content (`world.cue`, M4-06). Content sets **places** (towns, set pieces, hand-built stretches) into it at a fixed offset and depth. Cells nothing covers are the world's fill, solid clay until M5 generates them. A position is a cell; nothing has a map ID (M4-07).

**Everyone is an individual** (M4-01). Content places actual bodies. There are no families, counts, promotion or demotion.

**Things in places** (M4-02). A container is a body with an inventory and no actor (a cistern, a larder); `carry` fills it and `consume` takes from it. Nothing refills a container and nothing replaces a dead person.

**Recruits** (M4-03). The upkeep and flight systems are deleted. A hired recruit follows and fights; pay, asking, quitting and fleeing come with M6's desires and planner.

**Regions** (M4-08). A region is a named area drawn as boxes, carrying `Data`. The smallest box containing a cell wins. Formulas read the region of the body's own cell. A crawlspace is a region tagged `crawlspace`.

**Agents in M4** keep M3's behaviors, which find their targets within reach, as before. Knowing where things are, travelling for needs, far agents, the coarse chunk graph and the time budget are M6's.

**Interim.** After M4-07 the old maps are places side by side in clay, unjoined; M4-11 joins them into one sample world. Nobody travels between places for their needs until M6.

**Determinism.** Spawns run in place ID order, then layer order. Regions are listed in region ID order. Every task that changes a component bumps the save version.
