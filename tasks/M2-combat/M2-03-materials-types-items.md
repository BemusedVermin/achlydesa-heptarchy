---
id: M2-03
title: Materials, types, items and the party's pack
status: done
milestone: M2
kind: feature
depends_on: [M2-02]
discovered_by: null
size: S
---

# M2-03 · Materials, types, items and the party's pack

## Goal

Weaknesses, types and the items that consumable types burn are content, and the party carries a pack. Code knows nothing about Cut or Soak; it only knows a type may need an item and may have a condition.

## Context

- Game design: "Combat → Types" (the full table and its catches), "Core loop and time management → Time pressures" (where supplies come from).
- This milestone's README, "The combat model" (Moves).

## Deliverables

- `internal/content/types.go`: extend `Material`; add `DamageType`, `Item`, `Faction`; regenerate the schema.
- `internal/content/env.go`: add `TypeEnv`.
- `content/data/types.cue`, `items.cue`, `factions.cue`; extend `materials.cue`.
- `internal/core/comp/inventory.go`, plus a spawner for `Template.Inventory`.
- `internal/core/inventory/inventory.go` (foundation), `inventory_test.go`

## Spec

**Content:**

```go
type Material struct {
	Name     api.ContentID   `json:"name"`
	WeakTo   []api.ContentID `json:"weak_to"`   // type IDs
	ImmuneTo []api.ContentID `json:"immune_to"` // type IDs
	Tags     []string        `json:"tags"`      // e.g. "old_matter", "clay"
}

type DamageType struct {
	Name      api.ContentID  `json:"name"`
	Innate    bool           `json:"innate"`
	Consumes  *api.ContentID `json:"consumes,omitempty"`  // item ID used up per move
	Condition string         `json:"condition,omitempty"` // Expr bool over TypeEnv; empty means always
}

type Item struct {
	Name  api.ContentID `json:"name"`
	Kind  string        `json:"kind"`  // "consumable", "piece", "keepsake", "name"
	Bound bool          `json:"bound"` // tied to one person, e.g. a name or a keepsake
}

type Faction struct {
	Name      api.ContentID   `json:"name"`
	HostileTo []api.ContentID `json:"hostile_to"`
}

// TypeEnv is what a type's condition can see.
type TypeEnv struct {
	TargetUID   uint64 `expr:"target_uid"`
	ItemBoundTo uint64 `expr:"item_bound_to"` // 0 when the item isn't bound
	TargetTags  []string `expr:"target_tags"`
}
```

Bundle keys: `types.` (map `types`), `item.`, `faction.`. Hostility is not symmetric unless both sides list each other; validation warns, doesn't fail, on one-sided hostility.

**Sample content:** all eleven types from the design table, with `type.name` and `type.keepsake` conditioned on `target_uid == item_bound_to`. Items for the consumables: coal, salk dust, lake water, daylight jar, Violet blood and lees. Materials, read straight from the design table:

| Material | Weak to | Immune to |
| --- | --- | --- |
| `material.flesh` | Cut, Burn, Relic | none |
| `material.clay` | Soak, Relic | none |
| `material.old_matter` | Rot | Relic |
| `material.bone` (hollow: skulls, shells) | Press | none |
| `material.eye` | Light | none |
| `material.rift_flesh` | Violet | none |

Catches such as Burn hardening clay or Rot eating grafts are triggers (M2-05) and content (M2-11), not material fields. Where the table is ambiguous, pick the reading above and note it.

**Pack:**

```go
type ItemStack struct {
	Item    api.ContentID `json:"item"`
	Count   int           `json:"count"`
	BoundTo api.UID       `json:"bound_to,omitempty"`
}
type Inventory struct{ Stacks []ItemStack } // sorted by (Item, BoundTo)

func Add(w *sim.World, holder ecs.Entity, s ItemStack)
func Take(w *sim.World, holder ecs.Entity, item api.ContentID, boundTo api.UID, n int) bool // false, and no change, if short
func Count(w *sim.World, holder ecs.Entity, item api.ContentID, boundTo api.UID) int
```

`inventory` is already listed as foundation in `AGENTS.md` and the tech design.

## Acceptance criteria

- [x] All eleven types load; Name's condition compiles against `TypeEnv`.
- [x] `Take` more than the count fails without changing the pack; stacks merge and stay sorted.
- [x] The party template starts with a few of each consumable in its pack.
- [x] Save and load keeps the pack.
- [x] `go tool task check` passes.

## Out of scope

- Using types in moves (M2-06). Buying and finding items (later milestones).

## Notes

- Bundle keys use the singular prefixes `type.`, `item.`, `faction.` (the spec's `types.` read as a typo); the map fields are `types`, `items`, `factions`.
- `Template.Inventory` is `[]StackDef{Item, Count}` in content (no `BoundTo`, since a UID can't be authored). The `inventory` spawner lives in `internal/core/inventory` and is installed by `game.installAll` after anatomy. The pack sits on the party entity (`template.party.main`), three of each consumable.
- Sample types: Cut, Press, Relic, Name are innate; the rest consume their item. Name has no `consumes`: its "used up when spoken" catch is a trigger (M2-05). Keepsake consumes `item.keepsake`, a bound item of kind `keepsake` that I added so the type has something to burn; both Name and Keepsake use `target_uid == item_bound_to`.
- Materials follow the spec table; `material.old_matter`, `bone`, `eye`, `rift_flesh` are new with text IDs. Each carries one tag of its own name; tune later.
- Validation also rejects a type both weak and immune on one material, unknown types, items and factions, bad item kinds and bad template packs. One-sided hostility is a warning: `Registry.Warnings()` (new; the registry had no warning channel). Sample factions `faction.party` and `faction.deep` hostile to each other.
- Added `Registry.Type`, `TypeIDs`, `Item`, `Faction`, `TypeAllowed` (runs the compiled condition).
- Files outside the list: `comp.go` (registers `Inventory`), `game/install.go`, `content/load.go`, `registry.go`, `types_check.go` (new), `content/data/templates.cue`. `watch_go_gen.cue` appears because `task schema` generates it for M2-12's `watch.go`, which had no generated file.
- No follow-ups filed.
