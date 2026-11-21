---
id: M2-06
title: Forms, reach and move resolution
status: todo
milestone: M2
kind: feature
depends_on: [M2-04]
discovered_by: null
size: M
---

# M2-06 · Forms, reach and move resolution

## Goal

A move is resolved by one function: check that the form, type and target are valid and in reach, pay the type's cost, and apply the form's effects. All seven forms are content, including how far each reaches, so positioning on the grid decides what's possible.

## Context

- Game design: "Combat → Moves = form × type", "Combat → Forms".
- This milestone's README, "The combat model" (Moves).
- M2-03 types and the pack, M2-04 effects.

## Deliverables

- `internal/content/types.go`: `Form`; `Template.Combatant`; regenerate the schema.
- `content/data/forms.cue`: all seven forms.
- `internal/core/comp/comp.go`: `Combatant`, `Allegiance`.
- `internal/core/combat/move.go`, `move_test.go`

## Spec

**Content:**

```go
type Form struct {
	Name     api.ContentID `json:"name"`
	Target   string        `json:"target"`   // "part", "body", "self", "ally", "none"
	NeedsType bool         `json:"needs_type"`
	Range    int           `json:"range"`    // hexes; 1 is adjacent; 0 is self only
	Requires string        `json:"requires,omitempty"` // a component the source must have, e.g. "graftable"
	Effects  []EffectRef   `json:"effects"`
}

type CombatantDef struct {
	Faction   api.ContentID   `json:"faction"`
	Forms     []api.ContentID `json:"forms"`
	Types     []api.ContentID `json:"types"` // innate types this combatant can load
	Graftable bool            `json:"graftable"`
}
```

`Template` gains `Combatant *CombatantDef`; its spawner adds `comp.Combatant` and `comp.Allegiance`, plus `comp.Graftable` when set.

**Forms, provisionally** (the numbers are placeholders for M7):

| Form | Target | Effects |
| --- | --- | --- |
| Strike | part | `damage(amount: weak ? 999 : 4)` |
| Sweep | body | `damage(amount: weak ? 2 : 1)` on every part, then `reveal(claim: "claim.material")` |
| Sever | part | `damage(amount: 1)`, `sever(when: weak)` |
| Pin | part | `lock_part(turns: 1)` |
| Stand | self | `add_status(status: "status.standing", turns: 1, on: "self")`, `delay_turns(scope: "others", ticks: 10)`, `extra_turn(on: "self")` |
| Give | ally | `transfer_action` |
| Graft | ally | `attach_part(from: "pack")`; requires graftable |

Body-wide forms apply their effects once per attached part, in UID order. Strike's 999 stands in for "comes off in one blow"; file a `question` follow-up on whether Strike and Sever should differ on a weakness the way these placeholders do, and on what "everyone crawling" means for Stand.

**Resolution:**

```go
type Move struct {
	Source, SourcePart, Target, TargetPart ecs.Entity
	Form, Type api.ContentID
}
// Validate says whether a move is legal now, and why not.
func Validate(w *sim.World, m Move) (ok bool, reason api.Text)
// Resolve validates, takes the type's item if it consumes one, publishes MoveUsed, and applies the form's effects.
func Resolve(w *sim.World, m Move) error
```

Validation checks:

- the source can use the form and type (its own, or a consumable whose item is in its party's pack);
- the source has an action left this turn (`turns.Remaining`), and the source part is functioning, unlocked and grants an action;
- the target fits the form's target kind and is in reach: visible from the source, within the form's range in hexes, and within that many cells of depth (at least one);
- the target part isn't immune to the type;
- the type's condition holds.

Every form in the sample has range 1 except Stand (0). Ranged forms are content for later.

## Acceptance criteria

- [ ] Each form resolves on an adjacent test dummy with the expected events.
- [ ] A dummy two hexes away, or around a corner out of sight, is out of reach for every range-1 form.
- [ ] In a one-cell soop, the member behind the front one can't reach the creature beyond it.
- [ ] Strike with Soak on a clay part severs it and uses one lake water; with no lake water in the pack the move is invalid with a reason.
- [ ] Name on the wrong target is invalid; on its bound owner it's valid and uses up the name.
- [ ] Relic against old matter is invalid (immune).
- [ ] Graft from a non-graftable source is invalid.
- [ ] `go tool task check` passes.

## Out of scope

- Turn budgets (M1-15, M2-02), alert and death (M2-07), and commands, options and the creature controller (M2-08).

## Notes
