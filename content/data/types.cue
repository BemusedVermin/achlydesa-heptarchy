package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

types: [string]: content.#DamageType

// The eleven types of the design table. Which item a move burns, and when it may be used, is data.
// The timings and landing effects here are placeholders for M7's tuning; tests set their own numbers.
types: {
	"type.cut": {name: "text.type.cut", innate: true}
	// Press staggers: landing on a body that is winding up cancels its swings.
	"type.press": {
		name:   "text.type.press"
		innate: true
		on_land: [{do: "interrupt", params: {bodies: "target", when: "winding_up"}}]
	}
	"type.relic": {name: "text.type.relic", innate: true}
	// A name works only on its owner, and speaking it uses it up: the move spends a name the party has
	// learned, tied to one person (item.name).
	"type.name": {
		name:      "text.type.name"
		innate:    true
		consumes:  "item.name"
		condition: "target_uid == item_bound_to"
	}
	"type.burn": {name: "text.type.burn", innate: false, consumes: "item.coal"}
	"type.rot": {name: "text.type.rot", innate: false, consumes: "item.salk_dust"}
	// Soak is heavy: it adds to the recovery of any form it is loaded into.
	"type.soak": {name: "text.type.soak", innate: false, consumes: "item.lake_water", recovery: "2"}
	"type.light": {name: "text.type.light", innate: false, consumes: "item.daylight_jar"}
	"type.violet": {name: "text.type.violet", innate: false, consumes: "item.violet_blood"}
	"type.keepsake": {
		name:      "text.type.keepsake"
		innate:    false
		consumes:  "item.keepsake"
		condition: "target_uid == item_bound_to"
	}
	"type.lees": {name: "text.type.lees", innate: false, consumes: "item.lees"}
}

text: {
	"text.type.cut":      "Cut"
	"text.type.press":    "Press"
	"text.type.relic":    "Relic"
	"text.type.name":     "Name"
	"text.type.burn":     "Burn"
	"text.type.rot":      "Rot"
	"text.type.soak":     "Soak"
	"text.type.light":    "Light"
	"text.type.violet":   "Violet"
	"text.type.keepsake": "Keepsake"
	"text.type.lees":     "Lees"
}
