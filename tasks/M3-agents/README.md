# M3 · Agents

**Goal:** monsters and people act on their needs: on screen as full agents, off screen as region and town aggregates. Generic NPCs are replaced on timers, and the party can hire recruits.

**Design sources:** `docs/game-design.md` → "The living deep" (Monsters with needs, People with needs, Three tiers of people, Keeping it affordable), "What to do in town", "Party" (Food and water, Recruits, Exotic recruits). `docs/go-technical-design.md` → "Level of detail" and "Designing for pivots" → "Behaviors".

## Tasks

| ID | Title | Depends on |
| --- | --- | --- |
| [M3-01](M3-01-foundation-hostility-choosers.md) | Hostility, choosers and controller lookup in the foundation | none |
| [M3-02](M3-02-needs.md) | Needs: records, levels, a per-phase tick and the satisfy effect | none |
| [M3-03](M3-03-behaviors-agents-controller.md) | Behaviors and the agents controller | M3-02 |
| [M3-04](M3-04-more-behavior-primitives.md) | Behavior primitives: consume, carry, flee, guard and engage | M3-01, M3-03 |
| [M3-05](M3-05-agents-chooser-dodging.md) | The agents chooser for combat, and dodging | M3-04 |
| [M3-06](M3-06-gates.md) | Gates between maps | none |
| [M3-07](M3-07-playground.md) | A playground: the test maps joined by gates | M3-06 |
| [M3-08](M3-08-regions-towns.md) | Regions and towns as entities | M3-06 |
| [M3-09](M3-09-level-of-detail.md) | Level of detail: promotion and demotion | M3-03, M3-08 |
| [M3-10](M3-10-off-screen-tick.md) | The off-screen tick: drift, stocks and migration | M3-09 |
| [M3-11](M3-11-residents.md) | Residents: needs, roles and schedules in town | M3-04, M3-08 |
| [M3-12](M3-12-generic-npcs.md) | Generic NPCs and replacement timers | M3-11 |
| [M3-13](M3-13-party-food-water.md) | The party eats and drinks | M3-02 |
| [M3-14](M3-14-rest-heal.md) | Rest in town: heal integrity and set the rest place | M3-06 |
| [M3-15](M3-15-hiring-recruits.md) | Hiring recruits | M3-12 |
| [M3-16](M3-16-upkeep-desertion.md) | Upkeep, grumbling, desertion and flight | M3-04, M3-15 |
| [M3-17](M3-17-exotic-recruits.md) | Graftable and exotic recruits | M3-16, M3-18 |
| [M3-18](M3-18-monster-families.md) | Monster families as content | M3-05, M3-10 |
| [M3-19](M3-19-sample-world.md) | The sample world: two towns and the deep between them | M3-10, M3-11, M3-14, M3-18 |
| [M3-20](M3-20-harness-emergence.md) | Harness columns and the camping check | M3-16, M3-19 |

M3-01, M3-02 and M3-06 can start in parallel. After them, three lanes run side by side: agents (M3-03 → M3-04 → M3-05), maps (M3-07, M3-08 → M3-09 → M3-10, and M3-14), and the party's needs (M3-13). People (M3-11 → M3-12 → M3-15 → M3-16) opens once M3-04 and M3-08 are done.

M3-07 was added at the owner's request: a playground of the test maps joined by gates, which later tasks add their features to so they can be played by hand.

## Exit criteria

| Exit criterion | Covered by |
| --- | --- |
| Moving clay walks to a corpse the party left behind. | M3-18 (M3-04's `consume`) |
| A resident fetches water on a schedule, and a town short of water starts to lose people. | M3-11 (on screen), M3-10 (off screen), M3-20 (golden) |
| A recruit hired in town fights, then deserts when their wants go unmet. | M3-15, M3-16 |
| Off-screen populations drift and promote into full agents when the party arrives. | M3-09, M3-10, M3-19 |
| 100 harness seeds show no degenerate camping (one family sitting in one town). | M3-20 |

## Decisions

Settled with the owner while planning:

- **Healing:** resting restores a damaged part's integrity over days at a content rate; a severed part never regrows. Recorded in `docs/game-design.md` → "What to do in town" → "Healing". Used by M3-14.
- **The party needs food and water.** Members are slowed first, then lose vital integrity; drinking uses lake water, the same item as Soak. Recorded in `docs/game-design.md` → "Party" → "Food and water". Used by M3-13.
- **Recruit prices** are items: a bundle up front, scaled by what the recruit brings, then daily upkeep from the same wants; unpaid days raise discontent toward desertion. Recorded in `docs/game-design.md` → "Recruits" → "Paying them". Used by M3-15 and M3-16.
- **Regions before procgen:** one hand-made map per region and per town, joined by gates; full agents on the party's map, aggregates everywhere else. Recorded in `docs/go-technical-design.md` → "Level of detail" (which also answers the level-of-detail radius question there) and in M4's brief. Used by M3-06 to M3-10.
  - Superseded on 2026-11-25: the world is one continuous place with no gates. M4 converts regions, towns, level of detail and migration to areas of that world.

Still open, and not needed by M3: what enemies did while a wiped party slept, and whether dead (not downed) members ever come back (M2-07.F1.F1's Notes).

## The agents model

Every M3 task builds on this. Where the design docs are silent, the choice here is provisional and marked.

**Layers.** `agents` is a feature package, so it imports only foundation packages. M3-01 moves the pieces it needs out of `combat`: `Hostile` to `anatomy`, the chooser registry to `actions`, and a controller lookup to `timeline`. A creature that fights hands its decision to the `hunt` controller by name; it doesn't import `combat`.

**Needs** (M3-02). A need is a content record (`need.water`) with an Expr `rise` per day phase, an Expr `urge` from 0 to 1, and thresholds that apply effects each phase while the level is at or above them. Levels live in `comp.Needs`, from 0 (met) to 1 (desperate). The `satisfy` effect lowers a level; a negative amount raises it, which is how events such as terror push a need (provisional). Needs tick per phase, not per entry, to keep the timeline cheap (provisional).

**Behaviors and utility** (M3-03, M3-04). A behavior is a content record: one registered primitive, its parameters, an optional target (the nearest body, mark or cell matching a filter), and considerations: Expr floats whose product is the score. The `agents` controller scores the template's behaviors at each decision, keeps the current one unless another beats it by `agents.keep_bonus` (provisional hysteresis), and runs one step of the winner. Primitives are `idle`, `wander`, `move_toward`, `consume`, `carry`, `flee`, `guard`, `engage`, `dodge` and `leave`. Roles, schedules and customs are only behavior lists and considerations (over `phase` and the body's tags), never Go.

**Fighting** (M3-05). `engage` hands the decision to the `hunt` controller. A template's `actor.chooser` names the chooser its attacks are picked with: `score` (M2's, the default) or `agents`, which scores with a content rule that can read needs and `target.dying`. Dodging is a behavior that steps out of range of a swing winding up at the body.

**Regions, towns and gates** (M3-06, M3-08). A gate is a mark that leads to a mark on another map. Every map is a region entity (`comp.Region`), with content numbers such as `decay`. A town is a region with `comp.Town`, whose stock is the region entity's own `comp.Inventory`, filled and emptied at its stock marks. Neighbouring regions are the ones gates join.

**Level of detail** (M3-09, M3-10). Families (`family.moving_clay`) are content: the templates a member is promoted as, and Expr drift and attrition. Off screen, a region holds a count per family (`comp.Populations`). Entering a map promotes its counts into bodies, and leaving demotes them, except for bodies with `comp.Individual`, which stay as entities but leave the timeline. Each phase, off-screen regions drift: the fraction of each family that moves to each neighbour is computed from the state at the start of the phase and applied in region ID order. Dead bodies stay where they fell.

**Towns and people** (M3-11, M3-12). Residents are agents with people's needs; a role is a template listing behaviors. Haulers carry water from a well to the town's stock; residents drink from it. A town short of a stock loses people: on screen they `leave` through a gate, and off screen they drift to the neighbouring town with the most of it. Generics are replaced on a timer while the town has residents; an empty town replaces nobody.

**Recruits** (M3-15, M3-16). A hirable template has `recruit: {wants, worth, upkeep}`. Hiring pays `wants` × `worth` from the party's inventory and makes the recruit a member under the follow controller, up to `party.slots` members (provisional: no bench). Daily upkeep keeps discontent at bay; grumbling and desertion are content thresholds. A deserter becomes an `Individual` in a deserters' faction and keeps its grafts. Terror is a need raised by triggers; past its threshold the recruit flees under the `agents` controller until it falls.

**Determinism.** Agents draw from the `agents` stream, promotion and replacement from `lod`. Everything iterates in UID, region ID or family ID order. Every new component is saved, and each task that adds one bumps the save version.
