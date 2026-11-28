---
id: C2-04
title: Map the core as it is, and propose the simulation core
status: todo
milestone: C2
kind: debt
depends_on: [C2-02]
discovered_by: null
size: L
---

# C2-04 · Map the core as it is, and propose the simulation core

## Goal

Draw the core as it really is, with an arrow for every way its parts interact. Then propose what replaces it: a reusable **simulation core** for a discrete simulation of many agents, with the game living entirely in content. The proposal is written like a short research paper: a model with every term defined, then the architecture mapped onto the model, then how its cost and correctness are judged. The owner approves it in this task's pull request, and the task files the work that carries it out.

The owner's framing (2026-11-27):

- "I always think of software as a set of very simple services that interact over some API. Defining the API and component boundaries up front will make it easier to reason about."
- Map state is his main example of the problem: today it's "spread out across three or four packages and everything touches it".
- The engine should hold the core components "instead of allowing malleable things to seep into the engine".
- "What I really want is a simulation core. Agents (and this includes the player) all have some controller that makes decisions. All agents navigate some procedurally generated world. All agents may take actions, and a separate rules system decides when they are legal. This is more of a research/mathematical engine that is optimized for a discrete simulation of many agents. The content hooks into the engine by tuning the parameters, providing the content that dictates the behavior."
- Verbs like damage, step and afford "should be expressible FROM engine verbs. They absolutely should not be built into the engine. If something only has one variation or it is domain specific knowledge to the setting, it should be content."

## Context

- AGENTS.md, "Rules that never bend":
  - rules 1, 3, 5, 6, 7 and 10 still hold;
  - rule 2 (the layers) is what this redesigns;
  - rule 4 (content is data) is what it strengthens.
- `docs/go-technical-design.md`, read in this order:
  - "Designing for pivots", first, including "What stays concrete" and "The pivot test";
  - "Repository layout";
  - "Core packages";
  - "Content pipeline";
  - "Simulation loop and determinism".
- `docs/glossary.md`. Names come from there, and a new name is proposed as a rename.
- `.golangci.yml`, `depguard`: the import rules enforced today.
- The code: `internal/core/*`, `internal/content`, `internal/harness`, `internal/frontend/tui` and `cmd/`. Use the code as the source, not the docs.
- What exists that the model generalizes:
  - the controllers (hunt, still, the mind controller);
  - the timeline's busy-until ticks;
  - the procgen stages;
  - the harness's random policy;
  - rule 7's options in the view, which are the legal-action set;
  - the registered effect primitives in each package's `install.go`.
- C1's brief (`tasks/C1-cleanup/README.md`) has the rename practice: `gopls rename`, and a `renames: [old=new, ...]` list in each rename task's front matter.

## Deliverables

- `tasks/C2-cleanup/proposal.md`.
- Task files `tasks/C2-cleanup/C2-05-*.md` onward, plus `kind: question` tasks if any, all `status: proposed`. Each gets a row in `tasks/INDEX.md` and in the brief's task table.

## Decisions already made (owner, 2026-11-27)

- **Content declares the state schema.** "Parts have integrity" is content, built on generic shapes the engine provides. The engine knows no setting word.
- **Engine algorithms read content.** Pathfinding, sight and other hot algorithms stay in the engine, in Go, and read costs and rules from content.
- **Research-style learning agents.** The controller interface supports them: the percept as plain data, a legal-action mask, and a score or reward signal defined in content, so a learned or external policy can drive an agent.
- **The core stays under `internal/` for now.** Making it public later isn't ruled out, so its boundary must allow lifting it out: the core imports no game-specific package, and content is its only coupling to the game.

## Spec

`proposal.md` reads as a short paper, with its parts in this order. Every term is defined where it first appears. Names are descriptive: a single letter only where it has its usual meaning.

1. **The core as it is** (firm). A Mermaid flowchart with one node per package in `internal/core`, plus `internal/content`, the harness and the tui at the edges. It has one arrow for each interaction, and each arrow is labelled with its kind:
   - imports and calls;
   - reads a component;
   - writes a component;
   - emits an event, or handles one;
   - takes a command;
   - builds a view.

   Draw map state (cells, terrain, height, chunks, sight, reachability, the plan view) so its readers and writers stand out. Derive the graph from the code: imports, then reading the uses. If the full graph has more than about 25 arrows, add a one-screen summary beside it.
2. **What the graph shows** (firm). Each problem gets `file:line` examples:
   - state with more than one writer, map state first;
   - packages that reach into other packages' components;
   - setting knowledge in the engine, such as primitives, components or names that only mean something in this game;
   - code that must know another package's internals;
   - rules written in Go that a designer would tune (numbers, tables, formulas, ID lists, thresholds).
3. **The model** (firm that each piece is there; the definitions are the task's to write). This is the simulation the engine runs, defined precisely enough that two people would build the same thing from it. At least:
   - **World state:** entities, each holding named values of kinds that content declares, on generic shapes (numbers, tags, relations, a position and so on). Plus the map: cells with content-declared properties.
   - **Generation:** from a seed and parameters to an initial world state, as a pipeline of stages that content declares.
   - **Time:** discrete events. Each agent has a tick when it is next ready, and the scheduler advances the earliest, breaking ties by UID.
   - **Observation:** from the state and an agent to what that agent perceives (sight, knowledge).
   - **Controller:** from a percept, the agent's memory and its legal actions to one action. The player's controller is a human choosing from the view. Planners, scripted behaviors, the harness policy and learned policies are other controllers behind the same interface.
   - **Rules:** which actions are legal for an agent in a state, and how a chosen action resolves into a new state and events, using a named random stream.
   - **Measurement:** scores and rewards that content defines, read from the state, plus the harness reading outcomes from the state (rule 10).
   - **Simultaneity:** when several agents are ready on the same tick, does each one see the effects of the ones before it, or do they all decide from the same state? This changes outcomes, so file it as a `kind: question` (rule 9), with what each answer costs and allows. Deciding from the same state lets observation and decisions run in parallel.
4. **The divide between engine and content** (firm). Something belongs in the engine only if all three hold:
   - it can be described without any word from the glossary's Game section;
   - content can't do it well, because it's an algorithm or a guarantee (pathfinding, sight, spatial queries, the scheduler, random streams, saves, the one writer);
   - it's reused, since something with one variation is content.

   Reuse alone isn't enough, and reuse inside content is what the rule language's modules are for. This gives three layers:
   - **engine verbs,** in Go;
   - **a content library:** the reusable game mechanics such as damage, sever, step, climbing and dropping, and statuses, written from engine verbs;
   - **setting content:** clay, the forms, the fall injury, archons, and everything else that is one of a kind.

   Classify every effect primitive, component and rule in Go today into a layer, in a table.
5. **The architecture** (firm that each item is there; the shape itself is the task's to propose). Map each part of the model to a component. Aim for about 8 to 12 components, with a top-level graph that fits on one screen and arrows only through named APIs. For each component, give:
   - one sentence on what it does;
   - the state it owns. Every piece of state has exactly one owner, and map state has one.
   - its API: exported Go signatures (queries, commands, the events it emits and handles), marked firm;
   - what it may depend on.

   Internals stay unexported, under an `internal/` folder where that helps. The proposal also gives:
   - **The learning interface:** a step API that a research agent or an outside process can drive. It covers resetting from a seed and parameters, giving an agent's percept and legal-action mask, taking its action, and returning the next percept, the reward and whether the run is over. It also covers how many runs go side by side (rule 3: parallel, then in order). Compare its shape with the multi-agent research environments the owner may know, such as PettingZoo's agent–environment cycle or OpenSpiel's `legal_actions`, and say where it differs and why.
   - **The reuse boundary:** what a second game on the same core would supply, and the check that keeps the core free of this game's packages and names.
   - The new text of AGENTS.md rule 2, and the `depguard` rules that enforce the new graph.
   - The new text of the two tech-design positions this reverses: "Primitives in Go, compositions in data", with no scripting beyond Expr, and "A seam needs a feature". Reuse is now a goal in its own right.
6. **Choices** (firm). For each big choice, such as how the entity store holds content-declared state, give the options briefly, what each costs, and a recommendation. If the owner should decide one, file it as a `kind: question` task. The tasks that need its answer depend on it.
7. **The rule language** (firm that this part is there; the recommendation is the task's to make). The owner said "CUE is great for configuration, but expressing behavior in it is weird." Writing rules is awkward, some behavior needs steps, branches and loops, and he wants rules easy to organize. Today the content holds an unnamed rule language:
   - triggers written as CUE data, with `on`, `if` and `do` (`_fallInjury` in `content/data/templates.cue`);
   - conditions and amounts as Expr strings that nothing checks until load;
   - effect primitives in Go as the verbs, many of them setting-specific;
   - Go pickers named by string from content, such as `target: "flee"` in `content/data/operators.cue`.

   This part has four pieces.
   - **The engine vocabulary:** every engine verb, query a rule may read, event a rule may react to, and formula environment, each with its signature. Only setting-free verbs belong here, per part 4. Suggestion: define the vocabulary once in Go and generate its reference from that, so the docs can't drift.
   - **The options**, judged against the needs below. At least these:
     - CUE plus Expr as today, made better, for example with typed checks, named helpers, and errors that give a file and line;
     - an embedded scripting language for behavior, with CUE kept for declarations and parameters. Starlark (`go.starlark.net`) is the obvious candidate. Compare one or two others briefly, such as Lua or a Go interpreter, and say why each fits or doesn't;
     - a language of the game's own, with its own parser.
   - **The same rules in each option, side by side,** so the owner can compare them by reading:
     - `damage`, as a library rule built from engine verbs;
     - the fall injury trigger that uses it;
     - one action whose effect runs several steps or branches;
     - the flee operator with its picker.
   - **A recommendation,** including where the line between CUE declarations and behavior falls, and how behavior is organized into files and modules.

   The needs, in order:
   - deterministic under AGENTS.md rule 3: randomness only through the engine's streams, no clock, and a fixed iteration order;
   - sandboxed, and sure to stop;
   - errors found at load, with the file and line;
   - compiled once at load, and fast on the hot paths;
   - easy to organize: modules, imports, reuse, grouped by topic;
   - works for both content sets and with hot reload;
   - easy for the owner and for agents to read and write;
   - testable with the test content;
   - behavior state lives in the world state, never in the language, so saves are unchanged;
   - keeps rule 10: a rule may not become a system that forces an outcome.

   File the choice as a `kind: question` task for `/decide`. The implementation tasks that move rules depend on it. If the recommendation is a new language, the breakdown includes a task that adds it and ports the sample rules before the others move. Its Notes give the sentence rule 8 asks for.
8. **The cost model** (firm). Say where the time goes on one tick of the new design, and why it is fast. Cover:
   - **What is compiled at load:** content names resolved to dense indices, formulas to bytecode, rules to whatever the rule language compiles to, and tag sets to bit sets. Nothing on a hot path looks a name up by string.
   - **The data layout** for content-declared state.
   - **What runs in parallel:** for example observation and decisions, if they are pure, and what stays with the one writer.
   - **Level of detail** for far agents.
   - **Caches and what invalidates them,** for sight, reachability and paths.
   - **Where interpreting could dominate,** and the fallback if it does, such as translating content to Go at build time for release builds.

   Give an estimate for each benchmark from C2-03 against the baseline in C2-03's Notes, with the reasoning. Where it's a guess, say so.
9. **The move table** (firm). Every package in `internal/core` today, and every file where a package splits, goes in the table with its new home and its renames as `old=new`.
10. **Rules moving to content** (firm, owner decision 2026-11-27: they move in the same pass). This is a table of every rule in Go that a designer would tune, and every setting-specific primitive and component:
    - `file:line`;
    - what it becomes: a content declaration, a content library rule, or setting content;
    - the implementation task that moves it.

    Production content keeps today's values exactly, so the game plays the same. The test content gets values of its own. Include the content IDs that C2-01 listed as named in Go.
11. **The breakdown** (firm). These are the implementation tasks, C2-05 onward, with `kind: debt` and `status: proposed`. Each one is one boundary and a pull request small enough to review. Order them so that after each one:
    - the build is green;
    - every C2-03 game-level test passes unchanged;
    - the goldens are unchanged.

    Renames use `gopls rename` and are listed in front matter as `renames: [old=new, ...]`. Each task's acceptance criteria include:
    - no file of a C2-03 game-level test changed;
    - goldens unchanged;
    - the C2-03 benchmarks no more than 10% slower than before the task (median of `-count 6`, before and after on the same machine, with the numbers in Notes), unless the cost model's plan says a later task wins it back, in which case the task says which one;
    - `go tool task check-pr` and `go tool task check` pass.

    The last two tasks are fixed:
    - **docs:** the tech design's overview, layout, "Core packages" and "Designing for pivots" rewritten to the model and the new shape, and the glossary and content guide updated;
    - **the lock-in:** run in the owner's session, because it changes protected files. It covers AGENTS.md rule 2 and the Style line from C2-01's Notes, the `.golangci.yml` `depguard` rules, `.claude/rules/core.md`, the audit lenses, and `/audit`'s area list. It depends on every other C2 task.

## Acceptance criteria

- [ ] `proposal.md` has all eleven parts in order, every term in the model is defined where it first appears, and both graphs render on GitHub.
- [ ] The divide table puts every effect primitive, component and rule in Go today into one of the three layers.
- [ ] The engine vocabulary contains no word from the glossary's Game section.
- [ ] The learning interface gives its step API's signatures and how runs go side by side.
- [ ] The rule-language part writes the same four rules in each option, and its choice is filed as a `kind: question` task that the rule-moving tasks depend on. So is the simultaneity question.
- [ ] The cost model gives an estimate for each C2-03 benchmark, with its reasoning.
- [ ] Every package in `internal/core` today appears in the move table.
- [ ] In the proposed graph, every piece of state has one owner, map state included, and every arrow names the API it goes through.
- [ ] Every row of the rules table names the task that moves it.
- [ ] The implementation tasks, the questions, the docs task and the lock-in are filed as `proposed`, with rows in `tasks/INDEX.md` and the brief.
- [ ] `git diff --name-only origin/master...HEAD` lists only files under `tasks/`.
- [ ] `go tool task check-pr` passes.

## Out of scope

- Changing any code: the implementation tasks do it once the owner approves.
- Speed work beyond the cost model: `/optimize`, after C2.
- Making the core a public package or its own module: the boundary allows it, and the owner decides later.

## Notes
