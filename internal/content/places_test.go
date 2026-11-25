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
		marks         map[string]string
		wantFragments []string
	}{
		{
			name:          "one mark name for two characters",
			places:        map[string]PlaceAt{"place.a": {}},
			marks:         map[string]string{".": "start", "#": "start"},
			wantFragments: []string{`places["place.a"].marks["."]`, `mark "start" is also named by "#"`},
		},
		{
			name:          "out of bounds",
			places:        map[string]PlaceAt{"place.a": {Col: 68, Row: 0}},
			wantFragments: []string{`places["place.a"]`, "column 70, row 0, depth 0 is outside the world"},
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
			bundle.Start = Start{}
			bundle.Places = map[string]PlaceDef{}
			for id, at := range test.places {
				def := testPlace(id, at)
				def.Marks = test.marks
				bundle.Places[id] = def
				bundle.Text[string(def.Name)] = "a place"
			}
			_, err := New(bundle)
			requireErrContains(t, err, test.wantFragments...)
		})
	}
}

func TestPlaceOpeningsAreChecked(t *testing.T) {
	tests := []struct {
		name     string
		openings []string
		spawns   map[string]api.ContentID
		want     string
	}{
		{name: "not a mark", openings: []string{"x"}, want: "not one of the marks"},
		{name: "in no layer", openings: []string{"o"}, want: "the character is in no layer"},
		{
			name:     "also a spawn",
			openings: []string{"."},
			spawns:   map[string]api.ContentID{".": "template.test.dummy"},
			want:     "also a spawn",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bundle := sampleBundle(t)
			bundle.Start = Start{}
			def := testPlace("place.a", PlaceAt{})
			def.Marks = map[string]string{".": "floor", "o": "way_out"}
			def.Legend = map[string]api.ContentID{"#": "terrain.packed_clay"}
			def.Spawns = test.spawns
			def.Openings = test.openings
			bundle.Places = map[string]PlaceDef{"place.a": def}
			bundle.Text[string(def.Name)] = "a place"
			_, err := New(bundle)
			requireErrContains(t, err, `places["place.a"].openings[`, test.want)
		})
	}
}
