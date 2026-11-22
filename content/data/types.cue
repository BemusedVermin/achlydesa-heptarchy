package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

types: [string]: content.#DamageType

// A type that burns one item: it lists the item in lists.consumes and uses one up when it starts,
// whether the swing lands or not.
#Consumable: {
	_item: string
	lists: consumes: [_item]
	on_start: [{do: "consume", params: {item: _item}}]
	...
}

// The eleven types of the design table. Which body knows which type innately is on the templates
// (templates.cue); the rest need their item in the pack. The timings and landing effects here are
// placeholders for M7's tuning.
types: {
	"type.cut": {name: "text.type.cut"}
	// Press staggers: landing on a body that is winding up cancels the swing it is winding up.
	"type.press": {
		name: "text.type.press"
		on_land: [{do: "interrupt", params: {bodies: "target", when: "winding_up"}}]
	}
	"type.relic": {name: "text.type.relic"}
	// A name works only on its owner: it needs a name in the pack that is bound to the target.
	// Speaking it in the deep using it up is a trigger (M2-05).
	"type.name": {
		name:     "text.type.name"
		requires: "pack[\"item.name\"] > 0"
		refusal:  "text.type.name_refusal"
	}
	"type.burn": #Consumable & {_item: "item.coal", name: "text.type.burn"}
	"type.rot": #Consumable & {_item: "item.salk_dust", name: "text.type.rot"}
	// Soak is heavy: it adds to the recovery of any form it is loaded into.
	"type.soak": #Consumable & {_item: "item.lake_water", name: "text.type.soak", recovery: "2"}
	// Light has its consume and nothing more: it is gone once opened.
	"type.light": #Consumable & {_item: "item.daylight_jar", name: "text.type.light"}
	"type.violet": #Consumable & {_item: "item.violet_blood", name: "text.type.violet"}
	"type.keepsake": #Consumable & {
		_item:    "item.keepsake"
		name:     "text.type.keepsake"
		requires: "pack[\"item.keepsake\"] > 0"
		refusal:  "text.type.keepsake_refusal"
	}
	"type.lees": #Consumable & {_item: "item.lees", name: "text.type.lees"}
}

text: {
	"text.type.cut":              "Cut"
	"text.type.press":            "Press"
	"text.type.relic":            "Relic"
	"text.type.name":             "Name"
	"text.type.burn":             "Burn"
	"text.type.rot":              "Rot"
	"text.type.soak":             "Soak"
	"text.type.light":            "Light"
	"text.type.violet":           "Violet"
	"text.type.keepsake":         "Keepsake"
	"text.type.lees":             "Lees"
	"text.type.name_refusal":     "No name of theirs is held."
	"text.type.keepsake_refusal": "No keepsake of theirs is held."
}
