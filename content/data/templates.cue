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

// A creature that acts on its needs. It brings its own actions and behaviors.
_#Monster: {
	actor: {controller: "agents", ...}
	lists: types: ["type.cut"]
	faction: *"faction.wild" | string
	...
}

_residentBehaviors: [
	"behavior.resident.drink", "behavior.resident.eat", "behavior.resident.sleep",
	"behavior.resident.leave", "behavior.resident.wander",
]

// Townspeople. Safety, belonging, faith, duty and grief have placeholder curves and nothing meets them yet.
_#Resident: {
	actor: controller: "agents"
	body:    "body.human"
	faction: "faction.people"
	actions: ["action.step", "action.drink_stock", "action.eat_stock", "action.sleep"]
	needs: {
		"need.water": 0, "need.food": 0, "need.rest":  0, "need.safety": 0, "need.belonging": 0
		"need.faith": 0, "need.duty": 0, "need.grief": 0
	}
	...
}

// A generic townsperson: replaced on a timer, so it carries only the basic needs of food and safety.
_#Generic: {
	actor: controller: "agents"
	body:    "body.human"
	faction: "faction.people"
	actions: ["action.step", "action.eat_stock"]
	needs: {"need.food": 0, "need.safety": 0}
	behaviors: ["behavior.resident.eat", "behavior.resident.wander"]
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
	// A creature of the soop, all clay. It holds the dead and fights what it sees.
	"template.creature.moving_clay": _#Monster & {
		name: "text.template.creature.moving_clay"
		appearance: {glyph: "c", color: "yellow"}
		body: "body.moving_clay"
		actions: ["action.step", "action.strike", "action.hold"]
		needs: {"need.hold_dead": 0.5}
		behaviors: ["behavior.clay.hold", "behavior.monster.engage"]
	}
	// A thing of the deep that has never died. It leaves decayed ground, and a part cut off it
	// sometimes crawls away and fights. The 0.3 is a placeholder, for tuning.
	"template.creature.remnant": _#Monster & {
		name: "text.template.creature.remnant"
		appearance: {glyph: "r", color: "magenta"}
		body: "body.remnant"
		actions: ["action.step", "action.strike"]
		needs: {"need.undecayed": 0}
		behaviors: ["behavior.remnant.leave", "behavior.monster.engage"]
		triggers: [{
			on:   "combat.part_severed"
			role: "target"
			if:   "roll < 0.3"
			do: [{do: "spawn_from_part", params: {faction: "faction.crawling", controller: "hunt"}}]
		}]
	}
	// A husk that hunts the living, and the slowly dying first.
	"template.creature.soul_feeder": _#Monster & {
		name: "text.template.creature.soul_feeder"
		appearance: {glyph: "s", color: "red"}
		actor: chooser: "agents"
		body: "body.soul_feeder"
		actions: ["action.step", "action.strike"]
		needs: {"need.soul": 0.5}
		behaviors: ["behavior.feeder.hunt"]
	}
	// A creature of the archon: it guards the believers and attacks the party near them.
	"template.creature.archon_warden": _#Monster & {
		name: "text.template.creature.archon_warden"
		appearance: {glyph: "w", color: "red"}
		faction: "faction.archon"
		body:    "body.beast"
		actions: ["action.step", "action.strike"]
		behaviors: ["behavior.archon.guard"]
	}
	// One of the archon's worshippers.
	"template.resident.believer": {
		name: "text.template.resident.believer"
		appearance: {glyph: "b", color: "white"}
		actor: controller: "agents"
		body:    "body.human"
		faction: "faction.believers"
		actions: ["action.step"]
		tags: ["believer"]
		needs: {"need.faith": 0}
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
	// A body that goes for the slowly dying, picking its attacks with the agents chooser.
	"template.test.predator": {
		name: "text.template.test.predator"
		appearance: {glyph: "p", color: "red"}
		actor: {controller: "hunt", chooser: "agents"}
		body:    "body.moving_clay"
		faction: "faction.wild"
		actions: ["action.step", "action.strike"]
		lists: types: ["type.cut"]
	}
	// A body that steps out of reach of a swing aimed at it.
	"template.test.dodger": {
		name: "text.template.test.dodger"
		appearance: {glyph: "q", color: "green"}
		actor: controller: "agents"
		body:    "body.human"
		faction: "faction.wild"
		actions: ["action.step"]
		behaviors: ["behavior.test.dodge"]
	}
	"template.resident.plain": _#Resident & {
		name: "text.template.resident.plain"
		appearance: {glyph: "p", color: "white"}
		behaviors: _residentBehaviors
	}
	// Carries water from the well to the cistern by day, as well as living as a resident does.
	"template.resident.hauler": _#Resident & {
		name: "text.template.resident.hauler"
		appearance: {glyph: "h", color: "cyan"}
		behaviors: [for b in _residentBehaviors {b}, "behavior.resident.haul_water"]
	}
	"template.generic.guard": _#Generic & {
		name: "text.template.generic.guard"
		appearance: {glyph: "G", color: "white"}
	}
	"template.generic.field_hand": _#Generic & {
		name: "text.template.generic.field_hand"
		appearance: {glyph: "f", color: "green"}
	}
	// A well the haulers draw from: a thing with water in it, not a body to fight.
	"template.town.well": {
		name: "text.template.town.well"
		appearance: {glyph: "o", color: "blue"}
		tags: ["well"]
		inventory: [{item: "item.lake_water", count: 40}]
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
	"text.template.creature.remnant":       "The remnant"
	"text.template.creature.beast":         "The beast"
	"text.template.crawling.remnant_arm":   "A crawling arm"
	"text.template.test.dummy":             "Dummy"
	"text.template.creature.moving_clay":   "Moving clay"
	"text.template.creature.soul_feeder":   "The soul-feeder"
	"text.template.creature.archon_warden": "The archon's warden"
	"text.template.resident.believer":      "A believer"
	"text.template.test.crawler":           "Crawler"
	"text.template.test.drinker":           "Drinker"
	"text.template.test.eater":             "Eater"
	"text.template.test.hauler":            "Hauler"
	"text.template.test.well":              "Well"
	"text.template.test.cistern":           "Cistern"
	"text.template.test.runner":            "Runner"
	"text.template.test.guard":             "Guard"
	"text.template.test.engager":           "Fighter"
	"text.template.test.predator":          "Predator"
	"text.template.test.dodger":            "Dodger"
	"text.template.person.reaper":          "The Reaper"
	"text.template.person.salken":          "Salken"
	"text.template.generic.guard":          "A guard"
	"text.template.generic.field_hand":     "A field hand"
	"text.template.resident.plain":         "A resident"
	"text.template.resident.hauler":        "A water hauler"
	"text.template.town.well":              "The well"
	"text.template.unknown":                "Someone"
}
