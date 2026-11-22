package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

types: [string]: content.#DamageType

// The eleven types of the design table. Which item a move burns, and when it may be used, is data.
types: {
	"type.cut": {name: "text.type.cut", innate: true}
	"type.press": {name: "text.type.press", innate: true}
	"type.relic": {name: "text.type.relic", innate: true}
	// A name works only on its owner; speaking it in the deep using it up is a trigger (M2-05).
	"type.name": {
		name:      "text.type.name"
		innate:    true
		condition: "target_uid == item_bound_to"
	}
	"type.burn": {name: "text.type.burn", innate: false, consumes: "item.coal"}
	"type.rot": {name: "text.type.rot", innate: false, consumes: "item.salk_dust"}
	"type.soak": {name: "text.type.soak", innate: false, consumes: "item.lake_water"}
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
