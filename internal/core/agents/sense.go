package agents

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/knowledge"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

type sighting struct {
	subject api.UID
	at      api.Cell
	dead    bool
}

// sense has a living body learn what it sees of the other bodies within the content's agent sight: where
// each stands and, if so, that it is dead, as the claims content names. A mind calls it from its own
// timeline entry, so that an agent looks only when it decides and never on every entry of the world.
// What it saw before and no longer sees stays believed until a new sighting replaces it.
func sense(w *sim.World, body ecs.Entity) error {
	cfg := w.Content.Agents()
	positions := ecs.NewMap[comp.Pos](w.ECS)
	deaths := ecs.NewMap[comp.Dead](w.ECS)
	if (cfg.AtClaim == "" && cfg.DeadClaim == "") || w.Map == nil ||
		!w.Alive(body) || !positions.Has(body) || deaths.Has(body) {
		return nil
	}
	from := positions.Get(body).At
	var seen []sighting
	query := ecs.NewFilter2[comp.Pos, comp.Body](w.ECS).Query()
	for query.Next() {
		other := query.Entity()
		pos, _ := query.Get()
		if other == body {
			continue
		}
		visible, err := w.Map.Sees(from, pos.At, cfg.SightRadius, cfg.SightDepth)
		if err != nil {
			query.Close()
			return fmt.Errorf("agents: sight: %w", err)
		}
		if visible {
			seen = append(seen, sighting{subject: w.UIDOf(other), at: pos.At, dead: deaths.Has(other)})
		}
	}
	slices.SortFunc(seen, func(a, b sighting) int { return cmp.Compare(a.subject, b.subject) })
	for _, s := range seen {
		fact := api.Fact{Subject: s.subject, Origin: cfg.SeenOrigin, Confidence: cfg.SeenConfidence}
		if cfg.AtClaim != "" {
			fact.Claim, fact.Value = cfg.AtClaim, knowledge.CellValue(s.at)
			knowledge.Learn(w, body, fact)
		}
		if s.dead && cfg.DeadClaim != "" {
			fact.Claim, fact.Value = cfg.DeadClaim, ""
			knowledge.Learn(w, body, fact)
		}
	}
	return nil
}
