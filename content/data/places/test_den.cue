package data

// A small room at depth 0 for the other two sample creatures, the flesh beast and the pre-entropy
// remnant, so the hall keeps one creature. Tests start a game here by setting start.place.
places: "place.test.den": _#TestPlace & {
	name: "text.place.test.den"
	at: {col: 170, row: 230, depth: 20}
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

text: "text.place.test.den": "The test den"
