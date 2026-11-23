# Go technical design

Nov 20, 2026 · @Liam McGuire

## Overview

The game is one Go module: a pure engine core with no UI imports, content held as data, and thin frontends on top. Every subsystem starts from an existing library; custom code is limited to the game rules nothing off the shelf covers.

- **Reuse first.** Before writing a system, pick a maintained library and write only the glue and the game rules.
- **Pure core.** `internal/core` never imports a UI package, and a linter rule enforces it.
- **Content is data.** CUE files hold every material, form, creature, town and candidate; formulas inside them are Expr expressions.
- **Go structs are the source of truth.** CUE schemas are generated from them, so content and code can't drift apart.
- **Deterministic.** Seeded random streams and a single writer to the world make any run replayable. Parallel work is pure computation, applied in a fixed order.

&#91;embedded content: module layout · one core, content in, four consumers\]

Only the frontends know about screens, and content reaches the core already validated.

## Designing for pivots

The game will change once it's played, so every place the game design already expects to grow gets a seam from day one, and nowhere else does. Each seam below costs a few lines now and saves a refactor later.

### Seams cut from day one

| Seam | From day one | Leaves room for |
| --- | --- | --- |
| Maps | Positions are `Pos{Map, Cell}`, where a cell is a hex plus a depth as fine as Minecraft's blocks. Maps are stored in chunks so they can be infinite, and neighbors come from a topology | Soops crossing over and under each other, slopes and ledges, sinks, an infinite plain, impossible folds, towns and set pieces as their own maps |
| Plan view | The core projects the 3D world into one 2D hex view: for each hex, the visible cell nearest the party's depth, shaded only by how far above or below it lies. The view never holds a depth or a layer | A seamless deep in the terminal now, and the same view or a richer one for Ebitengine later |
| Entity templates | Content defines templates as bags of optional components, not one Go type per kind; one spawn function builds any of them | A person who becomes a monster, a creature that trades, a recruit who becomes a candidate, with no new content types |
| Parts as entities | Each part is its own entity, tied to its body by an Ark relation, `PartOf` | Severed parts that crawl off and fight, a hand that latches onto an ally, moving clay taking a part, grafting and reattaching |
| System scheduler | Each system registers with a cadence (per turn, per phase, per day, or on an event) and the systems it runs after; the order is sorted once, deterministically | The flock, dreams and export effects slot in without touching the tick loop |
| Time | Every body acts on one shared timeline of ticks, and the body and each of its parts is a channel with its own busy-until tick; phases per day, ticks per phase and costs are content | Exploration and fights on one clock, multi-limb actions, wind-ups, dodging, hit-and-run, staggering, retuning the calendar |
| Commands and events | Plain serializable structs with a stable kind name, kept with the view types in the api package; each package registers handlers for its own commands | New verbs without growing `core/game`, JSON replay logs now, protobuf later |
| Choices come from the core | The view lists what the player can do right now (moves, targets, routes) in the order to show it; frontends only render the list | Grafts shaping what's offered, recruits refusing orders, an interface that is quietly the Demiurge's |
| View sections | Each package contributes its own section of the view (map, party, combat, journal) | New screens and a second frontend, without one giant snapshot struct |
| Combat sides | Sides are allegiances to factions with content-defined hostility. There's no fight entity or combat mode: a fight is hostile bodies acting near each other, so anyone joins or leaves by walking in or out | Grafts acting on their own, recruits fleeing, monsters fighting each other, crawling parts, fights spilling into towns |
| Effects anywhere | Effect primitives act on entities through a generic context, not on combat state | Consumables used while exploring, seat prices, dreams with effects, traps |
| Actions | Every action, a step included, is content: which channels it occupies (by tag and count), wind-up and recovery formulas, target and range, and effects on start and on landing. The choices an action asks for, and what a chosen type or item changes about it, are content too; code adds only primitives | Two-handed and lunging moves, new forms, effects that hit several limbs or stop a whole body, alternative combat rules tried in CUE |
| Rules | Content records carry generic tags, numbers and lists. Named formulas (`rules`) and content-defined choices read them, and code names no content record, tag or rule | Reworking what weakness means, partial resistances, types that only work on pinned parts, new kinds of item or grant, all in CUE |
| Triggers | A `Triggers` component from content: on an event, if an Expr condition holds, apply effects | Graft autonomy, curdle modes, exotic recruit abilities and material reactions, all as data |
| Behaviors | Agent behaviors are registered primitives that content composes and scores with Expr considerations; monsters, people, recruits and candidates share one path | Simulating candidates later, new creature families as pure data |
| Knowledge | One fact store per knower (subject, claim, source, confidence); the piece, rumors, clues and dreams all write facts | New clue sources, false positives, NPCs who know things, the journal as a view over facts |
| Identity | Every individual gets a stable `UID`; content IDs are namespaced strings (`material.violet`); components hold IDs, never pointers into content | Persistent nemeses and deserters, rumors about the dead, hot reload, saves that survive content edits |
| Saves | A save header carries a save version and a content hash; migration functions run in order | Changing components without breaking old saves |
| Text and looks | Player-facing text lives in content by ID; an `Appearance` component holds the glyph, color and a sprite ID | Rewriting tone, translation, and the Ebitengine frontend reading the same content |
| Procgen stages | Generation is a pipeline of named stages (noise, carve, stitch gates, place set pieces, populate, verify) declared per geography in content | Crawlspaces with different geographies, new export effects |
| The party | The party is an entity with members; commands name the acting party; the slot count is content | A fifth slot, splitting the party, a second group to control |

### What stays concrete

- **The hex grid.** Hexes are decided, so there's no square-grid option, though the hex library still stays behind `hexmap`.
- **Turn-based, with one writer.** No real-time loop. Only one goroutine changes the world; anything parallel is pure computation applied in a fixed order.
- **Primitives in Go, compositions in data.** No embedded scripting language beyond Expr formulas; a new primitive is a deliberate code change.
- **No mod loader and no networking.** The `core/game` boundary already allows an out-of-process frontend if one is ever built.
- **A seam needs a feature.** No abstraction goes in unless the game design names something that needs it.

### The pivot test

A new feature should mean new components, newly registered primitives, behaviors or systems, and new content. If it means editing `core/game`, the scheduler or another package's internals, a seam is missing: file a task to cut it before building the feature.

## Library choices

Every subsystem below reuses a library; the last column is all that's left to write. Versions are as of the date above.

| Subsystem | Library | Why it fits | Left to write |
| --- | --- | --- | --- |
| Terminal UI | [Bubble Tea v2](https://github.com/charmbracelet/bubbletea), Bubbles, Lip Gloss ([v2 out of beta Feb 2026](https://charm.land/blog/v2/)) | Elm-style update loop suits turn-based play; ready-made lists, tables, viewports and layout | The game's screens |
| Later graphics | [Ebitengine](https://ebitengine.org/) | Pure-Go 2D in a real window; can draw sprites or an ASCII bitmap font | A second frontend, later |
| Entities and events | [Ark](https://mlange-42.github.io/ark) ECS | Archetype ECS with zero dependencies, entity relations, an event system, and world serialization via ark-serde | Component structs and queries |
| Content and validation | [CUE](https://cuelang.org/docs/integration/go/) | Schemas, defaults, cross-references and validation in one language; loads straight into Go structs | One load-and-decode function |
| Formulas in content | [Expr](https://github.com/expr-lang/expr) | Type-checked, sandboxed expressions that always terminate, compiled to bytecode | The variables each formula can see |
| Hex math | [gravitton/hexagon](https://pkg.go.dev/github.com/gravitton/hexagon) | Axial and offset coordinates, including double-width, which is exactly the ASCII layout | Chunked 3D storage and sight lines with depth |
| Pathfinding | [fzipp/astar](https://github.com/fzipp/astar) | Generic A\*: supply neighbours and costs | Hex neighbour and cost functions |
| Region graph | gonum graph | Connectivity and shortest paths for procgen guarantees | Building the graph from regions |
| Procgen noise | opensimplex-go | Seeded 3D noise for caves, and for depth and material fields | The generation rules |
| Randomness | `math/rand/v2` (PCG) | Seedable, deterministic, in the standard library | A registry of named streams |
| Saves | ark-serde | Serializes the whole ECS world | RNG states and content IDs |
| Hot reload | fsnotify | Watches the content folder in dev builds | The reload trigger |
| Rumor and dream text | `text/template` | Templates live in content | Nothing |
| CLI | cobra | `play`, `harness` and other subcommands | Nothing |
| Dev commands | [Task](https://taskfile.dev) (go-task), run as a Go tool | Pinned in go.mod and run with `go tool task`, so nothing to install; its built-in shell runs the same commands on Windows and Linux. The CUE CLI is a Go tool the same way | The Taskfile |
| Task front matter | `go.yaml.in/yaml/v3` | Parses the YAML front matter of task files in the agent tools; already in go.sum through Task | Nothing |
| Logs and metrics | `log/slog`, `encoding/csv` | Structured logs; CSV for harness output | Metric definitions |
| Testing | `testing`, rapid, native fuzzing, pprof | Property tests for invariants, fuzzed loaders, profiling | The invariants |
| Architecture checks | golangci-lint with depguard | Fails the build if the core imports a UI package | One rule |

Two notes:

- **Hex library:** gravitton/hexagon covers double-width layout, field of view and line drawing, and needs Go 1.27. [pmcxs/hexgrid](https://github.com/pmcxs/hexgrid/wiki) stays the fallback.
- **CUE** runs as a Go tool (`go tool cue`), pinned in go.mod at the same version as the library; it needs Go 1.25 or later.

## Repository layout

One Go module with two binaries. Engine code lives under `internal/core`, frontends under `internal/frontend`, and content outside the Go tree entirely.

```
achlydesa-heptarchy/
├── go.mod                  # github.com/BemusedVermin/achlydesa-heptarchy
├── cue.mod/                # the CUE module (same path as the Go module)
├── AGENTS.md               # rules for coding agents
├── CLAUDE.md               # one line, @AGENTS.md, so Claude Code loads the same rules
├── cmd/
│   ├── heptarchy/          # the game (cobra: play, new, load)
│   └── harness/            # headless seeded runs, CSV out
├── internal/
│   ├── core/               # pure engine: no UI imports (depguard)
│   │   ├── game/           # public API: command router, view assembly
│   │   ├── api/            # plain-data commands, events and view types
│   │   ├── sim/            # Ark world, event bus, system scheduler, registries
│   │   ├── comp/           # Ark component structs and relations
│   │   ├── anatomy/        # body and part queries
│   │   ├── clock/          # ticks, day phases, calendar from content
│   │   ├── timeline/       # channels, decisions, actions that land later
│   │   ├── rng/            # named math/rand/v2 streams
│   │   ├── hexmap/         # 3D hex cells in chunks, topology, sight, A*
│   │   ├── knowledge/      # fact store per knower
│   │   ├── inventory/      # inventories of item stacks
│   │   ├── effects/        # effect core, selectors, timing primitives, triggers
│   │   ├── actions/        # content-defined actions and the do command
│   │   ├── explore/        # party movement, plan-view projection
│   │   ├── procgen/        # stage pipeline, reseeding, guarantees
│   │   ├── combat/         # hostility, alert, moves with reach, death
│   │   ├── agents/         # behavior registry, needs scoring, level of detail
│   │   ├── seats/          # seat state machine, passing, the piece's readings
│   │   ├── cast/           # wounds, relationships, wish generator
│   │   ├── rumor/          # rumor spread and drift (writes facts)
│   │   ├── investigation/  # clues into facts about parts and weaknesses
│   │   └── save/           # versioned snapshots and migrations
│   ├── content/            # Go content structs, generated *_go_gen.cue schemas, loader, Expr compile, hot reload
│   └── frontend/
│       ├── tui/            # Bubble Tea screens
│       └── gui/            # Ebitengine, later
├── content/
│   └── data/               # templates, materials, forms, maps, text...
├── docs/                   # game design, this technical design, story system
├── tasks/                  # task specs for agents (start at tasks/README.md)
├── Taskfile.yml            # dev commands: go tool task check, run, schema...
└── .golangci.yml           # depguard rules: core is pure, layers hold
```

- **One API boundary.** Frontends and the harness talk only to `core/game`: they send commands and receive events plus a read-only view.
- **No package reaches sideways into another's internals.** Packages communicate through components and Ark events.
- **Two layers.** Foundation packages (`api`, `sim`, `comp`, `anatomy`, `rng`, `clock`, `timeline`, `hexmap`, `knowledge`, `inventory`, `effects`, `actions`) can be imported by anyone. Feature packages (`explore`, `procgen`, `combat`, `agents`, `seats`, `cast`, `rumor`, `investigation`) import only foundation packages, never each other; `game` and `save` sit on top. depguard enforces it.

## Core packages

Each package owns one part of the rules and leans on a library for everything else. The world itself is one Ark ECS world: towns, regions, creatures, people, parts and seats are all entities.

| Package | Owns | Reuses |
| --- | --- | --- |
| `game` | The public API: `Apply(cmd) []Event`, `View() Snapshot`, new game, load, save. Routes each command to the handler its package registered, and assembles the view from each package's section | `sim` and everything below |
| `api` | Plain-data commands, events and view types: the boundary that frontends, the harness and replay logs share | The standard library only |
| `sim` | The Ark world, the event bus, the system scheduler (cadence plus run-after ordering), and the command, event and view registries | Ark |
| `comp` | Plain component structs and relations: `Pos{Map, Cell}`, `Identity` (UID), `Body`, `Part` and `PartOf`, `Needs`, `Memory`, `Relationships`, `Seat`, `Graft`, `Appearance`, `Triggers` | Ark stores and queries them |
| `anatomy` | Queries over bodies and parts: attached parts, functioning parts, grants, party members | `comp` |
| `clock` | Ticks, day phases and the calendar read from content; per-phase and per-day systems as the timeline crosses boundaries; timers for generic NPC replacement and export scheduling | The `sim` scheduler |
| `timeline` | The shared timeline: channels (a body and each of its parts) and when they're free, decisions, actions that land later, controllers (player, follow, behaviors), and playing time forward until the player is needed | Ark resources, `clock` |
| `rng` | One named PCG stream per subsystem, seeded from the world seed (plus the export count for procgen) | `math/rand/v2` |
| `hexmap` | Maps keyed by map ID: 3D hex cells in chunks (maps can be infinite), the topology, stepping rules, 3D sight and pathing. Sight is line based and symmetric, not shadowcasting, so it differs slightly from the prototype's field of view around pillars | gravitton/hexagon, fzipp/astar |
| `knowledge` | Facts per knower: subject, claim, source, confidence. The journal is a view over the party's facts | Components only |
| `inventory` | Inventories of item stacks: add, take, count | Components only |
| `effects` | The effect core: the registry, selectors for parts and bodies, and the timing primitives (`lock`, `interrupt`, `delay`, `ready`); combat adds `damage`, `sever`, `add_status`, `attach_part`, `reveal` and more. Usable by any system, plus the trigger runner (on an event, if a condition holds, apply effects) | Expr for conditions |
| `actions` | Content-defined actions: channel requirements, wind-up and recovery, targets, effects on start and on landing, choices served from content and what a chosen record changes, the generic do command and its options | `timeline`, `effects` |
| `explore` | Maps from content, stepping, lead and follow, remembered cells and the plan-view projection; gates between maps later | `hexmap` |
| `procgen` | A pipeline of named stages per geography: generating the deep around fixed towns and set pieces, reseeding on exports, keeping the party's pocket | opensimplex-go, gonum graph for the guarantees |
| `combat` | Hostility between factions, the party's alert, combat primitives, the creature controller, death, graft autonomy. Forms × types is content | `effects`, Expr for content formulas |
| `agents` | The behavior primitive registry, needs scoring and behavior choice for every agent; full agents on screen, region aggregates off screen | Expr for need curves, Ark relations for grudges and relationships |
| `seats` | The seven seat state machines, stage events, seat passing, the piece's region-plus-progression readings | Ark events, `rng` |
| `cast` | Wound scoring, relationships, plot-armor flags, the wish generator (wound → wish → curdle mode → price → form) | Content tables, Expr |
| `rumor` | Facts carried between people and towns, drifting with each retelling | `knowledge`, `text/template` for the words |
| `investigation` | Turning clues (people, places, bodies, dreams) into facts about parts and weaknesses | `knowledge` |
| `save` | Whole-world snapshots with a version header, and the migration chain | ark-serde, plus RNG states |

## Content pipeline

Content goes from Go structs to CUE schemas to validated data, and almost none of that path is custom code.

1. **Go structs define the shapes.** Components (in `comp`) and content records (`Material`, `ActionDef`, `Template`, `CurdleMode`...) are plain structs. A template is a bag of optional components, so a creature, a person, a town and a candidate are all templates rather than separate types.
2. **`cue get go --local` generates the schemas** as `*_go_gen.cue` files beside the Go structs in `internal/content`, so schema and code never drift.
3. **Authors write data in CUE** under `content/data`. CUE fills defaults and checks references (a form naming an unknown primitive, a creature using an unknown material).
4. **`cue vet` runs in CI** through `go tool task check`, alongside the tests, so content errors fail before the game starts.
5. **The loader decodes into Go structs** with `cue/load` and `Value.Decode`, then compiles every Expr formula against a typed environment, with named rules inlined. A formula that references an unknown variable or rule fails at load, not mid-fight.
6. **The result is an immutable content registry** that the core reads.
7. **Dev builds hot-reload.** fsnotify watches `content/`, rebuilds the registry, and swaps it in between ticks.

What stays as code versus data:

- **Code:** effect primitives and the machinery that runs them (selectors, timing, actions), the needs scorer, seat rules, procgen algorithms. Code names no content record, tag or rule, only its own message text; where it needs a record, a config key says which.
- **Data:** materials, forms as compositions of primitives, the rules that tie them together (what makes a part weak, which types a body may load, what a loaded type does to a form), part and creature templates, needs and their curves, towns and gates, authored candidates and cast, curdle modes, tuning constants (seat speed, kill setback, export rate).

## Simulation loop and determinism

One goroutine changes the world, in a fixed order, so the same seed and the same inputs always produce the same run.

### Tick order

Systems register with the scheduler in `sim`, each declaring a cadence and what it runs after, so this is the default schedule rather than a hard-coded loop. Exploration and fights share it: every body, in the party or not, acts on one timeline.

1. The earliest entry on the timeline happens next, and the clock jumps to its tick: an attack landing, or a body deciding. At the same tick, landings go first, then decisions by UID.
2. A deciding body's controller picks actions: the player for party members (only the leader while no hostile is in sight), a follow controller for the rest of the party, and behaviors for everyone else. Each action occupies the channels it needs (a step the body, a swing one or more arms) for its wind-up and recovery; legs, terrain and climbing set how long a step takes.
3. After every entry, per-turn systems run: statuses expire and the party's alert updates.
4. At each day-phase boundary the clock crosses, the off-screen tick runs: region populations drift, town residents follow their routines, generic NPCs are replaced on their timers.
5. Once a day: seats advance, the export schedule is checked, rumors move between towns.
6. Ark events fire after each step (a death, a severed part, a passed seat), and subscribers update wounds, grudges and readings.

### Level of detail

- **On screen:** creatures and people near the party are full Ark entities with bodies, needs and memory.
- **Off screen:** each region holds aggregate populations per creature family; each town runs its residents on coarse schedules.
- **Promotion and demotion:** entering a region turns aggregates into full agents; leaving folds them back, keeping named agents and anyone with a grudge as individuals.

### Determinism rules

- **Named streams.** Every subsystem draws from its own PCG stream, so adding a feature in combat never changes procgen.
- **No hidden order.** Never let Go map iteration order or goroutines decide an outcome; sort by entity ID where order matters.
- **Parallel, then in order.** Work may run on other goroutines only as pure computation: it reads its inputs and returns results without touching the world. The world's goroutine then applies the results in a fixed order (by chunk key, region ID or seed). Procgen chunks, harness games and off-screen region updates fit this. The `sim` registries are safe for concurrent use; the ECS world, `World.Now` and the RNG streams are not.
- **Procgen seeds** come from the world seed plus the export count, so any export's layout can be regenerated exactly.

### Saves and replays

- **A save** is one JSON document: an ark-serde snapshot of the entities (resources are rebuilt by `Install`, and the turn queue is exported by `timeline`), every RNG stream's state, the UID counter, map edits, the command history and the message log. ark-serde's output is canonicalized (entities sorted by ECS ID, keys sorted) so equal worlds give equal bytes, and `Game.Hash` is the SHA-256 of those bytes.
- **A replay** is the world seed plus the player's command log. The harness uses replays to reproduce any bug or odd emergent run.
- **Versioned.** Every save header carries a save version and a content hash, and migration functions upgrade older saves one version at a time. Until v1.0.0, saves are disposable: a change bumps the save version without a migration, and older saves stop loading.

## Frontends

Frontends only draw the core's read-only view and send it commands, so a second frontend is new drawing code, not a port.

### The terminal frontend (now)

- **One root Bubble Tea model with a screen stack:** the map screen (towns and the deep alike, exploring and fighting), with overlays for party and grafts, investigation notes and dreams. Each is its own model.
- **Bubbles components for the plumbing:** `list` for menus, `table` for the party and inventories, `viewport` for logs and the investigation journal, `textinput` for naming, `help` and `key` for bindings.
- **No game text in Go.** The view carries every word the player reads, including channel state and menu group labels; only key help is written in the frontend. Hot reload reaches the frontend as an opaque function over the game, so the frontend imports only `game` and `api`.
- **Lip Gloss for layout and color.** The map, combat panel and message log are panes joined with Lip Gloss.
- **The hex map** is drawn from the hex library's double-width coordinates: one glyph every two columns, alternate rows offset by one. Depth shows only as shading relative to the party, never as a layer.
- **Combat** happens on the map screen: a side panel shows the deciding member's parts and what each is doing, a menu picks the action, type and target, and a look mode shows any visible body's parts. A timeline strip along the top shows upcoming ticks, an icon per tick with something on it and +N when several share one; selecting a tick lists who acts and what lands. There's no combat screen.

### The graphics frontend (later)

- **Ebitengine** drawing the same view: tiles and sprites, or an ASCII bitmap font to keep the look.
- **It reuses** the same `core/game` API and command set, so the core stays untouched.

### An out-of-process frontend (if Godot happens)

The `core/game` API stays plain Go types for now. If a Godot or other non-Go frontend is built, define that one boundary in protobuf and serve it over gRPC: commands in, a stream of events and view snapshots out. Everything already goes through `core/game`, so nothing else changes.

## Harness and testing

The harness tunes the game with data, and the tests guard the invariants the design depends on.

### The harness

- `harness run --seeds 500 --days 365` runs headless games with a scripted or idle party and writes one CSV row per event or per day: seat stages and passes, deaths per town, migrations, towns emptied, export effects.
- Analysis happens outside Go, in a spreadsheet or notebook, so no custom tooling is needed.
- `harness replay <file>` reproduces any run from its seed and command log.

### Tests

| Kind | Tool | Examples |
| --- | --- | --- |
| Property tests | rapid | Exactly seven seats exist at all times; every town reaches the deep after any reseed; the party's pocket survives an export; replaying a log gives the same world hash |
| Golden tests | `testing` | Procgen output for fixed seeds; combat results for scripted fights |
| Fuzzing | Go native fuzzing | The content loader and Expr formulas against malformed input |
| Content checks | `cue vet` | Every data file against the generated schema |
| Architecture | golangci-lint depguard | `internal/core` imports no UI package; feature packages import only foundation packages |
| Shared sample content | `contenttest.Sample` | One cached, immutable registry per test binary; tests that mutate content take a deep copy |
| Performance | `testing.B`, pprof | Off-screen ticks for a full deep stay fast enough to run 500 seeds overnight |

## Custom code inventory and build order

The custom code is the game's rules, plus thin glue to the libraries above. Nothing on this list has an off-the-shelf equivalent.

### What we write

- **Effect primitives and combat resolution:** channels and timing, actions defined in content, sever rules, graft autonomy.
- **The needs scorer:** a small utility-scoring loop; the curves themselves are Expr formulas in content.
- **Seat rules:** the state machine, passing, wound scoring, the piece's readings.
- **The wish generator:** a pipeline over content tables.
- **Procgen for the deep:** hex cave generation from noise and cellular rules, stitching to town gates, and the party's pocket on reseed.
- **Rumor spread and drift.**
- **The engine spine:** the system scheduler, the command, event and view registries, the trigger runner, the fact store, the action timeline and the action machinery.
- **Glue:** content decode and registry, the RNG stream registry, the Ark component set, the `core/game` API.
- **The Bubble Tea screens.**

### Build order

Each step ends with something playable or measurable.

1. **Skeleton:** module, depguard rule, content pipeline with one material, Ark world, `core/game` API, and the Bubble Tea frontend walking a hand-made hex map.
2. **Combat:** bodies of parts, a few materials and forms, the effect primitives, fighting on the map with the action menu.
3. **Agents:** needs scoring for creatures and people, on-screen and off-screen ticks, generic NPC timers, recruits.
4. **Seats:** one seat, the piece's readings, investigation, killing and passing, wound scoring.
5. **Procgen:** the generated deep around fixed towns, exports reseeding it, the pocket, the guarantees as property tests.
6. **Emergence:** rumors, grudges and memory, dreams.
7. **Tuning:** harness runs over hundreds of seeds; adjust content, not code.

After the MVP: post-MVP milestones live under `tasks/post-mvp/`, each with its own design doc. The first is the story system (`docs/story.md`). Nothing in the MVP depends on them.

Agent-ready task specs for each step live in the repo under `tasks/`, and `tasks/INDEX.md` tracks their status.

## Open questions

- [ ] CUE or YAML plus JSON Schema? CUE removes the most custom validation code, but it has a learning curve for content authoring.
- [x] Hex library: decided, gravitton/hexagon (double-width layout, field of view and line drawing).
- [ ] The deep is infinite, so: how big is the level-of-detail radius, and what's the chunk memory budget? Together they set the off-screen tick budget.
- [x] Stable content IDs: decided, namespaced string IDs plus versioned saves with migrations (see Designing for pivots).
- [ ] Terminal font: hexes look even only with a font about twice as tall as it is wide; pick a default for Windows Terminal.
- [x] Combat: decided, on the hex map and the same clock as exploring, with moves picked from menus.

### Sources

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) and the [v2 announcement](https://charm.land/blog/v2/)
- [Ark ECS](https://mlange-42.github.io/ark)
- [CUE Go integration](https://cuelang.org/docs/integration/go/)
- [Expr](https://github.com/expr-lang/expr)
- [gravitton/hexagon](https://pkg.go.dev/github.com/gravitton/hexagon) and [pmcxs/hexgrid](https://github.com/pmcxs/hexgrid/wiki)
- [fzipp/astar](https://github.com/fzipp/astar)
- [Ebitengine](https://ebitengine.org/)
