package hexmap

import "github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"

// ChunkSize is the number of cells along q, r and depth in one chunk.
const ChunkSize = 16

// Terrain is what one cell is made of.
type Terrain struct {
	ID api.ContentID
	// Blocks movement and can be stood on
	Solid bool
	// Blocks sight
	Opaque bool
}

// ChunkKey is a chunk's position: each cell coordinate floor-divided by ChunkSize.
type ChunkKey struct{ Q, R, Depth int }

// KeyOf is the key of the chunk holding c. It floors, so negative coordinates fall in the right chunk.
func KeyOf(c api.Cell) ChunkKey {
	return ChunkKey{
		Q:     floorDiv(c.Hex.Q, ChunkSize),
		R:     floorDiv(c.Hex.R, ChunkSize),
		Depth: floorDiv(c.Depth, ChunkSize),
	}
}

// Origin is the lowest cell coordinate in the chunk.
func (k ChunkKey) Origin() api.Cell {
	return api.Cell{Hex: api.Hex{Q: k.Q * ChunkSize, R: k.R * ChunkSize}, Depth: k.Depth * ChunkSize}
}

// Cell is the cell at a slot of the chunk, the inverse of [SlotOf].
func (k ChunkKey) Cell(slot int) api.Cell {
	origin := k.Origin()
	q, rest := slot%ChunkSize, slot/ChunkSize
	r, depth := rest%ChunkSize, rest/ChunkSize
	return api.Cell{Hex: api.Hex{Q: origin.Hex.Q + q, R: origin.Hex.R + r}, Depth: origin.Depth + depth}
}

// ChunkCells is the number of cells, and so of slots, in one chunk.
const ChunkCells = ChunkSize * ChunkSize * ChunkSize

// SlotOf is c's place within its chunk, in [0, ChunkCells): cells of a chunk ordered by depth, then R, then Q.
func SlotOf(c api.Cell) int {
	return index(floorMod(c.Hex.Q), floorMod(c.Hex.R), floorMod(c.Depth))
}

// ChunkDistance is how many chunks apart two chunks are: the hex distance between their keys, or their
// difference in depth when that is larger.
func ChunkDistance(a, b ChunkKey) int {
	return max(Distance(api.Hex{Q: a.Q, R: a.R}, api.Hex{Q: b.Q, R: b.R}), abs(a.Depth-b.Depth))
}

func abs(n int) int { return max(n, -n) }

// floorDiv divides rounding toward negative infinity, for a positive divisor.
func floorDiv(n, divisor int) int {
	q := n / divisor
	if n%divisor < 0 {
		q--
	}
	return q
}

func floorMod(n int) int {
	return n - floorDiv(n, ChunkSize)*ChunkSize
}

// Chunk holds ChunkSize³ cells as indexes into a small terrain palette. A chunk that is all its fill terrain
// stores no slots, so a uniform chunk costs about as much as one terrain.
type Chunk struct {
	palette []Terrain
	// cells is nil while every cell is palette[0].
	cells []uint16
}

// NewChunk returns a chunk filled with one terrain.
func NewChunk(fill Terrain) *Chunk {
	return &Chunk{palette: []Terrain{fill}}
}

// index maps local coordinates, each in [0, ChunkSize), to a slot; it panics outside that range.
func index(q, r, depth int) int {
	if q < 0 || q >= ChunkSize || r < 0 || r >= ChunkSize || depth < 0 || depth >= ChunkSize {
		panic("hexmap: local chunk coordinate out of range")
	}
	return (depth*ChunkSize+r)*ChunkSize + q
}

// Get returns the terrain at local coordinates, each in [0, ChunkSize).
func (c *Chunk) Get(q, r, depth int) Terrain {
	slot := index(q, r, depth)
	if c.cells == nil {
		return c.palette[0]
	}
	return c.palette[c.cells[slot]]
}

// Set changes the terrain at local coordinates, each in [0, ChunkSize).
func (c *Chunk) Set(q, r, depth int, t Terrain) {
	slot := index(q, r, depth)
	if c.cells == nil {
		if t == c.palette[0] {
			return
		}
		c.cells = make([]uint16, ChunkCells)
	}
	for i, p := range c.palette {
		if p == t {
			c.cells[slot] = uint16(i)
			return
		}
	}
	c.palette = append(c.palette, t)
	c.cells[slot] = uint16(len(c.palette) - 1)
}

// Source supplies chunks on demand.
type Source interface {
	Chunk(key ChunkKey) (*Chunk, error)
}

// FuncSource is a [Source] defined by a function from cell to terrain, for uniform worlds and tests.
type FuncSource func(api.Cell) Terrain

// Chunk implements [Source].
func (f FuncSource) Chunk(key ChunkKey) (*Chunk, error) {
	origin := key.Origin()
	chunk := NewChunk(f(origin))
	for d := range ChunkSize {
		for r := range ChunkSize {
			for q := range ChunkSize {
				cell := api.Cell{
					Hex:   api.Hex{Q: origin.Hex.Q + q, R: origin.Hex.R + r},
					Depth: origin.Depth + d,
				}
				chunk.Set(q, r, d, f(cell))
			}
		}
	}
	return chunk, nil
}
