# Glossary

One line per term: **term**: definition. Terms the cleanup milestone C1 contests are added when `/decide` settles them; the C1 questions list them.

## Game

- **action**: any timed doing. It occupies channels, winds up, and lands.
- **crawling part**: a severed part that fights on as its own body.
- **rest**: sleeping in town. It moves time on and sets where the party wakes after a wipe.

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
- **world**: the one continuous, bounded space of cells everything stands in (`sim.World.Map`, `world.cue`). There is no **map** as a unit of the world: `hexmap.Map` is only the store of its cells, and the map the player sees is the plan view of it.
- **place**: a piece of the world set at a fixed position, either hand-made or built from pieces by the settlement stage, such as a town (`place.` IDs, `places/`).
- **drive**: a need or a desire, a content record each agent weighs by its own weight. Personality is the set of weights.
- **need**: a drive that can kill, rising with time (water, food, rest) or with what happens to the body (safety, raised by being hit or by a neighbour dying).
- **desire**: a drive that does not kill (pay, belonging, faith, duty, a wish).
- **goal**: what an agent chose to work toward at a decision, from its drives.
- **plan**: the steps a planner found for a goal, each a timeline action.
- **operator**: a content record a planner uses: preconditions and effects over beliefs, and the action that carries it out.
- **recipe**: a plan written in content for HTN: a goal and its steps in order, each an operator or another recipe.
- **state**: the small set of symbolic facts both planners plan over (`agents.State`), built from what an agent carries and believes.
- **belief**: what an agent remembers of the world, from sight and from being told. Plans use beliefs, never the truth.
- **landmark**: a thing that advertises what it gives (`comp.Affordance`), such as a well, a bed or a corpse.
- **affordance**: what a landmark gives, an `affordance.` record such as water, rest, worship or the dead.
- **source**: a thing that refills a stock at a content rate (`comp.Source`), such as a well or a lake.
- **container**: anything with a place and an inventory, such as a cistern, a larder or a pile on the floor.
- **piece**: a content-defined sub-place (a house, a well) that the settlement stage lays out into a town.
- **move**: movement across the world (`api.Moved`, `move_cost`); never form x type.
- **type**: what a form carries into the part it lands on (`content.Type`). Not every type does damage.
- **still**: the controller for a creature that keeps still (`still_ticks`). Not rest.
- **inventory**: what a body carries, a list of item stacks (`comp.Inventory`, `inventory:`). The word "pack" is not used.
- **timeline**: the shared queue of entries (what was once called the turn queue).
- **per-entry**: runs after each timeline entry (`sim.PerEntry`).
- **systems due**: the timeline's flag that per-entry systems have not yet run (`SystemsDue`).
- **wild**: the hostile creatures as a side (`faction.wild`). The deep is the place.
- **text ID**: a record's name is `text.<record id>` and sits beside the record. An engine message is `text.<package>.<event>` in that package's file under `text/`.
- **chrome**: only key help. Every other word the player sees comes from the core's view.
