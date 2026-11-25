package data

// A room with a believer and the archon's warden that guards them.
maps: "map.play.shrine": _#TestMap & {
	name: "text.map.play.shrine"
	marks: {"a": "from_hub", "g": "gate_to_hub", "b": "believer", "w": "warden"}
	gates: "g": {to: "map.play.hub", mark: "from_shrine"}
	spawns: {
		"b": "template.resident.believer"
		"w": "template.creature.archon_warden"
	}
	layers: [
		{
			depth: 0
			rows: [
				"##############",
				"#............#",
				"#............#",
				"#.ga....w....#",
				"#.....b......#",
				"#............#",
				"##############",
			]
		},
	]
}
