package combat_test

import (
	"maps"
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
)

// TestEachCreatureOfTheDenSpawnsAndFightsBack starts on the den once for each of its creatures, the
// beast and the remnant, and checks it is there and starts a swing at the party.
func TestEachCreatureOfTheDenSpawnsAndFightsBack(t *testing.T) {
	for _, template := range []api.ContentID{"template.creature.beast", "template.creature.remnant"} {
		t.Run(string(template), func(t *testing.T) {
			f := newField(t, "map.test.den", func(b *content.Bundle) {
				for _, id := range []string{"template.person.reaper", "template.person.salken"} {
					person := b.Templates[id]
					person.Actions = []api.ContentID{"action.step", "action.strike", "action.sweep"}
					b.Templates[id] = person
				}
				// One creature at a time: with both, the party is gone before the slower one arrives.
				den := b.Maps["map.test.den"]
				den.Spawns = maps.Clone(den.Spawns)
				for mark, spawned := range den.Spawns {
					if spawned != template {
						delete(den.Spawns, mark)
					}
				}
				b.Maps["map.test.den"] = den
			})
			var creature api.UID
			query := ecs.NewFilter1[comp.Identity](f.w.ECS).Query()
			for query.Next() {
				if identity := query.Get(); identity.Template == template {
					creature = identity.UID
				}
			}
			if creature == 0 {
				t.Fatalf("no %s in the den", template)
			}
			if !f.until(3000, func() bool { return f.movingClayStarted(creature) }) {
				t.Errorf("the %s never swung at the party", template)
			}
		})
	}
}
