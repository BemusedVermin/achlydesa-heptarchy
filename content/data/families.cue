package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

// The kinds of creature a region counts while the party is away. A promoted member is one of the
// family's templates, picked by weight.
families: [string]: content.#FamilyDef

// For tests and the playground. Placeholder weights, for tuning.
families: {
	"family.test.clay": {
		name: "text.family.test.clay"
		templates: [{template: "template.creature.moving_clay", weight: 1}]
	}
	// A town's residents, counted when they leave through a gate or the party is away.
	"family.people": {
		name: "text.family.people"
		templates: [{template: "template.resident.plain", weight: 1}]
		marks: ["house"]
	}
	"family.people.haulers": {
		name: "text.family.people.haulers"
		templates: [{template: "template.resident.hauler", weight: 1}]
		marks: ["house"]
	}
	"family.people.guards": {
		name: "text.family.people.guards"
		templates: [{template: "template.generic.guard", weight: 1}]
		marks: ["house"]
	}
	"family.people.field_hands": {
		name: "text.family.people.field_hands"
		templates: [{template: "template.generic.field_hand", weight: 1}]
		marks: ["house"]
	}
	// The monster families. Drift and attrition are placeholder numbers, for tuning.
	"family.moving_clay": {
		name: "text.family.moving_clay"
		templates: [{template: "template.creature.moving_clay", weight: 1}]
		// Moves toward a neighbour with more dead bodies, by a tenth of the difference, at most half.
		drift: "neighbour.dead > region.dead ? min(0.5, 0.1 * (neighbour.dead - region.dead)) : 0"
	}
	"family.remnant": {
		name: "text.family.remnant"
		templates: [{template: "template.creature.remnant", weight: 1}]
		// Retreats to a neighbour with less decay, by half the difference.
		drift: "neighbour.data.props.decay < region.data.props.decay ? 0.5 * (region.data.props.decay - neighbour.data.props.decay) : 0"
	}
	"family.soul_feeder": {
		name: "text.family.soul_feeder"
		templates: [{template: "template.creature.soul_feeder", weight: 1}]
		// Goes where the living are: toward a neighbour with more people in it.
		drift: "neighbour.counts[\"family.people\"] > region.counts[\"family.people\"] ? 0.2 : 0"
	}
	"family.archon_wardens": {
		name: "text.family.archon_wardens"
		templates: [{template: "template.creature.archon_warden", weight: 1}]
	}
	"family.believers": {
		name: "text.family.believers"
		templates: [{template: "template.resident.believer", weight: 1}]
	}
	"family.test.mixed": {
		name: "text.family.test.mixed"
		templates: [
			{template: "template.creature.moving_clay", weight: 3},
			{template: "template.creature.beast", weight: 1},
		]
		marks: ["lair"]
	}
}

text: {
	"text.family.moving_clay":        "Moving clay"
	"text.family.remnant":            "Remnants"
	"text.family.soul_feeder":        "Soul-feeders"
	"text.family.archon_wardens":     "The archon's wardens"
	"text.family.believers":          "Believers"
	"text.family.people":             "People"
	"text.family.people.haulers":     "Water haulers"
	"text.family.people.guards":      "Guards"
	"text.family.people.field_hands": "Field hands"
	"text.family.test.clay":          "Clay creatures"
	"text.family.test.mixed":         "Mixed creatures"
}
