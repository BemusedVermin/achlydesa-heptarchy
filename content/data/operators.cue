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
