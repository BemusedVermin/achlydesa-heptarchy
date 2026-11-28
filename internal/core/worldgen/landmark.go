package worldgen

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/rng"
)

// landmarkTries is how many positions a landmark draws before PlaceLandmark gives up.
const landmarkTries = 256

// PlaceLandmark picks where the geography's landmark at index stands: at a drawn distance from its
// settlement's centre, depth and rotation, with every cell inside the world, outside the settlement's
// radius, and neither on nor beside a covered cell, so it keeps clear of every exit. It is a pure function
// of the world seed, the content and covered: the export count never enters it.
func PlaceLandmark(
	reg *content.Registry,
	geography api.ContentID,
	index int,
	seed uint64,
	covered func(api.Cell) bool,
) (Placement, error) {
	def, ok := reg.Geography(geography)
	if !ok || index < 0 || index >= len(def.Landmarks) {
		return Placement{}, fmt.Errorf("worldgen: landmark: %q has no landmark %d", geography, index)
	}
	landmark := def.Landmarks[index]
	town, ok := reg.Settlement(landmark.Near)
	if !ok {
		return Placement{}, fmt.Errorf("worldgen: landmark %d of %q: unknown settlement %q",
			index, geography, landmark.Near)
	}
	piece, ok := reg.Piece(landmark.Piece)
	if !ok {
		return Placement{}, fmt.Errorf("worldgen: landmark %d of %q: unknown piece %q",
			index, geography, landmark.Piece)
	}
	world := reg.World()
	bounds := hexmap.Bounds{
		MaxCol: world.Cols - 1, MaxRow: world.Rows - 1, MinDepth: world.MinDepth, MaxDepth: world.MaxDepth,
	}
	center := hexmap.FromOffsetOddR(town.At.Col, town.At.Row)
	shape := shapeOf(reg, piece)
	hash := fnv.New64a()
	_, _ = hash.Write(binary.LittleEndian.AppendUint64([]byte(geography), uint64(index)))
	gen := rng.New(seed).Derived(SettlementStream, hash.Sum64())
	roomy := func(cell api.Cell) bool {
		if !bounds.Contains(cell) || covered(cell) || hexmap.Distance(cell.Hex, center) <= town.Radius {
			return false
		}
		for dir := range api.Dir(api.DirCount) {
			beside := hexmap.Prisms{}.Neighbor(cell, dir).Cell
			if !bounds.Contains(beside) || covered(beside) {
				return false
			}
		}
		return true
	}
	for range landmarkTries {
		distance := landmark.MinDistance + gen.IntN(landmark.MaxDistance-landmark.MinDistance+1)
		placement := Placement{
			Piece: landmark.Piece,
			At: api.Cell{
				Hex:   onRing(center, distance, gen.IntN(api.DirCount*distance)),
				Depth: landmark.MinDepth + gen.IntN(landmark.MaxDepth-landmark.MinDepth+1),
			},
			Rotation: api.Dir(gen.IntN(api.DirCount)),
		}
		fits := true
		for _, cell := range shape.cells {
			if !roomy(cell.at(placement.At, placement.Rotation)) {
				fits = false
				break
			}
		}
		if fits {
			return placement, nil
		}
	}
	return Placement{}, fmt.Errorf("worldgen: landmark %d of %q: no room for %q near %q in %d tries",
		index, geography, landmark.Piece, landmark.Near, landmarkTries)
}

// onRing returns the hex at step along the ring of hexes at distance from center, counting from the hex
// distance steps east of it; step is below 6 x distance.
func onRing(center api.Hex, distance, step int) api.Hex {
	if distance == 0 {
		return center
	}
	corner := hexmap.Rotate(api.Hex{Q: distance}, api.Dir(step/distance))
	next := hexmap.Rotate(api.Hex{Q: distance}, api.Dir((step/distance+1)%api.DirCount))
	along := step % distance
	return api.Hex{
		Q: center.Q + corner.Q + (next.Q-corner.Q)*along/distance,
		R: center.R + corner.R + (next.R-corner.R)*along/distance,
	}
}
