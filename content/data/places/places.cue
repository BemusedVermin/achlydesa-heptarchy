package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

places: [string]: content.#PlaceDef

_#TestPlace: {
	legend: {
		"#": "terrain.packed_clay"
		".": "terrain.air"
	}
	...
}
