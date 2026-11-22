---
id: M2-01
title: "knowledge: the fact store"
status: done
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

- [x] Learning the same (Subject, Claim, Source) twice keeps one fact with the newer value.
- [x] `Best` prefers confidence, then the newer tick, then source ID; a test covers each tie-break.
- [x] A save and load keeps every fact (round-trip test through `game.Save`).
- [x] `go tool task check` passes.

## Out of scope

- The journal screen and investigation (M4); rumor drift (M6).

## Notes

- `Fact` lives in `internal/core/api/knowledge.go` as `api.Fact`, not in the knowledge package: `comp.Knowledge` needs it and `comp` cannot import `knowledge`. `Learn`, `About` and `Best` keep the spec's signatures (with `api.Fact`).
- A `Loggable` cannot be loggable only sometimes, so `Learn` publishes `api.FactLearned` (not loggable, always) and, when the knower is the controlled party or one of its members, `api.FactNoticed` (loggable). Its text ID is `text.knowledge.<claim without "claim.">`, via `api.KnowledgeTextID`; the value is passed raw as `{{.value}}`.
- Content: `claims` and `sources` maps in `content/data/knowledge.cue` (a new file, so concurrent tasks do not collide in `text.cue`), with `claim.material`, `claim.name`, `source.sweep`, `source.rumor`. Validation checks ID shape, name text, and that each claim has its `text.knowledge.*` text. `Registry.Claim` and `Registry.Source` look them up. `Learn` does not check IDs against content at run time.
- Files outside the list: `internal/content/{types,load,registry,load_test}.go`, generated `*_go_gen.cue`, `content/data/knowledge.cue`.
- Tie-break test for source ID is deterministic because `Learn` stamps `w.Now`, which the test holds fixed.
- Follow-up filed: M2-01.F1 (render fact values as text in the log).
