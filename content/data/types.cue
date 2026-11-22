package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

types: [string]: content.#DamageType

// A type that burns one item: it lists the item in lists.consumes and uses one up when it starts.
#Consumable: {
	_item: string
	lists: consumes: [_item]
	on_start: [{do: "consume", params: {item: _item}}]
	...
}

// The eleven types of the design table. Which item a move burns, and when it may be used, is data.
types: {
	"type.cut": {name: "text.type.cut", tags: ["innate"]}
	"type.press": {name: "text.type.press", tags: ["innate"]}
	"type.relic": {name: "text.type.relic", tags: ["innate"]}
	// A name works only on its owner: it needs a name in the pack, bound to the target. Speaking it in
	// the deep using it up is a trigger (M2-05).
	"type.name": {
		name:     "text.type.name"
		tags: ["innate"]
		requires: "pack[\"item.name\"] > 0"
	}
	"type.burn": #Consumable & {_item: "item.coal", name: "text.type.burn"}
	"type.rot": #Consumable & {_item: "item.salk_dust", name: "text.type.rot"}
	"type.soak": #Consumable & {_item: "item.lake_water", name: "text.type.soak"}
	"type.light": #Consumable & {_item: "item.daylight_jar", name: "text.type.light"}
	"type.violet": #Consumable & {_item: "item.violet_blood", name: "text.type.violet"}
	"type.keepsake": #Consumable & {
		_item:    "item.keepsake"
		name:     "text.type.keepsake"
		requires: "pack[\"item.keepsake\"] > 0"
	}
	"type.lees": #Consumable & {_item: "item.lees", name: "text.type.lees"}
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
