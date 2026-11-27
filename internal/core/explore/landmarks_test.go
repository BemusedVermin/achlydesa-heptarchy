package explore_test

import (
	"slices"
	"strings"
	"testing"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/game/gametest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
)

func TestEachCellOfALandmarkMarkGivesItsAffordances(t *testing.T) {
	const den = "place.test.den"
	var at content.PlaceAt
	_, w, _ := gametest.New(t, 1, func(bundle *content.Bundle) {
		contenttest.Edit(bundle.Places, den, func(place *content.PlaceDef) {
			place.Marks["g"] = "graves"
			place.Affordances = map[string][]api.ContentID{"g": {"affordance.the_dead", "affordance.rest"}}
			rows := slices.Clone(place.Layers[0].Rows)
			rows[1] = strings.Replace(rows[1], "..", "gg", 1)
			place.Layers[0].Rows = rows
			at = place.At
		})
	})
	var cells []api.Cell
	query := ecs.NewFilter2[comp.Pos, comp.Affordance](w.ECS).Without(ecs.C[comp.Body]()).Query()
	for query.Next() {
		pos, given := query.Get()
		if !slices.Equal(given.Tags, []api.ContentID{"affordance.rest", "affordance.the_dead"}) {
			t.Errorf("a grave gives %v, want rest and the dead in ID order", given.Tags)
		}
		cells = append(cells, pos.At)
	}
	slices.SortFunc(cells, hexmap.CompareCells)
	want := []api.Cell{
		{Hex: hexmap.FromOffsetOddR(at.Col+1, at.Row+1), Depth: at.Depth},
		{Hex: hexmap.FromOffsetOddR(at.Col+2, at.Row+1), Depth: at.Depth},
	}
	if !slices.Equal(cells, want) {
		t.Errorf("landmark marks at %v, want %v", cells, want)
	}
}
