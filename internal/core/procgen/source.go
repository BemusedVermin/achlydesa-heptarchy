package procgen

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash/fnv"
	"maps"
	"slices"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/rng"
)

// Fixed is what procgen needs of the hand-made layer: the fixed places set in the world's fill.
type Fixed interface {
	hexmap.Source
	// Covers reports whether a cell belongs to a place.
	Covers(api.Cell) bool
	// Places lists every fixed place's open cells, in a fixed order.
	Places() []hexmap.Footprint
}

// Opening is an open cell of a fixed place with a neighbor outside the place: a spot where the place meets
// the generated ground.
type Opening struct {
	Place api.ContentID
	Cell  api.Cell
	// Exit is the first neighbor of Cell, in direction order, that lies in the bounds outside every place.
	Exit api.Cell
}

// Config is what a [Source] generates from.
type Config struct {
	Seed        uint64
	ExportCount uint64
	Bounds      hexmap.Bounds
	Geography   content.GeographyDef
	Stages      *Stages
	Fixed       Fixed
	// Open is the terrain stages carve open cells with.
	Open hexmap.Terrain
}

// Context is all a stage may read while it builds one chunk.
type Context struct {
	Key    hexmap.ChunkKey
	Bounds hexmap.Bounds
	// Seed and ExportCount are the world's, for stages that sample world-space noise: a chunk's Rand differs
	// from its neighbor's, so noise seeded from it would show seams.
	Seed        uint64
	ExportCount uint64
	// Open is the terrain of a carved cell.
	Open hexmap.Terrain
	// Rand depends only on the world seed, the export count, the chunk key and the stage's place in the
	// geography, so a chunk comes out the same whatever order chunks are built in.
	Rand *rng.Generator
	// Params is the stage's own copy.
	Params map[string]float64
	// Fixed must only be read.
	Fixed Fixed
	// Openings are the places' openings, in a fixed order. Must only be read.
	Openings []Opening
	// Siblings holds the params of every stage of the geography by name, the first of a name winning. Must
	// only be read.
	Siblings map[string]map[string]float64
}

type boundStage struct {
	name   string
	stage  Stage
	params map[string]float64
}

// Source is a [hexmap.Source] that runs a geography's stages on the fixed layer's chunks. Fixed cells and
// cells outside the bounds always come out as the fixed layer has them. It never changes after NewSource, so
// chunks may be built on several goroutines at once.
type Source struct {
	seed        uint64
	streams     *rng.Streams
	open        hexmap.Terrain
	exportCount uint64
	bounds      hexmap.Bounds
	fixed       Fixed
	openings    []Opening
	siblings    map[string]map[string]float64
	stages      []boundStage
}

// NewSource checks the geography's stages against cfg.Stages and returns a source for them.
func NewSource(cfg Config) (*Source, error) {
	if cfg.Fixed == nil {
		return nil, errors.New("procgen: source: no fixed layer")
	}
	src := &Source{
		seed:        cfg.Seed,
		streams:     rng.New(cfg.Seed),
		open:        cfg.Open,
		exportCount: cfg.ExportCount,
		bounds:      cfg.Bounds,
		fixed:       cfg.Fixed,
		openings:    openingsOf(cfg.Fixed, cfg.Bounds),
		siblings:    map[string]map[string]float64{},
	}
	if len(cfg.Geography.Stages) > 0 && cfg.Stages == nil {
		return nil, errors.New("procgen: source: no stage registry")
	}
	for i, def := range cfg.Geography.Stages {
		stage, err := cfg.Stages.lookup(def)
		if err != nil {
			return nil, fmt.Errorf("procgen: source: stages[%d]: %w", i, err)
		}
		if _, seen := src.siblings[def.Stage]; !seen {
			src.siblings[def.Stage] = def.Params
		}
		src.stages = append(src.stages, boundStage{name: def.Stage, stage: stage, params: def.Params})
	}
	return src, nil
}

// Chunk implements [hexmap.Source].
func (s *Source) Chunk(key hexmap.ChunkKey) (*hexmap.Chunk, error) {
	fixed, err := s.fixed.Chunk(key)
	if err != nil {
		return nil, fmt.Errorf("procgen: fixed chunk %v: %w", key, err)
	}
	if len(s.stages) == 0 || !s.touchesBounds(key) {
		return fixed, nil
	}
	chunk := copyChunk(fixed)
	for i, bound := range s.stages {
		ctx := &Context{
			Key:         key,
			Bounds:      s.bounds,
			Seed:        s.seed,
			ExportCount: s.exportCount,
			Open:        s.open,
			Rand:        s.streams.Derived("procgen", s.salt(key, i, bound.name)),
			Params:      maps.Clone(bound.params),
			Fixed:       s.fixed,
			Openings:    s.openings,
			Siblings:    s.siblings,
		}
		if err := bound.stage.Run(ctx, chunk); err != nil {
			return nil, fmt.Errorf("procgen: stage %q on chunk %v: %w", bound.name, key, err)
		}
	}
	origin := key.Origin()
	forEachCell(func(q, r, depth int) {
		cell := api.Cell{Hex: api.Hex{Q: origin.Hex.Q + q, R: origin.Hex.R + r}, Depth: origin.Depth + depth}
		if !s.bounds.Contains(cell) || s.fixed.Covers(cell) {
			chunk.Set(q, r, depth, fixed.Get(q, r, depth))
		}
	})
	return chunk, nil
}

// Openings lists where the fixed places meet the generated ground, places in the fixed layer's order.
func (s *Source) Openings() []Opening { return slices.Clone(s.openings) }

func openingsOf(fixed Fixed, bounds hexmap.Bounds) []Opening {
	var openings []Opening
	for _, footprint := range fixed.Places() {
		for _, cell := range footprint.Open {
			for dir := range api.Dir(api.DirCount) {
				exit := hexmap.Prisms{}.Neighbor(cell, dir).Cell
				if bounds.Contains(exit) && !fixed.Covers(exit) {
					openings = append(openings, Opening{Place: footprint.Place, Cell: cell, Exit: exit})
					break
				}
			}
		}
	}
	return openings
}

// salt gives each stage of each chunk its own stream under the export count's.
func (s *Source) salt(key hexmap.ChunkKey, index int, name string) uint64 {
	hash := fnv.New64a()
	var buf []byte
	buf = binary.LittleEndian.AppendUint64(buf, s.exportCount)
	for _, n := range []int{key.Q, key.R, key.Depth, index} {
		buf = binary.LittleEndian.AppendUint64(buf, uint64(int64(n)))
	}
	buf = append(buf, name...)
	_, _ = hash.Write(buf)
	return hash.Sum64()
}

// touchesBounds reports whether any cell of the chunk lies inside the bounds.
func (s *Source) touchesBounds(key hexmap.ChunkKey) bool {
	origin := key.Origin()
	depth := max(origin.Depth, s.bounds.MinDepth)
	if depth >= origin.Depth+hexmap.ChunkSize || depth > s.bounds.MaxDepth {
		return false
	}
	for r := range hexmap.ChunkSize {
		for q := range hexmap.ChunkSize {
			if s.bounds.Contains(api.Cell{Hex: api.Hex{Q: origin.Hex.Q + q, R: origin.Hex.R + r}, Depth: depth}) {
				return true
			}
		}
	}
	return false
}

func copyChunk(from *hexmap.Chunk) *hexmap.Chunk {
	chunk := hexmap.NewChunk(from.Get(0, 0, 0))
	forEachCell(func(q, r, depth int) { chunk.Set(q, r, depth, from.Get(q, r, depth)) })
	return chunk
}

func forEachCell(visit func(q, r, depth int)) {
	for depth := range hexmap.ChunkSize {
		for r := range hexmap.ChunkSize {
			for q := range hexmap.ChunkSize {
				visit(q, r, depth)
			}
		}
	}
}
