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
	// Eating and drinking take one item and lower the need by a placeholder half. With a hostile in sight they stay in
	// the menu, disabled.
	"action.eat": {
		name: "text.action.eat"
		key:  "eat"
		uses: [{tag: "body", count: 1}]
		target:   "self"
		requires: "inventory[\"item.ration\"] > 0"
		allowed:  "!source.alert"
		refusal:  "text.actions.on_alert"
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
		allowed:  "!source.alert"
		refusal:  "text.actions.on_alert"
		windup:   "0"
		recovery: "20"
		on_start: [{do: "consume", params: {item: "item.lake_water"}}]
		on_land: [{do: "satisfy", params: {need: "need.water", amount: "0.5"}}]
		tags: ["supply"]
	}
	// Residents' versions: one item from the town stock, from beside a stock mark. Placeholder numbers.
	"action.drink_stock": {
		name: "text.action.drink_stock"
		key:  "drink_stock"
		uses: [{tag: "body", count: 1}]
		target:   "self"
		allowed:  "!source.alert"
		refusal:  "text.actions.on_alert"
		windup:   "0"
		recovery: "20"
		on_start: [{do: "take_stock", params: {item: "item.lake_water"}}]
		on_land: [{do: "satisfy", params: {need: "need.water", amount: "0.5"}}]
		tags: ["supply"]
	}
	"action.eat_stock": {
		name: "text.action.eat_stock"
		key:  "eat_stock"
		uses: [{tag: "body", count: 1}]
		target:   "self"
		allowed:  "!source.alert"
		refusal:  "text.actions.on_alert"
		windup:   "0"
		recovery: "20"
		on_start: [{do: "take_stock", params: {item: "item.ration"}}]
		on_land: [{do: "satisfy", params: {need: "need.food", amount: "0.5"}}]
		tags: ["supply"]
	}
	"action.sleep": {
		name: "text.action.sleep"
		key:  "sleep"
		uses: [{tag: "body", count: 1}]
		target:   "self"
		windup:   "0"
		recovery: "40"
		on_land: [{do: "satisfy", params: {need: "need.rest", amount: "0.3"}}]
	}
	// Moving clay laying hold of a body, which may be dead. Placeholder numbers.
	"action.hold": {
		name: "text.action.hold"
		key:  "hold"
		uses: [{tag: "body", count: 1}]
		target:   "body"
		range:    1
		windup:   "2"
		recovery: "20"
		on_land: [{do: "satisfy", params: {need: "need.hold_dead", amount: "0.3"}}]
		tags: ["corpse"]
	}
	// Rest in a town until the next dawn, mending one integrity per 120 ticks rested (placeholder), and
	// waking here after a wipe.
	"action.rest": {
		name: "text.action.rest"
		key:  "rest"
		uses: [{tag: "body", count: 1}]
		target:   "self"
		requires: "\"town\" in source.map_tags && !source.alert"
		windup:   "day_ticks - now % day_ticks"
		recovery: "0"
		on_land: [
			{do: "heal", params: {parts: "self:all", amount: "elapsed / 120"}},
			{do: "set_rest_place", params: {}},
		]
	}
	// For tests: a bite at a body next to the source, which may be a corpse.
	"action.gnaw": {
		name: "text.action.gnaw"
		key:  "gnaw"
		uses: [{tag: "body", count: 1}]
		target:   "body"
		range:    1
		windup:   "2"
		recovery: "5"
		tags: ["corpse"]
	}
}

text: {
	"text.action.gnaw":  "Gnaw"
	"text.action.hold":  "Hold"
	"text.action.step":  "Step"
	"text.action.eat":   "Eat"
	"text.action.drink": "Drink"

	"text.action.drink_stock": "Drink from the stock"
	"text.action.eat_stock":   "Eat from the stock"
	"text.action.sleep":       "Sleep"
	"text.action.rest":        "Rest until dawn"
}
