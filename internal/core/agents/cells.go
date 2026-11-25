package agents

import (
	"cmp"
	"fmt"
	"slices"

	"github.com/mlange-42/ark/ecs"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/comp"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/hexmap"
	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/sim"
)

// freeCellNear is the free standable cell nearest the anchor on the map, within the arrival
// radius; ties go to the lowest cell. Gates and occupied cells are never free. It reports false
// when none is free.
func freeCellNear(w *sim.World, id api.MapID, anchor api.Cell) (api.Cell, bool, error) {
	m := w.Maps.Get(id)
	if m == nil {
		return api.Cell{}, false, fmt.Errorf("unknown map %d", id)
	}
	taken := map[api.Cell]bool{}
	query := ecs.NewFilter1[comp.Pos](w.ECS).Query()
	for query.Next() {
		if pos := query.Get(); pos.Map == id {
			taken[pos.At] = true
		}
	}
	var free []api.Cell
	for _, cell := range m.CellsIn(anchor.Hex, w.Content.Agents().ArrivalRadius, anchor.Depth-1, anchor.Depth+1) {
		if taken[cell] {
			continue
		}
		standable, err := m.Standable(cell)
		if err != nil {
			return api.Cell{}, false, err
		}
		if _, gate := m.GateAt(cell); standable && !gate {
			free = append(free, cell)
		}
	}
	if len(free) == 0 {
		return api.Cell{}, false, nil
	}
	return slices.MinFunc(free, func(a, b api.Cell) int {
		return cmp.Or(
			cmp.Compare(hexmap.CellDistance(anchor, a), hexmap.CellDistance(anchor, b)),
			hexmap.CompareCells(a, b),
		)
	}), true, nil
}
