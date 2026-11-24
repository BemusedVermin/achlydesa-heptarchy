---
id: M3-15
title: Hiring recruits
status: todo
milestone: M3
kind: feature
depends_on: [M3-12]
discovered_by: null
size: M
---

# M3-15 · Hiring recruits

## Goal

The party can hire a generic person in town. The recruit names a price in items they want, scaled by what they bring; paying it from the party's inventory adds them to the party, where they fight like any member.

## Context

- `docs/game-design.md` → "Party" → "Recruits" (price and "Paying them", the owner's decision), "The fixed roster".
- The agents model in [README.md](README.md), "Recruits".
- `internal/core/explore` (party members, `MemberOf`, follow), `internal/core/inventory`, `comp.Party.Slots`; M3-12's generics.

## Deliverables

- `internal/content/types.go`: `RecruitDef` on `Template`; schemas regenerated.
- `internal/core/agents/recruits.go`: the hire option, the command handler, joining the party.
- `internal/core/api`: `Hire` command, `Hired` event, the option's label with the price.
- `content/data/templates.cue`: generic guard and field hand become hirable; `content/data/text/agents.cue`.
- `internal/core/agents/recruits_test.go`.

## Spec

**Firm:**

```go
// content
type RecruitDef struct {
    Wants  []StackDef // item kinds and base counts
    Worth  string     // Expr over BodyEnv: what they bring; scales the price
    Upkeep []StackDef // per day; used by M3-16
}
// api
type Hire struct { Party, Recruit api.UID }
type Hired struct { Party, Recruit api.UID; Paid []api.ItemCount }
```

- The price is each want's count times `Worth`, rounded up. The hire option is offered when the leader is next to a hirable generic on a town map, no hostile is in sight, and the party has fewer members than `party.slots`. It is disabled, with the reason, when the inventory can't pay.
- Hiring takes the items, removes `Generic` (so the town's replacement timer starts), adds `MemberOf`, sets the party's faction and the `follow` controller, and puts the recruit at the end of the party order.
- The recruit keeps their own body, needs and template; graftable recruits keep the `graftable` tag.

## Acceptance criteria

- [ ] A test hires a guard: the inventory loses exactly the price, the guard is a party member, and the town's guard is replaced later.
- [ ] Exit criterion, first half: the hired recruit fights in the next encounter, with options in the action menu like any member.
- [ ] With a full party, or too few items, the hire option is disabled with its reason.
- [ ] In the playground's town, a recruit can be hired from the action menu.
- [ ] `go tool task check` passes.

## Out of scope

- Upkeep, discontent, desertion and flight (M3-16).
- Caste friction (a proposal in the design, not decided).
- A bench of members beyond the four fighting slots.

## Notes


