package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

// What a body runs low on. A template lists the needs its bodies carry, with a starting level from
// 0 (met) to 1 (desperate); the level rises at every phase boundary. No template carries one yet.
needs: [string]: content.#NeedDef

// Placeholder numbers, for tuning.
needs: {
	"need.water": {
		name: "text.need.water"
		rise: "0.05"
		urge: "level"
	}
	"need.food": {
		name: "text.need.food"
		rise: "0.02"
		urge: "level"
	}
}

text: {
	"text.need.water": "water"
	"text.need.food":  "food"
}
