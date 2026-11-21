---
id: M1-10
title: "explore: maps from content, movement and the seamless plan view"
status: todo
milestone: M1
kind: feature
depends_on: [M1-09, M1-14]
discovered_by: null
size: L
---

# M1-10 · explore: maps from content, movement and the seamless plan view

## Goal

Party members step through a 3D map authored in CUE, one turn at a time: along soops, up and down slopes, off ledges, past tunnels that cross underneath. The core flattens what the party can see into one 2D hex view with no layers in it, so the deep reads as one continuous space, the way Minecraft's caves do.

## Context

- Game design: "The living deep → The hex grid" (seamless depth).
- Tech design: "Designing for pivots" (Maps, Plan view, Choices come from the core, The party).
- M1-05 cells and stepping, M1-07 spawning, M1-09 `game`, M1-14 sight, M1-15 turns.

## Deliverables

- `internal/content/types.go`: add `MapDef`, `LayerDef` and `Start`; extend `PartyConfig`; regenerate the schema.
- `content/data/maps/test_hall.cue`, `content/data/maps/test_descent.cue`, `content/data/start.cue`
- `internal/core/api/explore.go`: commands, events and view sections.
- `internal/core/comp/comp.go`: `Seen`, `Controlled`.
- `internal/core/explore/source.go`: a `hexmap.Source` built from a `MapDef`.
- `internal/core/explore/explore.go`, `project.go`, `install.go`, tests.
- `internal/core/game/install.go`: add `explore`; `game.New` runs the start scenario.

## Spec

**Content:**

```go
// MapDef is a hand-made map: a few authored layers embedded in endless fill.
type MapDef struct {
	Name   api.ContentID            `json:"name"`
	Fill   api.ContentID            `json:"fill"`   // terrain of every cell not authored, forever in every direction
	Open   api.ContentID            `json:"open"`   // terrain of a marked cell, usually air
	Layers []LayerDef               `json:"layers"`
	Legend map[string]api.ContentID `json:"legend"` // one character -> terrain ID
	Marks  map[string]string        `json:"marks"`  // one character -> mark name, e.g. "@": "party_start"
}

// LayerDef is one depth of a hand-made map. Rows use odd-r offset layout,
// one character per cell; a space leaves the fill.
type LayerDef struct {
	Depth int      `json:"depth"`
	Rows  []string `json:"rows"`
}

type Start struct {
	Map         api.ContentID `json:"map"`
	Party       api.ContentID `json:"party"` // template ID
	PartyAtMark string        `json:"party_at_mark"`
}
```

- `PartyConfig` gains `SightDepth int` (how many cells up and down the party sees; start at 6), `MaxStepUp`, `MaxStepDown`, `MaxDrop` (1, 1, 8), and `ClimbCost string`, an Expr in ticks added to each step up. Climbing costing more than descending is deliberate.
- `Bundle` gains `Maps` (`map.`) and `Start`.
- Validation: every character in every layer is in the legend or the marks, legend targets are terrains, each mark appears once, and the start's map, template and mark exist.
- Use the hex library's odd-r offset conversion; confirm its name with `go doc`.

**Test maps.** A throwaway generator is fine; don't commit it.

- `map.test.hall`: the prototype room at depth 0. Air inside, a ring of packed clay with the three pillars, fill of packed clay so depth 1 is the floor, and the party mark at the center.
- `map.test.descent`, for the tests:
  - a hall at depth 0;
  - a soop that slopes down one cell per hex to a chamber at depth 4;
  - an unconnected soop at depth 2 that crosses under the hall;
  - a ledge that drops three cells into the chamber.

**api:**

- `Step{ Actor api.UID; Dir api.Dir }`, kind `explore.step`.
- `Moved{ Actor api.UID; Map api.MapID; From, To api.Cell; Kind string }`, kind `explore.moved`. `Kind` is `level`, `up`, `down` or `drop`.
- `Fell{ Actor api.UID; Cells int }`, kind `explore.fell`, loggable. Injuries from falls come in M2 (file a follow-up linking them to leg damage).
- `Bumped{ Actor api.UID; At api.Cell }`, kind `explore.bumped`, loggable with `text.explore.bump`.
- The view sections:

```go
type MapView struct {
	Map      api.MapID
	Name     string
	Cells    []CellView   // sorted by R, Q
	Entities []EntityView // sorted by UID
}
type CellView struct {
	Hex     api.Hex
	Glyph   string
	Color   string
	Visible bool
	Below   int // the drawn cell's depth minus the party's: positive is lower, negative higher
}
type EntityView struct {
	UID   api.UID
	Hex   api.Hex
	Glyph string
	Color string
	Below int
}
type PartyView struct {
	UID     api.UID
	Map     api.MapID
	Hex     api.Hex // where to center the map: the current actor if it's a member, else the leader; a hex, never a depth
	Members []MemberView // in party order
}
type MemberView struct {
	UID     api.UID
	Name    string
	Glyph   string
	Color   string
	Leader  bool
	Current bool
}
```

`View` gains `Map *MapView` and `Party *PartyView`. **Neither contains an absolute depth or a layer index.** Frontends get only what to draw per hex and how far above or below the party it is, so they can shade smoothly without ever knowing layers exist. The harness reads absolute positions from components, not the view.

**Components:**

- `Controlled{}`: a marker for the party the player commands.
- `Seen{ ByMap map[api.MapID][]api.Cell }` on the party entity: cells any member has seen, each slice sorted and unique. Hide it behind `markSeen` and `seenIn(area)` helpers, because M5 will swap it for per-chunk bitsets.

**The plan view projection** (`project.go`; rules firm). For each hex within the party's sight radius, plus every hex with a seen cell nearby, draw at most one cell:

1. **Candidates** are cells in that hex's column that are either standable (drawn with the floor glyph of the solid cell beneath) or solid with a horizontally adjacent open cell (drawn with the wall glyph).
2. Among visible candidates (the union of every member's `Visible`, with the party's sight radius and `SightDepth`), take the one nearest the focus member in depth; ties go to the shallower one. Mark it visible. The focus member is the current actor if it's a party member, otherwise the leader, and `Below` is measured from it.
3. If there are none, apply the same rule to remembered candidates, marked not visible.
4. Entities are listed when their cell is visible, with `Below` set.

So a soop crossing under the party's hall is never drawn while the hall's floor is visible, and a slope reads as one continuous floor that gets gradually lower.

**explore:**

- `Install` registers `explore.step`, the view builders `explore.map` and `explore.party`, and an options builder.
- The options builder adds six options, `step.e` to `step.se`, when the current actor is a player-controlled party member that hasn't stepped this turn. Each is disabled, with the reason from `text.explore.bump`, when `Step` finds nothing in that direction.
- A start function, called by `game.New` after `Install`, builds every `MapDef` into a `hexmap.Map` with an `explore` source and `Prisms` topology, stored on the world (give `sim.World` a `Maps *hexmap.Maps` field, a registry by ID). Map IDs are assigned in sorted content-ID order from 1. It spawns the start party, places the leader at the mark and each other member in the nearest free standable cell (M1-16 refines this), marks the party `Controlled`, and records the members' first sight as seen.
- **Step:** rejected unless `Actor` is the current, player-controlled actor and it hasn't stepped this turn. `hexmap.Step` in the given direction. If there's a step, update `Pos`, spend the step (`turns.Spend`), publish `Moved` (plus `Fell` for a drop), add the new sight to the party's `Seen`, and return the cost: `MoveCost` of the floor terrain under the destination, plus `ClimbCost` for a step up. The cost lands in the actor's `Extra`, so climbing delays its next turn. Otherwise publish `Bumped` and return 0. For now a step always ends the turn; M1-16 makes that depend on the party's alert.

## Acceptance criteria

- [ ] `game.New` puts the leader at the hall's center and opens on its turn; the view's visible cells match the prototype's field of view there.
- [ ] On `map.test.descent`, walking the slope moves the leader down four depths in four steps with `Moved` kind `down`. The view at every step shows a continuous floor with `Below` changing smoothly, and no gap where the slope is.
- [ ] Standing in the hall, the soop crossing under it at depth 2 never appears in `MapView`, and no step reaches it.
- [ ] Stepping off the ledge publishes `Moved` kind `drop` and `Fell{Cells: 3}`. Climbing back that way is a disabled option.
- [ ] A step up delays the actor's next turn more than the same step down.
- [ ] Remembered cells stay in `MapView.Cells` with `Visible: false` after the party walks away.
- [ ] A `MapView` encoded to JSON contains no field named depth or layer (a test checks the encoded keys).
- [ ] Replaying 200 random steps gives the same hash as playing them.
- [ ] `go tool task check` passes.

## Out of scope

- The terminal rendering (M1-11).
- Followers trailing the leader, swapping places and changing leader (M1-16).
- Generated maps, gates between maps, towns and sinks (M5).
- Other entities moving on the map (M3), though `EntityView` already lists any that are visible.

## Notes

This task is L because the projection and the test maps go together. If it runs long, split the projection (`project.go` and its tests) into a follow-up and land movement first, with a temporary view of the party's own depth.
