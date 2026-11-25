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

// BodyEnvOf is what a formula sees of a body and one of its parts: their records and facts. An entity that is
// gone, zero or has no such component reads as an empty record.
func BodyEnvOf(w *sim.World, body, part ecs.Entity) content.BodyEnv {
	env := content.BodyEnv{
		Template: content.EmptyRecord(), Part: content.EmptyRecord(), Material: content.EmptyRecord(),
	}
	if w.Alive(body) {
		if identities := ecs.NewMap[comp.Identity](w.ECS); identities.Has(body) {
			env.Template, _ = w.Content.Record(identities.Get(body).Template)
		}
		env.WindingUp = len(timeline.PendingOf(w, w.UIDOf(body))) > 0
		if statuses := ecs.NewMap[comp.Statuses](w.ECS); statuses.Has(body) {
			for _, entry := range statuses.Get(body).Active {
				env.Statuses = append(env.Statuses, string(entry.Status))
			}
		}
		grafts := ecs.NewMap[comp.Graft](w.ECS)
		for _, attached := range anatomy.Parts(w, body) {
			if grafts.Has(attached) {
				env.Grafts++
			}
		}
	}
	if w.Alive(part) {
		if parts := ecs.NewMap[comp.Part](w.ECS); parts.Has(part) {
			data := parts.Get(part)
			env.Part, _ = w.Content.Record(data.Template)
			env.Material, _ = w.Content.Record(data.Material)
			env.Integrity, env.MaxIntegrity = data.Integrity, data.MaxIntegrity
			env.Graft = ecs.NewMap[comp.Graft](w.ECS).Has(part)
			env.Locked = timeline.Locked(w, timeline.Channel{Entity: part, Body: body})
		}
	}
	return env
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
