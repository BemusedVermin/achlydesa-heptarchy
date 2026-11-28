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

// A body that dies gives the dead from that tick, which moving clay needs.
_corpseGivesDead: {
	on:   "combat.body_died"
	if:   ""
	role: "target"
	do: [{do: "afford", params: {bodies: "target", affordance: "affordance.the_dead"}}]
}

// Being hit raises safety, which is terror. A template whose bodies carry drive.safety lists it.
_terrorOnHit: {
	on:   "combat.part_damaged"
	role: "target"
	do: [{do: "satisfy", params: {bodies: "target", drive: "drive.safety", amount: "-2"}}]
}

// Seeing a neighbour die raises safety in those who see it. The template of the body that dies lists
// it. Placeholder amount.
_terrorOnDeath: {
	on:   "combat.body_died"
	role: "target"
	do: [{do: "satisfy", params: {bodies: "witnesses", drive: "drive.safety", amount: "-3"}}]
}

// What a person knows: the seven forms (tagged form in forms.cue), a step, and the types innate to a
// body. The design gives everyone Cut and Name, and the Reaper Press (the thrukk grip). Relic is
// innate to grafts: the type choice offers it to a body that carries one.
_actions: [
	"action.step", "action.strike", "action.sweep", "action.sever", "action.pin", "action.stand",
	"action.give", "action.graft", "action.take", "action.put", "action.consume",
]
_innateTypes: ["type.cut", "type.name", "type.press"]

_#Person: {
	actions:   _actions
	body:      "body.human"
	protected: true
	faction:   "faction.party"
	tags: ["graftable"]
	lists: types: _innateTypes
	triggers: [_fallInjury, _corpseGivesDead]
	...
}

_#Hunter: {
	actor: controller: "hunt"
	actions: [...string] | *["action.step", "action.strike"]
	lists: types: ["type.cut"]
	faction: string | *"faction.wild"
	...
}

// A creature with a mind. It still runs on the hunt controller: a body on the mind controller fights
// only through goal.fight, which needs terror and sight to match the hunt it replaces.
_#Minded: _#Hunter & {
	mind: planner: "htn"
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
	"template.creature.moving_clay": _#Minded & {
		actions: ["action.step", "action.strike", "action.consume_corpse"]
		faction: "faction.clay"
		drives: "drive.the_dead": 1
		name: "text.template.creature.moving_clay"
		appearance: {glyph: "c", color: "yellow"}
		body: "body.moving_clay"
		triggers: [_corpseGivesDead]
	}
	// A thing of the deep that has never died. A part cut off it sometimes crawls away and fights.
	// The 0.3 is a placeholder, for tuning.
	"template.creature.remnant": _#Minded & {
		drives: "drive.old_ground": 1
		name: "text.template.creature.remnant"
		appearance: {glyph: "r", color: "magenta"}
		body: "body.remnant"
		triggers: [{
			on:   "combat.part_severed"
			role: "target"
			if:   "roll < 0.3"
			do: [{do: "spawn_from_part", params: {faction: "faction.crawling", controller: "hunt"}}]
		}, _corpseGivesDead]
	}
	// A flesh creature with a skull and an eye: Press cracks the bone, Light blinds the eye.
	"template.creature.beast": _#Minded & {
		name: "text.template.creature.beast"
		appearance: {glyph: "b", color: "red"}
		body: "body.beast"
		triggers: [_corpseGivesDead]
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
	// A well refills on its own. A cistern and a larder are containers: what is in them is what
	// someone carried there, so the cistern starts empty. The numbers are placeholders, for tuning.
	"template.source.well": {
		name: "text.template.source.well"
		appearance: {glyph: "o", color: "blue"}
		source: {item: "item.water", max: 20, refill: "2"}
		affordances: ["affordance.water"]
		tags: ["stock"]
	}
	"template.container.cistern": {
		name: "text.template.container.cistern"
		appearance: {glyph: "u", color: "blue"}
		container: true
		affordances: ["affordance.water", "affordance.store"]
		tags: ["stock"]
	}
	"template.source.lake": {
		name: "text.template.source.lake"
		appearance: {glyph: "~", color: "blue"}
		source: {item: "item.water", max: 200, refill: "10"}
		affordances: ["affordance.water"]
		tags: ["stock"]
	}
	"template.source.field": {
		name: "text.template.source.field"
		appearance: {glyph: "\"", color: "green"}
		source: {item: "item.ration", max: 30, refill: "1"}
		affordances: ["affordance.food"]
		tags: ["stock"]
	}
	"template.container.coal_pile": {
		name: "text.template.container.coal_pile"
		appearance: {glyph: "^", color: "white"}
		container: true
		affordances: ["affordance.store"]
		tags: ["stock"]
		inventory: [{item: "item.coal", count: 20}]
	}
	"template.container.larder": {
		name: "text.template.container.larder"
		appearance: {glyph: "u", color: "yellow"}
		container: true
		affordances: ["affordance.food", "affordance.store"]
		tags: ["stock"]
		inventory: [{item: "item.ration", count: 5}]
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
	"text.template.creature.moving_clay": "Moving clay"
	"text.template.test.crawler":         "Crawler"
	"text.template.person.reaper":        "The Reaper"
	"text.template.person.salken":        "Salken"
	"text.template.unknown":              "Someone"
	"text.template.source.well":          "A well"
	"text.template.container.cistern":    "A cistern"
	"text.template.container.larder":     "A larder"
	"text.template.source.lake":          "A lake"
	"text.template.source.field":         "A field"
	"text.template.container.coal_pile":  "A coal pile"
}
