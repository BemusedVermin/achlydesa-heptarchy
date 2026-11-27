package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

world: content.#WorldDef

// The one world: the places in places/ sit at fixed positions and procgen generates the deep around and
// between them. Sizes are first values, for tuning: 400 by 400 hexes and 64 depth steps.
world: {
	cols:      400
	rows:      400
	min_depth: 0
	max_depth: 63
	fill:      "terrain.packed_clay"
	edge:      "terrain.packed_clay"
	open:      "terrain.air"
	geography: "geography.deep"
	// Chunks far from every party member are unloaded past this many; first value, for tuning.
	chunk_budget: 256
	// A chunk this many chunks or fewer from any party member stays loaded whatever the budget.
	near_chunks: 2
	// An export keeps the chunks this many chunks or fewer from any party member as they were; first value.
	pocket_chunks: 1
	// The pocket is joined to the new deep by at most this many tunnels; first value.
	pocket_exits: 4
}
