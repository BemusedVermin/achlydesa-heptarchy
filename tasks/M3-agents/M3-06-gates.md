---
id: M3-06
title: Gates between maps
status: todo
milestone: M3
kind: feature
depends_on: []
discovered_by: null
size: M
---

# M3-06 · Gates between maps

## Goal

Maps connect: a gate cell leads to a mark on another map. The party walks through as a group, followers included, and any body can be moved through a gate, so towns and regions of the deep can be separate maps.

## Context

- The agents model in [README.md](README.md), "Regions, towns and gates".
- `docs/go-technical-design.md` → "Designing for pivots" → "Maps" (towns and set pieces as their own maps) and "Level of detail" (a region is a map); the `explore` row ("gates between maps later").
- `internal/content/types.go` (`MapDef`, marks), `internal/core/explore` (`Start`, `spawnParty`, follow, `wake.go`), `internal/core/hexmap/maps.go`.

## Deliverables

- `internal/content/types.go`: `MapDef.Gates`; validation; schemas regenerated.
- `internal/core/hexmap`: a gate table per map (foundation, so `agents` can read it).
- `internal/core/explore/gates.go` (new): the party crossing; `internal/core/api/explore.go`: `PartyCrossed`.
- `content/data/text/explore.cue`: the message for crossing.
- `internal/core/explore/gates_test.go`; two small test maps under `content/data/maps/` joined by a gate.

## Spec

**Firm:**

```go
// content
type GateDef struct { To api.ContentID; Mark string } // map ID, and a mark name on it
// MapDef.Gates maps a mark character to a GateDef: each such character in the layers is a gate.

// hexmap
type Gate struct { At api.Cell; To api.MapID; ToCell api.Cell }
func (m *Map) GateAt(c api.Cell) (Gate, bool)
func (m *Map) Gates() []Gate // sorted by cell

// api
type PartyCrossed struct { Party api.UID; From, To api.MapID }
```

- When the party's leader ends a step on a gate, every living member is moved to the destination mark (the leader on it, the others on the nearest free cells, as `spawnParty` places them), their pending actions are canceled, and `PartyCrossed` is emitted. Each map's seen cells are kept separately, as now.
- Gates go both ways only if both maps declare them. A gate whose map or mark doesn't exist fails at load.
- `explore.Wake` and `SetRestPlace` keep working across maps: a rest place on another map puts the party there.
- A body that is not in the party crossing a gate is `agents`' job (M3-11's `leave`); this task only exposes the table.

## Acceptance criteria

- [ ] A test walks the leader onto a gate; the whole party is on the other map, at the mark, and the plan view shows that map.
- [ ] Walking back through the return gate puts the party back, and the cells it saw on the first map are still remembered.
- [ ] A party wiped on map B whose rest place is on map A wakes on map A.
- [ ] A map naming an unknown gate destination fails `go tool task content`.
- [ ] `go tool task check` passes.

## Out of scope

- Joining the existing test maps into something playable (M3-07).
- Regions and promotion on crossing (M3-08, M3-09).

## Notes


