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
		"e": "from_feeders"
		"f": "from_shrine"
		"1": "gate_to_hall"
		"2": "gate_to_den"
		"3": "gate_to_descent"
		"4": "gate_to_town"
		"5": "gate_to_feeders"
		"6": "gate_to_shrine"
		"s": "stock"
	}
	gates: {
		"1": {to: "map.play.hall", mark: "from_hub"}
		"2": {to: "map.play.den", mark: "from_hub"}
		"3": {to: "map.play.descent", mark: "from_hub"}
		"4": {to: "map.test.town", mark: "from_hub"}
		"5": {to: "map.play.feeders", mark: "from_hub"}
		"6": {to: "map.play.shrine", mark: "from_hub"}
	}
	region: props: decay: 0.2
	// The hub is a town, so stocks can be tried by hand.
	region: town: {
		stock: [{item: "item.lake_water", count: 10}, {item: "item.ration", count: 5}]
		stock_marks: ["stock"]
	}
	layers: [
		{
			depth: 0
			rows: [
				"###############",
				"#...s.........#",
				"#.1.2.3.4.5.6.#",
				"#.a.b.c.d.e.f.#",
				"#...@.........#",
				"#.............#",
				"###############",
			]
		},
	]
}

text: {
	"text.map.play.hub":     "The playground hub"
	"text.map.play.hall":    "The playground hall"
	"text.map.play.den":     "The playground den"
	"text.map.play.descent": "The playground descent"
	"text.map.play.feeders": "The playground feeders' room"
	"text.map.play.shrine":  "The playground shrine"
}
