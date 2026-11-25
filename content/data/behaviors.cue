package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

// What agents choose between. A template under the agents controller lists the behaviors it weighs;
// each is scored by the product of its considerations, and the best runs.
behaviors: [string]: content.#BehaviorDef

agents: content.#AgentsConfig

// Placeholder numbers, for tuning.
agents: {
	// An agent keeps doing what it does unless another behavior beats it by this much.
	keep_bonus: 0.1
	// How far, in hexes and in depths, an agent looks for a cell of a terrain.
	cell_radius: 6
	// The rule that rates an attack for a creature whose template picks with the agents chooser.
	score_rule: "agents_score"
	// A body with a vital part at or below this fraction of its integrity is dying.
	dying_below: 0.5
	// How far from a mark or from the party a promoted body may stand, in hexes.
	arrival_radius: 8
}

// For tests: a body that wanders while it has water, and walks to the map's water mark once
// thirsty, and one that stands about.
behaviors: {
	"behavior.test.wander": {
		name: "text.behavior.test.wander"
		do:   "wander"
		consider: ["0.5"]
	}
	"behavior.test.to_water": {
		name: "text.behavior.test.to_water"
		do:   "move_toward"
		consider: ["source.needs.water"]
		target: {kind: "mark", mark: "water"}
	}
	"behavior.test.gnaw_corpse": {
		name: "text.behavior.test.gnaw_corpse"
		do:   "consume"
		params: action: "action.gnaw"
		consider: ["0.5"]
		target: {kind: "body", dead: true}
	}
	"behavior.test.haul_water": {
		name: "text.behavior.test.haul_water"
		do:   "carry"
		params: {item: "item.lake_water", count: "2", to: "cistern"}
		consider: ["0.5"]
		target: {kind: "body", tags: ["well"]}
	}
	// What a recruit does while terror drives it: run from whatever is hostile in sight.
	"behavior.recruit.flee": {
		name: "text.behavior.recruit.flee"
		do:   "flee"
		consider: ["1.0"]
	}
	"behavior.test.run": {
		name: "text.behavior.test.run"
		do:   "flee"
		consider: ["0.5"]
	}
	"behavior.test.guard_post": {
		name: "text.behavior.test.guard_post"
		do:   "guard"
		params: radius: "2"
		consider: ["0.5"]
		target: {kind: "mark", mark: "post"}
	}
	"behavior.test.fight": {
		name: "text.behavior.test.fight"
		do:   "engage"
		consider: ["0.5"]
	}
	"behavior.test.dodge": {
		name: "text.behavior.test.dodge"
		do:   "dodge"
		params: within: "4"
		consider: ["threatened >= 0 && threatened <= 4 ? 1 : 0"]
	}
	"behavior.test.idle": {
		name: "text.behavior.test.idle"
		do:   "idle"
		consider: ["0.5"]
	}
}

// What monsters do. Placeholder numbers, for tuning: a creature fights a hostile it is within
// sight of unless its own want is stronger.
behaviors: {
	"behavior.monster.engage": {
		name: "text.behavior.monster.engage"
		do:   "engage"
		consider: ["distance <= 8 ? 0.7 : 0"]
		target: {kind: "body", hostile: true, dead: false}
	}
	"behavior.clay.hold": {
		name: "text.behavior.clay.hold"
		do:   "consume"
		params: action: "action.hold"
		consider: ["0.2 + 0.6 * source.needs.hold_dead"]
		target: {kind: "body", dead: true}
	}
	"behavior.remnant.leave": {
		name: "text.behavior.remnant.leave"
		do:   "leave"
		consider: ["region.props.decay > 0.3 ? source.needs.undecayed : 0"]
		// The gate to the ground that eases the want most. The want is what makes it leave; the
		// ranking asks where that want would fall furthest.
		rank: "source.needs.undecayed * (0.3 - region.props.decay)"
	}
	"behavior.feeder.hunt": {
		name: "text.behavior.feeder.hunt"
		do:   "engage"
		consider: ["distance <= 8 ? 0.6 + 0.4 * source.needs.soul : 0"]
		target: {kind: "body", hostile: true, dead: false}
	}
	"behavior.archon.guard": {
		name: "text.behavior.archon.guard"
		do:   "guard"
		params: radius: "3"
		consider: ["0.8"]
		target: {kind: "body", tags: ["believer"]}
	}
}

// What people do in a town. Roles are the lists in templates.cue and schedules are the phase
// considerations here. Placeholder numbers, for tuning.
behaviors: {
	"behavior.resident.drink": {
		name: "text.behavior.resident.drink"
		do:   "consume"
		params: action: "action.drink_stock"
		consider: [
			"source.needs.water > 0.3 ? 0.5 + source.needs.water / 2 : 0",
			"stock[\"item.lake_water\"] > 0 ? 1 : 0",
		]
		target: {kind: "mark", mark: "cistern"}
	}
	"behavior.resident.eat": {
		name: "text.behavior.resident.eat"
		do:   "consume"
		params: action: "action.eat_stock"
		consider: [
			"source.needs.food > 0.3 ? 0.4 + source.needs.food / 2 : 0",
			"stock[\"item.ration\"] > 0 ? 1 : 0",
		]
		target: {kind: "mark", mark: "cistern"}
	}
	"behavior.resident.sleep": {
		name: "text.behavior.resident.sleep"
		do:   "consume"
		params: action: "action.sleep"
		consider: ["phase == \"phase.night\" ? 0.4 : 0"]
		target: {kind: "mark", mark: "house"}
	}
	"behavior.resident.haul_water": {
		name: "text.behavior.resident.haul_water"
		do:   "carry"
		params: {item: "item.lake_water", count: "2", to: "cistern"}
		consider: ["phase == \"phase.night\" ? 0 : 0.5"]
		target: {kind: "body", tags: ["well"]}
	}
	"behavior.resident.leave": {
		name: "text.behavior.resident.leave"
		do:   "leave"
		consider: ["source.needs.water > 0.5 && stock[\"item.lake_water\"] <= 0 ? 1 : 0"]
	}
	"behavior.resident.wander": {
		name: "text.behavior.resident.wander"
		do:   "wander"
		consider: ["0.1"]
	}
}

text: {
	"text.behavior.monster.engage":      "fighting"
	"text.behavior.clay.hold":           "holding the dead"
	"text.behavior.remnant.leave":       "leaving decayed ground"
	"text.behavior.feeder.hunt":         "hunting the living"
	"text.behavior.archon.guard":        "guarding the believers"
	"text.behavior.resident.drink":      "drinking"
	"text.behavior.resident.eat":        "eating"
	"text.behavior.resident.sleep":      "sleeping"
	"text.behavior.resident.haul_water": "hauling water"
	"text.behavior.resident.leave":      "leaving the town"
	"text.behavior.resident.wander":     "wandering"
	"text.behavior.test.wander":         "wandering"
	"text.behavior.test.to_water":       "going for water"
	"text.behavior.test.gnaw_corpse":    "gnawing a corpse"
	"text.behavior.test.haul_water":     "hauling water"
	"text.behavior.test.run":            "running"
	"text.behavior.recruit.flee":        "fleeing"
	"text.behavior.test.guard_post":     "guarding a post"
	"text.behavior.test.fight":          "fighting"
	"text.behavior.test.dodge":          "dodging"
	"text.behavior.test.idle":           "standing about"
}
