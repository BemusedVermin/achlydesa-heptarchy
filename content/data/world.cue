package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

world: content.#WorldDef

// The one world, small and hand-built at first: the places in places/ sit side by side in its clay,
// nothing joining them yet.
world: {
	cols:      70
	rows:      30
	min_depth: 0
	max_depth: 7
	fill:      "terrain.packed_clay"
	edge:      "terrain.packed_clay"
	open:      "terrain.air"
}
