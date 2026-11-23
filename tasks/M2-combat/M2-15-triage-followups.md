---
id: M2-15
title: "Triage the open follow-ups into milestones"
status: done
milestone: M2
kind: debt
depends_on: []
discovered_by: null
size: S
---

# M2-15 · Triage the open follow-ups into milestones

## Goal

Every open follow-up has a decision: it's `todo` in a milestone, folded into another task, or `obsolete` with a reason. After that, `agentkit ready` only offers work that should happen now. The owner has handed this task the triage he normally does himself; the pull request is the proposal, and merging it is his approval.

## Context

- `tasks/README.md`: statuses, front matter, and the Follow-ups section.
- `tasks/INDEX.md`: the Follow-ups table, and each milestone's tasks.
- The milestone briefs, `tasks/M*/README.md` and `tasks/post-mvp/*/README.md`: each one's goal and exit criteria decide what belongs where.
- `docs/game-design.md` and `docs/go-technical-design.md`, to judge what a follow-up is for.
- The open follow-ups when this task was filed:

  | ID | Status | Milestone | Kind |
  | --- | --- | --- | --- |
  | M2-02.F1 | proposed | M2 | bug |
  | M1-10.F1.F1.F1 | proposed | unplanned | debt |
  | M1-14.F1 | proposed | unplanned | seam |
  | M1-16.F1 | proposed | unplanned | seam |
  | M1-20.F2 | proposed | unplanned | seam |
  | M2-06.F4 | todo | unplanned | question |

  List them again when you start: agents may have filed more since.

## Deliverables

- The front matter, and a line in Notes, of each follow-up triaged.
- The matching rows of `tasks/INDEX.md`.
- The specs of any tasks a follow-up is folded into.

## Spec

1. **Which follow-ups.** Every follow-up with `status: proposed`, and every `todo` follow-up whose milestone is still `unplanned`. Leave alone any follow-up that already has a `task/<ID>` branch or an open pull request.
2. **Check it still applies.** Read the code the follow-up points at. A follow-up that another change has already fixed or made pointless becomes `obsolete`. Its Notes say which change did it.
3. **Decide one of three:**
   - **`todo` in a milestone.** Choose the earliest milestone whose goal, exit criteria or tasks need it. A bug that hurts current work goes in the current milestone. Set `milestone`, then set `depends_on` so `agentkit ready` offers it at the right time. For a milestone after the current one, depend on the last task of the milestone before it (for M3, that's M2-11).
   - **Folded into another task.** If it fits a task that isn't done yet, add its point to that task's Spec, Deliverables and acceptance criteria. Then set the follow-up to `obsolete` with a note naming the task.
   - **`obsolete`.** It no longer applies, or isn't worth doing. Give the reason.
4. **Questions** (`kind: question`). Don't decide the game design (AGENTS.md rule 9). If the follow-up's file already records the owner's decision, triage it like any other follow-up. If not, leave its status alone and list it in the pull request for the owner, with the options and your recommendation.
5. **Record each decision.** Add one line to the follow-up's Notes: `Triaged by M2-15: <decision>, because <reason>.` Change the file and its INDEX row in the same commit.
6. **Don't implement anything.** Only task files and `tasks/INDEX.md` change.
7. **Pull request.** The body has a table with one row per follow-up: ID, decision, milestone, reason. List the questions left for the owner under it.

This is the one task where an agent sets `todo` and `obsolete` on tasks other than its own.

## Acceptance criteria

- [x] No follow-up is left `proposed`, except questions that wait on the owner's decision. Each of those is listed in the pull request.
- [x] No `todo` follow-up is left with `milestone: unplanned`.
- [x] Each triaged follow-up has its `Triaged by M2-15` line, and its INDEX row matches its file.
- [x] `go run ./internal/tools/agentkit ready` reports no INDEX mismatches. Its Ready list holds only work meant for the current milestone.
- [x] `go tool task check` passes.

## Out of scope

- Implementing any follow-up.
- Breaking later milestones into tasks.
- Changing follow-ups that are `done`, or that an open pull request is working on.

## Notes

Triaged eight follow-ups (the six listed at filing had mostly been done since; M2-07 filed F1 to F3 later). M2-07.F1 is a question with no recorded owner decision, so it is left `proposed` for the owner (see the pull request). The three parked seams M1-14.F1, M1-16.F1 and M1-20.F2, and the measure-first M1-10.F1.F1.F1, are `obsolete`: no M2 to M7 milestone needs folds, cell targets or a reachability cache, and each says when to refile. M2-02.F1, M2-07.F2 and M2-07.F3 are `todo` in M2 and ready now; M2-07.F3 is best done before M2-08. No follow-ups filed.
