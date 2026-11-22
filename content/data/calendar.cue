package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

calendar: content.#Calendar

// Placeholder numbers, for tuning.
calendar: {
	ticks_per_phase: 60
	phases: [
		{id: "phase.dawn", name: "text.phase.dawn"},
		{id: "phase.day", name: "text.phase.day"},
		{id: "phase.dusk", name: "text.phase.dusk"},
		{id: "phase.night", name: "text.phase.night"},
	]
}
