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
		"p": "resident"
		"H": "hauler"
		"G": "guard"
		"F": "field_hand"
	}
	gates: "g": {to: "map.sample.deep_a", mark: "from_shallow"}
	spawns: {
		"w": "template.town.well"
		"p": "template.resident.plain"
		"H": "template.resident.hauler"
		"G": "template.generic.guard"
		"F": "template.generic.field_hand"
	}
	region: props: decay: 0.1
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

maps: "map.sample.deep_a": _#TestMap & {
	name: "text.map.sample.deep_a"
	marks: {
		"a": "from_shallow"
		"b": "from_deep_b"
		"s": "gate_to_shallow"
		"n": "gate_to_deep_b"
		"m": "moving_clay"
		"r": "remnant"
	}
	gates: {
		"s": {to: "map.sample.shallow", mark: "from_deep_a"}
		"n": {to: "map.sample.deep_b", mark: "from_deep_a"}
	}
	spawns: {
		"m": "template.creature.moving_clay"
		"r": "template.creature.remnant"
	}
	region: props: decay: 0.5
	layers: [
		{
			depth: 0
			rows: [
				"#############",
				"#.m...r...m.#",
				"#.as......bn#",
				"#...m...r...#",
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
		"m": "moving_clay"
		"r": "remnant"
		"f": "soul_feeder"
		"w": "archon_warden"
	}
	gates: {
		"s": {to: "map.sample.deep_a", mark: "from_deep_b"}
		"n": {to: "map.sample.lake", mark: "from_deep_b"}
	}
	spawns: {
		"m": "template.creature.moving_clay"
		"r": "template.creature.remnant"
		"f": "template.creature.soul_feeder"
		"w": "template.creature.archon_warden"
	}
	region: props: decay: 0.8
	layers: [
		{
			depth: 0
			rows: [
				"#############",
				"#.r...m...r.#",
				"#.as......bn#",
				"#..f..r..f..#",
				"#.m.r..w....#",
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
		"p": "resident"
		"T": "tamed_clay"
		"L": "leashed_feeder"
	}
	gates: "g": {to: "map.sample.deep_b", mark: "from_lake"}
	spawns: {
		"p": "template.resident.plain"
		"T": "template.generic.tamed_clay"
		"L": "template.generic.leashed_feeder"
	}
	region: props: decay: 0.1
	region: {
		town: {
			stock: [{item: "item.lake_water", count: 40}, {item: "item.ration", count: 6}]
			stock_marks: ["cistern"]
			residents: "template.resident.plain"
			generics: [
				{template: "template.generic.tamed_clay", count: 1, ticks: 80},
				{template: "template.generic.leashed_feeder", count: 1, ticks: 80},
			]
		}
	}
	layers: [
		{
			depth: 0
			rows: [
				"###############",
				"#.............#",
				"#.....c.......#",
				"#.@..h........#",
				"#....ppp..TL..#",
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
