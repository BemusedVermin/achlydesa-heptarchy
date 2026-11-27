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
		recovery: "step_ticks"
		on_land: [{do: "step", params: {}}]
		tags: ["movement"]
	}
	// Tell a near body one thing the speaker knows. The tell effect picks the fact when it lands.
	"action.chat": {
		name: "text.action.chat"
		key:  "chat"
		uses: [{tag: "body", count: 1}]
		target:   "body"
		range:    2
		sight:    true
		windup:   "0"
		recovery: "step_ticks"
		on_land: [{do: "tell", params: {}}]
		tags: ["speech"]
	}
}

text: {
	"text.action.step": "Step"
	"text.action.chat": "Chat"
}
