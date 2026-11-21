---
id: M2-01
title: "knowledge: the fact store"
status: todo
milestone: M2
kind: seam
depends_on: [M1-11, M1-13]
discovered_by: null
size: S
---

# M2-01 · knowledge: the fact store

## Goal

One store of facts per knower, which Sweep writes to in M2 and which the piece, rumors, investigation and dreams write to later. What the party knows about an enemy's parts is a query, not a field on the enemy.

## Context

- Tech design: "Designing for pivots" (Knowledge), "Core packages" (`knowledge`).
- Game design: "Core loop and time management → Investigation", "Combat → Forms" (Sweep reveals what each part is made of).

## Deliverables

- `internal/core/comp/comp.go`: add `Knowledge`.
- `internal/core/api/knowledge.go`: `FactLearned`.
- `internal/core/knowledge/knowledge.go`, `knowledge_test.go`

## Spec

```go
type Fact struct {
	Subject    api.UID       `json:"subject"`    // who or what the fact is about
	Claim      api.ContentID `json:"claim"`      // "claim.material", "claim.name", ...
	Value      string        `json:"value"`      // e.g. "material.clay"
	Source     api.ContentID `json:"source"`     // "source.sweep", "source.rumor", ...
	Confidence float64       `json:"confidence"` // 0..1
	Tick       int64         `json:"tick"`
}

// comp.Knowledge is the knower's facts, kept sorted by (Subject, Claim, Source).
type Knowledge struct{ Facts []Fact `json:"facts"` }

func Learn(w *sim.World, knower ecs.Entity, f Fact) // upsert on (Subject, Claim, Source); publishes FactLearned
func About(w *sim.World, knower ecs.Entity, subject api.UID) []Fact // sorted
func Best(w *sim.World, knower ecs.Entity, subject api.UID, claim api.ContentID) (Fact, bool) // highest confidence, then newest, then Source
```

- `Learn` adds `Knowledge` to the knower if missing and sets `Tick` to `w.Now`.
- `FactLearned` is loggable only when the knower is the controlled party; its text comes from `text.knowledge.<claim>`.
- Claims and sources are content IDs. Add `claims` and `sources` maps to content with names, and validate references.

## Acceptance criteria

- [ ] Learning the same (Subject, Claim, Source) twice keeps one fact with the newer value.
- [ ] `Best` prefers confidence, then the newer tick, then source ID; a test covers each tie-break.
- [ ] A save and load keeps every fact (round-trip test through `game.Save`).
- [ ] `go tool task check` passes.

## Out of scope

- The journal screen and investigation (M4); rumor drift (M6).

## Notes
