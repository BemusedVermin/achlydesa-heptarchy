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
	// Take, put and consume are open to anyone next to what they act on, agent or party alike.
	"action.take": {
		name: "text.action.take"
		key:  "take"
		uses: [{tag: "grant:action", count: 1}]
		target: "body"
		targets: {thing: true, tags: ["stock"]}
		range: 1
		choices: ["ware"]
		windup:   "0"
		recovery: "2"
		on_land: [{do: "take", params: {choice: "ware", count: "1"}}]
	}
	"action.put": {
		name: "text.action.put"
		key:  "put"
		uses: [{tag: "grant:action", count: 1}]
		target: "body"
		targets: {thing: true, tags: ["stock"]}
		range: 1
		choices: ["carried"]
		windup:   "0"
		recovery: "2"
		on_land: [{do: "put", params: {choice: "carried", count: "1"}}]
	}
	"action.consume": {
		name: "text.action.consume"
		key:  "consume"
		uses: [{tag: "grant:action", count: 1}]
		target: "self"
		choices: ["carried"]
		windup:   "0"
		recovery: "2"
		on_land: [{do: "consume", params: {choice: "carried"}}]
	}
}

text: {
	"text.action.chat":    "Chat"
	"text.action.step":    "Step"
	"text.action.take":    "Take"
	"text.action.put":     "Put"
	"text.action.consume": "Consume"
}
