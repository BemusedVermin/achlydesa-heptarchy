package hexmap

import "github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"

// Link is where a move leads. Map 0 means the same map.
type Link struct {
	Map  api.MapID
	Cell api.Cell
}

// Topology says which cells connect. Everything in hexmap asks it for neighbors, so impossible geometry is a new
// Topology rather than an edit.
type Topology interface {
	// Neighbor is the cell one horizontal step from at in direction dir.
	Neighbor(at api.Cell, dir api.Dir) Link
	// Above is the cell one step toward the surface (smaller depth).
	Above(at api.Cell) Link
	// Below is the cell one step deeper (larger depth).
	Below(at api.Cell) Link
}

// Prisms is the default topology: hex neighbors at the same depth, and depth ± 1 straight up and down.
type Prisms struct{}

// prismDirs are the axial offsets of each api.Dir, indexed by Dir.
var prismDirs = [api.DirCount]api.Hex{
	api.DirE:  {Q: 1, R: 0},
	api.DirNE: {Q: 1, R: -1},
	api.DirNW: {Q: 0, R: -1},
	api.DirW:  {Q: -1, R: 0},
	api.DirSW: {Q: -1, R: 1},
	api.DirSE: {Q: 0, R: 1},
}

// Neighbor implements [Topology]. An invalid dir yields a link to at itself.
func (Prisms) Neighbor(at api.Cell, dir api.Dir) Link {
	if !dir.Valid() {
		return Link{Cell: at}
	}
	off := prismDirs[dir]
	return Link{Cell: api.Cell{Hex: api.Hex{Q: at.Hex.Q + off.Q, R: at.Hex.R + off.R}, Depth: at.Depth}}
}

// Above implements [Topology].
func (Prisms) Above(at api.Cell) Link {
	return Link{Cell: api.Cell{Hex: at.Hex, Depth: at.Depth - 1}}
}

// Below implements [Topology].
func (Prisms) Below(at api.Cell) Link {
	return Link{Cell: api.Cell{Hex: at.Hex, Depth: at.Depth + 1}}
}
