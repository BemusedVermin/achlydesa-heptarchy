package content

import (
	"fmt"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
)

// checkSettlements validates every settlement: its pieces and street exist, its disc of ground lies
// inside the world, and the disc, with a hex of margin for the ways out, touches no place and no other
// settlement.
func (r *Registry) checkSettlements() []error {
	p := r.problems()
	world := r.bundle.World
	keys := sortedKeys(r.bundle.Settlements)
	centers := map[string]api.Hex{}
	for _, key := range keys {
		def := r.bundle.Settlements[key]
		path := fmt.Sprintf("settlements[%q]", key)
		p.text(path+".name", def.Name)
		centers[key] = hexmap.FromOffsetOddR(def.At.Col, def.At.Row)
		if def.Radius < 1 {
			p.addf("%s.radius: %d is less than 1", path, def.Radius)
		}
		if def.Exits < 0 {
			p.addf("%s.exits: %d is negative", path, def.Exits)
		}
		if _, ok := def.Pieces[string(def.Street)]; !ok {
			p.addf("%s.street: %q is not one of the pieces", path, def.Street)
		}
		depths := map[int]bool{def.At.Depth: true}
		for _, id := range sortedKeys(def.Pieces) {
			use := def.Pieces[id]
			at := fmt.Sprintf("%s.pieces[%q]", path, id)
			piece, _ := ref(p, r.bundle.Pieces, at, "piece", api.ContentID(id))
			if use.Weight < 1 || use.Count < 1 {
				p.addf("%s: weight and count must be at least 1, got %d and %d", at, use.Weight, use.Count)
			}
			for _, layer := range piece.Layers {
				depths[def.At.Depth+layer.Depth] = true
			}
		}
		for _, id := range sortedKeys(def.LandmarkMin) {
			at := fmt.Sprintf("%s.landmark_min[%q]", path, id)
			ref(p, r.bundle.Affordances, at, "affordance", api.ContentID(id))
			if def.LandmarkMin[id] < 1 {
				p.addf("%s: %d is less than 1", at, def.LandmarkMin[id])
			}
		}
		for depth := range depths {
			if depth < world.MinDepth || depth > world.MaxDepth {
				p.addf("%s: depth %d is outside the world", path, depth)
			}
		}
		if !r.discInside(centers[key], def.Radius) {
			p.addf("%s: radius %d reaches outside the world", path, def.Radius)
		}
		r.checkSettlementPlaces(p, path, def, centers[key], depths)
		for _, other := range keys {
			if other >= key {
				break
			}
			gap := hexmap.Distance(centers[key], centers[other])
			if gap <= def.Radius+r.bundle.Settlements[other].Radius+1 {
				p.addf("%s: overlaps settlement %q", path, other)
			}
		}
	}
	return p.errs
}

func (r *Registry) discInside(center api.Hex, radius int) bool {
	world := r.bundle.World
	for dr := -radius; dr <= radius; dr++ {
		for dq := -radius; dq <= radius; dq++ {
			if hexmap.Distance(api.Hex{}, api.Hex{Q: dq, R: dr}) > radius {
				continue
			}
			col, row := hexmap.ToOffsetOddR(api.Hex{Q: center.Q + dq, R: center.R + dr})
			if col < 0 || col >= world.Cols || row < 0 || row >= world.Rows {
				return false
			}
		}
	}
	return true
}

func (r *Registry) checkSettlementPlaces(
	p *problems,
	path string,
	def SettlementDef,
	center api.Hex,
	depths map[int]bool,
) {
	for _, key := range sortedKeys(r.bundle.Places) {
		place := r.bundle.Places[key]
		for _, layer := range place.Layers {
			if !depths[place.At.Depth+layer.Depth] {
				continue
			}
			for row, text := range layer.Rows {
				for col, char := range []rune(text) {
					cell := hexmap.FromOffsetOddR(place.At.Col+col, place.At.Row+row)
					if char != ' ' && hexmap.Distance(cell, center) <= def.Radius+1 {
						p.addf("%s: place %q lies within the settlement's radius", path, key)
						return
					}
				}
			}
		}
	}
}
