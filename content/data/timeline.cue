package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

timeline: content.#TimelineConfig

// Placeholder numbers, for tuning.
timeline: {
	step_ticks:    "speed >= 2 ? 10 : (speed == 1 ? 15 : 25)"
	default_speed: 2
	window_ticks:  30
}
