package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

// A hall, a soop sloping down to a chamber, a ledge, and a soop that crosses under the hall.
maps: "map.test.descent": content.#MapDef & {
	name: "text.map.test_descent"
	fill: "terrain.packed_clay"
	open: "terrain.air"
	legend: {
		"#": "terrain.packed_clay"
		".": "terrain.air"
	}
	marks: {"@": "party_start", "c": "creature"}
	spawns: "c": "template.creature.crawler"
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
				" #.....@.......",
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
