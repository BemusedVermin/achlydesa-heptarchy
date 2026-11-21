---
id: M1-05
title: "comp basics and hexmap: 3D cells, chunks, topology"
status: review
milestone: M1
kind: seam
depends_on: [M1-03]
discovered_by: null
size: M
---

# M1-05 · comp basics and hexmap: 3D cells, chunks, topology

## Goal

The world is three-dimensional: columns of hex cells with depth as fine as Minecraft's blocks, so soops can cross over and under each other, slope down gently, and drop off ledges. Maps are stored in chunks so they can be infinite, and neighbors come from a topology so impossible connections stay possible later. Players never see layers; that's M1-10's projection. This task builds the space underneath it.

## Context

- Game design: "The living deep → The hex grid" (depth is continuous and the plan view is seamless).
- Tech design: "Designing for pivots" (Maps, Plan view), "Core packages" (`comp`, `hexmap`), "Library choices" (Hex math).
- `internal/prototype/game.go` shows the gravitton/hexagon calls that work: `hex.Pt`, `Range`, `Ring`, `Neighbor`.

## Deliverables

- `internal/core/api/space.go`: `Dir` and `Cell`.
- `internal/core/comp/comp.go`: the first components.
- `internal/core/hexmap/chunk.go`, `map.go`, `topology.go`, `step.go`, and tests for each.

## Spec

**api** (firm):

```go
// Dir is one of the six horizontal directions, pointy-top, counter-clockwise from east.
type Dir uint8
const (DirE Dir = iota; DirNE; DirNW; DirW; DirSW; DirSE)

// Cell is one hex-prism cell of the world.
type Cell struct {
	Hex   Hex `json:"hex"`
	Depth int `json:"depth"` // 0 is the salk's surface; larger is deeper; one cell per step
}
```

**Components** (firm). Plain structs, JSON-tagged, with no methods beyond trivial helpers:

```go
type Pos struct {
	Map api.MapID `json:"map"`
	At  api.Cell  `json:"at"`
}
type Identity struct {
	UID      api.UID       `json:"uid"`
	Template api.ContentID `json:"template"`
}
type Appearance struct {
	Glyph  string `json:"glyph"`            // one grapheme, for the terminal
	Color  string `json:"color,omitempty"`  // a color name or #rrggbb
	Sprite string `json:"sprite,omitempty"` // for the graphics frontend later
}
```

**Cells and chunks** (behavior firm, names suggested):

```go
type Terrain struct {
	ID     api.ContentID
	Solid  bool // blocks movement and can be stood on
	Opaque bool // blocks sight
}

const ChunkSize = 16 // cells along q, r and depth

// ChunkKey is a chunk's position: each cell coordinate floor-divided by ChunkSize.
type ChunkKey struct{ Q, R, Depth int }

// Chunk holds ChunkSize³ cells as indexes into a small terrain palette.
type Chunk struct{ /* palette []Terrain; cells []uint16 */ }

// Source supplies chunks on demand: an authored map now, procgen in M5.
type Source interface {
	Chunk(key ChunkKey) (*Chunk, error)
}
```

- Use floor division for negative coordinates; a test covers cells on both sides of every chunk boundary.
- Hexes are axial `(Q, R)`. Chunks tile axial space as parallelograms; that's fine because nothing outside `hexmap` sees chunks.

**Map:**

```go
func NewMap(id api.MapID, name string, src Source, topo Topology) *Map
func (m *Map) ID() api.MapID
func (m *Map) Name() string
func (m *Map) At(c api.Cell) (Terrain, error)     // loads the chunk on first use and caches it
func (m *Map) Set(c api.Cell, t Terrain) error     // an edit, kept in an overlay so it can be saved and survives reloading the chunk
func (m *Map) Edits() []Edit                       // sorted by Depth, R, Q
func (m *Map) CellsIn(center api.Hex, radius, depthFrom, depthTo int) []api.Cell // sorted by Depth, R, Q
func (m *Map) Standable(c api.Cell) (bool, error)  // c is open and the cell below it is solid
```

There is deliberately no "list every cell": an infinite map can't.

**Topology:**

```go
// Link is where a move leads. Map 0 means the same map.
type Link struct {
	Map  api.MapID
	Cell api.Cell
}
type Topology interface {
	Neighbor(at api.Cell, dir api.Dir) Link
	Above(at api.Cell) Link
	Below(at api.Cell) Link
}
// Prisms is the default: hex neighbors at the same depth, and depth ± 1 straight up and down.
type Prisms struct{}
```

Everything in `hexmap` asks the topology for neighbors and never does hex arithmetic directly, so a fold or a sink with impossible geometry is a new `Topology`, not an edit.

**Stepping** (firm rules; bodies are one cell tall, as people crawl in the soops):

```go
type StepRules struct{ MaxStepUp, MaxStepDown, MaxDrop int } // defaults 1, 1, 8
type StepKind uint8
const (StepLevel StepKind = iota; StepUp; StepDown; StepDrop)
type Step struct {
	To   Link
	Kind StepKind
	Fall int // cells fallen, for StepDrop
}
func (m *Map) Step(from api.Cell, dir api.Dir, rules StepRules) (Step, bool, error)
```

Let `n` be `Neighbor(from, dir)`. In order:

1. If `n` is standable, it's a level step.
2. If `n` is solid: for `k` from 1 to `MaxStepUp`, if the cell `k` above `n` is standable and the `k` cells above `from` are all open (headroom), it's a step up to that cell.
3. If `n` is open but not standable: find the first standable cell below `n`, at most `MaxDrop` cells down. Within `MaxStepDown` it's a step down; beyond that it's a drop with `Fall` set. Deeper than `MaxDrop`, there's no step.
4. Otherwise there's no step.

Steps through a `Link` to another map are taken as given.

Conversions between `api` types and gravitton/hexagon happen only inside `hexmap`; no other package imports gravitton/hexagon or fzipp/astar.

## Acceptance criteria

- [x] `Prisms` neighbors, `Above` and `Below` match hand-computed cells for each `Dir`.
- [x] A `rapid` property: `Set` then `At` round-trips for random cells, including negative coordinates and chunk edges.
- [x] The source is asked for each chunk once; edits survive and `Edits()` is sorted.
- [x] Stepping: level, up one, refused up two, down one, a drop of three (`Fall` 3), refused past `MaxDrop`, and up refused without headroom.
- [x] Overlap: two tunnels crossing the same hexes at depths 0 and 4 never step into each other.
- [x] `CellsIn` is sorted and stable over 100 calls.
- [x] `go tool task check` passes.

## Out of scope

- Sight and pathing (M1-14). Maps from content and the plan view (M1-10). Generated sources (M5). Unloading chunks under a memory budget (M5).

## Notes

- **No gravitton/hexagon or fzipp/astar yet.** `Prisms` and `CellsIn` use a six-entry axial offset table and plain loops, which is all this task needs. The library conversions the spec mentions are left for M1-14 (sight and pathing), where they earn their place.
- **Additions beyond the spec:** `api.DirCount` and `Dir.Valid`; `hexmap.KeyOf` and `ChunkKey.Origin`; `Chunk.Get/Set/NewChunk` (local coordinates); `FuncSource` (a `Source` from a cell-to-terrain function, used by tests and handy for uniform maps); `DefaultStepRules()` and `ErrBadDir`.
- **Decisions on unspecified points:**
  - `StepDown` leaves `Fall` at 0; only `StepDrop` sets it. `Fall` counts cells below the neighbor cell, so a drop of three is `Fall` 3.
  - A link to another map counts as solid for `Standable` and headroom checks, since this map can't see it. A `Step` whose neighbor link leaves the map is returned as a level step as given.
  - `Step` returns `ErrBadDir` for a direction outside the six; `Prisms.Neighbor` returns the cell itself for one.
  - `Map.Set` always returns nil today; the error is kept for sources that may refuse edits later.
- Files touched outside the list: none.
- Follow-ups filed: none.
