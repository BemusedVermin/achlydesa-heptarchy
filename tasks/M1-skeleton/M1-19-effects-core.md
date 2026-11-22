---
id: M1-19
title: "effects: the effect core, selectors and timing primitives"
status: todo
milestone: M1
kind: seam
depends_on: [M1-06, M1-18]
discovered_by: null
size: M
---

# M1-19 · effects: the effect core, selectors and timing primitives

## Goal

The vocabulary every action, trigger and item composes: a registry of named effect primitives with Expr parameters, a context that assumes nothing about combat, and selectors that let one effect hit one limb, every limb with a tag, or whole bodies. The first primitives work only on time (lock, interrupt, delay, ready), so "stop everything this body is doing" is already expressible in content before any combat exists. M2-04 adds damage and the rest.

## Context

- Game design: "Combat → Timing", "Technical plan → The engine/data boundary".
- Tech design: "Designing for pivots" (Effects anywhere, Triggers).
- M1-17 channels and `Extend`, M1-18 pending actions and `Cancel`.

## Deliverables

- `internal/content/types.go`: `EffectRef`; `internal/content/env.go`: `EffectEnv`; regenerate the schema.
- `internal/core/effects/effects.go`: context, registry, parameters, apply.
- `internal/core/effects/selectors.go`: body and channel selectors.
- `internal/core/effects/timing.go`: the four timing primitives.
- Tests for each.

## Spec

**Content:**

```go
// EffectRef is one primitive invocation. Params are Expr sources evaluated with EffectEnv.
type EffectRef struct {
	Do     string            `json:"do"`
	Params map[string]string `json:"params"`
}

// EffectEnv is everything an effect formula can see. Later tasks add fields
// deliberately (M2-04 adds weak, immune, integrity, height, statuses...).
type EffectEnv struct {
	Now      int64   `expr:"now"`
	Roll     float64 `expr:"roll"`      // a fresh draw in [0,1) from the "effects" stream
	Distance int     `expr:"distance"`  // hexes from source to target, or 0
	WindingUp bool   `expr:"winding_up"` // the target body has a pending action
}
```

Formulas compile at content load against `EffectEnv`. The loader can't know primitive names, so `game.New` checks every `EffectRef` in content against the registry after `Install`, failing with a list of unknown primitives and unknown or missing parameters.

**Context and registry:**

```go
type Context struct {
	World      *sim.World
	Source     ecs.Entity         // the acting body, or zero
	Used       []timeline.Channel // the channels the action occupies, or none
	Target     ecs.Entity         // the target body, or zero
	TargetPart ecs.Entity         // the target part, or zero
	Dir        *api.Dir           // for direction-targeted actions
	Choices    map[string]string  // the action's chosen values, e.g. "type": "type.soak" (M1-20)
}

type Primitive struct {
	Name   string
	Params []string
	Apply  func(ctx *Context, p Params) error
	// Check, when set, says whether the effect could apply right now, and why not.
	// Actions call it before starting and again when landing (M1-20).
	Check func(ctx *Context, p Params) (ok bool, reason api.Text)
}
type Params interface {
	Int(name string) (int, error)
	Float(name string) (float64, error)
	Bool(name string) (bool, error)
	String(name string) (string, error)
	Bodies(name string) ([]ecs.Entity, error)          // a body selector, below
	Channels(name string) ([]timeline.Channel, error)  // a channel selector, below
}
func Register(w *sim.World, p Primitive) error
func Apply(ctx *Context, refs []content.EffectRef) error // in order; stops at the first error
func Check(ctx *Context, refs []content.EffectRef) (bool, api.Text)
```

**Selectors** are string parameters with a small fixed grammar, resolved against the context and returned sorted by UID:

- **Bodies:** `self`, `target`, and `allies`, `enemies`, `others` or `everyone`, all meaning "in sight of the source". Sight uses `hexmap.Visible`. Allies and enemies use a hostility hook that defaults to "nobody is hostile" until M2-07 sets it.
- **Channels:** `used` (the action's own channels), `target_part`, or `<bodies>:<tag>`. For example, `target:all` is every channel of the target, `target:arm` every channel of the target tagged `arm`, and `self:body` the source's body channel.

Keep the grammar in one parsing function, with errors at content load for anything malformed.

**Timing primitives:**

| Name | Params | Does |
| --- | --- | --- |
| `lock` | `channels`, `ticks` | Keeps each selected channel busy for `ticks` more and cancels any pending action it's winding up. Pin is `lock(channels: "target_part", ticks: 8)`; "stop all actions" is `lock(channels: "target:all", ticks: …)` |
| `interrupt` | `bodies` or `channels`, `when` | If `when` holds, cancels the selected pending actions without locking anything (a stagger) |
| `delay` | `channels`, `ticks` | Pushes back each selected channel's busy time (Stand's freeze is `delay(channels: "others:all", ticks: 8)`) |
| `ready` | `channels`, `count` | Frees up to `count` of the selected busy channels at once, in UID order (Stand's acting again; Give's ally) |

Each publishes an event (`ChannelsLocked`, `ActionsInterrupted`, `ChannelsDelayed`, `ChannelsReadied`) that names the affected UIDs, for logs and triggers.

## Acceptance criteria

- [ ] Each selector form resolves correctly in a table test, including `target:all` and a tag that matches nothing (an empty list, not an error).
- [ ] `lock(channels: "target:all")` on a body with two pending swings cancels both and delays its next decision by the lock.
- [ ] `interrupt` with `when: winding_up` does nothing to a body that's idle.
- [ ] `delay` on `others:all` pushes back every visible body's channels but not the source's.
- [ ] `ready` with `count: 1` frees only the lowest-UID busy channel.
- [ ] Content naming an unregistered primitive, a missing parameter or a malformed selector fails `game.New` with a clear list.
- [ ] `go tool task check` passes.

## Out of scope

- Damage, severing, statuses, items and knowledge (M2-04). Actions that use these (M1-20). Triggers (M2-05).

## Notes
