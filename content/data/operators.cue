package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

operators: [string]: content.#Operator

operators: {
	// Walk to a landmark that gives water, one step at a time, until it is near.
	"operator.go_to": {
		pre: [{fact: "known", tag: "affordance.water"}, {fact: "near", tag: "affordance.water", not: true}]
		eff: [{fact: "near", tag: "affordance.water"}]
		cost: 2
		does: {action: "action.step", target: "go_to"}
		kind: "travel"
	}
	// Step at random, to find what is not yet known.
	"operator.wander": {
		pre: [{fact: "known", tag: "affordance.water", not: true}]
		eff: [{fact: "known", tag: "affordance.water"}]
		cost: 5
		does: {action: "action.step", target: "wander"}
		kind: "travel"
	}
}

// Fighting and fleeing. Engage's picker hands the body to the hunt controller instead of starting the
// action; flee steps to the neighbouring cell farthest from the hostiles believed near.
operators: {
	"operator.engage": {
		eff: [{fact: "drive_met", tag: "drive.safety"}]
		cost: 1
		does: {action: "action.strike", target: "engage"}
		kind: "fight"
	}
	"operator.flee": {
		pre: [{fact: "drive_met", tag: "drive.safety", not: true}]
		eff: [{fact: "drive_met", tag: "drive.safety"}]
		cost: 1
		does: {action: "action.step", target: "flee"}
		kind: "travel"
	}
}

// Taking, putting away and consuming, for water and for food.
operators: {
	"operator.take": {
		pre: [{fact: "near", tag: "affordance.water"}]
		eff: [{fact: "has", tag: "drinkable", n: 1}]
		cost: 1
		does: {action: "action.take", target: "take"}
		kind: "work"
	}
	"operator.take_food": {
		pre: [{fact: "near", tag: "affordance.food"}]
		eff: [{fact: "has", tag: "edible", n: 1}]
		cost: 1
		does: {action: "action.take", target: "take"}
		kind: "work"
	}
	"operator.put": {
		pre: [{fact: "has", tag: "drinkable", n: 1}, {fact: "near", tag: "affordance.store"}]
		eff: [{fact: "has", tag: "drinkable", n: -1}]
		cost: 1
		does: {action: "action.put", target: "put"}
		kind: "work"
	}
	"operator.put_food": {
		pre: [{fact: "has", tag: "edible", n: 1}, {fact: "near", tag: "affordance.store"}]
		eff: [{fact: "has", tag: "edible", n: -1}]
		cost: 1
		does: {action: "action.put", target: "put"}
		kind: "work"
	}
	"operator.consume": {
		pre: [{fact: "has", tag: "drinkable", n: 1}]
		eff: [{fact: "has", tag: "drinkable", n: -1}, {fact: "drive_met", tag: "drive.water"}]
		cost: 1
		does: {action: "action.consume", target: "consume"}
		kind: "work"
	}
	"operator.eat": {
		pre: [{fact: "has", tag: "edible", n: 1}]
		eff: [{fact: "has", tag: "edible", n: -1}, {fact: "drive_met", tag: "drive.food"}]
		cost: 1
		does: {action: "action.consume", target: "consume"}
		kind: "work"
	}
}
