---
name: decide
description: Walk the owner through the open design questions in tasks/ one at a time as multiple-choice questions, then record each answer in its task file and in the design doc, in one pull request. Use when the owner wants to answer, decide or go through open design questions.
argument-hint: "[question IDs; default all open ones]"
---

# Decide open design questions

Arguments: $ARGUMENTS

Agents file design questions as follow-ups with `kind: question`, because only the owner decides the game design (AGENTS.md rule 9). This skill gets his answers quickly, often from his phone, and records them where agents will find them.

## 1. Find the open questions

1. `git fetch origin`, then read the task files on `origin/master`.
2. A question is open when its front matter has `kind: question`, its status is `proposed` or `todo`, and its Spec doesn't yet say **Decided by the owner**. If the arguments name IDs, take only those.
3. If none are open, say so in one line and stop.
4. Order them as `tasks/INDEX.md` lists them: earlier milestones first, and within one, the ones blocking the most work first.

## 2. Prepare each question

For each one, read the task file, the docs and code it points at, and what depends on it. Then work out:

- **The question**, in a sentence or two the owner can answer without opening anything. Name the forms, parts or rules involved; say what the game does today.
- **Two or three options**, each a real answer with its consequence in one line, such as what the player would notice or what content would change. Put the one you recommend first, with "(Recommended)" at the end of its label, and recommend only what the docs and the game's direction support.
- **"Not now"** as the last option, for questions he wants to leave open.

## 3. Ask

Ask with AskUserQuestion, up to four questions per call, in batches until all are asked. He may be on his phone and see only the question dialog, so each question stands alone: no "as above", no IDs without a description. Keep each header under 12 characters, such as the form or rule the question is about. He can always type his own answer instead of picking one.

## 4. Record each answer

On a new branch from `origin/master` named `decide/<today's date>`, for every question he answered (skip the "Not now" ones):

1. **Task file.** Add a paragraph to the Spec that starts with `**Decided by the owner:**` and gives the answer as he gave it, plus what follows for the content or code. Then:
   - If work follows, set `status: todo`. If `milestone` is `unplanned`, set it to the current milestone (the first one in `tasks/INDEX.md` that still has tasks not done).
   - If nothing needs to change (the answer keeps what the game does today), set `status: done` and say so in Notes.
   - Update its row in `tasks/INDEX.md` in the same commit.
2. **Design doc.** Write the decision into `docs/game-design.md` in the section it belongs to, in the doc's own voice, as a statement of how the game works, not as a change log. If it changes how the code is laid out, update `docs/go-technical-design.md` the same way. Tick the task's criterion about the design doc, if it has one.
3. Change nothing else: no content, no code. The work the answer calls for is the task's, for an agent to do.

Commit with a message like `Decide M2-06.F4: Name and Keepsake are never refused`, one commit per question.

## 5. Pull request

Push and open a pull request, as in step 7 of the do-task skill (body file outside the repository; `gh pr create`, or the `gh api` REST call if GraphQL is blocked). The body:

- **Summary:** a table with one row per answered question: ID, the decision in a few words, and what happens next (`todo` in M2, or `done`).
- **Left open:** the questions he answered "Not now".
- **Notes:** each section of the design docs you changed.

Then tell the owner in two or three lines: how many he decided, the pull request's link, and which tasks are now ready because of it.
