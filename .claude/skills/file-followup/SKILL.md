---
name: file-followup
description: File a follow-up task for something found outside the current task's scope (a bug, a missing seam, debt, a missing test, an unclear design point, a doc that needs changing), as AGENTS.md requires. Use instead of fixing it silently or leaving a bare TODO.
argument-hint: <what you found>
---

# File a follow-up

What you found: $ARGUMENTS

1. **Parent ID**: your task's ID, from your branch name `task/<ID>-<slug>`.
2. **Number**: the next free `F<n>` for that parent. List `tasks/followups/<ID>.F*` and count up from F1.
3. **File**: copy `tasks/TEMPLATE.md` to `tasks/followups/<ID>.F<n>-<short-slug>.md`.
4. **Front matter**:
   - `id: <ID>.F<n>`
   - `title`: short and imperative. Quote it if it contains a colon.
   - `status: proposed`
   - `milestone`: your best guess, or `unplanned`
   - `kind`: `feature`, `seam`, `bug`, `debt`, `test`, `content` or `question`
   - `depends_on`: the tasks it needs first, often `[<ID>]`
   - `discovered_by: <ID>`
   - `size`: `S`, `M` or `L`
5. **Sections**: fill in every one. Say what you saw, where (file and line), and why it matters.
   - For `kind: question`, give the options, what each costs, and a recommendation. Never settle a game-design question in code.
   - For a change to `docs/`, quote the passage and give the new wording.
6. **Index**: add a row at the top of the Follow-ups table in `tasks/INDEX.md`:
   `| [<ID>.F<n>](followups/<file>.md) | <title> | proposed | <kind> | <milestone> | <ID> |`
7. **Notes**: list the follow-up's ID in your task's Notes.
8. **Code**: if the gap shows in code, add `// TODO(<ID>.F<n>): <one line>` there.

Then go back to your task. Never start a proposed follow-up yourself: the owner triages them.
