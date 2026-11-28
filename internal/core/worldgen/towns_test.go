package worldgen_test

import (
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/content/contenttest"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/explore"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/worldgen"
)

var realTowns = []api.ContentID{
	"settlement.reaper", "settlement.reaper_lake", "settlement.salken_village", "settlement.lake",
}

func TestTheRealTownsGenerateWithTheirLandmarks(t *testing.T) {
	reg := contenttest.Full(t)
	for _, id := range realTowns {
		def, _ := reg.Settlement(id)
		for seed := uint64(1); seed <= 4; seed++ {
			layout, err := worldgen.LaySettlement(reg, id, seed)
			if err != nil {
				t.Fatalf("%s seed %d: %v", id, seed, err)
			}
			gives := map[string]int{}
			for _, placement := range layout.Placements {
				piece, _ := reg.Piece(placement.Piece)
				for _, row := range piece.Layers[0].Rows {
					for _, char := range row {
						for _, affordance := range piece.Affordances[string(char)] {
							gives[string(affordance)]++
						}
						if template, ok := reg.Template(piece.Spawns[string(char)]); ok {
							for _, affordance := range template.Affordances {
								gives[string(affordance)]++
							}
						}
					}
				}
			}
			for affordance, least := range def.LandmarkMin {
				if gives[affordance] < least {
					t.Errorf("%s seed %d: %d cells give %s, want %d", id, seed, gives[affordance], affordance, least)
				}
			}
		}
	}
}

func TestTheReapersCrawlspaceHoldsNoWaterSource(t *testing.T) {
	reg := contenttest.Full(t)
	def, _ := reg.Settlement("settlement.reaper")
	center := hexmap.FromOffsetOddR(def.At.Col, def.At.Row)
	src, err := explore.NewWorldSource(reg, 1)
	if err != nil {
		t.Fatal(err)
	}
	for _, spawn := range src.Spawns() {
		template, _ := reg.Template(spawn.Template)
		if template.Source != nil && hexmap.Distance(spawn.Cell.Hex, center) <= def.Radius {
			t.Errorf("%s stands inside the Reaper's crawlspace at %v", spawn.Template, spawn.Cell)
		}
	}
	lake, _ := reg.Settlement("settlement.reaper_lake")
	lakeCenter := hexmap.FromOffsetOddR(lake.At.Col, lake.At.Row)
	found := false
	for _, spawn := range src.Spawns() {
		template, _ := reg.Template(spawn.Template)
		found = found || template.Source != nil && hexmap.Distance(spawn.Cell.Hex, lakeCenter) <= lake.Radius
	}
	if !found {
		t.Error("no water source stands in the lake outside")
	}
}

func TestEveryRealTownConnectsToTheDeep(t *testing.T) {
	reg := contenttest.Full(t)
	src, m := verifiedWorldOf(t, reg, 1, 0, 0, 0)
	openings := src.Openings()
	frontier := []api.Cell{openings[0].Cell, openings[0].Exit}
	seen := map[api.Cell]bool{frontier[0]: true, frontier[1]: true}
	for len(frontier) > 0 {
		from := frontier[0]
		frontier = frontier[1:]
		for dir := range api.Dir(api.DirCount) {
			step, ok, err := m.Step(from, dir, verifyRules)
			if err != nil {
				t.Fatal(err)
			}
			if ok && !seen[step.To.Cell] {
				seen[step.To.Cell] = true
				frontier = append(frontier, step.To.Cell)
			}
		}
	}
	for _, town := range realTowns {
		reached := false
		for _, opening := range openings {
			reached = reached || opening.Place == town && (seen[opening.Cell] || seen[opening.Exit])
		}
		if !reached {
			t.Errorf("%s is not reachable from %v", town, openings[0].Place)
		}
	}
}
