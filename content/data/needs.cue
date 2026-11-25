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

// A need no behavior meets yet: it creeps up too slowly to reach any threshold, and has none.
_#Slow: {
	rise: "0.005"
	urge: "level"
	...
}

needs: {
	"need.water": _#Hunger & {name: "text.need.water"}
	"need.food": _#Hunger & {name: "text.need.food"}
	// People's needs. Rest builds by day and is not built at night, while sleeping.
	"need.rest": {
		name: "text.need.rest"
		rise: "phase == \"phase.night\" ? 0 : 0.0625"
		urge: "level"
	}
	// Moving clay wants to hold the dead, and the want builds.
	"need.hold_dead": {
		name: "text.need.hold_dead"
		rise: "0.1"
		urge: "level"
	}
	// A remnant's want builds where the ground is decayed and eases where it is not.
	"need.undecayed": {
		name: "text.need.undecayed"
		rise: "region.props.decay - 0.3"
		urge: "level"
	}
	// A soul-feeder's hunger; nothing feeds it yet.
	"need.soul": {
		name: "text.need.soul"
		rise: "0.0625"
		urge: "level"
	}
	"need.safety": _#Slow & {name: "text.need.safety"}
	"need.belonging": _#Slow & {name: "text.need.belonging"}
	"need.faith": _#Slow & {name: "text.need.faith"}
	"need.duty": _#Slow & {name: "text.need.duty"}
	"need.grief": _#Slow & {name: "text.need.grief"}
}

text: {
	"text.need.water":     "water"
	"text.need.food":      "food"
	"text.need.rest":      "rest"
	"text.need.hold_dead": "to hold the dead"
	"text.need.undecayed": "undecayed ground"
	"text.need.soul":      "soul"
	"text.need.safety":    "safety"
	"text.need.belonging": "belonging"
	"text.need.faith":     "faith"
	"text.need.duty":      "duty"
	"text.need.grief":     "grief"
}
