---
id: M1-20
title: "actions: content-defined actions"
status: done
milestone: M1
kind: seam
depends_on: [M1-07, M1-14, M1-19]
discovered_by: null
size: L
---

# M1-20 · actions: content-defined actions

## Goal

Everything a body does over time is an action defined in content. That covers a step, a two-handed swing, a lunge using legs and an arm, and a graft that takes both hands. An action says which channels it occupies (by tag and count), how long it winds up and recovers (as formulas), what it targets and how far, and which effects apply when it starts and when it lands. Code supplies only the machinery and the primitives, so combat rules can be tried and changed in CUE.

## Context

- Game design: "Combat → Fighting on the map", "Combat → Timing".
- Tech design: "Designing for pivots" (Choices come from the core, Effects anywhere).
- M1-14 sight, M1-17 channels and the usable hook, M1-18 pending actions, M1-19 effects.

## Deliverables

- `internal/content/types.go`: `ActionDef`, `Requirement`; `Template.Actions`; `ActionEnv`; regenerate the schema.
- `internal/core/api/actions.go`: `Do`, `ActionStarted`, `ActionMissed`.
- `internal/core/actions/` (a foundation package): `actions.go`, `choices.go`, `options.go`, `install.go`, tests.
- `AGENTS.md` and `docs/go-technical-design.md`: list `effects` (if not already) and `actions` as foundation packages.

## Spec

**Content:**

```go
type ActionDef struct {
	Name     api.ContentID `json:"name"`    // text ID
	Key      string        `json:"key"`     // short option-ID segment, e.g. "step", "strike"
	Uses     []Requirement `json:"uses"`    // channels it occupies; every requirement must be met by free channels
	Target   string        `json:"target"`  // "none", "self", "direction", "body", "part", "ally", "cell"
	Range    int           `json:"range"`   // hexes, for body/part/ally/cell targets
	Sight    bool          `json:"sight"`   // the target must be visible from the source
	Choices  []string      `json:"choices"` // extra choices the player makes, by provider name, e.g. "type" (M2)
	Requires string        `json:"requires,omitempty"` // Expr bool over ActionEnv
	Windup   string        `json:"windup"`   // Expr ticks before it lands; 0 lands at once
	Recovery string        `json:"recovery"` // Expr ticks the used channels stay busy after landing
	OnStart  []EffectRef   `json:"on_start"`
	OnLand   []EffectRef   `json:"on_land"`
	Connects string        `json:"connects,omitempty"` // Expr bool at landing; empty means "the target is still in range (and sight, if required)"
	Tags     []string      `json:"tags"`     // free tags for rules and content: "attack", "movement", ...
}
type Requirement struct {
	Tag   string `json:"tag"`   // a channel tag: "body", "arm", "legs", "grant:action", ...
	Count int    `json:"count"` // how many such channels; 0 means every one the body has (at least one)
}
// ActionEnv is what an action's formulas see. Grow it deliberately.
type ActionEnv struct {
	Speed    int    `expr:"speed"`
	Distance int    `expr:"distance"`
	Used     int    `expr:"used"`    // how many channels it occupies
	Choice   map[string]string `expr:"choice"`
}
```

Templates list the actions a body knows: `Template.Actions []api.ContentID`, with `action.` IDs. Party members know `action.step` from now on.

**Choices.** Choice providers are registered in code by name and offer values for a given actor:

- `direction` is built in: the six directions.
- `type` arrives in M2-06: the types a body can load.

When more channels match a requirement than it needs, which ones to use is a choice too: one option per combination, capped, in UID order.

**The machinery** (signatures are suggestions; behavior is firm):

```go
type Intent struct {
	Actor      ecs.Entity
	Action     api.ContentID
	Channels   []timeline.Channel // chosen; empty lets Validate pick the lowest-UID matches
	Target     ecs.Entity
	TargetPart ecs.Entity
	Dir        *api.Dir
	Choices    map[string]string
}
func Validate(w *sim.World, in Intent) (ok bool, reason api.Text)
func Start(w *sim.World, in Intent) error   // validates, applies OnStart, then lands at once or schedules the landing
func Startable(w *sim.World, actor ecs.Entity) []Intent // every valid intent right now, in option order
```

- **Validate:**
  - the actor knows the action;
  - its requirements can be met from free channels;
  - the target fits the target kind, and is within range and sight;
  - `Requires` holds;
  - every `Check` of its `OnStart` and, for windup 0, `OnLand` effects passes. This is how a step into packed clay comes out disabled with "Packed clay. No way through."
- **Start:**
  1. Applies `OnStart`.
  2. Occupies every used channel for windup plus recovery.
  3. With windup 0, lands now. Otherwise it schedules a pending action of kind `actions.land` at now plus windup, with a component holding the intent's IDs, and publishes `ActionStarted`.
- **Landing** (the `actions.land` resolver):
  - It re-checks that the actor and its used channels still exist.
  - It evaluates `Connects` (by default, still in range and sight), then the `OnLand` checks.
  - If all of that holds, it applies `OnLand`; if not, it publishes `ActionMissed` (loggable, e.g. "Clay-thing's Strike hits empty air").
  - Its label for the timeline view comes from the action's name and target.
- **Hooks for other packages,** so combat rules stay in combat:
  - `AddValidator(w, fn func(w, Intent) (bool, api.Text))` adds a check to `Validate` (combat uses it for immune materials and type conditions).
  - `AddModifier(w, fn func(w, Intent) Modifier)` lets an intent's choices add wind-up ticks, recovery ticks and extra `OnStart` or `OnLand` effects (combat uses it so a type changes any form it's loaded into).
  - Both run in registration order.
- **Usable hook:** a channel is usable when some startable intent uses it. So arms don't make a body stop and decide while nothing is in reach.
- **Command:** `Do{ Actor api.UID; Action api.ContentID; Channels []api.UID; Target, TargetPart api.UID; Dir *api.Dir; Choices map[string]string }`, kind `actions.do`. Rejected unless the actor is the deciding player body and `Validate` passes.
- **Options:** one per startable intent.
  - The ID is the action's key, then each choice, the channels and the target, slash-separated, for example `step/nw` or `strike/type.cut/<target uid>/<part uid>/<arm uid>`.
  - The label names the action and its choices, with "lands in N, busy M".
  - Order: action key, then choices, then UIDs, through one `orderOptions` function.
  - Cap the list at 500 and file a follow-up if a sample exceeds it.

**Sample content in this task:** waiting stays M1-17's `wait` options, not an action. Define `action.step` here with:

- `uses: [{tag: "body", count: 1}]`, `target: "direction"`, windup 0, recovery `step_ticks`;
- one `OnLand` effect, `step`, registered by M1-10.

Until M1-10 lands, test with a test primitive.

## Acceptance criteria

- [x] A test action needing two `arm` channels is startable only when both are free, occupies both, and offers one option per pair when three arms are free.
- [x] An action with windup 3 lands 3 ticks later; moving its target out of range before then makes it publish `ActionMissed`.
- [x] Locking one of its used channels before landing cancels it (through M1-19's `lock`).
- [x] A `Check` that fails makes the option disabled with that check's reason.
- [x] The usable hook keeps a body with a free but useless channel from stopping to decide.
- [x] `actions.do` from an option always validates; a hand-built `Do` with a busy channel is rejected with a reason.
- [x] Replaying 200 random options gives the same hash as playing them.
- [x] `go tool task check` passes.

## Out of scope

- The step primitive and moving through the world (M1-10). Forms, types and combat rules (M2-06). AI that chooses actions (M2-08, M3).

## Notes

This is L. If it runs long, split options and the usable hook into a follow-up and land `Validate`, `Start` and landing first.

**Implementer's notes**

- **Placeholder `step`.** The sample `action.step` names an effect `step`, and `game.New` rejects content naming unregistered primitives, so `actions.Install` registers a `step` that moves nobody (`install.go`, `registerPlaceholderStep`). A step still takes its time and its options appear. M1-10 should delete the placeholder when it registers the real one (the bump text `text.explore.bump` is its `Check`).
- **`ActionEnv.StepTicks`.** The spec's `recovery: step_ticks` needs a `step_ticks` name in `ActionEnv`, so it gained `StepTicks` (the existing `timeline.step_ticks` formula at the actor's speed).
- **Content field tags.** `choices`, `on_start`, `on_land`, `tags`, `range` and `sight` are `omitempty`, so a simple action doesn't have to write them out in CUE. Content checks add: `action.` IDs, a text `name`, a unique lower-case `key`, a known `target`, at least one `uses` requirement with a valid tag, `windup` and `recovery` present, and every formula compiling against `ActionEnv`. `game.New` also calls `actions.ValidateContent` (every `choices` entry needs a registered provider). Template `actions` must name real actions.
- **One pending per action.** A multi-channel action is one pending entity whose `Part` is its first channel; the new `comp.Landing` lists all of them. `effects.windingUp` now also matches a `Landing` that lists the channel, so `lock` and `interrupt` on any used channel cancel the action.
- **Usable hook** ignores busyness: a channel is usable when some action could use it were it free (otherwise a busy arm would look useless and the body would wake too early). Landing windup-0 actions publishes no `ActionStarted`, as specced; a miss at landing publishes `ActionMissed` with a `Line` for the log and a `Reason`.
- **Option IDs** include the channel UIDs only when the action has more than one way to pick channels (`step/nw`, but `grab/<arm>/<arm>` with three arms). `Do` always carries the chosen channels.
- **Cell targets.** `target: "cell"` loads but nothing is offered: no cell travels in the intent. Filed as M1-20.F2.
- **Limits.** At most 32 channel sets per action (`MaxChannelSets`) and 500 options (`MaxOptions`). The sample stays far below both, so no follow-up for the cap.
- **Files touched outside the list:**
  - `content`: `Load` split into `LoadBundle` and `New` (tests adjust a bundle), `actions.go`, `bundle.go`, `env.go`.
  - `comp`: `Landing`.
  - `hexmap`: `Distance`.
  - `timeline`: `Speed` and `CurrentPlayer`.
  - `effects`: `Hostile`, `hexmap.Distance` instead of its own copy, and `windingUp`.
  - `game`: install and `ValidateContent`.
  - Sample content: `actions.cue`, `templates.cue`, `text.cue`.
  - Regenerated schemas.
  - AGENTS.md and the tech design already list `effects` and `actions` as foundation packages, so they are unchanged.
- **Follow-ups filed:** M1-20.F1 (effect checks draw from the RNG while building views), M1-20.F2 (cell targets). `TestGameNewRejectsActionsContentCannotRun` now covers the unknown-primitive path end to end, so the owner may fold M1-19.F1 into this; its missing-parameter and malformed-selector cases are still open.
