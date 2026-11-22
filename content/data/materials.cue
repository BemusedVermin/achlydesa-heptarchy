package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

materials: [string]: content.#Material

// Read from the design table's Types section.
materials: {
	"material.clay": {
		name:      "text.material.clay"
		weak_to:   ["type.soak", "type.relic"]
		immune_to: []
		tags: ["clay"]
	}
	"material.flesh": {
		name:      "text.material.flesh"
		weak_to:   ["type.cut", "type.burn", "type.relic"]
		immune_to: []
		tags: ["flesh"]
	}
	"material.old_matter": {
		name:      "text.material.old_matter"
		weak_to:   ["type.rot"]
		immune_to: ["type.relic"]
		tags: ["old_matter"]
	}
	// Hollow things: skulls, shells.
	"material.bone": {
		name:      "text.material.bone"
		weak_to:   ["type.press"]
		immune_to: []
		tags: ["bone"]
	}
	"material.eye": {
		name:      "text.material.eye"
		weak_to:   ["type.light"]
		immune_to: []
		tags: ["eye"]
	}
	"material.rift_flesh": {
		name:      "text.material.rift_flesh"
		weak_to:   ["type.violet"]
		immune_to: []
		tags: ["rift"]
	}
}

text: {
	"text.material.old_matter": "old matter"
	"text.material.bone":       "bone"
	"text.material.eye":        "eye"
	"text.material.rift_flesh": "rift flesh"
}
