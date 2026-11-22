package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

templates: [string]: content.#Template

// A drop of two cells or more hurts the legs, by a cell less than it fell. Placeholder numbers.
_fallInjury: {
	on:   "explore.fell"
	role: "target"
	if:   "cells >= 2"
	do: [{do: "damage", params: {parts: "target:legs", amount: "cells - 1"}}]
}

templates: {
	"template.person.reaper": {
		name: "text.template.person.reaper"
		appearance: {glyph: "@", color: "red"}
		actor: controller: "player"
		actions:   ["action.step"]
		body:      "body.human"
		protected: true
		triggers:  [_fallInjury]
	}
	"template.person.salken": {
		name: "text.template.person.salken"
		appearance: {glyph: "@", color: "cyan"}
		actor: controller: "follow"
		actions:   ["action.step"]
		body:      "body.human"
		protected: true
		triggers:  [_fallInjury]
	}
	"template.party.main": {
		party: members: ["template.person.reaper", "template.person.salken"]
		// A few of each consumable to start with.
		inventory: [
			{item: "item.coal", count: 3},
			{item: "item.salk_dust", count: 3},
			{item: "item.lake_water", count: 3},
			{item: "item.daylight_jar", count: 3},
			{item: "item.violet_blood", count: 3},
			{item: "item.lees", count: 3},
		]
	}
}
