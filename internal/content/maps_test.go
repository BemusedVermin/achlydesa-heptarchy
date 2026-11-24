package content

import "testing"

func TestATownsStockAndStockMarksAreChecked(t *testing.T) {
	tests := []struct {
		name string
		town TownDef
		want string
	}{
		{"an unknown item", TownDef{Stock: []StackDef{{Item: "item.nope", Count: 1}}},
			`maps["map.test.gate_a"].region.town.stock[0].item: unknown item "item.nope"`},
		{"an empty stack", TownDef{Stock: []StackDef{{Item: "item.ration", Count: 0}}},
			`maps["map.test.gate_a"].region.town.stock[0].count`},
		{"an unknown mark", TownDef{StockMarks: []string{"nowhere"}},
			`maps["map.test.gate_a"].region.town.stock_marks[0]: "nowhere" is not a plain mark`},
		{"a gate", TownDef{StockMarks: []string{"gate_to_b"}},
			`maps["map.test.gate_a"].region.town.stock_marks[0]: "gate_to_b" is not a plain mark`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bundle := sampleBundle(t)
			def := bundle.Maps["map.test.gate_a"]
			def.Region = &RegionDef{Town: &tt.town}
			bundle.Maps["map.test.gate_a"] = def
			_, err := New(bundle)
			requireErrContains(t, err, tt.want)
		})
	}
}

func TestARegionsDataIsTheMapsRecord(t *testing.T) {
	bundle := sampleBundle(t)
	def := bundle.Maps["map.test.gate_a"]
	def.Region = &RegionDef{Data: Data{Props: map[string]float64{"decay": 2}}}
	bundle.Maps["map.test.gate_a"] = def
	reg, err := New(bundle)
	if err != nil {
		t.Fatal(err)
	}
	if got := reg.dataOf(CollectionMaps)["map.test.gate_a"].Props["decay"]; got != 2 {
		t.Errorf("decay %v, want 2", got)
	}
}
