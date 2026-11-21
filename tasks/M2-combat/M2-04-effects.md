---
id: M2-04
title: "effects: context, registry and the primitives"
status: todo
milestone: M2
kind: seam
depends_on: [M2-01, M2-03]
discovered_by: null
size: L
---

# M2-04 · effects: context, registry and the primitives

## Goal

The vocabulary content composes: a registry of named effect primitives with typed parameters, applied through a context that doesn't assume a fight, so the same primitives serve moves, triggers, seat prices and items.

## Context

- Game design: "Combat → Forms" and "Technical plan → The engine/data boundary".
- Tech design: "Designing for pivots" (Effects anywhere), "Core packages" (`effects`).
- This milestone's README, "The combat model".

## Deliverables

- `internal/content/types.go`: `EffectRef`; `internal/content/env.go`: `EffectEnv`; regenerate the schema.
- `internal/core/effects/effects.go`: context, registry, parameter evaluation.
- `internal/core/effects/primitives.go`: the primitives below.
- `internal/core/effects/effects_test.go`: one table test per primitive.
- `internal/core/api/combat.go`: the part and status events used here.

## Spec

**Content:**

```go
// EffectRef is one primitive invocation. Params are Expr sources evaluated with EffectEnv.
type EffectRef struct {
	Do     string            `json:"do"`     // primitive name, e.g. "damage"
	Params map[string]string `json:"params"` // name -> Expr
}

// EffectEnv is everything an effect formula can see. Grow it deliberately.
type EffectEnv struct {
	Weak, Immune, Pinned bool
	TargetIntegrity, TargetMaxIntegrity int
	SourceActions int
	Height int // target depth minus source depth: positive when the source is above
	Statuses []string // the target body's status IDs
	Roll float64      // a fresh draw in [0,1) from the "effects" stream
}
```

Tag every `EffectEnv` field with a snake_case `expr` name, as `MoveEnv` does (`expr:"weak"`). Formulas compile at content load against `EffectEnv`. The loader can't know primitive names, so `game.New` checks every `EffectRef.Do` in content against the registry after `Install` and fails with the list of unknown names.

**Context and registry:**

```go
type Context struct {
	World      *sim.World
	Source     ecs.Entity // acting body, or zero
	SourcePart ecs.Entity // part used, or zero
	Target     ecs.Entity // target body, or zero
	TargetPart ecs.Entity // target part, or zero
	Type       api.ContentID // the move's type, or ""
}

type Primitive struct {
	Name   string
	Params []string // required parameter names; extra ones are a content error
	Apply  func(ctx *Context, p Params) error
}
type Params interface {
	Int(name string) (int, error)
	Float(name string) (float64, error)
	Bool(name string) (bool, error)
	String(name string) (string, error) // for IDs; the Expr must evaluate to a string
}

func Register(w *sim.World, p Primitive) error
func Apply(ctx *Context, refs []content.EffectRef) error // in order; stops at the first error
```

`Params` builds `EffectEnv` from the context once per `Apply` call. Weak and Immune compare the target part's material with `ctx.Type`.

**Primitives** (behavior firm, parameter names suggested):

| Name | Params | Does |
| --- | --- | --- |
| `damage` | `amount` | Lowers the target part's integrity; at zero it's severed (worn down). Publishes `PartDamaged` and maybe `PartSevered{Cause: "worn"}` |
| `sever` | `when` | If true, removes the target part's `PartOf`; the part stays as an entity at the body's position. Publishes `PartSevered{Cause: "cut"}` |
| `lock_part` | `turns` | Adds `comp.Locked{Turns}` to the target part (Pin): it gives no action on its body's next N turns. Publishes `PartLocked` |
| `add_status` | `status`, `turns`, `on` (`"self"` or `"target"`) | Adds a content status to a body for its next N turns. Publishes `StatusAdded` |
| `grant_actions` | `amount`, `on` | Adds actions to the body's current turn, or its next if it isn't acting |
| `drain_actions` | `amount`, `scope` (`"target"`, `"allies"`, `"others"`) | Takes actions away from the next turn of each body in scope |
| `delay_turns` | `ticks`, `scope` (`"others"`, `"allies"`, `"enemies"`) | Pushes back the next turn of every body in scope within sight (Stand's freeze) |
| `extra_turn` | `on` | The body takes another turn right after this one ends (Stand's acting again) |
| `transfer_action` | none | The source loses one action; the target ally gains one (Give) |
| `attach_part` | `from` (`"pack"` or `"severed"`), `item` | Attaches a carried piece, or a severed part at the target's position, to the target body. Publishes `PartAttached` |
| `reveal` | `claim` | Writes a fact about each part of the target body for the source's party (Sweep). Uses `knowledge.Learn` |
| `spawn_from_part` | `faction` | Turns a severed part into a one-part body with that allegiance (crawling parts) |

- Add content `statuses` (`status.`) with names; status effects live in formulas that read `Statuses`, such as Stand's double damage.
- Add `comp.Locked`, `comp.Statuses` and `comp.ActionMods{Extra, Drained int}`.
- Per-turn counters (`Locked`, statuses and `ActionMods`) tick down when their owner's turn ends, through a `turns.turn_ended` subscriber. Extend M2-02's budget hook so locked parts give no action and `ActionMods` add or drain actions.

## Acceptance criteria

- [ ] One table test per primitive, including a no-op case (a `sever` whose `when` is false).
- [ ] `damage` past zero severs and publishes both events in that order.
- [ ] `reveal` writes one fact per attached part, and `knowledge.Best` returns each part's material.
- [ ] Content naming an unregistered primitive, or passing a parameter the primitive doesn't take, fails `game.New` with a clear list.
- [ ] Every primitive behaves the same whether or not anything hostile is around: there is no fight state to depend on.
- [ ] `go tool task check` passes.

## Out of scope

- Forms, targeting and turn flow (M2-06, M2-07). Triggers (M2-05).

## Notes

This is the largest M2 task. If it runs past about 600 lines, split it: registry and the first four primitives here, and the rest as a follow-up.
