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

// What a person knows: the seven forms (tagged form in forms.cue), a step, and the types innate to a
// body. The design gives everyone Cut and Name, and the Reaper Press (the thrukk grip). Relic is
// innate to grafts: the type choice offers it to a body that carries one.
_actions: [
	"action.step", "action.strike", "action.sweep", "action.sever", "action.pin", "action.stand",
	"action.give", "action.graft", "action.eat", "action.drink",
]
_innateTypes: ["type.cut", "type.name", "type.press"]

_#Person: {
	actions:   _actions
	body:      "body.human"
	protected: true
	faction:   "faction.party"
	needs: {"need.food": 0, "need.water": 0}
	tags: ["graftable"]
	lists: types: _innateTypes
	triggers: [_fallInjury]
	...
}

_#Hunter: {
	actor: controller: "hunt"
	actions: ["action.step", "action.strike"]
	lists: types: ["type.cut"]
	faction: "faction.wild"
	...
}

templates: {
	"template.person.reaper": _#Person & {
		name: "text.template.person.reaper"
		appearance: {glyph: "@", color: "red"}
		actor: controller: "player"
	}
	"template.person.salken": _#Person & {
		name: "text.template.person.salken"
		appearance: {glyph: "@", color: "cyan"}
		actor: controller: "follow"
	}
	// A creature of the soop, all clay.
	"template.creature.moving_clay": _#Hunter & {
		name: "text.template.creature.moving_clay"
		appearance: {glyph: "c", color: "yellow"}
		body: "body.moving_clay"
	}
	// A thing of the deep that has never died. A part cut off it sometimes crawls away and fights.
	// The 0.3 is a placeholder, for tuning.
	"template.creature.remnant": _#Hunter & {
		name: "text.template.creature.remnant"
		appearance: {glyph: "r", color: "magenta"}
		body: "body.remnant"
		triggers: [{
			on:   "combat.part_severed"
			role: "target"
			if:   "roll < 0.3"
			do: [{do: "spawn_from_part", params: {faction: "faction.crawling", controller: "hunt"}}]
		}]
	}
	// A flesh creature with a skull and an eye: Press cracks the bone, Light blinds the eye.
	"template.creature.beast": _#Hunter & {
		name: "text.template.creature.beast"
		appearance: {glyph: "b", color: "red"}
		body: "body.beast"
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
	// A body that wanders until thirsty, then walks to the map's water mark.
	"template.test.drinker": {
		name: "text.template.test.drinker"
		appearance: {glyph: "w", color: "blue"}
		actor: controller: "agents"
		body: "body.human"
		actions: ["action.step"]
		needs: {"need.water": 0}
		behaviors: ["behavior.test.wander", "behavior.test.to_water"]
	}
	// Bodies for the behavior primitives' tests.
	"template.test.eater": {
		name: "text.template.test.eater"
		appearance: {glyph: "e", color: "green"}
		actor: controller: "agents"
		body: "body.human"
		actions: ["action.step", "action.gnaw"]
		behaviors: ["behavior.test.gnaw_corpse"]
	}
	"template.test.hauler": {
		name: "text.template.test.hauler"
		appearance: {glyph: "h", color: "green"}
		actor: controller: "agents"
		body: "body.human"
		actions: ["action.step"]
		behaviors: ["behavior.test.haul_water"]
	}
	"template.test.well": {
		name: "text.template.test.well"
		appearance: {glyph: "o", color: "blue"}
		actor: controller: "idle"
		body: "body.human"
		tags: ["well"]
		inventory: [{item: "item.lake_water", count: 5}]
	}
	"template.test.cistern": {
		name: "text.template.test.cistern"
		appearance: {glyph: "u", color: "blue"}
		actor: controller: "idle"
		body: "body.human"
	}
	"template.test.runner": {
		name: "text.template.test.runner"
		appearance: {glyph: "r", color: "green"}
		actor: controller: "agents"
		body:    "body.human"
		faction: "faction.wild"
		actions: ["action.step"]
		behaviors: ["behavior.test.run"]
	}
	"template.test.guard": {
		name: "text.template.test.guard"
		appearance: {glyph: "g", color: "green"}
		actor: controller: "agents"
		body:    "body.human"
		faction: "faction.wild"
		actions: ["action.step", "action.strike"]
		lists: types: ["type.cut"]
		behaviors: ["behavior.test.guard_post"]
	}
	"template.test.engager": {
		name: "text.template.test.engager"
		appearance: {glyph: "f", color: "green"}
		actor: controller: "agents"
		body:    "body.moving_clay"
		faction: "faction.wild"
		actions: ["action.step", "action.strike"]
		lists: types: ["type.cut"]
		behaviors: ["behavior.test.fight"]
	}
	"template.party.main": {
		party: members: ["template.person.reaper", "template.person.salken"]
		// A few of each consumable to start with.
		inventory: [
			{item: "item.ration", count: 6},
			{item: "item.coal", count: 3},
			{item: "item.salk_dust", count: 3},
			{item: "item.lake_water", count: 6},
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
	"text.template.creature.moving_clay": "Moving clay"
	"text.template.test.crawler":         "Crawler"
	"text.template.test.drinker":         "Drinker"
	"text.template.test.eater":           "Eater"
	"text.template.test.hauler":          "Hauler"
	"text.template.test.well":            "Well"
	"text.template.test.cistern":         "Cistern"
	"text.template.test.runner":          "Runner"
	"text.template.test.guard":           "Guard"
	"text.template.test.engager":         "Fighter"
	"text.template.person.reaper":        "The Reaper"
	"text.template.person.salken":        "Salken"
	"text.template.unknown":              "Someone"
}
