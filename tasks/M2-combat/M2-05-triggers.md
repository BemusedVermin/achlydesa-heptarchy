---
id: M2-05
title: "Triggers: content reacting to events"
status: todo
milestone: M2
kind: seam
depends_on: [M2-04]
discovered_by: null
size: S
---

# M2-05 · Triggers: content reacting to events

## Goal

Content can say "when this happens to me, if this holds, do these effects". Graft autonomy, crawling parts, Burn hardening clay and exotic recruit abilities all become data on top of one small runner.

## Context

- Tech design: "Designing for pivots" (Triggers).
- Game design: "Combat → Types" (the catches), "Combat → Grafts", "Party → Exotic recruits".
- M2-04 `effects`.

## Deliverables

- `internal/content/types.go`: `Trigger`; `Template.Triggers`, `Material.Triggers`, `Status.Triggers`, `Item.Triggers`; regenerate the schema.
- `internal/core/api/api.go`: the `Subjected` interface.
- `internal/core/comp/comp.go`: `Triggers`.
- `internal/core/effects/triggers.go`, `triggers_test.go`

## Spec

```go
type Trigger struct {
	On   string              `json:"on"`   // an event kind, e.g. "combat.part_severed"
	If   string              `json:"if"`   // Expr bool over EffectEnv; empty means always
	Do   []EffectRef         `json:"do"`
	Role string              `json:"role"` // which subject of the event this entity must be: "source", "target", "part", or "any"
}

// Subjected events name the entities they're about, by role.
type Subjected interface{ Subjects() map[string]api.UID }
```

- The runner subscribes once per event kind that any content trigger uses.
- On an event, it collects the entities carrying a matching trigger:
  - the entity named in the trigger's role;
  - for material triggers, the parts made of that material;
  - for status triggers, the bodies carrying that status;
  - for item triggers, the holders of that item.
- It sorts the matches by UID, builds an `effects.Context` from the event's subjects, checks `If` and applies `Do`.
- Spawning adds `comp.Triggers{IDs}` from the template, so the runner finds template triggers without reading content per event.
- Make every combat event from M2-04 implement `Subjected`.
- A trigger whose `On` isn't a registered event kind fails `game.New`.
- Cascades are bounded by the bus's `MaxCascade`. Add a test that a trigger re-firing itself errors rather than hangs.

## Acceptance criteria

- [ ] A template trigger on `combat.part_severed` with role `target` fires only for that body's severed parts.
- [ ] A material trigger (`material.clay`, on being hit by `type.burn`, raise max integrity) fires for clay parts only. This is the Burn-hardens-clay catch as data.
- [ ] Two entities with triggers on the same event fire in UID order.
- [ ] A self-refiring trigger produces an error, not a hang.
- [ ] An unknown event kind in content fails `game.New`.
- [ ] `go tool task check` passes.

## Out of scope

- The graft and crawling content itself (M2-09).

## Notes
