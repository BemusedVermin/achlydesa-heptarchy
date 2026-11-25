package data

// A hall, a soop sloping down to a chamber, a ledge, and a soop that crosses under the hall.
places: "place.test.descent": _#TestPlace & {
	name: "text.place.test.descent"
	at: {col: 200, row: 200, depth: 10}
	marks: {"@": "party_start", "c": "creature", "o": "way_out"}
	openings: ["o"]
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
				"o.....c.....o",
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
				"                 .o",
			]
		},
	]
}

text: {
	"text.place.test.descent": "The test descent"
}
