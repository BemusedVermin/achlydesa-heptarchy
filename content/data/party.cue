package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

party: content.#PartyConfig

// Placeholder numbers, for tuning.
party: {
	slots:         4
	sight_radius:  6
	sight_depth:   6
	max_step_up:   1
	max_step_down: 1
	max_drop:      8
	// Climbing costs more than descending, on top of the step's own ticks.
	climb_cost: "5"
}
