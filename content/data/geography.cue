package data

import (
	"list"

	"github.com/bemusedvermin/achlydesa-heptarchy/internal/content"
)

geographies: [string]: content.#GeographyDef

// The stages that carve the deep between the fixed places: solid clay carved by noise into chambers,
// tunnels and crawl soops.
_#ground: [{
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

_#verify: {
	stage: "verify"
	params: {
		max_step_up:   1
		max_step_down: 1
		min_pocket:    24
		// Chunks the check holds at once; first value, for tuning.
		max_chunks: 256
	}
}

// The ground and the guarantees check.
geographies: "geography.deep": stages: list.Concat([_#ground, [_#verify]])

// The ground alone, for the large world: verify takes about a minute on a 400 by 400 world, too long for a
// game to start.
geographies: "geography.large": stages: _#ground

// What starts on the generated ground: the soop's clay near the surface, remnants and beasts deeper. The
// weights, bands and densities are first values, for tuning.
_#spawns: [{
	template:  "template.creature.moving_clay"
	weight:    3
	min_depth: 0
	max_depth: 24
	density:   0.004
}, {
	template:  "template.creature.remnant"
	weight:    1
	min_depth: 12
	max_depth: 63
	density:   0.003
}, {
	template:  "template.creature.beast"
	weight:    2
	min_depth: 20
	max_depth: 63
	density:   0.003
}]

geographies: "geography.deep": spawns:  _#spawns
geographies: "geography.large": spawns: _#spawns
