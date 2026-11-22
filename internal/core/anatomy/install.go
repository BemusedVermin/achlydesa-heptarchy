package anatomy

import (
	"errors"
	"fmt"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/timeline"
)

// bodyTag and partTag tag the channels of a body and of its parts.
const (
	bodyTag  = "body"
	partTag  = "part"
	grantTag = "grant:"
)

// Install registers the body and protected spawners and the timeline's channel and speed hooks.
// It must run after timeline.Install, and before any entity with a body is spawned.
func Install(w *sim.World) error {
	if w.Content == nil {
		return errors.New("anatomy: install: world has no content")
	}
	if err := w.RegisterSpawner(sim.Spawner{Name: "body", Apply: spawnBody}); err != nil {
		return err
	}
	if err := w.RegisterSpawner(sim.Spawner{Name: "protected", Apply: spawnProtected}); err != nil {
		return err
	}
	timeline.SetChannels(w, channelsOf)
	timeline.SetSpeed(w, speedOf)
	return nil
}

// spawnBody builds the body's parts from its plan: one part entity per plan entry, each PartOf the
// body, in plan order so UIDs follow it. Parts have no position while they are attached.
func spawnBody(w *sim.World, body ecs.Entity, t content.Template) error {
	if t.Body == nil {
		return nil
	}
	plan, ok := w.Content.Body(*t.Body)
	if !ok {
		return fmt.Errorf("unknown body plan %q", *t.Body)
	}
	bodies := ecs.NewMap[comp.Body](w.ECS)
	partData := ecs.NewMap[comp.Part](w.ECS)
	partOf := ecs.NewMap[comp.PartOf](w.ECS)
	bodies.Add(body, &comp.Body{Plan: *t.Body})
	for i, entry := range plan.Parts {
		def, ok := w.Content.Part(entry.Part)
		if !ok {
			return fmt.Errorf("body plan %q part %d: unknown part %q", *t.Body, i, entry.Part)
		}
		material := def.Material
		if entry.Material != "" {
			material = entry.Material
		}
		part := w.NewIdentified(entry.Part)
		partData.Add(part, &comp.Part{
			Template: entry.Part, Slot: def.Slot, Material: material,
			Integrity: def.Integrity, MaxIntegrity: def.Integrity,
			Grants: append([]comp.Grant(nil), def.Grants...), Vital: def.Vital,
		})
		partOf.Add(part, &comp.PartOf{}, body)
	}
	timeline.Reschedule(w, body)
	return nil
}

// spawnProtected gives plot armor to the templates that ask for it.
func spawnProtected(w *sim.World, e ecs.Entity, t content.Template) error {
	if t.Protected {
		ecs.NewMap[comp.Protected](w.ECS).Add(e, &comp.Protected{})
	}
	return nil
}

// channelsOf is the timeline's channel hook: the body, then one channel per attached part, tagged
// "part", its slot and "grant:<kind>" for each grant. An entity with no body has just its own channel.
func channelsOf(w *sim.World, body ecs.Entity) []timeline.Channel {
	channels := []timeline.Channel{{Entity: body, Tags: []string{bodyTag}}}
	parts := ecs.NewMap[comp.Part](w.ECS)
	for _, part := range Parts(w, body) {
		data := parts.Get(part)
		tags := []string{partTag, data.Slot}
		for _, grant := range data.Grants {
			tags = append(tags, grantTag+grant.Kind)
		}
		channels = append(channels, timeline.Channel{Entity: part, Body: body, Tags: tags})
	}
	return channels
}

// speedOf is the timeline's speed hook: the sum of the speed grants of the functioning parts, so a
// body with no legs has speed 0. An entity with no body keeps the configured default speed.
func speedOf(w *sim.World, body ecs.Entity) int {
	if !ecs.NewMap[comp.Body](w.ECS).Has(body) {
		return w.Content.Timeline().DefaultSpeed
	}
	return Grants(w, body, content.GrantSpeed)
}
