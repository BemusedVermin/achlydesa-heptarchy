---
paths:
  - "internal/tools/**"
---

# Working on the tools

- `internal/tools/agentkit` holds the hooks that guard agents. It is protected on task branches, so shared code that agents could change does not go in it. Its git helper is its own copy.
- `samebehavior` and `affected` share `internal/tools/internal/gitcmd`. Tests that need a repository use `internal/tools/internal/gittest`.
- Tools are infrastructure: a thin smoke test each, and no test that runs longer than the code it protects.
- Commands you write for the owner must work in PowerShell, and paths must work on Windows: build them with `filepath`, never by joining with `/`.
