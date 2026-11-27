# Content guide

How to add and change the game's content: what goes where, how records refer to each other, and how to check your work. The terms are the ones in [glossary.md](glossary.md). Why content is data at all is in [go-technical-design.md](go-technical-design.md).

Content is everything a designer would tune: numbers, names, text, glyphs and formulas. It lives in `content/data` as [CUE](https://cuelang.org), and the engine never names a record, a tag or a rule. If a change needs new Go, it is an engine change, not a content change.

## The layout of `content/data`

Every file is in `package data`, and CUE merges them all into one bundle, so a record may be split across files and any file may add to `text:`. Each collection is a map from ID to record, and a file states its schema once, such as `types: [string]: content.#Type`.

| File | Holds |
| --- | --- |
| `materials.cue` | `materials`: what parts are made of, their weaknesses and immunities |
| `terrains.cue` | `terrains`: what one cell of the world is made of |
| `templates.cue` | `templates`: people, creatures, crawling parts, test bodies, the party, and sources and containers. A container is any template with an `inventory` and a place to stand; a `source` (`item`, `max`, and a `refill` formula over `stock` and `max`) starts full and gets back `refill` of its item at each phase boundary, never above `max` |
| `bodies.cue` | `parts` (part templates) and `bodies` (body plans) |
| `forms.cue` | the seven forms, as `actions` tagged `form` |
| `actions.cue` | the other `actions`, such as `action.step` |
| `types.cue` | `types`, and the `#Consumable` helper |
| `items.cue` | `items`: what the inventory carries |
| `factions.cue` | `factions`: the sides, and who is hostile to whom |
| `statuses.cue` | `statuses`: conditions a body is under for a while |
| `drives.cue` | `drives`: needs and desires (see Drives below), and `agents`: the numbers every agent shares |
| `knowledge.cue` | `claims` (kinds of fact) and `origins` (where a fact came from) |
| `world.cue` | `world`: the one world's size in columns, rows and depths, the `fill` of every cell no place covers, the `edge` beyond the bounds (solid and opaque), the `open` terrain of a marked cell and the `geography` that generates the cells no place covers |
| `geography.cue` | `geographies`: kinds of generated ground. A geography lists its `stages` in order, each a registered stage's name (`stage`) and its numbers (`params`). Worldgen checks that every stage is registered and takes its params. Fixed cells always keep what their place authors. Its `spawns` list what starts on the generated ground: each entry names a `template`, a `weight`, a depth band (`min_depth`, `max_depth`) and a `density`. A cell that is open with solid ground below, inside the band and outside every place, draws once: an entry takes the share weight x density / (sum of the applying weights), and the rest of the draw is no body. A chunk is populated once, the first time it loads for an export count |
| `places/` | `places`: hand-made pieces of the world (towns, set pieces), one file each, with the schema and `_#TestPlace` in `places.cue`. A place has `layers`, `legend`, `marks`, `spawns` and `openings`, and an `at`: the world cell of its column 0, row 0 and depth 0. `at.row` must be even, and no two places may author the same cell. Its mark names are its own: another place may reuse them |
| `start.cue` | `start`: the place and mark a new game begins at, and the party template |
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
| `places` | `place.` | `place.test.hall` |
| `geographies` | `geography.` | `geography.deep` |
| `parts` | `part.` | `part.human_arm` |
| `claims` | `claim.` | `claim.material` |
| `origins` | `origin.` | `origin.sweep` |
| `bodies` | `body.` | `body.human` |
| `types` | `type.` | `type.soak` |
| `items` | `item.` | `item.lake_water` |
| `factions` | `faction.` | `faction.wild` |
| `statuses` | `status.` | `status.standing` |

The day's phases in `calendar.cue` have IDs too (`phase.dawn`), and are named `text.<phase id>` like any record.

Records refer to each other by ID, never by nesting one inside another. References in records are checked when content loads, and references inside effect params (such as an `add_status` effect's `status`) when a game starts, so a typo is an error that names the field, not a crash in a fight. Renaming an ID needs no save migration: saves are disposable until v1.0.0.

`test.` after the prefix (`template.test.dummy`, `part.test.crawler_maw`, `place.test.den`) marks a fixture that tests rely on. Change those only together with the tests that use them.

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

Strike's damage is then just `amount: "rule.hit"` in `forms.cue`. A formula that names an unknown variable or rule is not caught by `cue vet`: an action's `windup`, `recovery` and `requires`, a mod, an offer and a trigger's `if` fail when content loads, and an effect's params fail when a game starts.

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

## Effects

An effect is `{do: "<primitive>", params: {...}}`. The primitives are `damage`, `sever`, `lock`, `delay`, `ready`, `interrupt`, `add_status`, `attach_part`, `raise_max_integrity`, `reveal`, `consume`, `act_alone`, `spawn_from_part`, `step` and `satisfy`. A new primitive is an engine change. An unknown primitive, a missing param or a bad selector fails when a game starts.

Params that pick bodies or parts are selectors:

- bodies: `self`, `target`, `allies`, `enemies`, `others` or `everyone`, optionally narrowed to a template tag, as in `others[crawls]`;
- channels and parts: `used`, `target_part`, or `<bodies>:<tag>`, as in `target:all`, `target:legs` or `self:all`.

## Drives

A drive is a need or a desire. A body carries the drives its template lists in `drives`, each with a typical weight; when the body spawns, its own weight on each is drawn once by `agents.weight`, a formula over `typical` and a fresh `roll` in [0, 1), and kept for good. That set of weights is its personality.

```cue
"drive.water": {
	name: "text.drive.water"
	kind: "need"
	rise: "1"
	urge: "min(level / 12, 1)"
	thresholds: [{at: 12, do: [{do: "damage", params: {parts: "self:torso", amount: "1"}}]}]
}
```

- A level is 0 when the drive is met and never goes below it. Once a phase, every living body's levels rise by `rise`, a formula over `level` and `weight`. A need requires a rise; a desire without one keeps its level until something changes it.
- `urge` says how pressing the drive is at its level, from 0 to 1; anything outside that is a content bug.
- After the rise, each threshold whose `at` the level has reached applies its effects, with the body as both `self` and `target`. Only a need has thresholds: a desire never kills.
- The `satisfy` effect lowers a drive's level by `amount` for the `bodies` it selects (`drive` names the drive); a negative amount raises it.

## Recipes

### A creature, in `templates.cue`

1. Copy a hunter, such as `template.creature.beast`: `_#Hunter & {...}` gives it a step, Strike, Cut, the `hunt` controller and `faction.wild`.
2. Give it an ID under `template.creature.`, a `name` of `text.<its ID>`, an `appearance` (one glyph and a color) and a `body` (a `body.` ID from `bodies.cue`).
3. Add its name to the `text:` block at the bottom of `templates.cue`.
4. To meet it, put it in a place: a mark character in `marks` and the same character in `spawns` (see the place recipe).

Override the hunter's `actions` or `lists.types` to change what it fights with, and add `triggers` for anything it does on an event, as the remnant does.

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

### A place, like `places/test_den.cue`

1. Copy `places/test_den.cue` to a new file in `places/`, with a `place.` ID and a `name` of `text.<its ID>`.
2. Set `at` to a free part of the world: the cell of the place's column 0, row 0 and depth 0, on an even row, with every authored cell inside `world.cue`'s bounds and none shared with another place. Grow the world if it does not fit. Tests do not use these positions: `contenttest.Sample` puts the places side by side in a small world with no generated deep (`contenttest.Fixture`), and `contenttest.Full` loads the real world.
3. `_#TestPlace & {...}` gives `#` as clay and `.` as air in the `legend`. Everything the layers leave is the world's fill.
4. Draw one or more `layers`, each a `depth` (relative to `at.depth`) and `rows`: one character per cell, in odd-r offset layout (odd rows are shifted half a hex right). A space leaves the fill.
5. Name the special characters in `marks`; a mark name is unique within its place. Put a mark's character in `spawns` with a template ID to spawn a body of it at every such cell. A mark with no spawn must appear exactly once in the layers; a spawn's character must also be a mark, and may appear any number of times.
   To open the place onto the generated deep, list a mark's character in `openings`: worldgen carves a tunnel out of every such cell (the `join` stage in `geography.cue`). Like a spawn, it may appear any number of times, and it needs a neighbor outside the places. Nothing else makes a tunnel, so a place with no `openings` stays closed, and a wall-less edge is not an opening.
6. Add the name to the file's `text:` block.

To play it, point `start.place` and `start.mark` in `start.cue` at it for the session, and put them back before you commit unless the change is meant. Nothing joins the places yet, so a party started in one stays there.

## Checking your work

1. `go tool task content` checks every file against the schemas (`cue vet`). It catches a wrong field or a value of the wrong kind.
2. `go tool task test -- ./internal/content` loads the real content: it catches a dangling ID, a name that is not `text.<record id>`, a missing text, a value the vocabulary does not allow, and a formula outside effect params that does not compile.
3. `go tool task test -- ./internal/core/game` starts games on the real content: it catches an unknown effect, event kind or selector, and a bad formula or ID in an effect's params.
4. `go tool task run -- play --seed 1` plays it. Add `--dev` to reload the content each time you save a file, so a fight can be tuned without restarting. Look for `[missing text:` in what you see.

Before a pull request, `go tool task check` runs all of these and the rest of the tests.
