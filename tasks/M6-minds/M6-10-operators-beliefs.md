---
id: M6-10
title: "Operators and beliefs: the state both planners plan over"
status: done
milestone: M6
kind: seam
depends_on: [M6-03, M6-04, M6-07]
discovered_by: null
size: M
---

# M6-10 · Operators and beliefs: the state both planners plan over

## Goal

One symbolic model that GOAP (M6-15) and HTN (M6-16) both use: what an agent believes, the operators that change it, and how a plan step becomes a timeline action. Nothing here searches yet.

## Context

- `docs/game-design.md` → "How minds work" (Goal, then plan; Beliefs; Landmarks).
- Prior art: Orkin, "Three States and a Plan" (GOAP); the Ubisoft GOAP postmortem on what it cost in production; HTN used in large-crowd games.
- `internal/core/actions`: content actions with channels, wind-up and effects. A plan step is one such action.

## Deliverables

- `internal/content/`: `Operator` (`operator.` IDs): `pre` and `eff` over symbolic facts, `cost`, `does` (the action ID and how its target is picked), `kind` tag. `Goal` (`goal.` IDs): the drive it serves and the state it wants. `Recipe` (`recipe.` IDs): a goal and an ordered list of operators and sub-recipes (read by M6-16).
- `internal/core/agents/state.go`: `State`, a set of symbolic facts of three forms: `has(tag, n)`, `near(tag)`, `known(tag)`, `drive_met(id)`; `StateOf(w, body)` builds it from the body's inventory, position and memory, never from the world. `Apply(op, state)` and `Holds(op, state)`.
- `internal/core/agents/plan.go`: `Plan{Goal, Steps}`, `Step{Operator, Target}`.
- Tests: a state built from beliefs ignores a well the agent has not seen; applying an operator's effects changes the state as written; content naming an unknown operator or tag fails at load.

## Spec

- **Firm:** operators and goals are content, so a new job is content. Code names no operator or tag.
- **Firm:** the state is symbolic and small, so search is cheap. A fact about a thing carries its UID only inside a step's `Target`, never inside the state.
- **Firm:** the same operators serve GOAP and HTN, so a recipe step is an operator.
- **Suggestion:** state is a sorted slice of facts, comparable and hashable for caching.

## Acceptance criteria

- [x] The tests above pass.
- [x] `go tool task check` passes.

## Out of scope

- Searching: M6-15. Recipes expanding: M6-16.
- Executing a plan: M6-11.

## Notes

- Fact tags are existing content IDs, so no new vocabulary: `has` takes an `item.` ID, `near` and `known` an `affordance.` ID, `drive_met` a `drive.` ID. The loader checks each against its collection. Matching by item tag instead is filed as the question M6-10.F1.
- A condition is `{fact, tag, n, not}`. As a precondition `has` holds at `n` or more; as an effect `has` adds `n` (negative to take). `not` negates a precondition and makes an effect remove the fact. Neither form is in the spec; they are the smallest that let an operator undo a fact (leaving a well stops `near`).
- Two new numbers in `agents`, first values for tuning: `near_radius: 1` (how close a believed landmark must be for `near`) and `met: "level < 1"` (a formula over `DriveEnv`, when a drive counts as met). `met` is required once any operator or goal names `drive_met`.
- `does` is `{action, target}`, where `target` is a picker name. Pickers come with M6-12 and M6-13, so the loader only checks the name isn't empty; those tasks, which register the pickers, should check content's names against them when the game starts.
- `Goal` has `drive` and `want` only; M6-14 adds `when` and `weight`. Recipes are checked for references and that each step names exactly one operator or recipe; cycle checks are M6-16's.
- Operators, goals and recipes are collections for ID prefixes (`operator.`, `goal.`, `recipe.`) but not `Collection` constants, as with geographies: they carry no record data.
- `Plan`, `Step` and `Target` live in `agents/plan.go` as the spec says. M6-11's `comp.Mind` holds a plan, and `comp` cannot import `agents`, so M6-11 will likely move these types into `comp`.
- `State` is a sorted `[]Fact`; a slice is not comparable, so `State.Key()` gives a string for cache keys.
- Docs: `docs/content-guide.md` (ID table, the `agents` numbers, a new "Operators, goals and recipes" section, and its old "Recipes" section renamed "How-tos" so the word means only HTN recipes) and `docs/glossary.md` (recipe, state).
- Follow-ups: M6-10.F1 (question: `has` by item ID or by item tag).
