package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

// What a body runs low on. A template lists the needs its bodies carry, with a starting level from
// 0 (met) to 1 (desperate); the level rises at every phase boundary. No template carries one yet.
needs: [string]: content.#NeedDef

// A day is four phases. Placeholder numbers, for tuning: a body is slowed after about two days
// without (eight phases at 0.0625) and loses a point of torso integrity each phase from about four.
// The slowed status outlasts the longest phase, so each phase boundary keeps it going.
_#Hunger: {
	rise: "0.0625"
	urge: "level"
	thresholds: [{
		at: 0.5
		do: [{do: "add_status", params: {bodies: "self", status: "status.slowed", ticks: "120"}}]
	}, {
		at: 0.9
		do: [{do: "damage", params: {parts: "self:torso", amount: "1"}}]
	}]
	...
}

needs: {
	"need.water": _#Hunger & {name: "text.need.water"}
	"need.food": _#Hunger & {name: "text.need.food"}
}

text: {
	"text.need.water": "water"
	"text.need.food":  "food"
}
