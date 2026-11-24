package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

maps: [string]: content.#MapDef

// A test map: air carved out of endless packed clay.
_#TestMap: {
	fill: "terrain.packed_clay"
	open: "terrain.air"
	legend: {
		"#": "terrain.packed_clay"
		".": "terrain.air"
	}
	...
}
