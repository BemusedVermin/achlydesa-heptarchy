---
paths:
  - "tasks/**"
---

# Editing task files

- A task file's front matter is the source of truth, and `tasks/INDEX.md` mirrors it. When you change a status, title or dependency, change both in the same commit.
- Statuses are `proposed`, `todo`, `in-progress`, `blocked`, `done` and `obsolete`. Agents set `in-progress`, `blocked` and `done` on their own task, and `proposed` on the follow-ups they file. `todo` and `obsolete` are the owner's, set when triaging. The exception is a triage task the owner assigned, such as M2-15: it sets them on the follow-ups it triages, and merging its pull request is the owner's approval.
- Quote a title that contains a colon. `depends_on` is a flow list, such as `[M2-05, M2-08]`, or `[]` for none.
- Follow-up IDs are `<parent ID>.F<n>`, numbered from F1 within the parent. Their files go in `tasks/followups/`, and their rows go at the top of the Follow-ups table. The file-followup skill has the steps.
- When implementing a task, tick a criterion (`- [x]`) only after running the test or command that shows it holds. Don't rewrite the task's Goal, Deliverables, Spec or acceptance criteria to match what you built: record the deviation in Notes.
- `go run ./internal/tools/agentkit ready` lists where `INDEX.md` disagrees with the task files.
