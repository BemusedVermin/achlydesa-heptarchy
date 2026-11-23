package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

// The prototype's room at depth 0: air inside a ring of packed clay with three pillars.
maps: "map.test.hall": content.#MapDef & {
	name: "text.map.test.hall"
	fill: "terrain.packed_clay"
	open: "terrain.air"
	legend: {
		"#": "terrain.packed_clay"
		".": "terrain.air"
	}
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
