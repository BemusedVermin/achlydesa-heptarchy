package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

calendar: content.#Calendar

// Placeholder numbers, for tuning.
calendar: {
	phases: [
		{id: "phase.dawn", name: "text.phase.dawn", ticks: 20},
		{id: "phase.day", name: "text.phase.day", ticks: 100},
		{id: "phase.dusk", name: "text.phase.dusk", ticks: 30},
		{id: "phase.night", name: "text.phase.night", ticks: 90},
	]
}
