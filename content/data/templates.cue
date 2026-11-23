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

// What a person knows: the seven forms (tagged form in forms.cue) and a step, and the types innate to a body. The design gives everyone
// Cut and Name, and the Reaper Press (the thrukk grip). Relic is innate to grafts: the type choice
// offers it to a body that carries one.
_actions: [
	"action.step", "action.strike", "action.sweep", "action.sever", "action.pin", "action.stand",
	"action.give", "action.graft",
]
_innateTypes: ["type.cut", "type.name", "type.press"]

templates: {
	"template.person.reaper": {
		name: "text.template.person.reaper"
		appearance: {glyph: "@", color: "red"}
		actor: controller: "player"
		actions:   _actions
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
		actions:   _actions
		body:      "body.human"
		protected: true
		faction:   "faction.party"
		tags: ["graftable"]
		lists: types: _innateTypes
		triggers: [_fallInjury]
	}
	// A creature of the soop, all clay. It hunts the party and bites.
	"template.creature.moving_clay": {
		name: "text.template.creature.moving_clay"
		appearance: {glyph: "c", color: "yellow"}
		actor: controller: "hunt"
		actions: ["action.step", "action.strike"]
		lists: types: ["type.cut"]
		body:    "body.moving_clay"
		faction: "faction.wild"
	}
	// A thing of the deep that has never died. A part cut off it sometimes crawls away and fights.
	// The 0.3 is a placeholder for M7's tuning.
	"template.creature.remnant": {
		name: "text.template.creature.remnant"
		appearance: {glyph: "r", color: "magenta"}
		actor: controller: "hunt"
		body:    "body.remnant"
		faction: "faction.wild"
		actions: ["action.step", "action.strike"]
		lists: types: ["type.cut"]
		triggers: [{
			on:   "combat.part_severed"
			role: "target"
			if:   "roll < 0.3"
			do: [{do: "spawn_from_part", params: {faction: "faction.crawling", controller: "hunt"}}]
		}]
	}
	// A flesh creature with a skull and an eye: Press cracks the bone, Light blinds the eye.
	"template.creature.beast": {
		name: "text.template.creature.beast"
		appearance: {glyph: "b", color: "red"}
		actor: controller: "hunt"
		actions: ["action.step", "action.strike"]
		lists: types: ["type.cut"]
		body:    "body.beast"
		faction: "faction.wild"
	}
	// What a remnant's arm fights as once it crawls off (its part's crawls_as): the arm is its whole
	// body, so it has none of its own, and spawn_from_part gives its faction and controller. It knows
	// every form its one action grant can pay for (Strike, Sever, Pin), swung with Cut, and has no
	// crawls tag, so Stand does not freeze it.
	"template.crawling.remnant_arm": {
		name: "text.template.crawling.remnant_arm"
		appearance: {glyph: "a", color: "magenta"}
		actions: ["action.strike", "action.sever", "action.pin"]
		lists: types: ["type.cut"]
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
	"text.template.creature.remnant":     "The remnant"
	"text.template.creature.beast":       "The beast"
	"text.template.crawling.remnant_arm": "A crawling arm"
	"text.template.test.dummy":           "Dummy"
	"text.template.creature.moving_clay": "The moving clay"
	"text.template.test.crawler":         "Crawler"
	"text.template.person.reaper":        "The Reaper"
	"text.template.person.salken":        "Salken"
	"text.template.unknown":              "Someone"
}
