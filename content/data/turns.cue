package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

turns: content.#TurnsConfig

// Placeholder numbers, for tuning.
turns: {
	turn_ticks:    "speed >= 2 ? 10 : (speed == 1 ? 15 : 25)"
	default_speed: 2
}
