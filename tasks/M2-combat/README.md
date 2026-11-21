# M2 · Combat

**Goal:** fights happen on the map, on the same turns as exploring, with every form and every effect primitive, a handful of materials and types, and an action menu. Content composes moves; code only supplies primitives.

**Exit criteria:**

- Walking into sight of a hostile creature puts the party on alert, on the same map and the same timeline. Members step into position and attack through menus: Strike, Sweep, Sever, Pin, Stand, Give and Graft, with innate and consumable types.
- Severing an arm takes an action off the body's next turn; severing legs makes its turns come less often; losing its head makes it pick targets at random.
- In a soop one cell wide, only the front member can reach.
- A pre-entropy creature's severed part can crawl off and take its own turns; a graft sometimes acts by itself.
- Golden encounters replay identically, and the combat property tests hold.

## Tasks

| ID | Title | Depends on |
| --- | --- | --- |
| [M2-01](M2-01-knowledge.md) | knowledge: the fact store | M1 (M1-11, M1-13) |
| [M2-02](M2-02-bodies-parts.md) | Bodies and parts as entities, with turn budgets and speed | M1 (M1-11, M1-13) |
| [M2-03](M2-03-materials-types-items.md) | Materials, types, items and the party's pack | M2-02 |
| [M2-04](M2-04-effects.md) | effects: context, registry and the primitives | M2-01, M2-03 |
| [M2-05](M2-05-triggers.md) | Triggers: content reacting to events | M2-04 |
| [M2-06](M2-06-forms-moves.md) | Forms, reach and move resolution | M2-04 |
| [M2-07](M2-07-hostility-alert-death.md) | Hostility, alert and death | M2-06 |
| [M2-08](M2-08-act-options-controller.md) | The act command, options and the creature controller | M2-07 |
| [M2-09](M2-09-grafts-crawlers.md) | Graft autonomy and crawling parts | M2-05, M2-08 |
| [M2-10](M2-10-action-menu.md) | Action menu and look mode on the map screen | M2-08 |
| [M2-11](M2-11-sample-content-tests.md) | Sample content, golden encounters and properties | M2-09 |
| [M2-12](M2-12-hot-reload.md) | Content hot reload in dev builds | M1 (M1-11, M1-13) |

M2-01, M2-02 and M2-12 can start in parallel.

## The combat model

Every M2 task builds on this. The source is `docs/game-design.md` → "Combat"; where it's silent, the choice below is provisional and marked.

**Fighting on the map.** There's no combat mode, no fight entity and no combat screen. Every body takes turns on M1-15's timeline, on the same hex grid it explores. A fight is just hostile bodies taking turns near each other; anyone joins or leaves by walking in or out of it.

**Bodies.** A body is an entity with `comp.Body`. Each part is its own entity with `comp.Part` and the Ark relation `comp.PartOf` pointing at its body. Party members are bodies on the map, each with its own position, related to the party by `comp.MemberOf` (M1-07).

**Grants.** A part grants any of:

- `action`: one use per turn, as arms do.
- `targeting`: the body picks its own targets, as a head does. Without one, its targets are picked at random.
- `speed`: summed over the body's legs, and fed to `turns.turn_ticks` to set how often its turns come.

Content decides which part grants what, so strange bodies can have strange parts. A part is *functioning* while attached with integrity above zero.

**A turn** is one step plus one use of each functioning, unlocked action part, in any order (M1-15's budget, filled by M2-02). Each use is one move.

**Moves.** A move is a form loaded with a type, used through one acting part. The form is a content composition of effect primitives with a range; the type is innate or consumes an item. A target is in reach when it's visible, within the form's range in hexes, and within that many cells of depth (at least one). So in a one-cell soop only the front body can reach. Formulas see whether the target part is weak or immune to the type, whether it's pinned, and `height`, which is positive when the attacker is above (provisional; content decides whether height matters).

**Alert.** When a hostile body is visible to any party member, the party is on alert (M1-16's `Party.Alert`), and every member's turn comes to the player. After `combat.calm_turns` of the leader's turns with no hostile in sight, the party relaxes back into lead and follow.

**Sides.** Each body that fights has `comp.Allegiance{Faction}`. Content factions list which factions they're hostile to; two bodies are enemies when either faction lists the other.

**Death (provisional; the design doesn't settle it).** Parts can be marked `vital`. A body dies when it has vital parts and all of them are gone, or when it has none and no parts are left. A body with `comp.Protected` (plot armor) is `Downed` instead. Dead bodies stay as corpses and leave the timeline. File a `question` follow-up if this rule fights the content.

**Events** live in `internal/core/api/combat.go`: `AlertStarted`, `AlertEnded`, `MoveUsed`, `PartDamaged`, `PartSevered`, `PartLocked`, `PartAttached`, `StatusAdded`, `BodyDied`, `BodyDowned`, `PartyLost`. Turn events come from `turns`. Every event about an entity carries its UID, never an Ark entity handle.
