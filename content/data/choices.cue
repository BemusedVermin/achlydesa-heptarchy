package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

choices: [string]: content.#ChoiceDef

choices: {
	// The type a form is loaded with: one the body knows innately, one whose item the pack holds, or
	// Relic once the body carries a graft (the design: Relic is innate to grafts).
	type: {
		from:  "types"
		offer: "record.id in source.template.lists.types || (record.id == \"type.relic\" && source.grafts > 0) || any(record.lists.consumes, {pack[#] > 0})"
	}
	// The piece a graft fits: any piece in the pack.
	piece: {
		from:  "items"
		offer: "\"piece\" in record.tags && pack[record.id] > 0"
	}
}
