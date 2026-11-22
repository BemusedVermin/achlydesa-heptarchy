---
id: M1-07
title: Templates as component bags, and spawning
status: done
milestone: M1
kind: seam
depends_on: [M1-04, M1-05, M1-06]
discovered_by: null
size: M
---

# M1-07 · Templates as component bags, and spawning

## Goal

Every entity, whether creature, person, party or town, is spawned from one kind of content template: a bag of optional components. Later features add a component field and a spawner hook; they never add a new content type or edit the spawn function.

## Context

- Tech design: "Designing for pivots" (Entity templates, Identity, The party), "Content pipeline" step 1.
- Game design: "Party" (four fight slots, a fifth to playtest).
- M1-04 `sim.World`, M1-05 `comp`, M1-06 `content`.

## Deliverables

- `internal/content/types.go`: add `Template`, `PartyDef` and the `party` config; regenerate the CUE schema.
- `internal/core/comp/comp.go`: add `Party`.
- `internal/core/sim/spawn.go`, `spawn_test.go`
- `internal/core/sim/world.go`: add the `Content` field and the UID index.
- `content/data/templates.cue`, `content/data/party.cue`

## Spec

**Content:**

```go
// Template is a bag of optional components. A nil field means "no such component".
type Template struct {
	Appearance *comp.Appearance `json:"appearance,omitempty"`
	Party      *PartyDef        `json:"party,omitempty"`
	// Later: Body, Combatant, Needs, Inventory, Triggers, ...
}

type PartyDef struct {
	Members []api.ContentID `json:"members"` // member templates, in party order; the first leads
}

type PartyConfig struct {
	Slots       int `json:"slots"`        // fighters at a time; 4 to start
	SightRadius int `json:"sight_radius"` // hexes; 6 to start
}
```

`Bundle` gains `Templates map[string]Template` (IDs start with `template.`) and `Party PartyConfig` (`json:"party"`). `content` may import `comp` so components and content share one shape; `comp` must never import `content`.

**Components:** `Party{ Slots int; Leader api.UID }` on the party entity, and the Ark relation `MemberOf` from each member to its party. Slots are copied from content at spawn so a save keeps its own value.

**World additions:**

```go
Content *content.Registry // set by game.New

func (w *World) Spawn(template api.ContentID, at *comp.Pos) (ecs.Entity, error)
func (w *World) Despawn(e ecs.Entity)
func (w *World) ByUID(uid api.UID) (ecs.Entity, bool)
func (w *World) UIDOf(e ecs.Entity) api.UID

// A Spawner adds the components for one template field. Spawners run in
// registration order after the core components.
type Spawner struct {
	Name  string
	Apply func(w *World, e ecs.Entity, t content.Template) error
}
func (w *World) RegisterSpawner(s Spawner) error
```

`Spawn`:

1. Looks up the template; an unknown ID is an error.
2. Creates the entity with `Identity` (a new UID and the template ID), `Pos` when `at` isn't nil, and `Appearance` when it's set.
3. Runs every spawner in order.
4. Indexes the UID.

`Despawn` removes the entity and its index entry. The UID is never reused.

This task registers two spawners. `party` adds `comp.Party` with the configured slot count. `party_members` then spawns each member template in order with `MemberOf` pointing at the party, and sets `Leader` to the first member's UID. Members spawn without a position; M1-10 and M1-16 place them on the map. The party entity itself never gets a position: it holds what the whole party shares, such as the pack, knowledge and seen cells.

## Acceptance criteria

- [x] Spawning `template.party.main` gives a party entity with `Identity` and `Party{Slots: 4, Leader: <first member's UID>}`, plus two member entities (`template.person.reaper`, `template.person.salken`) with `Identity`, `Appearance` (glyph `@`, different colors) and `MemberOf` the party, in template order.
- [x] `ByUID` finds it; after `Despawn` it doesn't, and the next spawn gets a new UID.
- [x] Spawners run in registration order (a test with two recording spawners).
- [x] An unknown template ID is an error; a template with no components still gets `Identity`.
- [x] Spawning the same templates in the same order on two worlds with the same seed gives the same UIDs.
- [x] `go tool task check` passes.

## Out of scope

- Bodies and parts (M2-02).
- Placing the party on a map at game start (M1-10).

## Notes

- Spawners `party` and `party_members` are registered by `NewWorld`, so every world has them. They read `World.Content`, which must be set before spawning (`Spawn` errors if it's nil).
- `Spawn` removes the entity again if a spawner fails, and `party_members` removes members it already spawned, so a failed spawn leaves nothing behind.
- Content checks added beyond the spec: template IDs start with `template.`, a party needs at least one member, members must exist, and a member can't itself be a party (rejects cycles; easy to relax later). `party.slots` and `party.sight_radius` must be at least 1, so the existing content tests' sample data gained a `party` block.
- `comp.MemberOf` embeds `ecs.RelationMarker` with `json:"-"` so the generated CUE schema doesn't import Ark.
- Files outside the list: `Taskfile.yml` (the `schema` task now includes `./internal/core/comp`), `internal/content/{load,registry}.go` (validation, `Template` and `Party` accessors), `internal/content/load_test.go`, and the generated `*_go_gen.cue` files for `comp` and `content`.
- No follow-ups filed.
