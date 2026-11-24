package data

// The test descent joined to the hub by a gate. A hall, a soop sloping down to a chamber, a ledge, and a soop that crosses under the hall.
maps: "map.play.descent": _#TestMap & {
	name: "text.map.play.descent"
	marks: {"a": "from_hub", "g": "gate_to_hub", "c": "creature"}
	gates: "g": {to: "map.play.hub", mark: "from_descent"}
	spawns: "c": "template.creature.moving_clay"
	layers: [
		{
			depth: 0
			rows: [
				"    #######",
				"   #......#",
				"   #.......#",
				"  #........#",
				"  #.........#",
				" #..........#",
				" #...ga.......",
				" #..........#",
				"  #.........#",
				"  #........#",
				"   #.......#",
				"   #......#",
				"    #######",
			]
		},
		{
			depth: 1
			rows: [
				"",
				"",
				"",
				"",
				"",
				"              .",
				"              ..",
			]
		},
		{
			depth: 2
			rows: [
				"",
				"",
				"",
				"",
				"",
				"              .",
				"               ..",
				"",
				"",
				"......c......",
			]
		},
		{
			depth: 3
			rows: [
				"",
				"",
				"",
				"",
				"",
				"              .",
				"                ..",
			]
		},
		{
			depth: 4
			rows: [
				"",
				"",
				"",
				"",
				"",
				"              .....",
				"                 ...",
				"                 ..",
			]
		},
	]
}
