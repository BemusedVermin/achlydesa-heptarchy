package data

// The test den joined to the hub by a gate.
maps: "map.play.den": _#TestMap & {
	name: "text.map.play.den"
	marks: {"a": "from_hub", "g": "gate_to_hub", "r": "remnant", "b": "beast"}
	gates: "g": {to: "map.play.hub", mark: "from_den"}
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
				"#.ga.....#",
				"#........#",
				"##########",
			]
		},
	]
}
