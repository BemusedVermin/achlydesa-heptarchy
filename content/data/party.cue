package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

party: content.#PartyConfig

// Placeholder numbers, for tuning.
party: {
	slots:        4
	sight_radius: 6
	sight_depth:  6
	// The plan view draws floor out to this many steps of walking from the party; walls show elsewhere.
	plan_range:    12
	max_step_up:   1
	max_step_down: 1
	max_drop:      8
	// Climbing costs more than descending, on top of the step's own ticks.
	climb_cost: "5"
	// An alert puts every member under the player's hand, and gives single file back when it ends.
	individual_on_alert: true
	// A party that is wiped out wakes at its last rest place this many ticks later.
	wake_ticks: 600
	// A follower steps with this action.
	follow_action: "action.step"
}

recruits: content.#RecruitsConfig

// Placeholder numbers, for tuning. A recruit grumbles after two unpaid days and deserts after four;
// from half terror up it runs.
recruits: {
	grumble_at:        2
	desert_at:         4
	flee_at:           0.5
	terror_need:       "need.terror"
	flee_behavior:     "behavior.recruit.flee"
	deserters_faction: "faction.deserters"
}
