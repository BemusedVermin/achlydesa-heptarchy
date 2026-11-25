package data

// The sample world: a shallow town, two regions of the deep and a lake town, joined in a chain by
// gates. Every number is a placeholder for tuning.
maps: "map.sample.shallow": _#TestMap & {
	name: "text.map.sample.shallow"
	tags: ["town"]
	marks: {
		"@": "party_start"
		"w": "well"
		"c": "cistern"
		"h": "house"
		"a": "from_deep_a"
		"g": "gate_to_deep_a"
	}
	gates: "g": {to: "map.sample.deep_a", mark: "from_shallow"}
	spawns: "w": "template.town.well"
	region: props: decay: 0.1
	region: {
		town: {
			stock: [{item: "item.lake_water", count: 6}, {item: "item.ration", count: 10}]
			stock_marks: ["cistern"]
			residents: "family.people"
			generics: [
				{template: "template.generic.guard", count: 2, ticks: 40},
				{template: "template.generic.field_hand", count: 1, ticks: 60},
			]
		}
		populations: {
			"family.people":        2, "family.people.haulers":     1
			"family.people.guards": 2, "family.people.field_hands": 1
		}
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

maps: "map.sample.deep_a": _#TestMap & {
	name: "text.map.sample.deep_a"
	marks: {
		"a": "from_shallow"
		"b": "from_deep_b"
		"s": "gate_to_shallow"
		"n": "gate_to_deep_b"
	}
	gates: {
		"s": {to: "map.sample.shallow", mark: "from_deep_a"}
		"n": {to: "map.sample.deep_b", mark: "from_deep_a"}
	}
	region: props: decay: 0.5
	region: populations: {"family.moving_clay": 3, "family.remnant": 2}
	layers: [
		{
			depth: 0
			rows: [
				"#############",
				"#...........#",
				"#.as......bn#",
				"#...........#",
				"#...........#",
				"#############",
			]
		},
	]
}

maps: "map.sample.deep_b": _#TestMap & {
	name: "text.map.sample.deep_b"
	marks: {
		"a": "from_deep_a"
		"b": "from_lake"
		"s": "gate_to_deep_a"
		"n": "gate_to_lake"
	}
	gates: {
		"s": {to: "map.sample.deep_a", mark: "from_deep_b"}
		"n": {to: "map.sample.lake", mark: "from_deep_b"}
	}
	region: props: decay: 0.8
	region: populations: {
		"family.remnant":     4, "family.moving_clay":    2
		"family.soul_feeder": 2, "family.archon_wardens": 1
	}
	layers: [
		{
			depth: 0
			rows: [
				"#############",
				"#...........#",
				"#.as......bn#",
				"#...........#",
				"#...........#",
				"#############",
			]
		},
	]
}

maps: "map.sample.lake": _#TestMap & {
	name: "text.map.sample.lake"
	tags: ["town"]
	marks: {
		"@": "party_start"
		"c": "cistern"
		"h": "house"
		"a": "from_deep_b"
		"g": "gate_to_deep_b"
	}
	gates: "g": {to: "map.sample.deep_b", mark: "from_lake"}
	region: props: decay: 0.1
	region: {
		town: {
			stock: [{item: "item.lake_water", count: 40}, {item: "item.ration", count: 6}]
			stock_marks: ["cistern"]
			residents: "family.people"
			refill: [{item: "item.lake_water", count: 4}]
		}
		populations: {"family.people": 3}
	}
	layers: [
		{
			depth: 0
			rows: [
				"###############",
				"#.............#",
				"#.....c.......#",
				"#.@..h........#",
				"#.............#",
				"#...........ag#",
				"###############",
			]
		},
	]
}

text: {
	"text.map.sample.shallow": "The shallow town"
	"text.map.sample.deep_a":  "The upper deep"
	"text.map.sample.deep_b":  "The lower deep"
	"text.map.sample.lake":    "The lake town"
}
