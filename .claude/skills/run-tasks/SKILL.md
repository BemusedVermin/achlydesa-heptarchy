---
name: run-tasks
description: Start a task-implementer agent on each of several ready tasks at once, each in its own worktree, and report their pull requests.
argument-hint: "[task IDs, or how many tasks; default 2]"
disable-model-invocation: true
---

# Run several tasks at once

Arguments: $ARGUMENTS

You coordinate. You don't implement, commit, push or merge anything yourself.

1. Run `git fetch origin`, then `go run ./internal/tools/agentkit ready`.
2. Pick the tasks:
   - If the arguments name task IDs, use those. Each must be listed under **Ready**; for any that isn't, say why and leave it out.
   - If the arguments are a number, take that many ready tasks from the top of the list.
   - With no arguments, take two.
   - Never start more than four at once.
3. Check for overlap. Read each picked task's Deliverables. If two tasks name the same file (other than `tasks/INDEX.md`), keep the first and defer the other: they would conflict.
4. Say which tasks you're starting, and which you deferred and why.
5. Start one `task-implementer` agent per task, all at once, in the background. Each prompt is just `Implement task <ID>.`: the agent already has the workflow.
6. As each agent finishes, relay its report: the task ID, the PR link or why it stopped, the follow-ups it filed, and anything the owner has to decide.

Parallel pull requests often conflict on `tasks/INDEX.md` once the first of them merges. To fix one, open a session and ask Claude to update that PR from master. The do-task skill's last section covers it.
