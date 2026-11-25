package content

import "testing"

func TestFamiliesAndPopulationsAreChecked(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Bundle)
		want   string
	}{
		{"no templates", func(b *Bundle) {
			b.Families["family.test.clay"] = FamilyDef{Name: "text.family.test.clay"}
		}, `families["family.test.clay"].templates: at least one is required`},
		{"an unknown template", func(b *Bundle) {
			b.Families["family.test.clay"] = FamilyDef{
				Name: "text.family.test.clay", Templates: []Weighted{{Template: "template.nope", Weight: 1}},
			}
		}, `unknown template "template.nope"`},
		{"a zero weight", func(b *Bundle) {
			b.Families["family.test.clay"] = FamilyDef{
				Name: "text.family.test.clay", Templates: []Weighted{{Template: "template.creature.beast"}},
			}
		}, `weight: must be at least 1`},
		{"an unknown family in a population", func(b *Bundle) {
			def := b.Maps["map.test.gate_a"]
			def.Region = &RegionDef{Populations: map[string]int{"family.nope": 1}}
			b.Maps["map.test.gate_a"] = def
		}, `unknown family "family.nope"`},
		{"a negative population", func(b *Bundle) {
			def := b.Maps["map.test.gate_a"]
			def.Region = &RegionDef{Populations: map[string]int{"family.test.clay": -1}}
			b.Maps["map.test.gate_a"] = def
		}, `must not be negative`},
		{"a drift formula that does not compile", func(b *Bundle) {
			def := b.Families["family.test.clay"]
			def.Drift = "nope +"
			b.Families["family.test.clay"] = def
		}, `families["family.test.clay"].drift`},
		{"an attrition formula over an unknown variable", func(b *Bundle) {
			def := b.Families["family.test.clay"]
			def.Attrition = "nothing"
			b.Families["family.test.clay"] = def
		}, `families["family.test.clay"].attrition`},
		{"a town that uses stock without residents", func(b *Bundle) {
			def := b.Maps["map.play.hub"]
			def.Region.Town.Use = []StackDef{{Item: "item.lake_water", Count: 1}}
			b.Maps["map.play.hub"] = def
		}, `needs residents`},
		{"an exodus above one", func(b *Bundle) {
			def := b.Maps["map.play.hub"]
			def.Region.Town.Exodus = 2
			b.Maps["map.play.hub"] = def
		}, `exodus: must be from 0 to 1`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bundle := sampleBundle(t)
			tt.mutate(&bundle)
			_, err := New(bundle)
			requireErrContains(t, err, tt.want)
		})
	}
}
