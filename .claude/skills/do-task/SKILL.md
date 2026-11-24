---
name: do-task
description: Implement one task from tasks/ by its ID (like M2-07), from a fresh task branch to an open pull request, following AGENTS.md. Use when asked to do, implement or work on a task.
argument-hint: <task ID, e.g. M2-07>
---

# Do a task

The task ID is in the arguments ($ARGUMENTS) or in the prompt you were given.

You take one task from a fresh branch to an open pull request, then stop: the owner reviews and merges. `AGENTS.md` holds the rules; this is the order to apply them in.

Guardrails run as hooks (`internal/tools/agentkit`). When one refuses a command, read its reason and do what it says. Don't look for another command that does the same thing.

## 1. Check the task can start

1. `git fetch origin`
2. `go run ./internal/tools/agentkit ready`. The task must be listed under **Ready**. If you weren't given an ID, take the first ready task.
   - Under **Claimed**: another agent has a branch for it. Stop and report the branch.
   - Under **Waiting**: stop and report which dependencies aren't done.
   - Status `proposed`, `done` or `obsolete`, or not listed: stop and report. Never start a proposed task.
3. Read the task file, then every doc section it links, then the code it builds on. Also read `.claude/rules/comments.md`, and the file in `.claude/rules/` for each area you'll touch (`core.md`, `content.md`, `frontend.md`, `tasks.md`). They may not have loaded on their own.

## 2. Make the task branch

Branch from the latest master, as `task/<ID>-<slug>` with a slug of two to four words from the title:

```
git switch -c task/M2-07-hostility-alert origin/master
```

- If you're resuming and the branch already exists, switch to it instead.
- If `git status` shows changes you didn't make, stop and ask: they're the owner's.

Set the task's status to `in-progress` in its front matter and in its row of `tasks/INDEX.md`.

## 3. Implement

- Build only what the Deliverables and Spec cover. Keep the spec's names; if one has to change, say why in Notes.
- Every line follows "Rules that never bend" in `AGENTS.md`.
- Where the spec is silent, pick the most reversible option and note it. A game-design question becomes a follow-up of kind `question` (rule 9).
- Work outside the task's scope becomes a follow-up (the `file-followup` skill). A `// TODO(<follow-up ID>)` is fine; a bare TODO is not.
- On a task branch you can't edit `AGENTS.md`, `CLAUDE.md`, `.claude/`, `.github/`, `Taskfile.yml`, `.golangci.yml` or `internal/tools/agentkit/`. If the task needs one of them changed, file a follow-up.
- If the work changes what the design docs in `docs/` say, update them as part of the task and say so in Notes.
- Commit as you go, with messages like `M2-07: alert spreads to nearby bodies`.
- While you work, test only the packages you're changing: `go tool task test -- ./internal/core/combat/...`. The whole suite is slow, and other agents may be running tests on the same machine.
- **Blocked?** Set the status to `blocked` in the task file and the index, say why in Notes, commit, and stop. Report what blocks you. Don't push a blocked branch.

## 4. Verify

1. Run `go tool task fmt`, then `go tool task check-ci BASE=origin/master`, until it passes. It runs every check but tests only the packages your change can affect, as CI does. Fix the cause: never weaken a test or a lint rule to get through.
   - If golangci-lint is missing, install the version that `.github/workflows/ci.yml` pins: `go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@<version>`.
2. For each acceptance criterion, run the test or command that shows it holds, then tick it (`- [x]`). If you can't meet a criterion, that's a blocker (step 3), not a tick.

## 5. Get a review

Ask the `task-reviewer` agent to review the task, giving it the task ID and your branch. Fix every blocking finding. Fix the should-fix findings too, or say in Notes why not. If the fixes changed code or content, run `go tool task check-ci BASE=origin/master` again.

If you can't start an agent, work through the checklist in `.claude/agents/task-reviewer.md` yourself.

## 6. Finish the task file

- Notes: assumptions, deviations, files touched outside the Deliverables, follow-ups filed (IDs), and anything the owner should look at.
- Set the status to `done` in the task file and in `tasks/INDEX.md`.
- Commit. `git status` must be clean.

## 7. Push and open the pull request

1. `git fetch origin`, then `git rebase origin/master`. If `tasks/INDEX.md` conflicts, keep both sides' rows, with your task's status on yours. If a conflict was in code or content, run `go tool task check-ci BASE=origin/master` again after resolving it.
2. Run `go tool task check-pr`. It checks that `tasks/INDEX.md` agrees with the task files and that your task is `done`. On a cleanup task, it also checks that behavior didn't change. Fix whatever it reports.
3. `git push -u origin HEAD`. CI runs the check, and check-pr, on the pull request; if either fails, fix it, commit and push again.
4. Write the PR body to a file outside the repository, such as in your system's temp folder. Use the layout of `.github/pull_request_template.md`: Summary, Notes, and a ticked Test plan.
5. Open the PR:
   ```
   gh pr create --base master --head <branch> --title "<ID>: <title>" --body-file <file>
   ```
   If gh says GraphQL isn't available (cloud sessions block it), use the REST API instead:
   ```
   gh api repos/BemusedVermin/achlydesa-heptarchy/pulls -f title="<ID>: <title>" -f head=<branch> -f base=master -F body=@<file>
   ```
6. Stop there. Don't merge, approve or close the PR, and don't delete the branch.

## 8. Report

Give the PR link, the follow-ups you filed, and anything the owner needs to decide or check.

## Updating an open pull request

Stay on the same branch. Commit the changes, then `git push`. If the PR conflicts because master has moved on, run `git fetch origin` and `git merge origin/master`, resolve the conflicts, run `go tool task check-ci BASE=origin/master`, and push. Never rebase or amend commits you've already pushed, and never force-push.
