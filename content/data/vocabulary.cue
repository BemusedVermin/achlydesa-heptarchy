package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

vocabulary: content.#Vocabulary

// What content closes: anything not declared here stays open.
vocabulary: {
	tags: items: ["consumable", "piece", "keepsake", "name", "drinkable", "edible"]
	grants: ["action", "targeting", "speed"]
	lists: {
		"materials.weak_to":   "types"
		"materials.immune_to": "types"
		"types.consumes":      "items"
	}
}
