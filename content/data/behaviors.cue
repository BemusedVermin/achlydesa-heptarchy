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
	"behavior.test.idle": {
		name: "text.behavior.test.idle"
		do:   "idle"
		consider: ["0.5"]
	}
}

text: {
	"text.behavior.test.wander":   "wandering"
	"text.behavior.test.to_water": "going for water"
	"text.behavior.test.idle":     "standing about"
}
