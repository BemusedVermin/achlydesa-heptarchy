package api

// Dir is one of the six horizontal directions, pointy-top, counter-clockwise from east.
type Dir uint8

// The six horizontal directions.
const (
	DirE Dir = iota
	DirNE
	DirNW
	DirW
	DirSW
	DirSE
)

// DirCount is the number of horizontal directions.
const DirCount = 6

// Valid reports whether d is one of the six directions.
func (d Dir) Valid() bool { return d < DirCount }

// Cell is one hex-prism cell of the world.
type Cell struct {
	Hex   Hex `json:"hex"`
	Depth int `json:"depth"` // 0 is the salk's surface; larger is deeper; one cell per step
}
