package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

choices: [string]: content.#ChoiceDef

choices: {
	// The type a form is loaded with: any innate type, or one whose item the pack holds.
	type: {
		from:  "types"
		offer: "\"innate\" in record.tags || any(record.lists.consumes, {pack[#] > 0})"
	}
}
