package tui_test

import (
	"testing"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/frontend/tui"
)

func heightView() *api.MapView {
	m := &api.MapView{}
	for q := range 7 {
		m.Cells = append(
			m.Cells,
			api.CellView{Hex: api.Hex{Q: q}, Glyph: ".", Color: "green", Visible: true, Below: q - 3},
		)
	}
	m.Cells = append(
		m.Cells,
		api.CellView{Hex: api.Hex{Q: 7}, Glyph: "#", Color: "brown", Visible: true, Wall: true},
		api.CellView{Hex: api.Hex{Q: 8}, Glyph: ".", Color: "green", Visible: false, Below: 2},
	)
	return m
}

func TestRenderMapHeightsOverlay(t *testing.T) {
	out := tui.RenderMapOverlay(heightView(), nil, tui.MapOverlay{Heights: true}, 20, 1)
	golden(t, "map_heights", out)
}

func TestRenderMovesReadout(t *testing.T) {
	m := &api.MapView{
		Elevation: "Depth 1",
		Moves: []api.MoveView{
			{Dir: api.DirE, Kind: api.MoveDown, Below: 1, Label: "east: step down 1"},
			{Dir: api.DirNE, Kind: api.MoveDrop, Below: 3, Label: "northeast: drop 3"},
			{Dir: api.DirNW, Kind: api.MoveBlocked, Label: "northwest: ledge 2 up, too high"},
			{Dir: api.DirW, Kind: api.MoveUp, Below: -1, Label: "west: climb 1"},
			{Dir: api.DirSW, Kind: api.MoveBlocked, Label: "southwest: wall"},
			{Dir: api.DirSE, Kind: api.MoveLevel, Label: "southeast: level"},
		},
	}
	golden(t, "moves_readout", tui.RenderMoves(m))
}
