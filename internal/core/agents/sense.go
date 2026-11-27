package agents

import (
	"cmp"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/knowledge"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

type sighting struct {
	subject api.UID
	at      api.Cell
	dead    bool
	gives   []api.ContentID
}

// sense has a living body learn what it sees within the content's agent sight, as the claims content
// names: where each other body and each landmark stands, which bodies are dead, what each landmark
// gives, and how decayed the ground is where it stands and at the open cells it sees. A mind calls it
// from its own timeline entry, so that an agent looks only when it decides and never on every entry
// of the world. What it saw before and no longer sees stays believed until a new sighting replaces it.
func sense(w *sim.World, body ecs.Entity) error {
	cfg := w.Content.Agents()
	positions := ecs.NewMap[comp.Pos](w.ECS)
	deaths := ecs.NewMap[comp.Dead](w.ECS)
	if w.Map == nil || !w.Alive(body) || !positions.Has(body) || deaths.Has(body) {
		return nil
	}
	from := positions.Get(body).At
	if cfg.AtClaim != "" || cfg.DeadClaim != "" {
		seen, err := sightings(w, body, from, cfg)
		if err != nil {
			return err
		}
		for _, s := range seen {
			learnSighting(w, body, cfg, s)
		}
	}
	if cfg.GroundClaim != "" {
		return senseGround(w, body, from, cfg)
	}
	return nil
}

// sightings lists the bodies and landmarks other than the looker that it sees, in UID order.
func sightings(w *sim.World, body ecs.Entity, from api.Cell, cfg content.AgentsConfig) ([]sighting, error) {
	deaths := ecs.NewMap[comp.Dead](w.ECS)
	affordances := ecs.NewMap[comp.Affordance](w.ECS)
	var seen []sighting
	look := func(other ecs.Entity, at api.Cell) error {
		if other == body {
			return nil
		}
		visible, err := w.Map.Sees(from, at, cfg.SightRadius, cfg.SightDepth)
		if err != nil {
			return fmt.Errorf("agents: sight: %w", err)
		}
		if visible {
			s := sighting{subject: w.UIDOf(other), at: at, dead: deaths.Has(other)}
			if affordances.Has(other) {
				s.gives = affordances.Get(other).Tags
			}
			seen = append(seen, s)
		}
		return nil
	}
	bodies := ecs.NewFilter2[comp.Pos, comp.Body](w.ECS).Query()
	for bodies.Next() {
		pos, _ := bodies.Get()
		if err := look(bodies.Entity(), pos.At); err != nil {
			bodies.Close()
			return nil, err
		}
	}
	if cfg.GivesClaim != "" {
		marks := ecs.NewFilter2[comp.Pos, comp.Affordance](w.ECS).Without(ecs.C[comp.Body]()).Query()
		for marks.Next() {
			pos, _ := marks.Get()
			if err := look(marks.Entity(), pos.At); err != nil {
				marks.Close()
				return nil, err
			}
		}
	}
	slices.SortFunc(seen, func(a, b sighting) int { return cmp.Compare(a.subject, b.subject) })
	return seen, nil
}

func learnSighting(w *sim.World, body ecs.Entity, cfg content.AgentsConfig, s sighting) {
	fact := api.Fact{Subject: s.subject, Origin: cfg.SeenOrigin, Confidence: cfg.SeenConfidence}
	if cfg.AtClaim != "" {
		fact.Claim, fact.Value = cfg.AtClaim, knowledge.CellValue(s.at)
		knowledge.Learn(w, body, fact)
	}
	if s.dead && cfg.DeadClaim != "" {
		fact.Claim, fact.Value = cfg.DeadClaim, ""
		knowledge.Learn(w, body, fact)
	}
	if len(s.gives) > 0 && cfg.GivesClaim != "" {
		fact.Claim, fact.Value = cfg.GivesClaim, givesValue(s.gives)
		knowledge.Learn(w, body, fact)
	}
}

// givesSeparator joins the affordance IDs of a gives fact's value; an ID never holds one.
const givesSeparator = ","

func givesValue(tags []api.ContentID) string {
	parts := make([]string, len(tags))
	for i, tag := range tags {
		parts[i] = string(tag)
	}
	return strings.Join(parts, givesSeparator)
}

// senseGround learns the decay level of the cell the body stands in and, for each other level among
// the depths it sees, of the nearest open cell it sees at that level (ties by cell order). Recording
// one cell per level, not every cell in sight, keeps a glance from filling the memory: decay varies only
// with depth, so the nearest cell is all a mind needs to head for that ground.
func senseGround(w *sim.World, body ecs.Entity, from api.Cell, cfg content.AgentsConfig) error {
	geography := w.Content.World().Geography
	if geography == "" {
		return nil
	}
	levels := map[int]int64{}
	unfound := map[int64]bool{}
	for depth := from.Depth - cfg.SightDepth; depth <= from.Depth+cfg.SightDepth; depth++ {
		level, err := w.Content.DecayLevel(geography, content.DecayEnv{Depth: depth, Exports: max(w.ExportCount, 0)})
		if err != nil {
			return fmt.Errorf("agents: ground: %w", err)
		}
		levels[depth] = level
		unfound[level] = true
	}
	var candidates []api.Cell
	for distance := 0; distance <= cfg.SightRadius+cfg.SightDepth && len(unfound) > 0; distance++ {
		candidates = candidates[:0]
		for offset := -min(distance, cfg.SightDepth); offset <= min(distance, cfg.SightDepth); offset++ {
			depth, ring := from.Depth+offset, distance-max(offset, -offset)
			if ring <= cfg.SightRadius && unfound[levels[depth]] {
				candidates = appendRing(candidates, api.Cell{Hex: from.Hex, Depth: depth}, ring)
			}
		}
		slices.SortFunc(candidates, hexmap.CompareCells)
		for _, cell := range candidates {
			level := levels[cell.Depth]
			if !unfound[level] {
				continue
			}
			terrain, err := w.Map.At(cell)
			if err != nil {
				return fmt.Errorf("agents: ground: %w", err)
			}
			if terrain.Solid && cell != from {
				continue
			}
			visible, err := w.Map.Sees(from, cell, cfg.SightRadius, cfg.SightDepth)
			if err != nil {
				return fmt.Errorf("agents: sight: %w", err)
			}
			if !visible {
				continue
			}
			delete(unfound, level)
			knowledge.Learn(w, body, api.Fact{
				Cell: &cell, Claim: cfg.GroundClaim, Value: strconv.FormatInt(level, 10),
				Origin: cfg.SeenOrigin, Confidence: cfg.SeenConfidence,
			})
		}
	}
	return nil
}

// ringSteps walks a hex ring's six sides, in axial coordinates.
var ringSteps = [6]api.Hex{{Q: 1, R: -1}, {Q: 1, R: 0}, {Q: 0, R: 1}, {Q: -1, R: 1}, {Q: -1, R: 0}, {Q: 0, R: -1}}

// appendRing appends the cells of center's layer exactly radius hexes from it.
func appendRing(cells []api.Cell, center api.Cell, radius int) []api.Cell {
	if radius == 0 {
		return append(cells, center)
	}
	at := api.Hex{Q: center.Hex.Q - radius, R: center.Hex.R}
	for _, step := range ringSteps {
		for range radius {
			cells = append(cells, api.Cell{Hex: at, Depth: center.Depth})
			at = api.Hex{Q: at.Q + step.Q, R: at.R + step.R}
		}
	}
	return cells
}
