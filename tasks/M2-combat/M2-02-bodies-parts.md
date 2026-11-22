---
id: M2-02
title: Bodies and parts as entities, as channels
status: done
milestone: M2
kind: feature
depends_on: [M1-11, M1-13]
discovered_by: null
size: M
---

# M2-02 · Bodies and parts as entities, as channels

## Goal

Bodies are made of part entities joined by a relation, built from content body plans, so severing, grafting, reattaching and a part fighting on its own are all just relation changes. Each attached part becomes a timeline channel tagged by its slot and grants, so actions can ask for "one arm", "two arms" or "legs", and legs set how fast a body steps.

## Context

- Game design: "Combat → Bodies and timing", "Party", "Living deep → Three tiers of people" (plot armor).
- Tech design: "Designing for pivots" (Parts as entities, Entity templates, The party).
- This milestone's README, "The combat model" (Bodies, Grants, Death).
- Ark relations: the user guide's relations chapter and `go doc github.com/mlange-42/ark/ecs RelationMarker`.

## Deliverables

- `internal/content/types.go`: `PartTemplate`, `BodyPlan`, `Template.Body`, `Template.Protected`; regenerate the schema.
- `content/data/bodies.cue`: `body.human` and one creature plan; `content/data/templates.cue`: the two protagonists' templates get human bodies.
- `internal/core/comp/body.go`: the components and relations.
- `internal/core/sim/` or a new foundation file `internal/core/comp/bodyquery.go`: body queries (see Spec for where).
- Spawners `body` and `protected`, the timeline's channel and speed hooks, with tests.
- From M1-17.F1: `internal/core/timeline` gains an exported `Reschedule(w *sim.World, body ecs.Entity)` that recomputes the body's decision tick and requeues it (a no-op for a body with no `comp.Actor`). Call it when parts are lost or attached. Test: `SetUsable` starts rejecting the only busy channel, and `Reschedule` brings the decision forward.

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

**Timeline hooks:** register `timeline.SetChannels` to give each body its own channel (tagged `body`) plus one per attached part. A part's channel is tagged `part`, its slot (`arm`, `legs`, `head`, ...), and `grant:<kind>` for each grant. Register `timeline.SetSpeed` to sum the speed grants of functioning parts. A body with no legs has speed 0, so `step_ticks` makes its steps slow; it can still crawl.

## Acceptance criteria

- [x] Spawning `template.party.main` gives each of the two members five human parts (head, two arms, legs, torso as vital), all with UIDs.
- [x] A human has six channels: its body plus its five parts, each tagged with its slot and grants. Removing an arm's `PartOf` removes that channel; with its legs removed, a step keeps it busy longer (`step_ticks` with speed 0).
- [x] `Grants(member, "action")` is 2 and `Grants(member, "targeting")` is 1 for a human.
- [x] A material override in a plan changes that part's material.
- [x] Removing a part's `PartOf` relation drops it from `Parts` and `Grants`.
- [x] Spawning is deterministic: UIDs and part order match across two worlds.
- [x] Save and load keeps bodies, parts and relations (round trip through `game.Save`).
- [x] `go tool task check` passes.

## Out of scope

- Damage, severing and death handling (M2-04, M2-07).
- Recruits joining the party (M3).

## Notes

Deviations:

- `comp.Grant` lives in `comp`, not `content`. `content` imports `comp` (for `Appearance`), so `comp.Part.Grants` cannot name a `content` type without a cycle. `content.PartTemplate.Grants` is `[]comp.Grant`. The grant kind constants (`content.GrantAction`, `GrantTargeting`, `GrantSpeed`, `GrantKinds`) are in `content`.
- Body queries and install are in the new package `internal/core/anatomy` (as the Spec says), not `comp/bodyquery.go`. `anatomy.Install` runs in `game.installAll` after `timeline.Install`.
- `timeline.Reschedule` already existed in `timeline` (M1), so it was not added again; I only extended its doc comment. `Reschedule` is called when a body is spawned with parts. Losing or attaching parts later is for the tasks that do it (M2-04, M2-07, M2-09). The Reschedule test uses a hook that goes from "only the arm is usable" to "every channel is usable", because a body decides as soon as any usable channel is free, so rejecting the only busy channel never brings a decision forward.
- Added `sim.World.NewIdentified(template)`: part entities need an `Identity` with a UID that is in the UID index, and the index is private to `sim`.
- `Grants` and `speed` count only functioning parts (attached, integrity above 0). Channels exist for every attached part, whatever its integrity.
- An entity with no `comp.Body` keeps one body channel and the configured `default_speed`, so test fixtures and bodiless templates behave as before.
- Both protagonists are `protected: true`; the docs say plot armor is set per authored character, and this is easy to change in `templates.cue`.
- `TestASlowerBodysStepsTakeLonger` (explore) set `Timeline.DefaultSpeed`; bodies now take speed from their legs, so it edits `part.human_legs` instead.
- Text IDs added: `text.part.*` (part names).

Content: `bodies.cue` has `body.human` and `body.crawler` (clay core, one slow leg part, a maw), seven part templates, all with placeholder numbers.

Files touched outside the list: `internal/content/{load,registry}.go`, `content/data/text.cue`, `internal/core/game/install.go`, `internal/core/sim/spawn.go`, `internal/core/explore/explore_test.go`, the regenerated `*_go_gen.cue`.

Follow-ups filed: [M2-02.F1](../followups/M2-02.F1-despawn-body-parts.md) (despawning a body orphans its parts).
