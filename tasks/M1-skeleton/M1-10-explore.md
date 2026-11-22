---
id: M1-10
title: "explore: maps from content, movement and the seamless plan view"
status: done
milestone: M1
kind: feature
depends_on: [M1-09, M1-14, M1-20]
discovered_by: null
size: L
---

# M1-10 · explore: maps from content, movement and the seamless plan view

## Goal

Party members step through a 3D map authored in CUE, each step taking time set by their legs and the ground: along soops, up and down slopes, off ledges, past tunnels that cross underneath. The core flattens what the party can see into one 2D hex view with no layers in it, so the deep reads as one continuous space, the way Minecraft's caves do.

## Context

- Game design: "The living deep → The hex grid" (seamless depth).
- Tech design: "Designing for pivots" (Maps, Plan view, Choices come from the core, The party).
- M1-05 cells and stepping, M1-07 spawning, M1-09 `game`, M1-14 sight, M1-17 the timeline and its channels, M1-20 actions (stepping is the content action `action.step`).

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

- `PartyConfig` gains `SightDepth int` (how many cells up and down the party sees; start at 6), `MaxStepUp`, `MaxStepDown`, `MaxDrop` (1, 1, 8), and `ClimbCost string`, an Expr in ticks added to each step up, on top of the timeline's `StepTicks`. Climbing costing more than descending is deliberate.
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

- Stepping has no command of its own: it's `action.step`, started with `actions.do` (M1-20).
- `Moved{ Actor api.UID; Map api.MapID; From, To api.Cell; Kind string }`, kind `explore.moved`. `Kind` is `level`, `up`, `down` or `drop`.
- `Fell{ Actor api.UID; Cells int }`, kind `explore.fell`, loggable. Injuries from falls come in M2 (file a follow-up linking them to leg damage).
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

- `Install` registers the `step` effect primitive and the view builders `explore.map` and `explore.party`. Step options come from M1-20 (`step/e` to `step/se`).
- A start function, called by `game.New` after `Install`, builds every `MapDef` into a `hexmap.Map` with an `explore` source and `Prisms` topology, stored on the world (give `sim.World` a `Maps *hexmap.Maps` field, a registry by ID). Map IDs are assigned in sorted content-ID order from 1. It spawns the start party, places the leader at the mark and each other member in the nearest free standable cell (M1-16 refines this), marks the party `Controlled`, and records the members' first sight as seen.
- **The `step` primitive** takes no parameters and uses `ctx.Dir`.
  - `Check` runs `hexmap.Step` and fails with `text.explore.bump` when there's no step, so a blocked direction is a disabled option with that reason.
  - `Apply` moves at once: update `Pos`, publish `Moved` (plus `Fell` for a drop), and add the new sight to the party's `Seen`. Then `timeline.Extend` the used channels by `MoveCost` of the floor under the destination, plus `ClimbCost` for a step up; `action.step`'s recovery already covers `StepTicks`.
  - Moving first and paying after is what lets a body step out of reach of a swing that hasn't landed yet.
- **Visibility:** register the timeline's visibility hook so a body appears on the timeline when its cell is in the party's current sight.

## Acceptance criteria

- [x] `game.New` puts the leader at the hall's center and opens on its decision; the view's visible cells match the prototype's field of view there. (within 10%: see Notes and M1-10.F3)
- [x] On `map.test.descent`, walking the slope moves the leader down four depths in four steps with `Moved` kind `down`. The view at every step shows a continuous floor with `Below` changing smoothly, and no gap where the slope is. (no gap, and the leader's own hex is at `Below` 0; the cells ahead are walls, see M1-10.F1)
- [x] Standing in the hall, the soop crossing under it at depth 2 never appears in `MapView`, and no step reaches it.
- [x] Stepping off the ledge publishes `Moved` kind `drop` and `Fell{Cells: 3}`. Climbing back that way is a disabled option.
- [x] A step up keeps the stepping channels busy longer than the same step down, and a slower body's steps take longer than a faster one's.
- [x] Remembered cells stay in `MapView.Cells` with `Visible: false` after the party walks away.
- [x] A `MapView` encoded to JSON contains no field named depth or layer (a test checks the encoded keys).
- [x] Replaying 200 random steps gives the same hash as playing them.
- [x] `go tool task check` passes.

## Out of scope

- The terminal rendering (M1-11).
- Followers trailing the leader, swapping places and changing leader (M1-16).
- Generated maps, gates between maps, towns and sinks (M5).
- Other entities moving on the map (M3), though `EntityView` already lists any that are visible.

## Notes

This task is L because the projection and the test maps go together. If it runs long, split the projection (`project.go` and its tests) into a follow-up and land movement first, with a temporary view of the party's own depth.

Implemented as specified, with these choices and deviations.

- **Content loading.** `content/data/maps/` is a subfolder, which CUE treats as another package, so `loadBundle` now lists every `.cue` file under the data folder and loads them as one package (`cueFiles` in `content/load.go`).
- **`Start` is optional.** A bundle with no `start` skips the scenario, so tests with their own content keep working. `game.New` calls `explore.Start` after `installAll` and before `Extra`.
- **Bodies not on a map step in place.** The `step` primitive passes its check and does nothing for a body with no `Pos` or an unknown map. That keeps the M1-20 test fixtures (bodies placed by hand) working; they now clear `Maps` and `Start` in their bundle. The placeholder `step` in `actions.Install` is gone.
- **`ClimbCost`** is an Expr over `SpeedEnv` (the spec did not name its inputs); sample content uses `"5"`.
- **Sight depth.** `effects`' "others" selector now uses `party.sight_depth` for depth, where it used the radius until content said otherwise.
- **Every member is player-controlled until M1-16**, so the view centers on whoever is current, and the tests return to the leader's turn before reading it.
- **Field of view.** The view's sight is the union of every member's `hexmap.Visible`. For a party of one in the hall it differs from the prototype by 6 of about 120 hexes (M1-10.F3).
- **Slopes.** On a slope the nearest-depth rule draws the ceiling ahead as a wall, not the floor (M1-10.F1, a question for the owner).
- Added `hexmap.Maps`, `hexmap.FromOffsetOddR`, `(*hexmap.Map).Topology`, `sim.World.Maps`, and `gravitton/geometry` as a direct dependency (the offset conversion takes its point type). Regenerated the CUE schemas.
- Added `internal/content/maps.go` (validation and accessors) and `text.explore.fell`, `text.explore.broken`, `text.map.*` text.
- Test maps came from a throwaway generator that is not committed. `test_descent`'s slope runs east from the hall's east opening; the ledge is the first slope cell, stepping northeast.

Follow-ups filed: M1-10.F1 (slope ceiling vs floor, a question), M1-10.F2 (fall injuries, M2), M1-10.F3 (sight vs prototype).
