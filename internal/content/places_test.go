package content

import (
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

func testPlace(id string, at PlaceAt) PlaceDef {
	return PlaceDef{
		Name:   api.ContentID("text." + id),
		At:     at,
		Layers: []LayerDef{{Depth: 0, Rows: []string{"#..", "#.."}}},
		Legend: map[string]api.ContentID{"#": "terrain.packed_clay", ".": "terrain.air"},
	}
}

func TestPlacesAreChecked(t *testing.T) {
	tests := []struct {
		name          string
		places        map[string]PlaceAt
		wantFragments []string
	}{
		{
			name:          "out of bounds",
			places:        map[string]PlaceAt{"place.a": {Col: 62, Row: 0}},
			wantFragments: []string{`places["place.a"]`, "column 64, row 0, depth 0 is outside the world"},
		},
		{
			name:          "too deep",
			places:        map[string]PlaceAt{"place.a": {Depth: 8}},
			wantFragments: []string{`places["place.a"]`, "outside the world"},
		},
		{
			name:          "overlapping",
			places:        map[string]PlaceAt{"place.a": {Col: 4, Row: 4}, "place.b": {Col: 6, Row: 4}},
			wantFragments: []string{`places["place.b"]`, `also authored by place "place.a"`},
		},
		{
			name:          "odd row",
			places:        map[string]PlaceAt{"place.a": {Row: 3}},
			wantFragments: []string{`places["place.a"].at.row`, "odd"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bundle := sampleBundle(t)
			bundle.Places = map[string]PlaceDef{}
			for id, at := range test.places {
				def := testPlace(id, at)
				bundle.Places[id] = def
				bundle.Text[string(def.Name)] = "a place"
			}
			_, err := New(bundle)
			requireErrContains(t, err, test.wantFragments...)
		})
	}
}
