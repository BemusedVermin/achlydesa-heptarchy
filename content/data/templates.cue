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

// What a person knows: the seven forms, and the types innate to a body. The design gives everyone
// Cut and Name, and the Reaper Press (the thrukk grip). Relic is innate to grafts, which an offer
// can't see yet.
_forms: [
	"action.step", "action.strike", "action.sweep", "action.sever", "action.pin", "action.stand",
	"action.give", "action.graft",
]
_innateTypes: ["type.cut", "type.name", "type.press"]

templates: {
	"template.person.reaper": {
		name: "text.template.person.reaper"
		appearance: {glyph: "@", color: "red"}
		actor: controller: "player"
		actions:   _forms
		body:      "body.human"
		protected: true
		tags: ["graftable"]
		lists: types: _innateTypes
		triggers: [_fallInjury]
	}
	"template.person.salken": {
		name: "text.template.person.salken"
		appearance: {glyph: "@", color: "cyan"}
		actor: controller: "follow"
		actions:   _forms
		body:      "body.human"
		protected: true
		tags: ["graftable"]
		lists: types: _innateTypes
		triggers: [_fallInjury]
	}
	// A human body with no controller of its own, for tests to swing at.
	"template.test.dummy": {
		name: "text.template.test.dummy"
		appearance: {glyph: "d", color: "white"}
		actor: controller: "idle"
		body: "body.human"
		tags: ["graftable"]
	}
	// A body that crawls, for Stand to freeze.
	"template.test.crawler": {
		name: "text.template.test.crawler"
		appearance: {glyph: "c", color: "white"}
		actor: controller: "idle"
		body: "body.human"
		tags: ["crawls"]
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

text: "text.template.test.dummy": "Dummy"
text: "text.template.test.crawler": "Crawler"
