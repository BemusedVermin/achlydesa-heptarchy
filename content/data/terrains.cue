package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

terrains: [string]: content.#Terrain

terrains: {
	"terrain.air": {
		name:   "text.terrain.air"
		solid:  false
		opaque: false
	}
	"terrain.packed_clay": {
		name:        "text.terrain.packed_clay"
		solid:       true
		opaque:      true
		floor_glyph: "."
		wall_glyph:  "#"
		move_cost:   "1"
	}
	"terrain.decayed_clay": {
		name:        "text.terrain.decayed_clay"
		solid:       true
		opaque:      true
		floor_glyph: ","
		wall_glyph:  "%"
		move_cost:   "1"
	}
	"terrain.rotten_clay": {
		name:        "text.terrain.rotten_clay"
		solid:       true
		opaque:      true
		floor_glyph: ";"
		wall_glyph:  "&"
		move_cost:   "1"
	}
}

text: {
	"text.terrain.decayed_clay": "decayed clay"
	"text.terrain.rotten_clay":  "rotten clay"
	"text.terrain.air":          "air"
	"text.terrain.packed_clay":  "packed clay"
}
