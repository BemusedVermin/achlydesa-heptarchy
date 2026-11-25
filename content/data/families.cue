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
	"text.family.test.clay":  "Clay creatures"
	"text.family.test.mixed": "Mixed creatures"
}
