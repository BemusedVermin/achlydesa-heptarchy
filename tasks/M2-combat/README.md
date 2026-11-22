# M2 · Combat

**Goal:** fights happen on the map, on the same clock as exploring. Every body part acts on its own timing, attacks wind up and can be dodged or staggered, and fights are short and lethal. The seven forms and every combat rule are content built from M1's generic actions and effects, so they can be reworked in CUE without touching code.

**Exit criteria:**

- Walking into sight of a hostile creature puts the party on alert on the same map. Members position themselves and attack through menus using all seven forms, with innate and consumable types.
- Wind-ups show on the timeline strip. Stepping out of reach before a swing lands makes it miss. Cutting off the arm that's winding up cancels the swing, and a staggering hit cancels a body's pending swings.
- Severing legs makes a body's steps slower; losing its head makes it pick targets at random.
- In a one-cell soop, only the front member can reach.
- A party that goes in blind (no knowledge of materials, no matching types) loses most fights against the sample creatures; an informed one wins most. The harness measures both.
- Golden encounters replay identically, and the combat property tests hold.

## Tasks

| ID | Title | Depends on |
| --- | --- | --- |
| [M2-01](M2-01-knowledge.md) | knowledge: the fact store | M1 (M1-11, M1-13) |
| [M2-02](M2-02-bodies-parts.md) | Bodies and parts as entities, as channels | M1 (M1-11, M1-13) |
| [M2-03](M2-03-materials-types-items.md) | Materials, types, items and the party's pack | M2-02 |
| [M2-04](M2-04-effects.md) | Combat effect primitives | M2-01, M2-03 |
| [M2-05](M2-05-triggers.md) | Triggers: content reacting to events | M2-04 |
| [M2-06](M2-06-forms-moves.md) | The seven forms as actions, with types and timing | M2-04 |
| [M2-07](M2-07-hostility-alert-death.md) | Hostility, alert and death | M2-06 |
| [M2-08](M2-08-act-options-controller.md) | The bodies view and the creature controller | M2-07 |
| [M2-09](M2-09-grafts-crawlers.md) | Graft autonomy and crawling parts | M2-05, M2-08 |
| [M2-10](M2-10-action-menu.md) | Action menu and look mode on the map screen | M2-08 |
| [M2-11](M2-11-sample-content-tests.md) | Sample content, lethality targets and golden encounters | M2-09 |
| [M2-12](M2-12-hot-reload.md) | Content hot reload in dev builds | M1 (M1-11, M1-13) |

M2-01, M2-02 and M2-12 can start in parallel.

## The combat model

Every M2 task builds on this. The source is `docs/game-design.md` → "Combat"; where it's silent, the choice below is provisional and marked.

**Fighting on the map.** There's no combat mode, no fight entity and no combat screen. Every body acts on M1's timeline, on the same hex grid it explores. A fight is just hostile bodies acting near each other; anyone joins or leaves by walking in or out of it.

**Bodies and channels.** A body is an entity with `comp.Body`. Each part is its own entity with `comp.Part` and the Ark relation `comp.PartOf`. Party members are bodies on the map related to the party by `comp.MemberOf` (M1-07). The body and each attached part is a timeline channel (M1-17), tagged with the part's slot and grants:

- `grant:action` for arms;
- `grant:targeting` for a head (without one, a body's targets are picked at random);
- `grant:speed` for legs (summed into the speed that sets step time).

**Actions are content (M1-20).** Each says which channels it occupies, by tag and count, plus its wind-up and recovery formulas, its target and range, and its effects on start and on landing. Some examples, all provisional:

- Strike uses one arm, Sweep both, and Graft two arms (one holds the piece).
- A step occupies the body channel, so a body can swing while its legs recover; that's hit-and-run.
- Effects pick their targets with selectors, from one part up to every channel of every body in sight (M1-19).

**Timing.**

- **Wind-up and dodging:** an attack is aimed when it starts and lands after its wind-up. At landing it checks again: if the target is out of range or out of sight, or the acting arm is gone, it misses.
- **Staggers and locks:** `interrupt` cancels pending swings (a stagger), and `lock` freezes channels (Pin; "stop everything" is a lock on all of a body's channels).
- **What changes timing:** wind-up and recovery come from the form, the type (lake water is heavy), the part's material and grafts.

**Lethality** is a design requirement, not a tuning afterthought. The sample numbers below must meet it, and M2-11 measures it.

- **Few hits:** a part comes off in one to three ordinary hits, and in one on its weakness. Vital parts take two or three.
- **Short fights:** a fight is decided in a handful of landings, measured in tens of ticks, not hundreds.
- **Blind parties lose.** Without knowing the materials or carrying the matching types, the party wears parts down while the enemy takes theirs off. Planning is the investigation loop: learn materials, fetch consumables, choose the ground and time the approach.

**Alert.** When a hostile body is visible to any party member, the party is on alert (M1-16's `Party.Alert`), and every member's decisions come to the player. After `combat.calm_ticks` with no hostile in sight, the party relaxes back into lead and follow.

**Sides.** Each body that fights has `comp.Allegiance{Faction}`. Content factions list which factions they're hostile to; two bodies are enemies when either faction lists the other.

**Death (provisional; the design doesn't settle it).** Parts can be marked `vital`. A body dies when it has vital parts and all of them are gone, or when it has none and no parts are left. A body with `comp.Protected` (plot armor) is `Downed` instead. Dead and downed bodies leave the timeline, and their pending actions are canceled. File a `question` follow-up if this rule fights the content.

**Events** live in `internal/core/api/combat.go`: `AlertStarted`, `AlertEnded`, `PartDamaged`, `PartSevered`, `PartAttached`, `StatusAdded`, `BodyDied`, `BodyDowned`, `PartyLost`. Starting, landing, missing and canceling come from `timeline` and `actions`. Every event about an entity carries its UID, never an Ark entity handle.
