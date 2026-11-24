package hexmap

import (
	"slices"

	"github.com/BemusedVermin/achlydesa-heptarchy/internal/core/api"
)

// Gate is a cell that leads to a cell of another map.
type Gate struct {
	At     api.Cell
	To     api.MapID
	ToCell api.Cell
}

// SetGates replaces the map's gate table. Gates are derived from content, so they are neither
// saved nor hashed.
func (m *Map) SetGates(gates []Gate) {
	m.gates = make(map[api.Cell]Gate, len(gates))
	for _, g := range gates {
		m.gates[g.At] = g
	}
}

// GateAt returns the gate in the cell, if there is one.
func (m *Map) GateAt(c api.Cell) (Gate, bool) {
	g, ok := m.gates[c]
	return g, ok
}

// Gates lists every gate, sorted by cell.
func (m *Map) Gates() []Gate {
	out := make([]Gate, 0, len(m.gates))
	for _, g := range m.gates {
		out = append(out, g)
	}
	slices.SortFunc(out, func(a, b Gate) int { return CompareCells(a.At, b.At) })
	return out
}
