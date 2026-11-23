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
// Cut and Name, and the Reaper Press (the thrukk grip). Relic is innate to grafts: the type choice
// offers it to a body that carries one.
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
		faction:   "faction.party"
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
		faction:   "faction.party"
		tags: ["graftable"]
		lists: types: _innateTypes
		triggers: [_fallInjury]
	}
	// A clay-cored creature of the soop. It hunts the party and bites.
	"template.creature.crawler": {
		name: "text.template.creature.crawler"
		appearance: {glyph: "c", color: "yellow"}
		actor: controller: "hunt"
		actions: ["action.step", "action.strike"]
		lists: types: ["type.cut"]
		body:    "body.crawler"
		faction: "faction.deep"
	}
	// A thing of the deep that has never died. A part cut off it sometimes crawls away and fights.
	// The 0.3 is a placeholder for M7's tuning.
	"template.creature.remnant": {
		name: "text.template.creature.remnant"
		appearance: {glyph: "r", color: "magenta"}
		actor: controller: "hunt"
		body:    "body.remnant"
		faction: "faction.deep"
		triggers: [{
			on:   "combat.part_severed"
			role: "target"
			if:   "roll < 0.3"
			do: [{do: "spawn_from_part", params: {faction: "faction.crawling", controller: "hunt"}}]
		}]
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

text: {
	"text.template.creature.remnant": "The remnant"
	"text.template.test.dummy":      "Dummy"
	"text.template.creature.crawler": "The crawler"
	"text.template.test.crawler":     "Crawler"
}
