package data

// The room the party starts in, with one gate to each playground map.
maps: "map.play.hub": _#TestMap & {
	name: "text.map.play.hub"
	tags: ["town"]
	marks: {
		"@": "party_start"
		"a": "from_hall"
		"b": "from_den"
		"c": "from_descent"
		"d": "from_town"
		"1": "gate_to_hall"
		"2": "gate_to_den"
		"3": "gate_to_descent"
		"4": "gate_to_town"
		"s": "stock"
	}
	gates: {
		"1": {to: "map.play.hall", mark: "from_hub"}
		"2": {to: "map.play.den", mark: "from_hub"}
		"3": {to: "map.play.descent", mark: "from_hub"}
		"4": {to: "map.test.town", mark: "from_hub"}
	}
	// The hub is a town, so stocks can be tried by hand.
	region: town: {
		stock: [{item: "item.lake_water", count: 10}, {item: "item.ration", count: 5}]
		stock_marks: ["stock"]
	}
	layers: [
		{
			depth: 0
			rows: [
				"###########",
				"#...s.....#",
				"#.1.2.3.4.#",
				"#.a.b.c.d.#",
				"#...@.....#",
				"#.........#",
				"###########",
			]
		},
	]
}

text: {
	"text.map.play.hub":     "The playground hub"
	"text.map.play.hall":    "The playground hall"
	"text.map.play.den":     "The playground den"
	"text.map.play.descent": "The playground descent"
}
