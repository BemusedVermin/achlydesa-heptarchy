---
name: audit
description: Audit the codebase between milestones for comment bloat, duplication, vocabulary, library use and Go and CUE conventions, and file the cleanup as a C<n> milestone of proposed tasks. "/audit lock-in" turns a finished cleanup's decisions into lint rules and rules for agents. Use when the owner wants a code audit or a cleanup pass.
argument-hint: "[lock-in]"
disable-model-invocation: true
---

# Audit the code, then lock in the cleanup

Arguments: $ARGUMENTS

A cleanup runs between milestones, and repeats after each one:

1. `/audit`: read-only auditors find the problems, and this skill files them as a cleanup milestone `C<n>`.
2. `/decide`: the owner settles the names and the new libraries.
3. `/triage`: the owner accepts, merges or drops the cleanup tasks.
4. `/run-tasks`: agents do them. The dependencies keep renames one at a time.
5. `/audit lock-in`: what was decided becomes lint rules and rules for agents, so the next audit finds less.

With `lock-in` in the arguments, skip to "Lock-in" at the end.

## 1. Before starting

1. `git fetch origin`. Everything below reads `origin/master`.
2. List open pull requests and remote `task/` branches. If any are open, tell the owner: cleanup tasks will conflict with them. Carry on only if he says so.
3. The cleanup milestone is `C<n>`, one higher than the last `tasks/C*-cleanup/` folder (`C1` if there is none). Read the last one's brief, if any, for what it locked in.

## 2. Split the code into areas

An area is about 2,000 to 4,000 lines of non-test Go, or the content, with related packages kept together and layers not mixed. Count with `git ls-files`. Today that gives:

1. **Content:** `content/` and `internal/content`.
2. **Actions:** `internal/core/actions`, `effects` and `timeline`.
3. **Features:** `internal/core/combat` and `explore`, plus any newer feature packages.
4. **Foundation:** `internal/core/api`, `sim`, `comp`, `anatomy`, `hexmap`, `inventory`, `clock`, `knowledge` and `rng`.
5. **Game and frontend:** `internal/core/game` and `save`, `internal/harness`, `internal/frontend` and `cmd/`.
6. **Tools:** `internal/tools`.

Regroup when packages have grown or appeared since.

## 3. Start the auditors

Start them all at once, in the background, each a `code-auditor` agent:

- **One per area,** with the lenses `comments`, `duplication`, `libraries` and `conventions`. The prompt is just: `Audit <area name>: <its packages and folders>. Lenses: comments, duplication, libraries, conventions.`
- **One for vocabulary,** across the whole codebase and the docs: `Audit the vocabulary of the whole codebase and docs. Lens: vocabulary.`

Say what you started. Wait for every report.

## 4. Turn the findings into tasks

1. **Merge** findings that several auditors reported. Duplication across areas becomes one task, at the lowest layer both sides may import.
2. **Drop** what isn't worth a task, and say so in the brief.
3. **Questions** (`kind: question`), for the owner to answer with `/decide`:
   - each contested name, with the options and your recommendation;
   - each new third-party library;
   - each convention that conflicts with `AGENTS.md`.
4. **Tasks** (`kind: debt`), in this order:
   - **Renames.** One task per contested name, depending on its question, and each depending on the rename before it, so they run one at a time. Agents rename Go identifiers with `gopls rename`, never by hand, and content IDs everywhere they appear.
   - **Duplication and libraries.** One task per area, or per cross-area finding, each depending on the last rename.
   - **Comments.** One task per area, depending on that area's duplication task, so they run once the code has settled. The task quotes the comments lens's policy and the auditor's examples.
   - **Lock-in.** Last, depending on every other task. Its Spec says to run `/audit lock-in` in the owner's session, because it changes files agents can't touch.
5. **Every cleanup task's acceptance criteria** include:
   - `go tool task same-behavior` passes. For a rename, add `-- --rename old=new` for each rename it makes, all listed in the task.
   - `go tool task check` passes.
   - The task's findings, one per line, each resolved.

   Saves are disposable before v1.0.0, so renaming content IDs needs no save migration.

## 5. Write it down

On a new branch from `origin/master` named `audit/C<n>`:

- **`tasks/C<n>-cleanup/README.md`,** the milestone brief:
  - the goal;
  - a table of findings by area and lens, with what was dropped and why;
  - the ordering above;
  - the exit criteria: every task done or obsolete, and the lock-in merged.
- **One task file per task:** `C<n>-NN-<slug>.md`, from `tasks/TEMPLATE.md`, with `status: proposed`, `milestone: C<n>`, a kind, a size and `depends_on`.
- **`tasks/INDEX.md`:** a `C<n> · Cleanup` section after the milestone just finished, one row per task.
- **`docs/glossary.md`:** create it if it doesn't exist, with `## Game` and `## Code` sections and one line per term: `**term**: definition`. Add the settled terms. Contested ones wait for `/decide`.

Commit, push, and open a pull request as in step 7 of the do-task skill. Its body:

- the counts by area and lens;
- the questions;
- what was dropped;
- the next steps: `/decide`, `/triage`, `/run-tasks`, then `/audit lock-in`.

Tell the owner in three or four lines: how many tasks and questions, the pull request's link, and that `/decide` comes next.

## Lock-in

Run when every other task of the latest `C<n>` is done or obsolete. On a new branch from `origin/master` named `setup/C<n>-lock-in`:

1. **Lint rules.** For each "Lint:" rule the auditors named, and each convention the owner decided, try enabling it in `.golangci.yml`. Keep it if `golangci-lint run` passes on the code. If it doesn't, fix the few remaining cases, or leave the rule off and say why.
2. **`AGENTS.md`.** Add the decided conventions to Style. Add `docs/glossary.md` to "Read before touching code" if it isn't there.
3. **`.claude/rules/`.** Add each area's conventions to its rule file, and point the rule files at the glossary.
4. **Lens files.** Add what this round found and locked in, so the next audit checks those first and doesn't report them again.
5. **The lock-in task.** Set it to `done`, with Notes listing what changed. Change its `INDEX.md` row in the same commit.

Run `go tool task check`, then open the pull request. Tell the owner what is now enforced.
