---
name: next-tasks
description: Show which tasks can start now, which can safely run side by side, and why the rest are waiting. Use when asked what to work on next or what's ready.
---

# What can start now

Look and report. Don't change anything.

1. Run `git fetch origin`, so that other agents' task branches show up.
2. Run `go run ./internal/tools/agentkit ready`. It reads every task file's front matter, the task branches, and `tasks/INDEX.md`.
3. For each ready task, read its Deliverables to see which files it creates or changes. Two tasks can run side by side when they share no files other than `tasks/INDEX.md`.
4. Report:
   - the ready tasks, each with its title and size
   - which of them can run in parallel
   - claimed tasks and their branches
   - what the waiting tasks need
   - any place where `INDEX.md` disagrees with a task file
