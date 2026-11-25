# M4 · One world

**Status:** outline. Break it into task files once M3 is done.

**Goal:** the salk, the towns and the deep are one bounded, continuous world at human scale, and everyone in it is an individual agent all the time. The party walks from anywhere to anywhere with nothing loading in between. What happens in the world comes from agents meeting their needs. M3's separate maps, gates, counts and outcome systems are deleted, not wrapped.

**Design sources:** `docs/game-design.md`:

- "World and towns": One world, no seams; authored and emergent crawlspaces; the bounded world.
- "The living deep": Outcomes emerge; Procgen and exports; The hex grid; People with needs; Three tiers of people; Keeping it affordable.

Tech design → "Designing for pivots" (The world, Outcomes emerge), "Level of detail". AGENTS.md rule 10.

## Why this milestone exists

M3 built stand-ins that contradict the design:

- **Separate maps.** Regions and towns were hand-made maps joined by gates.
- **Counts instead of creatures.** Off-screen populations were per-region counts, turned into bodies only when the party arrived.
- **A system per outcome.** Migration, towns emptying, NPC respawns, town stock and recruit desertion each had a system and events of their own.

M4 deletes all of that before procgen and seats build on it. What's left is one world and one agent model.

## Planned tasks

- **One world.** There is one `hexmap` world with edges, and its size is content. Content places each town, set piece and hand-built stretch of the deep at a fixed position in it (a hex offset and a depth). Cells nothing covers are solid clay until M5 generates them. A position is a cell, and the start puts the party at a cell of the world.
- **Delete separate maps.** `api.MapID`, `Pos.Map`, `hexmap.Maps`, the content registry of maps, `Seen.ByMap`, per-map edits in saves, and every "which map" check.
- **Delete gates.** `hexmap.Gate`, `GateDef`, `GateSite`, `explore/gates.go`, `crossGate`, `rankedGate`, gate marks, the playground hub and the gate-chained starts.
- **Everyone is an individual.** Content places actual bodies at the start, never family counts.
  - Delete `FamilyCount` populations and the off-screen drift of counts (`offscreen.go`).
  - Delete promotion and demotion (`lod.go`), and `api.Promoted` and `api.Demoted`.
  - Agents far from the party decide less often and travel a coarse graph of the world (chunks and the openings between them); near the party they act through the timeline cell by cell. One agent model, at two fidelities.
- **Delete the outcome systems.** Each outcome below becomes a choice agents make from their needs, through behaviors and content:
  - `api.Migrated`, `api.TownEmptied` and `api.PopulationDied`;
  - the generic NPC replacement timers (`generics.go`, `comp.Town`'s timers);
  - the abstract town stock and its refill (`stock.go`): water and food become items in places, carried there by people;
  - upkeep's own grumble, desert and flee machinery (`upkeep.go`, `api.Grumbled`, `api.Deserted`, `api.Fled`). A recruit leaves when the deal stops meeting their needs, through the same behaviors as anyone else.

  Events say what an agent did, such as a member leaving the party, never why a system decided it.
- **Regions describe ground.** A region is an area of the world drawn in content, carrying tags, props such as decay, and lists. It holds no population, stock, residents or generics. A crawlspace is a region where people happen to live. Formulas and the region panel read the region a cell lies in.
- **Leaving is a behavior.** `leave` and migration become ordinary behaviors: an agent whose needs aren't met where it is picks somewhere they could be, ranked by content formulas, and walks there, or travels the coarse graph when far from the party.
- **Human scale.** One hex and one depth step are 1.5 m. Rebuild the test maps, the playground and the sample world as one continuous world at that scale:
  - buildings with walls, rooms and doors;
  - wells as shafts, and lakes as water;
  - towns hundreds of hexes across;
  - a deep between the towns that takes real time to cross.

  The content guide states the scale and how to build a structure.
- **The harness measures from state.** Replace M3-20's event-counted `migrations` and `towns_emptied` columns with counts taken from the world: deaths, people per crawlspace, crawlspaces nobody lives in. The camping check counts positions.
- **Saves and the view follow.** A save holds individuals and cells in one world (bump the save version). The plan view never names a map.
- **Docs.** The glossary drops gate, aggregate and promote/demote, and redefines region. The content guide describes placing towns, regions and bodies in the world.

## Exit criteria

- Walk from the shallow town down through the deep to the lake town and back: no gate, no load, no jump in the view.
- Every creature and person in the sample world is an entity from the first tick to its death; nothing is ever a count.
- Off screen, residents fetch water and food, monsters roam and hunt, people leave a crawlspace whose water fails, and a recruit walks off when unpaid. All of it comes from behaviors, with no system, timer or event dedicated to the outcome.
- No cell in the sample world stands for a whole structure: every house, well and cistern is built of cells at 1.5 m.
- `git grep -i -w -E "gates?|MapID|ByMap|TownEmptied|Migrated|PopulationDied|Promoted|Demoted|Grumbled|Deserted|FamilyCount"` over `internal/`, `content/` and `docs/` finds nothing.
- A year of the sample world in the harness stays inside the time budget the questions below settle.

## Questions to settle before breaking it down

- **Headroom.** Does a standing body need two cells of height (3 m), with one cell enough for crawling? People crawl inside shelter, so low passages could be the norm.
- **How big the world is,** and the distances between the fixed places, which set how long crossing the deep takes.
- **How far is far:** the radius around the party, how often far agents decide, and how coarse their graph is. The harness's time budget settles it.
- **How content draws a region:** whole chunks, or painted cells.
- **How recruits and the party deal:** what a recruit's needs are while hired, so that leaving falls out of them.
