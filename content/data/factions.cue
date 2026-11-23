package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

factions: [string]: content.#Faction

factions: {
	"faction.party": {name: "text.faction.party", hostile_to: ["faction.deep", "faction.crawling"]}
	"faction.deep": {name: "text.faction.deep", hostile_to: ["faction.party", "faction.crawling"]}
	// A severed part that crawls off: hostile to everyone, itself included.
	"faction.crawling": {
		name: "text.faction.crawling"
		hostile_to: ["faction.party", "faction.deep", "faction.crawling"]
	}
}

text: {
	"text.faction.crawling": "the crawling"
	"text.faction.party": "the party"
	"text.faction.deep":  "the deep"
}
