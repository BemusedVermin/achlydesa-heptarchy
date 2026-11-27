package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

pieces: [string]: content.#PieceDef

// _#Piece draws "#" as packed clay, "." as air and "+" as a door.
_#Piece: {
	legend: {
		"#": "terrain.packed_clay"
		".": "terrain.air"
	}
	marks: {"+": "door", ...}
	doors: ["+"]
	...
}

pieces: {
	// A length of street: doors at both ends and along each side, where pieces join it.
	"piece.street": _#Piece & {
		name: "text.piece.street"
		layers: [{
			depth: 0
			rows: [
				"#########",
				"+...+...+",
				"#########",
			]
		}]
	}
	// One room with a bed.
	"piece.house": _#Piece & {
		name: "text.piece.house"
		marks: "b": "bed"
		affordances: "b": ["affordance.rest"]
		layers: [{
			depth: 0
			rows: [
				"#####",
				"#.b.#",
				"#...#",
				"##+##",
			]
		}]
	}
	"piece.well": _#Piece & {
		name: "text.piece.well"
		marks: "w":  "well"
		spawns: "w": "template.source.well"
		layers: [{
			depth: 0
			rows: [
				" ### ",
				"+.w.+",
				" ### ",
			]
		}]
	}
	"piece.larder": _#Piece & {
		name: "text.piece.larder"
		marks: "l":  "larder"
		spawns: "l": "template.container.larder"
		layers: [{
			depth: 0
			rows: [
				"####",
				"#.l#",
				"#+##",
			]
		}]
	}
	"piece.shrine": _#Piece & {
		name: "text.piece.shrine"
		marks: "a": "altar"
		affordances: "a": ["affordance.worship"]
		layers: [{
			depth: 0
			rows: [
				"#####",
				"#.a.#",
				"#...#",
				"##+##",
			]
		}]
	}
	// Ground for the dead: every grave cell gives them.
	"piece.graveyard": _#Piece & {
		name: "text.piece.graveyard"
		marks: "g": "grave"
		affordances: "g": ["affordance.the_dead"]
		layers: [{
			depth: 0
			rows: [
				"#######",
				"#g.g.g#",
				"#.....#",
				"#g.g.g#",
				"###+###",
			]
		}]
	}
}

text: {
	"text.piece.street":    "A street"
	"text.piece.house":     "A house"
	"text.piece.well":      "A well"
	"text.piece.larder":    "A larder"
	"text.piece.shrine":    "A shrine"
	"text.piece.graveyard": "A graveyard"
}
