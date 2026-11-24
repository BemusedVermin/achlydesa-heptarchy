package data

// A hall, a soop sloping down to a chamber, a ledge, and a soop that crosses under the hall.
maps: "map.test.descent": _#TestMap & {
	name: "text.map.test.descent"
	marks: {"@": "party_start", "c": "creature"}
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

text: {
	"text.map.test.descent": "The test descent"
}
