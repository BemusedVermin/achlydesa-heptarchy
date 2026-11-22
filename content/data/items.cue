package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

items: [string]: content.#Item

items: {
	"item.coal": {name: "text.item.coal", kind: "consumable", bound: false}
	"item.salk_dust": {name: "text.item.salk_dust", kind: "consumable", bound: false}
	"item.lake_water": {name: "text.item.lake_water", kind: "consumable", bound: false}
	"item.daylight_jar": {name: "text.item.daylight_jar", kind: "consumable", bound: false}
	"item.violet_blood": {name: "text.item.violet_blood", kind: "consumable", bound: false}
	"item.lees": {name: "text.item.lees", kind: "consumable", bound: false}
	// A belonging of one candidate, burned by the keepsake type against that candidate.
	"item.keepsake": {name: "text.item.keepsake", kind: "keepsake", bound: true}
	// A name the party has learned, tied to its owner and used up when spoken.
	"item.name": {name: "text.item.name", kind: "name", bound: true}
}

text: {
	"text.item.coal":         "coal"
	"text.item.salk_dust":    "salk dust"
	"text.item.lake_water":   "lake water"
	"text.item.daylight_jar": "daylight jar"
	"text.item.violet_blood": "Violet blood"
	"text.item.lees":         "lees"
	"text.item.keepsake":     "keepsake"
	"text.item.name":         "name"
}
