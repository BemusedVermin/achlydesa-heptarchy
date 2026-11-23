package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

// A small room at depth 0 for the other two sample creatures, the flesh beast and the pre-entropy
// remnant, so the hall keeps one creature. Tests start a game here by setting start.map.
maps: "map.test.den": content.#MapDef & {
	name: "text.map.test_den"
	fill: "terrain.packed_clay"
	open: "terrain.air"
	legend: {
		"#": "terrain.packed_clay"
		".": "terrain.air"
	}
	marks: {"@": "party_start", "r": "remnant", "b": "beast"}
	spawns: {
		"r": "template.creature.remnant"
		"b": "template.creature.beast"
	}
	layers: [
		{
			depth: 0
			rows: [
				"##########",
				"#........#",
				"#.r....b.#",
				"#........#",
				"#..@.....#",
				"#........#",
				"##########",
			]
		},
	]
}

text: "text.map.test_den": "The test den"
