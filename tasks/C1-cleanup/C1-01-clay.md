---
id: C1-01
title: "What do we call the clay creature that moves, and the body Stand freezes?"
status: done
milestone: C1
kind: question
depends_on: []
discovered_by: null
size: S
---

# C1-01 · What do we call the clay creature that moves, and the body Stand freezes?

## Goal

The owner picks one option, and the answer is recorded here and where the glossary or design doc needs it.

## Context

- Vocabulary audit finding 1. `docs/game-design.md` lines 137 and 491.

## Deliverables

- The answer, written in this task's Notes and in `docs/glossary.md` (and `docs/` where the design says otherwise).

## Spec

"Crawler" names three things: `template.creature.crawler` (a clay creature with no `crawls` tag, so Stand never freezes it), `template.test.crawler` (the test body that does crawl) and `body.crawler` (an unused body plan, described as clay but made of flesh parts). The design calls the creature "Moving clay"; `bodies.cue` calls its body `clay_thing`.

- **moving_clay**: `template.creature.moving_clay`, `body.moving_clay`, `part.moving_clay_*`. Matches the design. Delete `body.crawler`, or move it to the test bundle as `body.test.*`. Keep "crawl" only for the `crawls` tag and crawling parts.
- **clay_thing**: Keep the current body name and rename only the template.
- **crawler**: Keep it, and glossary-define it as the creature that crawls. Needs the creature to gain `crawls`, which is a design change.

**Recommendation:** `moving_clay`. The player-facing name is also the owner's call.

**Decided by the owner:** the creature is `moving_clay`, as the design says. `template.creature.crawler` becomes `template.creature.moving_clay`, `body.clay_thing` and `body.crawler` become `body.moving_clay` (the unused flesh `body.crawler` is deleted or moved to the test bundle as `body.test.*`), and the parts become `part.moving_clay_*`. "Crawl" is kept only for the `crawls` tag and crawling parts. The player-facing name stays as the design has it. Applied by C1-17.

## Acceptance criteria

- [x] The owner has answered, through `/decide`.
- [x] The answer is in Notes and in the glossary.

## Out of scope

- Doing the work: the cleanup task named in the milestone brief does it.

## Notes
Decided 2026-11-23: `moving_clay`. In the glossary.
