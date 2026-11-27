package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

choices: [string]: content.#ChoiceDef

choices: {
	// The type a form is loaded with: one the body knows innately, one whose item the inventory holds, or
	// Relic once the body carries a graft (the design: Relic is innate to grafts).
	type: {
		from:  "types"
		offer: "record.id in source.template.lists.types || (record.id == \"type.relic\" && source.grafts > 0) || any(record.lists.consumes, {inventory[#] > 0})"
	}
	// A thing to eat or drink, to take from a source or container.
	ware: {
		from:  "items"
		offer: "\"drinkable\" in record.tags || \"edible\" in record.tags"
	}
	// A ware the body carries, to put away or consume.
	carried: {
		from:  "items"
		offer: "(\"drinkable\" in record.tags || \"edible\" in record.tags) && inventory[record.id] > 0"
	}
	// The piece a graft fits: any piece in the inventory.
	piece: {
		from:  "items"
		offer: "\"piece\" in record.tags && inventory[record.id] > 0"
	}
}
