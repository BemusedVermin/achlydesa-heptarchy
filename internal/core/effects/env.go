package effects

import (
	"strings"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

// needPrefix is dropped from a need's ID to give its key in BodyEnv.Needs.
const needPrefix = "need."

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
		if needs := ecs.NewMap[comp.Needs](w.ECS); needs.Has(body) {
			levels := needs.Get(body).Levels
			env.Needs = make(map[string]float64, len(levels))
			for _, level := range levels {
				if _, known := w.Content.Need(level.Need); !known {
					continue
				}
				env.Needs[strings.TrimPrefix(string(level.Need), needPrefix)] = level.Level
			}
		}
		if party, ok := anatomy.PartyOf(w, body); ok {
			if parties := ecs.NewMap[comp.Party](w.ECS); parties.Has(party) {
				env.Alert = parties.Get(party).Alert
			}
		}
		grafts := ecs.NewMap[comp.Graft](w.ECS)
		partData := ecs.NewMap[comp.Part](w.ECS)
		dyingBelow := w.Content.Agents().DyingBelow
		for _, attached := range anatomy.Parts(w, body) {
			if grafts.Has(attached) {
				env.Grafts++
			}
			if !partData.Has(attached) {
				continue
			}
			data := partData.Get(attached)
			if data.Vital && isDying(data, dyingBelow) {
				env.Dying = true
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

// isDying says whether the part's integrity is at or below the fraction of its maximum.
func isDying(part *comp.Part, fraction float64) bool {
	return float64(part.Integrity) <= fraction*float64(part.MaxIntegrity)
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
