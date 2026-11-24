---
id: M3-05
title: The agents chooser for combat, and dodging
status: done
milestone: M3
kind: feature
depends_on: [M3-04]
discovered_by: null
size: M
---

# M3-05 · The agents chooser for combat, and dodging

## Goal

Creatures with needs fight according to them. A registered `agents` chooser scores attacks with the body's needs as well as M2's score, and a creature can step out of reach of a swing that is winding up at it, which M2 left for M3.

## Context

- The agents model in [README.md](README.md), "Fighting".
- `docs/game-design.md` → "Combat" → "Timing" (wind-ups, dodging) and "Monsters with needs" (soul-feeders hunt the slowly dying).
- `tasks/M2-combat/M2-08-act-options-controller.md` (choosers registered by name; "hunting creatures don't dodge yet; reading wind-ups is M3 behavior").
- `internal/core/combat/controller.go` (`ScoreChooser`, `ScoreEnv`), `internal/content` (`AIScore`), `comp.Pending`/`comp.Landing`.

## Deliverables

- `internal/core/agents/chooser.go`: the `agents` chooser, registered through `actions.RegisterChooser`.
- `internal/content/env.go`: `ScoreEnv` gains `source.needs` (through `BodyEnv`) and `target.dying`; the content rule `agents_score`.
- `internal/core/agents/primitives.go`: a `dodge` primitive.
- `content/data/rules.cue`, `content/data/behaviors.cue`: `rule.agents_score` and a dodge behavior.
- Tests beside each.

## Spec

**Firm:**

- The `agents` chooser scores each intent with the content formula named by the config key `agents.score_rule` over `ScoreEnv`; ties are broken by the `agents` stream. M2's `score` stays the default for bodies whose template names no chooser.
- `target.dying` is true for a body with a vital part at or below a content fraction of its max integrity (`agents.dying_below`), so content can make soul-feeders prefer the slowly dying.
- **`dodge`**: when a hostile's pending action targets this body and would land within `params.within` ticks, step to an open neighbour out of that action's range. A consideration can read `ConsiderEnv.threatened` (ticks until the soonest landing aimed at the body, or -1), which this task adds.

**Suggestion:** read pending actions through the `timeline` export already used by the timeline strip, not by scanning `comp.Pending` directly.

## Acceptance criteria

- [x] A test creature with `actor.chooser: "agents"` and a rule preferring `target.dying` attacks a nearly dead body over a whole one at equal range.
- [x] A test creature with a dodge behavior steps out of range of a winding-up Strike, and the Strike misses on landing.
- [x] Creatures without a chooser in their template still fight exactly as before: `go tool task same-behavior` shows no change on the default seeds.
- [x] `go tool task check` passes.

## Out of scope

- The families' real content (M3-18).
- Dodging for party members: the player does that.

## Notes

- `agents.score_rule` names a rule (`agents_score`); `agents.dying_below` is a fraction (0.5). Both are required content. The old `combat.ai_score` text moved to `rule.attack_score`, which `ai_score` and `agents_score` both use, so the default chooser rates as before.
- `target.dying` is `BodyEnv.Dying`, filled by `effects.BodyEnvOf` for every formula.
- The score environment is now built by `actions.ScoreEnvOf`, shared by both choosers; `combat.scoreOf` calls it. `actions.ChooserName` is newly exported.
- Deviation in `combat`: the hunter used to offer only attacks on its nearest hostile, so no chooser could weigh two bodies at equal range. A body whose template names a non-default chooser may now also target hostiles exactly as near as its prey. The default chooser's bodies are unchanged.
- `dodge` (param `within`) picks the safe step farthest from the threats; a threat's range is its action's `range` in hexes. It reads pending actions through `timeline.AimedAt`, the call the timeline view uses. It acts only if the behavior is chosen; `threatened` in `ConsiderEnv` is -1 with no hostile swing aimed at the body.
- `go tool task same-behavior` reports a changed final hash only because the save header holds the content hash and content was added; the harness CSV is identical.
- Test content added: `behavior.test.dodge`, `template.test.predator`, `template.test.dodger`. Docs: `docs/content-guide.md` updated.

