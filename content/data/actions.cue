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
	// Sleep where there is a place to rest: the long recovery is the sleep. First values, for tuning.
	"action.sleep": {
		name: "text.action.sleep"
		key:  "sleep"
		uses: [{tag: "body", count: 1}]
		target:   "self"
		windup:   "0"
		recovery: "12"
		on_land: [{do: "satisfy", params: {bodies: "self", drive: "drive.rest", amount: "100"}}]
	}
	// Consume a corpse next to the body: it leaves the world and satisfies the want of the dead.
	"action.consume_corpse": {
		name: "text.action.consume_corpse"
		key:  "consume_corpse"
		uses: [{tag: "body", count: 1}]
		target: "body"
		targets: {corpse: true}
		range:    1
		windup:   "0"
		recovery: "6"
		on_land: [
			{do: "satisfy", params: {bodies: "self", drive: "drive.the_dead", amount: "100"}},
			{do: "remove_corpse", params: {}},
		]
	}
}

text: {
	"text.action.chat":           "Chat"
	"text.action.step":           "Step"
	"text.action.take":           "Take"
	"text.action.put":            "Put"
	"text.action.consume":        "Consume"
	"text.action.sleep":          "Sleep"
	"text.action.consume_corpse": "Consume a corpse"
}
