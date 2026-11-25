package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

geographies: [string]: content.#GeographyDef

// The deep between the fixed places: solid clay carved by noise into chambers, tunnels and crawl soops.
geographies: "geography.deep": {
	stages: [{
		stage: "caves"
		params: {
			cheese_frequency:    0.05
			cheese_threshold:    0.6
			spaghetti_frequency: 0.04
			spaghetti_width:     0.12
			noodle_frequency:    0.09
			noodle_width:        0.07
			vertical_scale:      1.5
			solid_layers:        1
			fade_layers:         2
			deep_rarity:         0.3
		}
	}, {
		stage: "join"
		params: {
			width: 2
			reach: 24
			bias:  1.5
		}
	}, {
		stage: "sinks"
		params: {
			spacing:      44
			jitter:       0.4
			density:      0.5
			riser_chance: 0.3
			rim_depth:    1
			lap_height:   3
			laps:         2
			headroom:     2
			shaft_radius: 1.5
			outer_radius: 8
		}
	}]
}
