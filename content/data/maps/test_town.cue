package data

// A small town: a well, a cistern (the stock mark), a house to sleep at and a gate to the
// playground hub. Its residents are counted in the region and appear when the party arrives.
maps: "map.test.town": _#TestMap & {
	name: "text.map.test.town"
	marks: {
		"@": "party_start"
		"w": "well"
		"c": "cistern"
		"h": "house"
		"a": "from_hub"
		"g": "gate_to_hub"
	}
	gates: "g": {to: "map.play.hub", mark: "from_town"}
	spawns: "w": "template.town.well"
	region: {
		town: {
			stock: [{item: "item.lake_water", count: 6}, {item: "item.ration", count: 10}]
			stock_marks: ["cistern"]
		}
		populations: {"family.people": 2, "family.people.haulers": 1}
	}
	layers: [
		{
			depth: 0
			rows: [
				"###############",
				"#.............#",
				"#.w.....c.....#",
				"#.............#",
				"#.@..h........#",
				"#.............#",
				"#...........ag#",
				"###############",
			]
		},
	]
}

text: {
	"text.map.test.town": "The test town"
}
