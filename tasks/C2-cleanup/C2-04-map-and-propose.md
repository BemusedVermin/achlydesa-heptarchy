---
id: C2-04
title: Map the core as it is, and propose its new shape
status: todo
milestone: C2
kind: debt
depends_on: [C2-02]
discovered_by: null
size: L
---

# C2-04 · Map the core as it is, and propose its new shape

## Goal

Draw the core as it really is, with an arrow for every way its parts interact. Then propose a smaller shape a person can follow at a glance. Each component in it owns its own state, has one job you can say in a sentence, and talks to the others only through an API it defines. The engine keeps the mechanisms, and everything a designer would tune is content. The owner approves the proposal in this task's pull request, and the task files the work that carries it out.

The owner's framing: "I always think of software as a set of very simple services that interact over some API. Defining the API and component boundaries up front will make it easier to reason about." Map state is his main example: today it's "spread out across three or four packages and everything touches it". He likes the CUE content because it lets new rules and content in without engine changes. So the work is to zero in on the core engine components, "instead of allowing malleable things to seep into the engine".

## Context

- AGENTS.md, "Rules that never bend". Rules 1, 3, 5, 6, 7 and 10 still hold. Rule 2 (the layers) is what this redesigns. Rule 4 (content is data) is what it strengthens.
- `docs/go-technical-design.md`, read in this order:
  - "Designing for pivots", first;
  - "Repository layout";
  - "Core packages";
  - "Content pipeline";
  - "Simulation loop and determinism".
- `docs/glossary.md`. Names come from there, and a new name is proposed as a rename.
- `.golangci.yml`, `depguard`: the import rules enforced today.
- The code: `internal/core/*`, `internal/content`, `internal/harness`, `internal/frontend/tui` and `cmd/`. Use the code as the source, not the docs.
- C1's brief (`tasks/C1-cleanup/README.md`) has the rename practice: `gopls rename`, and a `renames: [old=new, ...]` list in each rename task's front matter.

## Deliverables

- `tasks/C2-cleanup/proposal.md`.
- Task files `tasks/C2-cleanup/C2-05-*.md` onward, plus `kind: question` tasks if any, all `status: proposed`. Each gets a row in `tasks/INDEX.md` and in the brief's task table.

## Spec

`proposal.md` has these parts, in this order.

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
   - foundation packages that hold game rules;
   - code that must know another package's internals;
   - rules written in Go that a designer would tune (numbers, tables, formulas, ID lists, thresholds).
3. **The engine and the content** (firm). Draw the line. The engine holds mechanisms, for example how a step resolves, how an effect applies, how sight and paths are computed, the timeline, the random streams and saves. The content holds the policy: every number, table, formula, threshold and piece of text a designer would tune.
4. **The proposed shape** (firm that each item is there; the shape itself is the task's to propose). Aim for about 8 to 12 components, with a top-level graph that fits on one screen and arrows only through named APIs. For each component, give:
   - one sentence on what it does;
   - the state it owns. Every piece of state has exactly one owner, and map state has one.
   - its API: exported Go signatures (queries, commands, the events it emits and handles), marked firm;
   - what it may depend on.

   Internals stay unexported, under an `internal/` folder where that helps. The proposal also gives:
   - the new text of AGENTS.md rule 2;
   - the `depguard` rules that enforce the new graph.
5. **Choices** (firm). For each big choice, such as whether to keep the entity store as it is, give the options briefly, what each costs, and a recommendation. If the owner should decide one, file it as a `kind: question` task. The tasks that need its answer depend on it.
6. **The rule language** (firm that this part is there; the recommendation is the task's to make). The owner, on 2026-11-27, said: "CUE is great for configuration, but expressing behavior in it is weird." Writing rules is awkward, and some behavior needs steps, branches and loops. He also wants rules easy to organize. Today the content holds an unnamed rule language:
   - triggers written as CUE data, with `on`, `if` and `do` (`_fallInjury` in `content/data/templates.cue`);
   - conditions and amounts as Expr strings that nothing checks until load;
   - effect primitives in Go as the verbs;
   - Go pickers named by string from content, such as `target: "flee"` in `content/data/operators.cue`.

   This part has four pieces.
   - **The vocabulary.** The API content programs against: every verb (effect primitive), every query a rule may read, every event a rule may react to, and every Expr environment, each with its signature. Say which are missing and which should be merged. Suggestion: define the vocabulary once in Go and generate its reference from that, so the docs can't drift.
   - **The options**, judged against the needs below. At least these:
     - CUE plus Expr as today, made better, for example with typed checks, named helpers, and errors that give a file and line;
     - an embedded scripting language for behavior, with CUE kept for data. Starlark (`go.starlark.net`) is the obvious candidate. Compare one or two others briefly, such as Lua or a Go interpreter, and say why each fits or doesn't;
     - a language of the game's own, with its own parser.
   - **The same rules in each option.** Write three real rules in each option, side by side, so the owner can compare them by reading:
     - the fall injury trigger;
     - one action whose effect runs several steps or branches;
     - one operator with its picker, such as flee.
   - **A recommendation,** including where the line between CUE data and behavior falls, and how behavior is organized into files and modules.

   The needs, in order:
   - deterministic under AGENTS.md rule 3: randomness only through the engine's streams, no clock, and a fixed iteration order;
   - sandboxed, and sure to stop;
   - errors found at load, with the file and line;
   - compiled once at load, and fast on the hot paths;
   - easy to organize: modules, imports, reuse, grouped by topic;
   - works for both content sets and with hot reload;
   - easy for the owner and for agents to read and write;
   - testable with the test content;
   - behavior state lives in components, never in the language, so saves are unchanged;
   - keeps rule 10: a rule may not become a system that forces an outcome.

   File the choice as a `kind: question` task for `/decide`. The implementation tasks that move rules depend on it. If the recommendation is a new language, the breakdown includes a task that adds it and ports the three rules before the others move. Its Notes give the sentence rule 8 asks for.
7. **The move table** (firm). Every package in `internal/core` today, and every file where a package splits, goes in the table with its new home and its renames as `old=new`.
8. **Rules moving to content** (firm, owner decision 2026-11-27: they move in the same pass). This is a table of every rule in Go that a designer would tune:
   - `file:line`;
   - what it becomes: a content kind and field, or a behavior in the rule language from part 6;
   - the implementation task that moves it.

   Production content keeps today's values exactly, so the game plays the same. The test content gets values of its own. Include the content IDs that C2-01 listed as named in Go.
9. **The breakdown** (firm). These are the implementation tasks, C2-05 onward, with `kind: debt` and `status: proposed`. Each one is one boundary and a pull request small enough to review. Order them so that after each one:
   - the build is green;
   - every C2-03 game-level test passes unchanged;
   - the goldens are unchanged.

   Renames use `gopls rename` and are listed in front matter as `renames: [old=new, ...]`. Each task's acceptance criteria include:
   - no file of a C2-03 game-level test changed;
   - goldens unchanged;
   - the C2-03 benchmarks no more than 10% slower than before the task (median of `-count 6`, before and after on the same machine, with the numbers in Notes);
   - `go tool task check-pr` and `go tool task check` pass.

   The last two tasks are fixed:
   - **docs:** the tech design's layout, "Core packages" and "Designing for pivots" rewritten to the new shape, and the glossary and content guide updated;
   - **the lock-in:** run in the owner's session, because it changes protected files. It covers AGENTS.md rule 2 and the Style line from C2-01's Notes, the `.golangci.yml` `depguard` rules, `.claude/rules/core.md`, the audit lenses, and `/audit`'s area list. It depends on every other C2 task.

## Acceptance criteria

- [ ] `proposal.md` has all nine parts, and both graphs render on GitHub.
- [ ] The rule-language part lists the whole vocabulary, writes the same three rules in each option, and its choice is filed as a `kind: question` task that the rule-moving tasks depend on.
- [ ] Every package in `internal/core` today appears in the move table.
- [ ] In the proposed graph, every piece of state has one owner, map state included, and every arrow names the API it goes through.
- [ ] Every rule in the rules table names the task that moves it.
- [ ] The implementation tasks, the questions, the docs task and the lock-in are filed as `proposed`, with rows in `tasks/INDEX.md` and the brief.
- [ ] `git diff --name-only origin/master...HEAD` lists only files under `tasks/`.
- [ ] `go tool task check-pr` passes.

## Out of scope

- Changing any code: the implementation tasks do it once the owner approves.
- Speed work: `/optimize`, after C2. Note what you see in passing, but don't plan it.

## Notes
