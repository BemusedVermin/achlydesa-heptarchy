package data

// Two small rooms joined by a gate each way. The arrival marks sit beside the gates, so a party
// that has just crossed is not standing on one.
maps: "map.test.gate_a": _#TestMap & {
	name: "text.map.test.gate_a"
	marks: {"@": "party_start", "g": "gate_to_b", "r": "from_b"}
	gates: "g": {to: "map.test.gate_b", mark: "from_a"}
	layers: [
		{
			depth: 0
			rows: [
				"#########",
				"#.......#",
				"#.@..rg.#",
				"#.......#",
				"#########",
			]
		},
	]
}

maps: "map.test.gate_b": _#TestMap & {
	name: "text.map.test.gate_b"
	marks: {"@": "party_start", "g": "gate_to_a", "r": "from_a"}
	gates: "g": {to: "map.test.gate_a", mark: "from_b"}
	layers: [
		{
			depth: 0
			rows: [
				"#########",
				"#.......#",
				"#.g.r@..#",
				"#.......#",
				"#########",
			]
		},
	]
}

text: {
	"text.map.test.gate_a": "The first gate room"
	"text.map.test.gate_b": "The second gate room"
}
