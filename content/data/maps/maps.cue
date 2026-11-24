package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

maps: [string]: content.#MapDef

_#TestMap: {
	fill: "terrain.packed_clay"
	open: "terrain.air"
	legend: {
		"#": "terrain.packed_clay"
		".": "terrain.air"
	}
	...
}
