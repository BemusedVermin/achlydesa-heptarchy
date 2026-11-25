package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

world: content.#WorldDef
places: [string]: content.#PlaceDef

// The one world. Its size is a placeholder, small and hand-built at first: the places go in places/.
world: {
	cols:      64
	rows:      64
	min_depth: 0
	max_depth: 7
	fill:      "terrain.packed_clay"
	edge:      "terrain.packed_clay"
	open:      "terrain.air"
}
