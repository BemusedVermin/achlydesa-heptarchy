---
id: M2-02
title: Bodies and parts as entities, with turn budgets and speed
status: todo
milestone: M2
kind: feature
depends_on: [M1-11, M1-13]
discovered_by: null
size: M
---

# M2-02 · Bodies and parts as entities, with turn budgets and speed

## Goal

Bodies are made of part entities joined by a relation, built from content body plans, so severing, grafting, reattaching and a part fighting on its own are all just relation changes. Arms set how many actions a body gets each turn, and legs set how often its turns come.

## Context

- Game design: "Combat → Bodies and turns", "Party", "Living deep → Three tiers of people" (plot armor).
- Tech design: "Designing for pivots" (Parts as entities, Entity templates, The party).
- This milestone's README, "The combat model" (Bodies, Grants, Death).
- Ark relations: the user guide's relations chapter and `go doc github.com/mlange-42/ark/ecs RelationMarker`.

## Deliverables

- `internal/content/types.go`: `PartTemplate`, `BodyPlan`, `Template.Body`, `Template.Protected`; regenerate the schema.
- `content/data/bodies.cue`: `body.human` and one creature plan; `content/data/templates.cue`: the two protagonists' templates get human bodies.
- `internal/core/comp/body.go`: the components and relations.
- `internal/core/sim/` or a new foundation file `internal/core/comp/bodyquery.go`: body queries (see Spec for where).
- Spawners `body` and `protected`, the `turns` budget and speed hooks, with tests.

## Spec

**Content:**

```go
type PartTemplate struct {
	Name      api.ContentID `json:"name"`      // text ID
	Slot      string        `json:"slot"`      // "head", "arm", "legs", ...: free text, for display and grafting
	Material  api.ContentID `json:"material"`
	Integrity int           `json:"integrity"`
	Grants    []Grant       `json:"grants"`
	Vital     bool          `json:"vital,omitempty"`
}
type Grant struct {
	Kind  string `json:"kind"`            // "action" | "targeting" | "speed"
	Value int    `json:"value,omitempty"` // speed amount; ignored for the others
}
type BodyPlan struct {
	Parts []PlanPart `json:"parts"`
}
type PlanPart struct {
	Part     api.ContentID `json:"part"`               // a PartTemplate ID
	Material api.ContentID `json:"material,omitempty"` // override, e.g. a clay arm
}
```

`Bundle` gains `Parts` (`part.`) and `Bodies` (`body.`). `Template` gains `Body *api.ContentID` (a plan), `Protected bool` and `Name *api.ContentID`. Validation covers every reference and the grant kinds.

**Components:**

```go
type Body struct{ Plan api.ContentID }
type Part struct {
	Template  api.ContentID
	Slot      string
	Material  api.ContentID
	Integrity, MaxIntegrity int
	Grants    []content.Grant // copied from content at spawn
	Vital     bool
}
type PartOf struct{ ecs.RelationMarker }   // part -> body
type Protected struct{}
type Dead struct{ Tick int64 }
type Downed struct{}
```

Use Ark's actual relation API; the shape above is the intent. Parts get `Identity` (their own UIDs) so commands, facts and events can name them.

**Queries** (foundation, so combat, agents and the frontend views can share them):

```go
func Parts(w *sim.World, body ecs.Entity) []ecs.Entity // attached parts, sorted by UID
func Functioning(w *sim.World, part ecs.Entity) bool    // attached and integrity > 0
func Grants(w *sim.World, body ecs.Entity, kind string) int // speed sums values; action and targeting count parts
func Members(w *sim.World, party ecs.Entity) []ecs.Entity  // sorted by UID
```

Put these in the new package `internal/core/anatomy`, which `AGENTS.md` and the tech design already list as foundation.

**Spawners:** `body` spawns one part entity per plan entry with `PartOf` the new body; parts have no `Pos` of their own while attached. `protected` adds `Protected` when the template sets it. `MemberOf` and party members already exist from M1-07.

**Turn hooks:** register `turns.SetBudget` to give one step and one action per functioning action part, and `turns.SetSpeed` to sum the speed grants of functioning parts. A body with no legs has speed 0 and slow turns; it can still crawl one step. M2-04 extends the budget for locked parts and granted or drained actions.

## Acceptance criteria

- [ ] Spawning `template.party.main` gives each of the two members five human parts (head, two arms, legs, torso as vital), all with UIDs.
- [ ] A human's budget is one step and two actions; with an arm's `PartOf` removed, one action on its next turn; with its legs removed, its turns come further apart (`turn_ticks` with speed 0).
- [ ] `Grants(member, "action")` is 2 and `Grants(member, "targeting")` is 1 for a human.
- [ ] A material override in a plan changes that part's material.
- [ ] Removing a part's `PartOf` relation drops it from `Parts` and `Grants`.
- [ ] Spawning is deterministic: UIDs and part order match across two worlds.
- [ ] Save and load keeps bodies, parts and relations (round trip through `game.Save`).
- [ ] `go tool task check` passes.

## Out of scope

- Damage, severing and death handling (M2-04, M2-07).
- Recruits joining the party (M3).

## Notes
