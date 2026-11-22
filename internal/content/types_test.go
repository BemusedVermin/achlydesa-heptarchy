package content

import (
	"strings"
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

func TestTypeTicksAndSampleForms(t *testing.T) {
	// The sample content's numbers are placeholders, so the test sets the ones it checks.
	bundle, err := LoadBundle(sampleDir)
	if err != nil {
		t.Fatal(err)
	}
	heavy := bundle.Types["type.soak"]
	heavy.Windup, heavy.Recovery = "1", "2"
	bundle.Types["type.soak"] = heavy
	plain := bundle.Types["type.light"]
	plain.Windup, plain.Recovery = "", ""
	bundle.Types["type.light"] = plain
	reg, err := New(bundle)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		id               api.ContentID
		windup, recovery int64
	}{
		{"type.soak", 1, 2},
		{"type.light", 0, 0},
	}
	for _, tt := range tests {
		windup, recovery, err := reg.TypeTicks(tt.id, ActionEnv{})
		if err != nil || windup != tt.windup || recovery != tt.recovery {
			t.Errorf("%s ticks = %d, %d, %v; want %d, %d", tt.id, windup, recovery, err, tt.windup, tt.recovery)
		}
	}
	for _, id := range []api.ContentID{
		"action.strike", "action.sweep", "action.sever", "action.pin", "action.stand", "action.give", "action.graft",
	} {
		if _, ok := reg.Action(id); !ok {
			t.Errorf("the sample content has no %s", id)
		}
	}
}

func TestLoadRejectsBadTypesAndCombatants(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Bundle)
		want   string
	}{
		{"type windup does not compile", func(b *Bundle) {
			def := b.Types["type.soak"]
			def.Windup = "2 +"
			b.Types["type.soak"] = def
		}, `types["type.soak"].windup`},
		{"type recovery is not an integer", func(b *Bundle) {
			def := b.Types["type.soak"]
			def.Recovery = "true"
			b.Types["type.soak"] = def
		}, `types["type.soak"].recovery`},
		{"combatant of an unknown faction", func(b *Bundle) {
			template := b.Templates["template.person.reaper"]
			template.Combatant = &CombatantDef{Faction: "faction.nope"}
			b.Templates["template.person.reaper"] = template
		}, `unknown faction "faction.nope"`},
		{"combatant with an unknown type", func(b *Bundle) {
			template := b.Templates["template.person.reaper"]
			template.Combatant = &CombatantDef{Faction: "faction.party", Types: []api.ContentID{"type.nope"}}
			b.Templates["template.person.reaper"] = template
		}, `unknown type "type.nope"`},
		{"combatant with a consumable type as innate", func(b *Bundle) {
			template := b.Templates["template.person.reaper"]
			template.Combatant = &CombatantDef{Faction: "faction.party", Types: []api.ContentID{"type.soak"}}
			b.Templates["template.person.reaper"] = template
		}, `type "type.soak" is not innate`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bundle, err := LoadBundle(sampleDir)
			if err != nil {
				t.Fatal(err)
			}
			tt.mutate(&bundle)
			if _, err := New(bundle); err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Errorf("err = %v, want it to contain %q", err, tt.want)
			}
		})
	}
}
