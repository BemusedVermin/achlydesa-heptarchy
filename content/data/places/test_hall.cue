package data

// The prototype's room at depth 0: air inside a ring of packed clay with three pillars.
places: "place.test.hall": _#TestPlace & {
	name: "text.place.test.hall"
	at: {col: 4, row: 4, depth: 0}
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
	"text.place.test.hall": "The test hall"
}
