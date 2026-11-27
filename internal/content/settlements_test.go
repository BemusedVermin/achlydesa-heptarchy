package content

import (
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

func TestSettlementsAreChecked(t *testing.T) {
	tests := []struct {
		name string
		edit func(*SettlementDef)
		want string
	}{
		{
			"street not among the pieces",
			func(s *SettlementDef) { s.Street = "piece.house" },
			`.street: "piece.house" is not one of the pieces`,
		},
		{
			"unknown piece",
			func(s *SettlementDef) { s.Pieces["piece.nope"] = SettlementPiece{Weight: 1, Count: 1} },
			`unknown piece "piece.nope"`,
		},
		{
			"zero weight",
			func(s *SettlementDef) { s.Pieces["piece.street"] = SettlementPiece{Count: 1} },
			"weight and count must be at least 1",
		},
		{"too wide for the world", func(s *SettlementDef) { s.Radius = 500 }, "reaches outside the world"},
		{
			"unknown landmark",
			func(s *SettlementDef) { s.LandmarkMin["affordance.nope"] = 1 },
			`unknown affordance "affordance.nope"`,
		},
		{
			"on a place",
			func(s *SettlementDef) { s.At = SettlementAt{Col: 160, Row: 190, Depth: 6} },
			"lies within the settlement's radius",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bundle := sampleBundle(t)
			bundle.Pieces = map[string]PieceDef{"piece.street": {
				Name:   "text.piece.street",
				Layers: []LayerDef{{Depth: 0, Rows: []string{"+.+"}}},
				Legend: map[string]api.ContentID{".": "terrain.air"},
				Marks:  map[string]string{"+": "door"},
				Doors:  []string{"+"},
			}}
			bundle.Text["text.piece.street"] = "a street"
			bundle.Text["text.settlement.a"] = "a town"
			def := SettlementDef{
				Name: "text.settlement.a", At: SettlementAt{Col: 50, Row: 20}, Radius: 5, Street: "piece.street",
				Pieces:      map[string]SettlementPiece{"piece.street": {Weight: 1, Count: 1}},
				LandmarkMin: map[string]int{},
			}
			test.edit(&def)
			bundle.Settlements = map[string]SettlementDef{"settlement.a": def}
			_, err := New(bundle)
			requireErrContains(t, err, test.want)
		})
	}
}
