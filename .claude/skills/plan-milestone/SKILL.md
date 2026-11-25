---
name: plan-milestone
description: Break a milestone's outline brief (such as tasks/M3-agents/README.md) into task files. Settles the brief's open questions with the owner first, shows him the breakdown before writing specs, and wires the follow-ups already aimed at the milestone to its real tasks. Use when a milestone is done and the next one needs planning.
argument-hint: "<milestone, e.g. M3; default the next outline>"
disable-model-invocation: true
---

# Plan a milestone

Arguments: $ARGUMENTS

A milestone stays an outline until the one before it is done, so its tasks reflect what the code has become. This skill turns the outline into tasks an agent can pick up. It's judgment-heavy, so it runs best in an Opus session.

## 1. Before starting

1. `git fetch origin`. Everything below reads `origin/master`.
2. **Which milestone.** The arguments name it. Otherwise it's the first one under "Later milestones (outlines)" in `tasks/INDEX.md`. Its brief is `tasks/M<n>-*/README.md`.
3. **The milestone before it.** If it still has tasks that aren't `done` or `obsolete`, list them and ask the owner whether to plan anyway.
4. **Cleanup.** If a cleanup milestone `C<n>` has tasks that aren't done, every task of the new milestone that has no other dependency depends on that cleanup's lock-in task. Its renames would otherwise collide with new work. If no cleanup has run since the milestone before ended, mention `/audit` in one line, and carry on.

## 2. Read

Read all of this before drafting:

- **The brief,** and every doc section it names as a design source.
- **The design docs:** `docs/game-design.md`, and `docs/go-technical-design.md` ("Designing for pivots", the layout, the core packages and the content pipeline). Also `docs/glossary.md` and `docs/content-guide.md` if they exist.
- **What the last milestone taught:** its brief, and the Notes of its tasks and follow-ups. Notes record deviations, provisional choices and things left for later.
- **The code the new tasks will build on.** Check what exists before planning anything new: packages, components, controllers, content kinds.
- **Every task aimed at this milestone:** each follow-up or task with `milestone: M<n>`, whatever its status. Also any that depend on the milestone's ID (`depends_on: [M<n>]`), the placeholder `/triage` uses until the milestone is planned.
- **`tasks/README.md`,** "Writing a task spec", and `tasks/TEMPLATE.md`.

## 3. Settle the open questions

Collect the questions the specs can't be written without:

- the brief's "Questions to settle" list;
- any `kind: question` follow-up aimed at this milestone that has no decision yet;
- any gap you found while reading, where the docs don't say how the game should behave (AGENTS.md rule 9).
- any stand-in a task would need because a later milestone hasn't built something yet. Ask whether to reorder the milestones instead. A stand-in that contradicts the design, such as separate maps joined by gates in a world the design calls seamless, is never yours to choose.

@.claude/skills/shared/asking-the-owner.md

Follow that guidance (if it isn't attached above, read the file) and ask them all before drafting. For each answer:

- write the decision into the design doc where it belongs, as `/decide` does;
- for a `kind: question` follow-up, record it in its file the same way.

A question he answers "Not now" becomes a `kind: question` task in the milestone. The tasks that need its answer depend on it.

## 4. Draft the breakdown and show it

Draft every task as one line: ID (`M<n>-01` upward), title, kind, size, `depends_on`, and what it delivers. Follow "Writing a task spec":

- **Small enough to review.** Each task is one change of size S or M; split anything larger.
- **Seams first.** Shared foundations come before the features that use them. Independent work sits in lanes that can run side by side.
- **Every exit criterion in the brief is covered** by at least one task's acceptance criteria.

Fit in the work already aimed at the milestone. For each such task or follow-up, either:

- fold it into a new task, which then lists it under "Folded in"; or
- keep it, and give it a real `depends_on` on the new tasks it needs. That replaces a placeholder such as `[M<n>]`, or the last task of the milestone before.

Send the owner the outline, with the lanes that can run in parallel and where each existing follow-up went. Then ask, with AskUserQuestion, whether to write the specs:

- "Write them (Recommended)";
- "Change the breakdown": he types what to change, and you redraft and ask again.

## 5. Write it down

On a new branch from `origin/master` named `plan/M<n>`:

1. **A task file per task:** `tasks/M<n>-<name>/M<n>-NN-<slug>.md`, from `tasks/TEMPLATE.md`, with `status: todo` and `milestone: M<n>`. Write the Goal, Context (exact doc sections), Deliverables (files created and changed), Spec (key shapes, marked firm or suggestion), acceptance criteria that are commands or observable behavior, and Out of scope (each tempting neighbour, with the task that owns it). Use the glossary's terms.
2. **Follow-ups and existing tasks:**
   - **Folded ones:** set to `obsolete`, with a Notes line naming the task that took them in.
   - **Kept ones:** give their new `depends_on`.
   - Change each one's `INDEX.md` row in the same commit.
3. **The brief:**
   - Replace "Status: outline" and "Planned tasks" with a Tasks table (ID, title, depends on) and a line on which tasks can start in parallel.
   - Add a table mapping each exit criterion to the tasks that cover it.
   - Turn "Questions to settle" into "Decisions": each answer with where it's recorded, and the questions left open.
   - If the tasks share a model that the design docs don't spell out (how M2's brief has "The combat model"), write it here once. Mark each provisional choice as provisional.
4. **`tasks/INDEX.md`:** move the milestone out of "Later milestones (outlines)" into its own section after the last planned milestone, with the brief link and one row per task.

Then:

- Run `go run ./internal/tools/agentkit ready`. There must be no `INDEX.md` mismatches, and the milestone's first tasks must be listed as ready (or waiting only on the cleanup's lock-in).
- Run `go tool task check`.
- Commit and open a pull request as in step 7 of the do-task skill. The body covers:
  - the tasks, with sizes and lanes;
  - the decisions and where they're recorded;
  - the follow-ups folded in or rewired;
  - the questions left open.

Tell the owner in three or four lines: the number of tasks, which can start first, the pull request's link, and that `/run-tasks` picks them up once it's merged.
