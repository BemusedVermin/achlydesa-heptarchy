package data

// A room with a soul-feeder, which hunts whoever enters.
maps: "map.play.feeders": _#TestMap & {
	name: "text.map.play.feeders"
	marks: {"a": "from_hub", "g": "gate_to_hub", "s": "feeder"}
	gates: "g": {to: "map.play.hub", mark: "from_feeders"}
	spawns: "s": "template.creature.soul_feeder"
	layers: [
		{
			depth: 0
			rows: [
				"##############",
				"#............#",
				"#............#",
				"#.ga......s..#",
				"#............#",
				"#............#",
				"##############",
			]
		},
	]
}
