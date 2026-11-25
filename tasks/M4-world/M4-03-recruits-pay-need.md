---
id: M4-03
title: Recruits on a pay need
status: todo
milestone: M4
kind: feature
depends_on: [M4-02]
discovered_by: null
size: M
---

# M4-03 · Recruits on a pay need

## Goal

A recruit stays, asks for its pay, flees or quits because of its own needs, through the same behaviors as anyone else. Quitting is an act, like hiring. The upkeep system, discontent, the flight system and the events that announced their decisions are deleted.

## Context

- `docs/game-design.md` → "Party" → "Recruits" (Needs and wants; Paying them, including the decision on personality, recorded on 2026-11-25).
- `docs/go-technical-design.md` → "Designing for pivots" → "What stays concrete" (Outcomes emerge). AGENTS.md rule 10.
- The brief's "The one-world model" → "Recruits".
- What exists: `internal/core/agents/upkeep.go` (`tickUpkeep`, `desert`, `tickFlight`), `recruits.go` (hiring), `comp.Discontent`, `comp.Fleeing`, `content.RecruitsConfig`, `api.Grumbled`, `api.Deserted`, `api.Fled`, the follow controller in `internal/core/explore/follow.go`, and `_severedTerror` in `templates.cue`.

## Deliverables

- Delete `upkeep.go` and its tests, the `agents.upkeep` and `agents.flight` systems, `comp.Discontent`, `comp.Fleeing`, `api.Grumbled`, `api.Deserted`, `api.Fled`, and `RecruitsConfig`'s `grumble_at`, `desert_at`, `flee_at`, `terror_need` and `flee_behavior`.
- `content/data/needs.cue`: `need.pay`, rising once a day (its `rise` reads `phase`), carried by every hirable template.
- New behavior primitives in `internal/core/agents` (one file, such as `party.go`), registered like the others:
  - `follow_party`: step toward the party member ahead of this one in the party order, as the follow controller does. It scores through its considerations like any behavior.
  - `take_wage`: take the template's `recruit.upkeep` stacks from the party's inventory, if all are there, and `satisfy` `need.pay`. The amount is a param.
  - `quit`: the act of ending the deal, as hiring is the act of making it. It takes the body out of its party. It keeps its body, grafts and inventory, joins `recruits.deserters_faction`, and from then on meets its needs like anyone else.
- `internal/core/api`: `Quit{Party, Body UID}`, the one event: what an agent did. It has a log text.
- `content/data/actions.cue` and `behaviors.cue`: `ask`, the general way any agent asks someone for what it needs. An action with a log line naming the need, aimed at a body that could meet it, used by an `ask` behavior. A recruit asks the party for its pay; later, residents ask the party for help with theirs, which is where quests come from. Run it through `consume` if that can aim a self-voiced action at another body, or through a small primitive; say which in Notes. Nothing in it is specific to recruits or pay.
- `content/data/behaviors.cue` and `templates.cue`: each hirable template lists `follow_party`, `take_wage`, `ask`, `quit` and `flee`, with considerations: an unmet pay need makes asking score, a higher one makes quitting beat following, and terror makes fleeing beat following. The numbers are placeholders for M8.
- `internal/core/explore`: a hired recruit runs under the agents controller while the party is in follow mode. When the player takes individual control, the player decides for it as for any member. A test for each.
- `internal/core/save`: bump `CurrentVersion`.
- `docs/content-guide.md`: rewrite the recruits paragraph of "Needs" and "Behaviors". `docs/glossary.md`: delete **discontent**; **recruit** is "kept by its pay need", not "kept by daily upkeep".

## Spec

- **Personality is later.** Robbery, personality and loyalty are M4-03.F1. Here a recruit only asks, flees and quits.
- **No system decides.** Nothing counts unpaid days or compares a level to a threshold in Go. Every choice is a behavior's score. Firm.
- **Who leads.** A recruit never becomes the party's leader by quitting; `quit` on the leader is refused (it scores 0 for the leader) rather than handled. Suggestion: give `quit` a consideration that reads whether the body leads, through an existing env field, or add one to `BehaviorEnv`.
- `follow_party` replaces nothing in explore for the fixed roster: protagonists keep the follow controller.

## Acceptance criteria

- [ ] `git grep -w -E "Grumbled|Deserted|Fled|Discontent|Fleeing|tickUpkeep|desert_at|grumble_at" -- internal content docs` finds nothing.
- [ ] A test: a recruit whose party pays keeps following for days.
- [ ] A test: an unpaid recruit asks the party for its pay (the action's log line appears), then quits (`api.Quit`) and goes about its needs as a body of the deserters' faction, keeping its grafts.
- [ ] A test: terror raised by a severing makes a recruit flee in a fight, and it falls back in once terror falls.
- [ ] Changed goldens are regenerated and named in Notes.
- [ ] `go tool task check` passes.

## Out of scope

- Robbery, personality and loyalty: M4-03.F1.
- Where a recruit goes once it has quit: its needs decide, through M4-15.

## Notes

Filled in by the implementer.

- Follow-up filed while planning: M4-03.F1 (personality, robbery and loyalty).
