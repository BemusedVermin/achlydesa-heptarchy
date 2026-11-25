package data

// A small town: a well, a cistern (the stock mark), a house to sleep at and a gate to the
// playground hub. Its residents are bodies from the first tick.
maps: "map.test.town": _#TestMap & {
	name: "text.map.test.town"
	tags: ["town"]
	marks: {
		"@": "party_start"
		"w": "well"
		"c": "cistern"
		"h": "house"
		"a": "from_hub"
		"g": "gate_to_hub"
		"p": "resident"
		"H": "hauler"
		"G": "guard"
		"F": "field_hand"
	}
	gates: "g": {to: "map.play.hub", mark: "from_town"}
	spawns: {
		"w": "template.town.well"
		"p": "template.resident.plain"
		"H": "template.resident.hauler"
		"G": "template.generic.guard"
		"F": "template.generic.field_hand"
	}
	region: {
		town: {
			stock: [{item: "item.lake_water", count: 6}, {item: "item.ration", count: 10}]
			stock_marks: ["cistern"]
			residents: "template.resident.plain"
			generics: [
				{template: "template.generic.guard", count: 2, ticks: 40},
				{template: "template.generic.field_hand", count: 1, ticks: 60},
			]
		}
	}
	layers: [
		{
			depth: 0
			rows: [
				"###############",
				"#.............#",
				"#.w.....c.....#",
				"#....pp.......#",
				"#.@..h........#",
				"#....HGG.F....#",
				"#...........ag#",
				"###############",
			]
		},
	]
}

text: {
	"text.map.test.town": "The test town"
}
