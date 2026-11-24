package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

actions: [string]: content.#ActionDef

actions: {
	// A step: one body channel, lands at once, then the body recovers for a step's worth of ticks.
	"action.step": {
		name: "text.action.step"
		key:  "step"
		uses: [{tag: "body", count: 1}]
		target:   "direction"
		windup:   "0"
		recovery: "rule.slowed ? step_ticks * 2 : step_ticks"
		on_land: [{do: "step", params: {}}]
		tags: ["movement"]
	}
	// Eating and drinking take one item and lower the need by a placeholder half.
	"action.eat": {
		name: "text.action.eat"
		key:  "eat"
		uses: [{tag: "body", count: 1}]
		target:   "self"
		requires: "inventory[\"item.ration\"] > 0"
		windup:   "0"
		recovery: "20"
		on_start: [{do: "consume", params: {item: "item.ration"}}]
		on_land: [{do: "satisfy", params: {need: "need.food", amount: "0.5"}}]
		tags: ["supply"]
	}
	"action.drink": {
		name: "text.action.drink"
		key:  "drink"
		uses: [{tag: "body", count: 1}]
		target:   "self"
		requires: "inventory[\"item.lake_water\"] > 0"
		windup:   "0"
		recovery: "20"
		on_start: [{do: "consume", params: {item: "item.lake_water"}}]
		on_land: [{do: "satisfy", params: {need: "need.water", amount: "0.5"}}]
		tags: ["supply"]
	}
}

text: {
	"text.action.step":  "Step"
	"text.action.eat":   "Eat"
	"text.action.drink": "Drink"
}
