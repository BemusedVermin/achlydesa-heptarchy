package effects

import (
	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

// SideOf is what a formula sees of a body and one of its parts: their records. An entity that is
// gone, zero or has no such component reads as an empty record.
func SideOf(w *sim.World, body, part ecs.Entity) content.Side {
	side := content.Side{
		Template: content.EmptyRecord(), Part: content.EmptyRecord(), Material: content.EmptyRecord(),
	}
	if alive(w, body) {
		if identities := ecs.NewMap[comp.Identity](w.ECS); identities.Has(body) {
			side.Template, _ = w.Content.Record(identities.Get(body).Template)
		}
		grafts := ecs.NewMap[comp.Graft](w.ECS)
		for _, attached := range anatomy.Parts(w, body) {
			if grafts.Has(attached) {
				side.Grafts++
			}
		}
	}
	if alive(w, part) {
		if parts := ecs.NewMap[comp.Part](w.ECS); parts.Has(part) {
			data := parts.Get(part)
			side.Part, _ = w.Content.Record(data.Template)
			side.Material, _ = w.Content.Record(data.Material)
		}
	}
	return side
}

// ChosenRecords reads each choice's value as the record it names. A value that names no record,
// such as a direction, is a record with just that ID.
func ChosenRecords(w *sim.World, choices map[string]string) map[string]content.Record {
	out := make(map[string]content.Record, len(choices))
	for name, value := range choices {
		out[name], _ = w.Content.Record(api.ContentID(value))
	}
	return out
}

// firstPart is the first channel that is a part of a body, or the zero entity.
func firstPart(channels []timeline.Channel) ecs.Entity {
	for _, c := range channels {
		if !c.Body.IsZero() && c.Entity != c.Body {
			return c.Entity
		}
	}
	return ecs.Entity{}
}
