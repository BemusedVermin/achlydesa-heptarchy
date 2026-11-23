---
name: triage
description: Triage the proposed tasks in tasks/, follow-ups and audit cleanup tasks alike. Marks the ones that no longer apply obsolete, asks the owner as multiple choice what to do with each of the rest, and records it all in one pull request. Use when the owner wants to triage, sort or plan the follow-ups. Design questions go through /decide instead.
argument-hint: "[follow-up IDs; default all proposed ones]"
---

# Triage the proposed follow-ups

Arguments: $ARGUMENTS

Agents file follow-ups as `proposed`, and nothing starts them until the owner triages them. This skill does the reading for him: it settles the ones that no longer apply on its own, and asks him about the rest. The rules for each decision are in `tasks/README.md`, under "How a follow-up is triaged".

## 1. Find them

1. `git fetch origin`, then read the task files on `origin/master`.
2. Take every task with `status: proposed`: follow-ups in `tasks/followups/`, and the cleanup tasks an `/audit` filed in `tasks/C*-cleanup/`. Leave out `kind: question`: those are `/decide`'s. If the arguments name IDs, take only those.
3. Leave out any follow-up that already has a `task/<ID>` branch on the remote or an open pull request.
4. If none are left, say so in one line, mention how many questions are open for `/decide`, and stop.

## 2. Check each one against the code

For each follow-up, read its file, then the code, content and docs it points at as they are on `origin/master` now. Look at what changed there since it was filed: `git log origin/master --since=<the date its file was added> -- <the paths it names>`. Sort it:

- **Obsolete:** a later change already did what it asks, the code it points at is gone, or another task or follow-up covers the same thing. Note the evidence: the commit or pull request that did it, or the task it duplicates.
- **Live:** it still applies. Work out where it would fit:
  - the milestone it belongs in;
  - a task that isn't done yet which it could fold into;
  - whether it blocks anything.

When you can't tell whether it still applies, treat it as live and say so in the question.

## 3. Ask about the live ones

@.claude/skills/shared/asking-the-owner.md

Follow that guidance (if it isn't attached above, read the file). For each live follow-up, the question says what's wrong or missing and what it would take, in a sentence or two. The options:

- **Your recommendation first:** usually "Todo in M2" for the current milestone, or a later milestone. A cleanup task already has its milestone, so the recommendation is just "Todo".
- **One alternative,** such as another milestone, or "Fold into M2-11" when a task that isn't done covers the same ground.
- **"Won't do",** which marks it obsolete. Record the reason he gives, or "the owner decided not to do it".
- **"Not now",** which leaves it `proposed`.

Don't ask about the obsolete ones. They go in the pull request with their evidence, so he can see each one and push back in review.

## 4. Record the decisions

On a new branch from `origin/master` named `triage/<today's date>`, apply each decision as `tasks/README.md` describes:

- status, milestone and `depends_on`;
- for a fold, the target task's Spec, Deliverables and acceptance criteria;
- the one-line `Triaged <date>: ...` note in the follow-up's Notes;
- its `INDEX.md` row, in the same commit.

Change nothing else: no code, no content, no docs. Commit once per follow-up, with a message such as `Triage M2-08.F2: todo in M2`.

Then run `go run ./internal/tools/agentkit ready`. It must report no `INDEX.md` mismatches. Note which tasks it now lists as ready.

## 5. Pull request

Push and open a pull request, as in step 7 of the do-task skill (body file outside the repository; `gh pr create`, or the `gh api` REST call if GraphQL is blocked). The body:

- **Summary:** a table with one row per follow-up: ID, decision, milestone, reason. The obsolete ones you found come first, each with its evidence.
- **Left proposed:** the ones he answered "Not now".
- **Questions:** how many `kind: question` follow-ups are open, with "run /decide".

Then tell him in two or three lines: how many follow-ups became `todo`, how many were obsolete, the pull request's link, and which tasks are ready now.
