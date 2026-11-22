package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

party: content.#PartyConfig

// Placeholder numbers, for tuning.
party: {
	slots:         4
	sight_radius:  6
	sight_depth:   6
	// A standable cell this close (4 covers the test descent's whole slope) to the party's depth hides the walls in its plan-view hex.
	plan_floor_depth: 4
	max_step_up:   1
	max_step_down: 1
	max_drop:      8
	// Climbing costs more than descending, on top of the step's own ticks.
	climb_cost: "5"
	// An alert puts every member under the player's hand, and gives single file back when it ends.
	individual_on_alert: true
}
