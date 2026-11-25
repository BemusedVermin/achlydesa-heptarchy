---
id: M4-01
title: Revert M3
status: done
milestone: M4
kind: debt
depends_on: []
discovered_by: null
size: L
---

# M4-01 · Revert M3

## Goal

The code and content go back to just before M3. M3 built agents before the world existed, so it hard-coded behavior around stand-ins: maps as regions joined by gates, counts instead of creatures, and a system per outcome. Minds (M6) designs how agents decide, once the world and procgen exist.

## What was done

- `internal/`, `cmd/` and `content/` restored to `234bf8d` (the last commit before M3), then C1-31 (#125) and C1-31.F1 (#129) re-applied, minus the needs collection only M3 used.
- `docs/content-guide.md` and `docs/glossary.md` restored the same way, with C1-31's note. The game design keeps every decision M3 recorded.
- The tools, CI and agent config are unchanged.

## Lost with M3, to add back later

Mostly in M6:

- the party eating and drinking (M3-13), and resting to heal (M3-14);
- hiring and exotic recruits (M3-15, M3-17);
- the monster families as content (M3-18);
- an action saying whether it may target a corpse, and what it may target, by tag (M3-04.F2, M3-04.F2.F1).

## Notes

Done by the owner's decision, 2026-11-25, in the pull request that replanned M4.
