package data

// The prototype's room at depth 0: air inside a ring of packed clay with three pillars.
maps: "map.test.hall": _#TestMap & {
	name: "text.map.test.hall"
	marks: {"@": "party_start", "c": "creature"}
	spawns: "c": "template.creature.moving_clay"
	layers: [
		{
			depth: 0
			rows: [
				"     ###########",
				"    #..........#",
				"    #...........#",
				"   #............#",
				"   #.............#",
				"  #..............#",
				"  #...............#",
				" #................#",
				" #.................#",
				"#...........#......#",
				"#.........@.......c.#",
				"#..................#",
				" #..........#......#",
				" #................#",
				"  #.......#.......#",
				"  #..............#",
				"   #.............#",
				"   #............#",
				"    #...........#",
				"    #..........#",
				"     ###########",
			]
		},
	]
}

text: {
	"text.map.test.hall": "The test hall"
}
