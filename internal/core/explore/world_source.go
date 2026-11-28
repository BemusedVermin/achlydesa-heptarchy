package explore

import (
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/worldgen"
)

// ErrOverlap is returned by [WorldSource.Stamp] when the piece would cover a cell that a place or an
// earlier piece already covers.
var ErrOverlap = errors.New("explore: the piece overlaps what is already stamped")

// ErrOutside is returned by [WorldSource.Stamp] when the piece would cover a cell outside the world.
var ErrOutside = errors.New("explore: the piece reaches outside the world")

// WorldSource is a [hexmap.Source] for the one world: every place's layers at its position, and every
// stamped piece, set in the world's fill. A [hexmap.Map] over it takes [WorldSource.Bounds] and
// [WorldSource.Edge].
type WorldSource struct {
	fill, edge, open hexmap.Terrain
	bounds           hexmap.Bounds
	authored         map[hexmap.ChunkKey][]placed
	covered          map[api.Cell]bool
	places           []hexmap.Footprint
	marks            map[placeMark]api.Cell
	spawns           []Spawn
	landmarks        []Landmark
}

type placeMark struct {
	place api.ContentID
	name  string
}

// Stamped is where a stamped piece's marks fell in the world.
type Stamped struct {
	// Marks maps each mark name that is not a spawn, landmark or door to its cell.
	Marks map[string]api.Cell
	// Doors lists the door cells, layers as authored, then rows and columns.
	Doors []api.Cell
}

// NewWorldSource builds the world of the content's world, places, settlements and the landmarks of its
// geography. Marked cells hold the world's Open terrain. Each settlement is laid out from the world seed
// alone and stamped after the places, in ID order; the landmarks follow, in the geography's order.
func NewWorldSource(reg *content.Registry, seed uint64) (*WorldSource, error) {
	world := reg.World()
	src := &WorldSource{
		bounds: hexmap.Bounds{
			MaxCol: world.Cols - 1, MaxRow: world.Rows - 1,
			MinDepth: world.MinDepth, MaxDepth: world.MaxDepth,
		},
		authored: map[hexmap.ChunkKey][]placed{},
		marks:    map[placeMark]api.Cell{},
		covered:  map[api.Cell]bool{},
	}
	var err error
	if src.fill, err = Terrain(reg, world.Fill); err != nil {
		return nil, fmt.Errorf("world fill: %w", err)
	}
	if src.edge, err = Terrain(reg, world.Edge); err != nil {
		return nil, fmt.Errorf("world edge: %w", err)
	}
	if src.open, err = Terrain(reg, world.Open); err != nil {
		return nil, fmt.Errorf("world open: %w", err)
	}
	for _, id := range reg.PlaceIDs() {
		def, _ := reg.Place(id)
		at := def.At
		drawn, err := draw(reg, def.Layout(), src.open, func(col, row, depth int) api.Cell {
			return api.Cell{Hex: hexmap.FromOffsetOddR(at.Col+col, at.Row+row), Depth: at.Depth + depth}
		})
		if err != nil {
			return nil, fmt.Errorf("place %q: %w", id, err)
		}
		src.add(id, drawn)
		for name, cell := range drawn.marks {
			src.marks[placeMark{place: id, name: name}] = cell
		}
		slices.SortFunc(drawn.openings, hexmap.CompareCells)
		src.places = append(src.places, hexmap.Footprint{Place: id, Openings: drawn.openings})
	}
	for _, id := range reg.SettlementIDs() {
		if err := src.stampSettlement(reg, id, seed); err != nil {
			return nil, err
		}
	}
	if geography, ok := reg.Geography(world.Geography); ok {
		for i := range geography.Landmarks {
			if err := src.stampLandmark(reg, world.Geography, i, seed); err != nil {
				return nil, err
			}
		}
	}
	return src, nil
}

// stampLandmark places and stamps one of the geography's landmarks; its doors are the openings of its
// footprint.
func (s *WorldSource) stampLandmark(reg *content.Registry, geography api.ContentID, index int, seed uint64) error {
	placement, err := worldgen.PlaceLandmark(reg, geography, index, seed, s.Covers)
	if err != nil {
		return err
	}
	stamped, err := s.Stamp(reg, placement.Piece, placement.At, placement.Rotation)
	if err != nil {
		return fmt.Errorf("landmark %d of %q: %w", index, geography, err)
	}
	doors := slices.Clone(stamped.Doors)
	slices.SortFunc(doors, hexmap.CompareCells)
	s.places = append(s.places, hexmap.Footprint{Place: placement.Piece, Openings: doors})
	return nil
}

// stampSettlement lays out and stamps a settlement; its exits are the openings of its footprint. A mark
// of its pieces is the settlement's too, at the first piece stamped with it.
func (s *WorldSource) stampSettlement(reg *content.Registry, id api.ContentID, seed uint64) error {
	layout, err := worldgen.LaySettlement(reg, id, seed)
	if err != nil {
		return err
	}
	for _, placement := range layout.Placements {
		stamped, err := s.Stamp(reg, placement.Piece, placement.At, placement.Rotation)
		if err != nil {
			return fmt.Errorf("settlement %q: %w", id, err)
		}
		for name, cell := range stamped.Marks {
			key := placeMark{place: id, name: name}
			if _, taken := s.marks[key]; !taken {
				s.marks[key] = cell
			}
		}
	}
	s.places = append(s.places, hexmap.Footprint{Place: id, Openings: layout.Exits})
	return nil
}

// Stamp sets a piece into the world with its column 0, row 0 and depth 0 at the cell at, turned about
// that cell as [hexmap.Rotate] turns a hex. Its spawns and landmarks follow everything stamped before
// it, so callers stamp in a fixed order. A refused piece changes nothing. Stamp before a map loads any
// chunk from the source, since a loaded chunk keeps what it was loaded with.
func (s *WorldSource) Stamp(
	reg *content.Registry,
	piece api.ContentID,
	at api.Cell,
	rotation api.Dir,
) (Stamped, error) {
	def, ok := reg.Piece(piece)
	if !ok {
		return Stamped{}, fmt.Errorf("explore: stamp: unknown piece %q", piece)
	}
	if !rotation.Valid() {
		return Stamped{}, fmt.Errorf("explore: stamp %q: rotation %d is not a direction", piece, rotation)
	}
	drawn, err := draw(reg, def.Layout(), s.open, func(col, row, depth int) api.Cell {
		turned := hexmap.Rotate(hexmap.FromOffsetOddR(col, row), rotation)
		return api.Cell{Hex: api.Hex{Q: at.Hex.Q + turned.Q, R: at.Hex.R + turned.R}, Depth: at.Depth + depth}
	})
	if err != nil {
		return Stamped{}, fmt.Errorf("explore: stamp %q: %w", piece, err)
	}
	for _, p := range drawn.cells {
		if !s.bounds.Contains(p.cell) {
			return Stamped{}, fmt.Errorf("explore: stamp %q at %v: %w", piece, p.cell, ErrOutside)
		}
		if s.covered[p.cell] {
			return Stamped{}, fmt.Errorf("explore: stamp %q at %v: %w", piece, p.cell, ErrOverlap)
		}
	}
	s.add(piece, drawn)
	return Stamped{Marks: drawn.marks, Doors: drawn.doors}, nil
}

// add sets a drawing's cells into the world and keeps its spawns and landmarks as those of id.
func (s *WorldSource) add(id api.ContentID, drawn drawing) {
	for _, p := range drawn.cells {
		key := hexmap.KeyOf(p.cell)
		s.authored[key] = append(s.authored[key], p)
		s.covered[p.cell] = true
	}
	s.spawns = append(s.spawns, drawn.spawns...)
	for _, landmark := range drawn.landmarks {
		landmark.Place = id
		s.landmarks = append(s.landmarks, landmark)
	}
}

// Covers reports whether a place or a stamped piece authors the cell. It makes WorldSource the fixed
// layer of worldgen.
func (s *WorldSource) Covers(cell api.Cell) bool { return s.covered[cell] }

// Places lists the openings of every place, settlement and landmark: places in ID order, then
// settlements, then landmarks.
func (s *WorldSource) Places() []hexmap.Footprint { return slices.Clone(s.places) }

// Chunk implements [hexmap.Source].
func (s *WorldSource) Chunk(key hexmap.ChunkKey) (*hexmap.Chunk, error) {
	return filledChunk(key, s.fill, s.authored[key]), nil
}

// Bounds is the box of cells the world holds.
func (s *WorldSource) Bounds() hexmap.Bounds { return s.bounds }

// Edge is the terrain beyond the bounds.
func (s *WorldSource) Edge() hexmap.Terrain { return s.edge }

// Mark returns the world cell of the named mark of a place.
func (s *WorldSource) Mark(place api.ContentID, name string) (api.Cell, bool) {
	cell, ok := s.marks[placeMark{place: place, name: name}]
	return cell, ok
}

// Spawns lists every spawn mark: places in ID order, then pieces in the order stamped, and within each,
// layers as authored, rows and columns.
func (s *WorldSource) Spawns() []Spawn { return slices.Clone(s.spawns) }

// Landmarks lists every landmark mark, in the order of [WorldSource.Spawns].
func (s *WorldSource) Landmarks() []Landmark { return slices.Clone(s.landmarks) }

// ChunkKeys lists the chunks any place or stamped piece covers, sorted by [hexmap.CompareKeys].
func (s *WorldSource) ChunkKeys() []hexmap.ChunkKey {
	keys := slices.Collect(maps.Keys(s.authored))
	slices.SortFunc(keys, hexmap.CompareKeys)
	return keys
}
