package anatomy_test

import (
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/anatomy"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

func TestHostileIsSymmetricWhenOnlyOneFactionListsTheOther(t *testing.T) {
	r := newRig(t, func(b *content.Bundle) {
		b.Factions["faction.a"] = content.Faction{Name: "text.faction.a", HostileTo: []api.ContentID{"faction.b"}}
		b.Factions["faction.b"] = content.Faction{Name: "text.faction.b"}
		b.Factions["faction.c"] = content.Faction{Name: "text.faction.c"}
		b.Text["text.faction.a"], b.Text["text.faction.b"], b.Text["text.faction.c"] = "a", "b", "c"
		for _, id := range []string{"a", "b", "c"} {
			faction := api.ContentID("faction." + id)
			b.Templates["template.test.side_"+id] = content.Template{Faction: &faction}
		}
		b.Templates["template.test.nobody"] = content.Template{}
	})
	spawn := func(id string) ecs.Entity {
		e, err := r.w.Spawn(api.ContentID("template.test."+id), nil)
		if err != nil {
			t.Fatal(err)
		}
		return e
	}
	a, b, c, other, nobody := spawn("side_a"), spawn("side_b"), spawn("side_c"), spawn("side_a"), spawn("nobody")
	tests := []struct {
		name string
		x, y ecs.Entity
		want bool
	}{
		{"the lister to the listed", a, b, true},
		{"the listed to the lister", b, a, true},
		{"unrelated sides", b, c, false},
		{"the same side", a, other, false},
		{"no allegiance", a, nobody, false},
		{"no allegiance either way", nobody, a, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := anatomy.Hostile(r.w, tt.x, tt.y); got != tt.want {
				t.Errorf("Hostile = %v, want %v", got, tt.want)
			}
		})
	}
}
