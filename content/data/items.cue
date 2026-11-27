package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

items: [string]: content.#Item & {bound: *false | bool}

items: {
	"item.coal": {name: "text.item.coal", tags: ["consumable"]}
	"item.salk_dust": {name: "text.item.salk_dust", tags: ["consumable"]}
	"item.lake_water": {name: "text.item.lake_water", tags: ["consumable"]}
	"item.daylight_jar": {name: "text.item.daylight_jar", tags: ["consumable"]}
	"item.violet_blood": {name: "text.item.violet_blood", tags: ["consumable"]}
	"item.lees": {name: "text.item.lees", tags: ["consumable"]}
	// A belonging of one candidate, burned by the keepsake type against that candidate.
	"item.keepsake": {name: "text.item.keepsake", tags: ["keepsake"], bound: true}
}

// Supplies: what people drink and eat, drawn from sources and kept in containers.
items: {
	"item.water": {name: "text.item.water"}
	"item.ration": {name: "text.item.ration"}
}

// A name: the thing a name type is spoken with, tied to its owner.
items: "item.name": {name: "text.item.name", tags: ["name"], bound: true}

// A piece of Serilmen: an arm that sometimes swings on its own once grafted on. The 0.15 is a
// placeholder, for tuning.
items: "item.serilmen_arm": {
	name: "text.item.serilmen_arm"
	tags: ["piece"]
	part: "part.serilmen_arm"
	triggers: [{
		on:   "timeline.ready"
		role: "source"
		if:   "roll < 0.15"
		do: [{do: "act_alone"}]
	}]
}

text: {
	"text.item.serilmen_arm": "a piece of Serilmen"
	"text.item.name":         "a name"
	"text.item.coal":         "coal"
	"text.item.salk_dust":    "salk dust"
	"text.item.lake_water":   "lake water"
	"text.item.daylight_jar": "daylight jar"
	"text.item.violet_blood": "Violet blood"
	"text.item.lees":         "lees"
	"text.item.keepsake":     "keepsake"
	"text.item.water":        "water"
	"text.item.ration":       "ration"
}
