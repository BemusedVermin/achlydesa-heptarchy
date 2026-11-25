# Content guide

How to add and change the game's content: what goes where, how records refer to each other, and how to check your work. The terms are the ones in [glossary.md](glossary.md). Why content is data at all is in [go-technical-design.md](go-technical-design.md).

Content is everything a designer would tune: numbers, names, text, glyphs and formulas. It lives in `content/data` as [CUE](https://cuelang.org), and the engine never names a record, a tag or a rule. If a change needs new Go, it is an engine change, not a content change.

## The layout of `content/data`

Every file is in `package data`, and CUE merges them all into one bundle, so a record may be split across files and any file may add to `text:`. Each collection is a map from ID to record, and a file states its schema once, such as `types: [string]: content.#Type`.

| File | Holds |
| --- | --- |
| `materials.cue` | `materials`: what parts are made of, their weaknesses and immunities |
| `terrains.cue` | `terrains`: what one cell of the world is made of |
| `templates.cue` | `templates`: people, creatures, crawling parts, test bodies and the party |
| `bodies.cue` | `parts` (part templates) and `bodies` (body plans) |
| `forms.cue` | the seven forms, as `actions` tagged `form` |
| `actions.cue` | the other `actions`, such as `action.step` |
| `types.cue` | `types`, and the `#Consumable` helper |
| `items.cue` | `items`: what the inventory carries |
| `factions.cue` | `factions`: the sides, and who is hostile to whom |
| `statuses.cue` | `statuses`: conditions a body is under for a while |
| `needs.cue` | `needs`: what bodies run low on, how fast, and the thresholds that hurt them |
| `behaviors.cue` | `behaviors`: what agents choose between, and `agents`, the numbers every agent shares |
| `knowledge.cue` | `claims` (kinds of fact) and `origins` (where a fact came from) |
| `maps/` | `maps`: `maps.cue` holds the schema and `_#TestMap`, one file per map (`test_hall.cue`, `test_den.cue`, `test_descent.cue`) |
| `start.cue` | `start`: the map, party template and mark a new game begins with |
| `rules.cue` | `rules`: named formulas |
| `choices.cue` | `choices`: what an action can ask the player to pick |
| `vocabulary.cue` | `vocabulary`: which tags, grant kinds and list values are allowed |
| `party.cue`, `combat.cue`, `timeline.cue`, `calendar.cue` | one config record each: the numbers every party shares, alert, sight and how creatures hunt, the timeline, the day's phases |
| `text.cue` | only the schema of `text` |
| `text/` | engine messages, one file per core package (`text/combat.cue`, `text/actions.cue`, ...) |

The schemas themselves (`content.#Type` and the rest) are generated from the Go structs in `internal/content` into the `*_go_gen.cue` files. Never edit those by hand; a field the schema lacks needs an engine change.

A field with a fixed set of values, such as an action's `target`, takes a named Go type, and the schema accepts only its constants. To add a value, add a constant of that type in `internal/content` and run `go tool task schema`.

## IDs

An ID is lower-case words joined by dots, such as `material.clay` or `template.creature.beast`: letters, digits and `_`, each word starting with a letter. The first word says the collection, and the loader rejects a record whose ID has the wrong prefix.

| Collection | Prefix | Example |
| --- | --- | --- |
| `materials` | `material.` | `material.clay` |
| `terrains` | `terrain.` | `terrain.packed_clay` |
| `text` | `text.` | `text.material.clay` |
| `templates` | `template.` | `template.creature.beast` |
| `actions` | `action.` | `action.strike` |
| `maps` | `map.` | `map.test.den` |
| `parts` | `part.` | `part.human_arm` |
| `claims` | `claim.` | `claim.material` |
| `origins` | `origin.` | `origin.sweep` |
| `bodies` | `body.` | `body.human` |
| `types` | `type.` | `type.soak` |
| `items` | `item.` | `item.lake_water` |
| `factions` | `faction.` | `faction.wild` |
| `statuses` | `status.` | `status.standing` |
| `needs` | `need.` | `need.water` |
| `behaviors` | `behavior.` | `behavior.test.wander` |

The day's phases in `calendar.cue` have IDs too (`phase.dawn`), and are named `text.<phase id>` like any record.

Records refer to each other by ID, never by nesting one inside another. References in records are checked when content loads, and references inside effect params (such as an `add_status` effect's `status`) when a game starts, so a typo is an error that names the field, not a crash in a fight. Renaming an ID needs no save migration: saves are disposable until v1.0.0.

`test.` after the prefix (`template.test.dummy`, `part.test.crawler_maw`, `map.test.den`) marks a fixture that tests rely on. Change those only together with the tests that use them.

## Text IDs

Every word the player sees is a text ID that names an entry in `text:`. Records hold the ID, never the string.

- **A record's name** is `text.<record id>` and sits in a `text:` block in the same file as the record: `material.clay` is named `text.material.clay`. The loader rejects any other name. Parts are the exception: several part templates may share one name, so `part.human_legs`, `part.beast_legs` and `part.moving_clay_legs` are all `text.part.legs`.
- **An engine message** is `text.<package>.<event>` and goes in that package's file under `text/`: `text.combat.body_died` is in `text/combat.cue`.
- **A claim** also needs `text.knowledge.<claim name>` in `text/knowledge.cue`, the line that reports it in the log (`claim.material` needs `text.knowledge.material`).
- **Placeholders** such as `{{.name}}` are filled from the message's arguments: `"text.combat.body_died": "{{.name}} dies."`.

A text ID with no entry renders as `[missing text: <id>]`. Look for that whenever you play after a change.

## Records carry data; formulas read it

Most records embed `Data`: free `tags`, named numbers in `props` and named `lists`. Code never reads a tag, prop or list; formulas do. `material.clay` lists `weak_to: ["type.soak", "type.relic"]`, and `rule.weak` reads it as `target.material.lists.weak_to`.

Formulas are [Expr](https://expr-lang.org) strings, such as an action's `windup`, a trigger's `if`, or an effect's `amount`. What each formula can see is a struct in `internal/content/env.go`: `source` and `target` (a `BodyEnv` each), `chosen.<x>.id` (the records the player picked), `inventory`, `roll`, and a trigger event's params such as `cells`. An effect param that holds text is a literal, or a formula when it starts with `=`, as in `value: "= target.material.id"`.

### `rules.cue` and `rule.<name>`

A condition used in more than one place is a named formula in `rules.cue`. Any formula uses it as `rule.<name>`, and the compiler inlines it, so rules may use each other:

```cue
rules: {
	weak:   "chosen.type.id in target.material.lists.weak_to"
	immune: "(chosen.type.id in target.material.lists.immune_to) || rule.unheld"
	hit:    "rule.immune ? 0 : (rule.weak ? 6 : 3)"
}
```

Strike's damage is then just `amount: "rule.hit"` in `forms.cue`. A formula that names an unknown variable or rule is not caught by `cue vet`: an action's `windup`, `recovery`, `requires` and `allowed`, a mod, an offer and a trigger's `if` fail when content loads, and an effect's params fail when a game starts.

An action's `requires` leaves it out of the menu when false, as `action.eat` does with no ration in the inventory. Its `allowed` keeps it in the menu but disabled, with its `refusal` text as the reason (`text.actions.requires` when it names none): `action.eat` and `action.drink` are `allowed: "!source.alert"`, so they are refused while the body's party has a hostile in sight. `source.alert` is false for a body outside a party.

An action aimed at a body, an ally or a part refuses a dead body unless it carries the tag named by `combat.corpse_tag` (`corpse`), as the test action `action.gnaw` does so that `consume` can eat a corpse. Corpses are never offered in the menu either way.

### Choices and mods

A form can ask the player to pick a record as well as a target. `choices.cue` defines each choice: the collection it draws from (`from`) and an `offer` formula that says which records to offer. The `type` choice offers the types a body knows innately (its template's `lists.types`), Relic to a body with a graft, and any type whose item the inventory holds. An action lists the choices it asks for, as Strike does with `choices: ["type"]`.

The record picked then changes the action through its **mods**, the fields a type or an item shares with an action: `requires` (with a `refusal` text), extra `windup` and `recovery`, and more effects `on_start` and `on_land`. `type.soak` adds `recovery: "2"`; `type.press` adds an `interrupt` on landing. The action's formulas read the pick as `chosen.type.id`.

### `vocabulary.cue`

The vocabulary closes what would otherwise be open words. Anything it does not declare stays open.

- `tags: items: [...]`: the only tags an item may carry. Add a tag here before you use it on an item.
- `grants: [...]`: the grant kinds a part may give (`action`, `targeting`, `speed`).
- `lists: {"materials.weak_to": "types", ...}`: every value of that list must be an ID in that collection, so `weak_to: ["type.sok"]` fails on load.

## Triggers

A trigger says "when this event happens to me, if this holds, do these effects". It sits on a template, a material, a status or an item, and the entities that carry that thing run it: a body spawned from the template, a part made of the material, a body under the status, a body whose inventory holds the item (or a graft made from the item).

```cue
triggers: [{
	on:   "combat.part_severed"
	role: "target"
	if:   "roll < 0.3"
	do: [{do: "spawn_from_part", params: {faction: "faction.crawling", controller: "hunt"}}]
}]
```

- `on` is an event kind, `<package>.<event>`, such as `combat.part_damaged`, `explore.fell` or `timeline.ready`. The kinds are the event `...Kind` constants in `internal/core/api`; an unknown one fails when a game starts.
- `role` says which subject of the event the carrier must be: `source` (the acting body), `target` (the body acted on), `part` (the part acted on) or `any`. The remnant's trigger above runs when the remnant is the target of a severing; clay's trigger in `materials.cue` uses `part`, because it is the part that is clay.
- `if` is a formula; empty means always. It sees the event's subjects as `source` and `target`, its params as `chosen.<name>.id` (`chosen.type.id` on `combat.part_damaged`, as clay's trigger reads), and a fall's drop as `cells`.
- `do` is a list of effects, the same as an action's `on_land`.

Several triggers that match one event fire in UID order of their carriers, then in the order content lists them.

## Needs

A need's level runs from 0 (met) to 1 (desperate). A template lists the needs its bodies carry, with starting levels: `needs: {"need.water": 0.2}`. At every day-phase boundary each living body's needs rise, in ID order, by their `rise` formula, and the level is kept within 0 to 1. Then each of the need's `thresholds` (in ascending `at`) whose `at` the level has reached applies its `do` effects, with the body as source and target. `rise` and `urge` see `level`, `phase` (the phase starting, such as `"phase.day"`), `source` and `region` (the record of the region the body stands in, such as `region.props.decay`); any formula reads a body's levels as `source.needs.water`. The `satisfy` effect lowers the level by `amount` (a negative amount raises it) in the target, or in the `bodies` it names (`bodies: "self"` feeds the source, such as a trigger on a hit that eases the attacker's hunger), and does nothing to a body without that need.

A hirable template's `recruit.upkeep` is taken from the party's inventory each day. The `recruits` block in `party.cue` holds the numbers that keep recruits: an unpaid day adds one to a recruit's discontent and a paid day takes one off; at `grumble_at` the recruit grumbles each day, and at `desert_at` it leaves the party, keeping its body, grafts and inventory, as an individual of `deserters_faction`. A recruit whose `terror_need` is at or above `flee_at` runs under the agents controller, doing `flee_behavior`, until the level falls below. Terror is raised by triggers that `satisfy` it with a negative amount (see `_severedTerror` in `templates.cue`).

## Behaviors

A body whose template has `actor: controller: "agents"` acts on the behaviors its template lists: `behaviors: ["behavior.test.wander", "behavior.test.to_water"]`. At each of its decisions every listed behavior is scored: the product of its `consider` formulas (1 when it has none). They see `source` (the agent, with its needs), `target` (the target body, when there is one), `distance` (hexes plus depths to the target), `phase` (the current phase ID), `now`, `doing` (the ID of the behavior it is doing, or `""`), `region` (as for needs) and `threatened` (ticks until the soonest landing of a hostile's action aimed at the agent, or -1; the `dodge` primitive takes `within` ticks and steps out of range). A behavior may also have a `rank` formula, which a primitive that picks among options uses to score each one: `leave` crosses the gate whose `rank` is highest, nearest first on a tie. It sees `source` (with its needs), `region` (the record of the region the option leads to) and `distance` (hexes to the option). The behavior it is already doing gets `agents.keep_bonus` added, so it doesn't switch back and forth between behaviors that score about the same. The best wins, and ties are drawn at random. A behavior that scores 0 or less is never chosen; with none left, the body waits.

A template with `recruit: {wants, worth, upkeep}` can be hired when it is a town's generic. `wants` are item stacks and `worth` is a formula over the body (`template`, `needs`, `integrity` and so on): the price is each want's count times `worth`, rounded up. `upkeep` is what the recruit wants each day after hiring.

An exotic recruit is a generic with a strange want and an ability, and no `graftable` tag, so Graft is never offered onto it (`action.graft` requires the tag). The leashed soul-feeder in `templates.cue` wants rations and lees, and its trigger `_feederGrows` adds `status.feeding` when it hurts a body for which `dying` holds; `rule.hit` in `rules.cue` reads the status to hit harder. Keep it hirable by listing its template in a town's `generics`, with a family of its own.

`do` names the behavior primitive that carries it out, one step per decision: `idle` waits, `wander` steps to a random open neighbour, and `move_toward` walks to the target with `combat.step_action`, until it stands on the cell (or next to the body). `params` give the primitive's parameters; one that a primitive doesn't take fails when a game starts.

`target` says what a behavior goes for: the nearest match on the agent's map, or the behavior is not scored. `kind: "body"` takes another body, filtered by its template `tags`, `dead` and `hostile` (to the agent); `kind: "mark"` takes the `mark` of that name on the map; `kind: "cell"` takes the nearest cell of `terrain` within `agents.cell_radius`. While a behavior goes on, it keeps the body or cell it chose as long as that still matches.

## Regions and towns

Every map is a region, named by the map's `name`, and its neighbours are the maps its gates lead to. A map's `region` holds the data formulas read about it (`tags`, `props` such as a decay, and `lists`), as the map's record. A region with `town` is a town: its `stock` is what the town starts with, and goods a `carry` behavior takes to one of its `stock_marks` (plain marks of the map) go into the stock, even when a body stands on the mark.

```cue
region: town: {
	stock: [{item: "item.lake_water", count: 10}]
	stock_marks: ["stock"]
}
```

## Families

A family (`families.cue`, `family.*`) is a kind of creature a region counts while the party is away. When the party enters a map, each count in that region becomes bodies: each is one of the family's `templates`, picked by `weight`, standing beside one of its `marks` (map marks, in turn) or, with no marks on the map, at random around the party. Bodies stand out of the party's sight when a cell is free, within `agents.promote_radius` hexes. A region starts with `region: populations: {"family.test.clay": 3}`. On leaving, living bodies of a family without `Individual` are counted back; the dead stay on the map.

## Effects

An effect is `{do: "<primitive>", params: {...}}`. The primitives are `damage`, `sever`, `lock`, `delay`, `ready`, `interrupt`, `add_status`, `attach_part`, `raise_max_integrity`, `heal`, `set_rest_place`, `reveal`, `consume`, `act_alone`, `spawn_from_part`, `step` and `satisfy`. A new primitive is an engine change. An unknown primitive, a missing param or a bad selector fails when a game starts.

Params that pick bodies or parts are selectors:

- bodies: `self`, `target`, `allies`, `enemies`, `others`, `everyone` or `witnesses` (in sight of the target, as of someone maimed), optionally narrowed to a template tag, as in `others[crawls]`;
- channels and parts: `used`, `target_part`, or `<bodies>:<tag>`, as in `target:all`, `target:legs` or `self:all`.

## Recipes

### A creature, in `templates.cue`

1. Copy a hunter, such as `template.creature.beast`: `_#Hunter & {...}` gives it a step, Strike, Cut, the `hunt` controller and `faction.wild`.
2. Give it an ID under `template.creature.`, a `name` of `text.<its ID>`, an `appearance` (one glyph and a color) and a `body` (a `body.` ID from `bodies.cue`).
3. Add its name to the `text:` block at the bottom of `templates.cue`.
4. To meet it, put it on a map: a mark character in `marks` and the same character in `spawns` (see the map recipe).

Override the hunter's `actions` or `lists.types` to change what it fights with, and add `triggers` for anything it does on an event, as the remnant does. `actor.chooser` names the chooser its attacks are picked with; left out, it is `score`, which rates each attack with `ai_score`. `agents` rates with the rule `agents.score_rule` names, which can read `source.needs` and `target.dying` (a vital part at or below `agents.dying_below` of its integrity), and lets the hunter pick among hostiles as near as its prey.

### A type, in `types.cue` with `#Consumable`

A type that burns an item:

1. Add the item to `items.cue`: `"item.<name>": {name: "text.item.<name>", tags: ["consumable"]}`, and its name to the `text:` block there.
2. Add the type: `"type.<name>": #Consumable & {_item: "item.<name>", name: "text.type.<name>"}`. `#Consumable` lists the item in `lists.consumes` and uses one up when the swing starts. Add mods for anything more: `recovery`, `on_land` effects.
3. Add its name to the `text:` block of `types.cue`.
4. Say what it does to what: add it to a material's `weak_to` or `immune_to` in `materials.cue`. `rules.cue` turns that into damage.
5. To have the party carry some, add a stack to `template.party.main`'s `inventory` in `templates.cue`.

The type choice offers it whenever the inventory holds its item. A type that needs no item (Cut, Press, Name) skips `#Consumable`; it is still defined in `types.cue`, and a body knows it innately when its template lists it in `lists.types`.

### A part, in `bodies.cue`

1. Add the part template under `parts`: a `part.` ID, a `name` (a new `text.part.<name>` or a shared one such as `text.part.legs`), a `slot`, a `material`, an `integrity`, and `grants` from the vocabulary (`action`, `targeting`, `speed`). Mark the parts a body cannot live without `vital: true`: a body with vital parts is out of the fight once none of them is left.
2. Add it to a body plan under `bodies`: `{part: "part.<name>"}`. A plan entry may override the material, as `body.test.crawler` does for its core.
3. Add a new name to the `text:` block of `bodies.cue`.

Give it `crawls_as: "template.crawling.<name>"` to have it fight on as a crawling part once severed, through a `spawn_from_part` trigger on its creature.

### A map, like `maps/test_den.cue`

1. Copy `maps/test_den.cue` to a new file in `maps/`, with a `map.` ID and a `name` of `text.<its ID>`.
2. `_#TestMap & {...}` fills everything outside the layers with packed clay, and gives `#` as clay and `.` as air in the `legend`.
3. Draw one or more `layers`, each a `depth` and `rows`: one character per cell, in odd-r offset layout (odd rows are shifted half a hex right). A space leaves the fill.
4. Name the special characters in `marks`; the mark that `start.party_at_mark` names, `party_start`, is where a new game puts the party (`"@": "party_start"`). Put a mark's character in `spawns` with a template ID to spawn a body of it at every such cell. A mark with no spawn must appear exactly once in the layers; a spawn's character must also be a mark, and may appear any number of times.
5. Add the name to the file's `text:` block.

To play it, point `start.map` in `start.cue` at it for the session, and put it back before you commit unless the change is meant.

## The playground

`go tool task run -- play --seed 1` opens in `map.play.hub` (`maps/play_hub.cue`), a small room with a gate to each of three maps that copy the test maps. Each has a gate back. A feature task adds a line here for what it puts in the playground.

| Gate | Map | What is there |
| --- | --- | --- |
| 1 | `map.play.hall` | The moving clay creature in the hall. |
| 2 | `map.play.den` | The remnant and the flesh beast. |
| 3 | `map.play.descent` | The soops down to the deep, and a moving clay creature at depth 2. |

The hub is a town, with lake water and rations in its stock and a stock mark (`s`) in its top row.

Tests keep their own start maps: the sample content in `contenttest` starts on `map.test.hall`, whatever `start.cue` says.

## The sample world

`starts` in `start.cue` holds named scenarios beside the default `start`. `start.sample` begins in `maps/sample_world.cue`: four maps in a chain, joined by gates, with placeholder numbers for M7 to tune. Play it with `go run ./cmd/harness run --start start.sample --steps 40`; a test picks it with `bundle.Start = bundle.Starts["start.sample"]`.

| Map | Decay | What is there |
| --- | --- | --- |
| `map.sample.shallow` | 0.1 | A town with a well, a small stock, residents and generics. The party starts here. |
| `map.sample.deep_a` | 0.5 | Moving clay and remnants. Gates to the shallow town and `deep_b`. |
| `map.sample.deep_b` | 0.8 | Remnants, moving clay, soul-feeders and an archon warden. Gates to `deep_a` and the lake town. |
| `map.sample.lake` | 0.1 | A town with water to spare, refilled each phase while the party is away. |

The two towns are generic test towns, not the design's crawlspaces. They are not neighbours, so no town migration happens between them; drift is of the monster families.

## Checking your work

1. `go tool task content` checks every file against the schemas (`cue vet`). It catches a wrong field or a value of the wrong kind.
2. `go tool task test -- ./internal/content` loads the real content: it catches a dangling ID, a name that is not `text.<record id>`, a missing text, a value the vocabulary does not allow, and a formula outside effect params that does not compile.
3. `go tool task test -- ./internal/core/game` starts games on the real content: it catches an unknown effect, event kind or selector, and a bad formula or ID in an effect's params.
4. `go tool task run -- play --seed 1` plays it. Add `--dev` to reload the content each time you save a file, so a fight can be tuned without restarting. Look for `[missing text:` in what you see.

Before a pull request, `go tool task check` runs all of these and the rest of the tests.
