package worldgen

import (
	"cmp"
	"fmt"
	"hash/fnv"
	"slices"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/rng"
)

// SettlementStream names the random stream a settlement's layout draws from.
const SettlementStream = "worldgen"

// Placement is one piece of a settlement: stamp the piece with its column 0, row 0 and depth 0 at At, turned
// by Rotation.
type Placement struct {
	Piece    api.ContentID
	At       api.Cell
	Rotation api.Dir
}

// SettlementLayout is where a settlement's pieces go.
type SettlementLayout struct {
	// Placements are in the order to stamp them; the first is the street the town grows from.
	Placements []Placement
	// Exits are the street door cells left open as ways out to the deep.
	Exits []api.Cell
}

// layAttempts is how many times a layout that fails is drawn again, the stream moving on each time, so that a
// settlement that fits does not hang on one unlucky seed.
const layAttempts = 20

type pieceShape struct {
	cells []shapeCell
	doors []shapeCell
	gives map[api.ContentID]int
}

type shapeCell struct {
	hex   api.Hex
	depth int
}

func shapeOf(reg *content.Registry, def content.PieceDef) pieceShape {
	shape := pieceShape{gives: map[api.ContentID]int{}}
	for _, layer := range def.Layers {
		for row, text := range layer.Rows {
			for col, char := range []rune(text) {
				if char == ' ' {
					continue
				}
				cell := shapeCell{hex: hexmap.FromOffsetOddR(col, row), depth: layer.Depth}
				shape.cells = append(shape.cells, cell)
				if slices.Contains(def.Doors, string(char)) {
					shape.doors = append(shape.doors, cell)
				}
				for _, affordance := range def.Affordances[string(char)] {
					shape.gives[affordance]++
				}
				if template, ok := reg.Template(def.Spawns[string(char)]); ok {
					for _, affordance := range template.Affordances {
						shape.gives[affordance]++
					}
				}
			}
		}
	}
	return shape
}

func (s shapeCell) at(origin api.Cell, rotation api.Dir) api.Cell {
	turned := hexmap.Rotate(s.hex, rotation)
	return api.Cell{Hex: api.Hex{Q: origin.Hex.Q + turned.Q, R: origin.Hex.R + turned.R}, Depth: origin.Depth + s.depth}
}

type openDoor struct {
	cell   api.Cell
	street bool
}

type settler struct {
	def      content.SettlementDef
	street   api.ContentID
	center   api.Hex
	bounds   hexmap.Bounds
	gen      *rng.Generator
	shapes   map[api.ContentID]pieceShape
	occupied map[api.Cell]bool
	open     []openDoor
	used     map[api.ContentID]int
	gives    map[api.ContentID]int
	out      SettlementLayout
}

// LaySettlement lays out the settlement as a street graph grown from its centre, with pieces attached at
// their doors, and picks its exits from the street doors left open. It is a pure function of the world seed
// and the content: the export count never enters it, so an export cannot rearrange a town. It fails when the
// pieces cannot satisfy the settlement's landmark minimums or exits.
func LaySettlement(reg *content.Registry, id api.ContentID, seed uint64) (SettlementLayout, error) {
	def, ok := reg.Settlement(id)
	if !ok {
		return SettlementLayout{}, fmt.Errorf("worldgen: settlement: unknown settlement %q", id)
	}
	world := reg.World()
	hash := fnv.New64a()
	_, _ = hash.Write([]byte(id))
	s := &settler{
		def:    def,
		street: def.Street,
		center: hexmap.FromOffsetOddR(def.At.Col, def.At.Row),
		bounds: hexmap.Bounds{
			MaxCol: world.Cols - 1, MaxRow: world.Rows - 1, MinDepth: world.MinDepth, MaxDepth: world.MaxDepth,
		},
		gen:      rng.New(seed).Derived(SettlementStream, hash.Sum64()),
		shapes:   map[api.ContentID]pieceShape{},
		occupied: map[api.Cell]bool{},
		used:     map[api.ContentID]int{},
		gives:    map[api.ContentID]int{},
	}
	for pieceID := range def.Pieces {
		piece, ok := reg.Piece(api.ContentID(pieceID))
		if !ok {
			return SettlementLayout{}, fmt.Errorf("worldgen: settlement %q: unknown piece %q", id, pieceID)
		}
		s.shapes[api.ContentID(pieceID)] = shapeOf(reg, piece)
	}
	var err error
	for range layAttempts {
		if err = s.lay(); err == nil {
			return s.out, nil
		}
		s.occupied, s.used, s.gives = map[api.Cell]bool{}, map[api.ContentID]int{}, map[api.ContentID]int{}
		s.open, s.out = nil, SettlementLayout{}
	}
	return SettlementLayout{}, fmt.Errorf("worldgen: settlement %q: %w", id, err)
}

func (s *settler) lay() error {
	root := Placement{
		Piece:    s.street,
		At:       api.Cell{Hex: s.center, Depth: s.def.At.Depth},
		Rotation: api.Dir(s.gen.IntN(api.DirCount)),
	}
	shape := s.shapes[s.street]
	if !s.fits(shape, root.At, root.Rotation) {
		return fmt.Errorf("the street %q does not fit at the centre", s.street)
	}
	s.place(root, shape, -1, nil)
	for {
		candidates := s.candidates()
		if deficit := s.deficit(); len(deficit) > 0 {
			scarce := s.scarcest(deficit)
			candidates = slices.DeleteFunc(candidates, func(piece api.ContentID) bool {
				return s.shapes[piece].gives[scarce] == 0
			})
			if len(candidates) == 0 {
				return fmt.Errorf("the pieces cannot give the landmarks %v", deficit)
			}
		}
		placed := false
		for len(candidates) > 0 && !placed {
			pick := s.pickPiece(candidates)
			placed = s.attach(candidates[pick])
			candidates = slices.Delete(candidates, pick, pick+1)
		}
		// A street gives no landmark, so without this fallback it would never grow once a landmark is lacking.
		if !placed && s.used[s.street] < s.def.Pieces[string(s.street)].Count {
			placed = s.attach(s.street)
		}
		if !placed {
			if deficit := s.deficit(); len(deficit) > 0 {
				return fmt.Errorf("no room left for the landmarks %v", deficit)
			}
			break
		}
	}
	return s.chooseExits()
}

// scarcest returns the affordance of the deficit that the settlement's pieces can give the fewest of in all,
// so the landmark with the least room to spare is laid before the plentiful ones take the doors.
func (s *settler) scarcest(deficit []api.ContentID) api.ContentID {
	supply := func(affordance api.ContentID) int {
		total := 0
		for pieceID, use := range s.def.Pieces {
			total += use.Count * s.shapes[api.ContentID(pieceID)].gives[affordance]
		}
		return total
	}
	return slices.MinFunc(deficit, func(a, b api.ContentID) int {
		return cmp.Or(cmp.Compare(supply(a), supply(b)), cmp.Compare(a, b))
	})
}

// candidates lists the pieces with uses left, in ID order.
func (s *settler) candidates() []api.ContentID {
	var ids []api.ContentID
	for pieceID, use := range s.def.Pieces {
		if s.used[api.ContentID(pieceID)] < use.Count {
			ids = append(ids, api.ContentID(pieceID))
		}
	}
	slices.Sort(ids)
	return ids
}

// deficit lists the affordances the layout still lacks landmarks for, in ID order.
func (s *settler) deficit() []api.ContentID {
	var ids []api.ContentID
	for affordance, least := range s.def.LandmarkMin {
		if s.gives[api.ContentID(affordance)] < least {
			ids = append(ids, api.ContentID(affordance))
		}
	}
	slices.Sort(ids)
	return ids
}

func (s *settler) pickPiece(candidates []api.ContentID) int {
	total := 0
	for _, piece := range candidates {
		total += s.def.Pieces[string(piece)].Weight
	}
	roll := s.gen.IntN(total)
	for i, piece := range candidates {
		roll -= s.def.Pieces[string(piece)].Weight
		if roll < 0 {
			return i
		}
	}
	return len(candidates) - 1
}

type attachment struct {
	door      int
	placement Placement
	joined    int
}

// attach stamps the piece with one of its doors beside an open door, chosen from every way it fits. It
// reports false when it fits nowhere.
func (s *settler) attach(piece api.ContentID) bool {
	shape := s.shapes[piece]
	var options []attachment
	for i, door := range s.open {
		for rotation := range api.Dir(api.DirCount) {
			for joined, own := range shape.doors {
				for dir := range api.Dir(api.DirCount) {
					beside := hexmap.Prisms{}.Neighbor(door.cell, dir).Cell
					turned := hexmap.Rotate(own.hex, rotation)
					at := api.Cell{
						Hex:   api.Hex{Q: beside.Hex.Q - turned.Q, R: beside.Hex.R - turned.R},
						Depth: beside.Depth - own.depth,
					}
					if s.fits(shape, at, rotation) {
						options = append(options, attachment{
							door:      i,
							placement: Placement{Piece: piece, At: at, Rotation: rotation},
							joined:    joined,
						})
					}
				}
			}
		}
	}
	if len(options) == 0 {
		return false
	}
	pick := options[s.gen.IntN(len(options))]
	s.place(pick.placement, shape, pick.door, &pick.joined)
	return true
}

// fits reports whether every cell of the shape is inside the world and the radius and free.
func (s *settler) fits(shape pieceShape, at api.Cell, rotation api.Dir) bool {
	for _, cell := range shape.cells {
		world := cell.at(at, rotation)
		if !s.bounds.Contains(world) || s.occupied[world] ||
			hexmap.Distance(world.Hex, s.center) > s.def.Radius {
			return false
		}
	}
	return true
}

// place records the placement; door is the open door it joins, or -1 for the first piece, and joined is the
// index of its own door that does the joining.
func (s *settler) place(placement Placement, shape pieceShape, door int, joined *int) {
	for _, cell := range shape.cells {
		s.occupied[cell.at(placement.At, placement.Rotation)] = true
	}
	if door >= 0 {
		s.open = slices.Delete(s.open, door, door+1)
	}
	for i, own := range shape.doors {
		if joined != nil && i == *joined {
			continue
		}
		s.open = append(
			s.open,
			openDoor{cell: own.at(placement.At, placement.Rotation), street: placement.Piece == s.street},
		)
	}
	s.used[placement.Piece]++
	for affordance, n := range shape.gives {
		s.gives[affordance] += n
	}
	s.out.Placements = append(s.out.Placements, placement)
}

// chooseExits picks the exits from the open street doors that have a free neighbor in the world.
func (s *settler) chooseExits() error {
	var doors []api.Cell
	for _, door := range s.open {
		if !door.street {
			continue
		}
		for dir := range api.Dir(api.DirCount) {
			outside := hexmap.Prisms{}.Neighbor(door.cell, dir).Cell
			if s.bounds.Contains(outside) && !s.occupied[outside] {
				doors = append(doors, door.cell)
				break
			}
		}
	}
	if len(doors) < s.def.Exits {
		return fmt.Errorf("only %d street doors are free for %d exits", len(doors), s.def.Exits)
	}
	for i := range s.def.Exits {
		j := i + s.gen.IntN(len(doors)-i)
		doors[i], doors[j] = doors[j], doors[i]
	}
	s.out.Exits = doors[:s.def.Exits]
	slices.SortFunc(s.out.Exits, hexmap.CompareCells)
	return nil
}
