package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

actions: [string]: content.#ActionDef

actions: {
	// A step: one body channel, lands at once, then the body recovers for a step's worth of ticks.
	// The step effect arrives with M1-10; until then it is a placeholder that moves nobody.
	"action.step": {
		name: "text.action.step"
		key:  "step"
		uses: [{tag: "body", count: 1}]
		target:   "direction"
		windup:   "0"
		recovery: "step_ticks"
		on_land: [{do: "step", params: {}}]
		tags: ["movement"]
	}
}

text: {
	"text.action.step": "Step"
}
