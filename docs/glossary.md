# Glossary

One line per term: **term**: definition. Terms the cleanup milestone C1 contests are added when `/decide` settles them; the C1 questions list them.

## Game

- **action**: any timed doing. It occupies channels, winds up, and lands.
- **crawling part**: a severed part that fights on as its own body.
- **rest**: sleeping in town. It moves time on, heals integrity (never a severed part) and sets where the party wakes after a wipe.
- **need**: something an agent runs low on (water, food, safety...), a level that rises with time and is met by behaviors (`need.water`, `comp.Needs`).
- **behavior**: a content composition of one behavior primitive and its considerations, which an agent picks by utility (`behavior.fetch_water`).
- **family**: a kind of creature or person counted as one population off screen (`family.moving_clay`).
- **region**: one map and what lives on it, for level of detail. A town is a region with residents and stocks.
- **gate**: a cell that leads to a mark on another map.
- **recruit**: a generic person hired into the party for a price in items, kept by daily upkeep.
- **discontent**: how far a recruit's unpaid upkeep has gone; it turns into grumbling, then desertion.

## Code

- **pending**: an action in flight (`comp.Pending`).
- **moving clay**: clay that has held the dead long enough to move. In content it is `moving_clay`; "crawl" means only the `crawls` tag and crawling parts.
- **graft**: a fitted piece of Serilmen. A severed part of the body's own that is put back is a reattach, not a graft.
- **fall**: dropping off a ledge, and only that (`api.Fell`).
- **out of the fight**: dead or downed (`outOfFight()`).
- **party wipe**: every member is out of the fight. It is recoverable: the party wakes where it last rested (`api.PartyWiped`).
- **lost**: the unrecoverable end, a wipe where the party cannot wake.
- **source**: the acting body, in event fields and in formulas.
- **target**: the body or part acted on, in event fields and in formulas.
- **origin**: where a fact came from (`origin.sweep`); knowledge's old `Source`.
- **BodyEnv**: what a formula sees of one body (`content.BodyEnv`, formerly `Side`). A side is a faction.
- **chosen**: in an action's formulas, the records the player picked (`chosen.<x>.id`).
- **params**: a trigger event's numbers, such as `cells` (`api.Params`).
- **form**: one of the seven combat actions, tagged `form` in `forms.cue`. Not a separate record.
- **move**: movement across the map (`api.Moved`, `move_cost`); never form x type.
- **type**: what a form carries into the part it lands on (`content.Type`). Not every type does damage.
- **still**: the controller for a creature that keeps still (`still_ticks`). Not rest.
- **inventory**: what a body carries, a list of item stacks (`comp.Inventory`, `inventory:`). The word "pack" is not used.
- **timeline**: the shared queue of entries (what was once called the turn queue).
- **per-entry**: runs after each timeline entry (`sim.PerEntry`).
- **systems due**: the timeline's flag that per-entry systems have not yet run (`SystemsDue`).
- **wild**: the hostile creatures as a side (`faction.wild`). The deep is the place.
- **text ID**: a record's name is `text.<record id>` and sits beside the record. An engine message is `text.<package>.<event>` in that package's file under `text/`.
- **aggregate**: a region's off-screen count of one family, with no bodies.
- **promote / demote**: turning a region's aggregates into full agents when the party enters its map, and folding them back when it leaves.
- **individual**: an agent that is never demoted into a count (`comp.Individual`): named, holding a grudge, or a deserter.
- **chrome**: only key help. Every other word the player sees comes from the core's view.
