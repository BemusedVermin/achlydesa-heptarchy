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
	// A name works only on its owner, but it is not refused elsewhere: the swing lands and does
	// nothing (rule.unheld, in rules.cue), so the menu never tells whose name the party holds.
	// Speaking it in the deep using it up is a trigger (M2-05).
	"type.name": {name: "text.type.name"}
	"type.burn": #Consumable & {_item: "item.coal", name: "text.type.burn"}
	"type.rot": #Consumable & {_item: "item.salk_dust", name: "text.type.rot"}
	// Soak is heavy: it adds to the recovery of any form it is loaded into.
	"type.soak": #Consumable & {_item: "item.lake_water", name: "text.type.soak", recovery: "2"}
	// Light has its consume and nothing more: it is gone once opened.
	"type.light": #Consumable & {_item: "item.daylight_jar", name: "text.type.light"}
	"type.violet": #Consumable & {_item: "item.violet_blood", name: "text.type.violet"}
	// Like Name, a keepsake on the wrong body is not refused: the swing lands and does nothing. It is
	// used up when the swing lands, after the form's own effects, not when it starts: rule.unheld reads
	// the pack, so the keepsake must still be there when the damage is worked out.
	"type.keepsake": {
		name: "text.type.keepsake"
		lists: consumes: ["item.keepsake"]
		on_land: [{do: "consume", params: {item: "item.keepsake"}}]
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
}
