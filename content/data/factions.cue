package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

factions: [string]: content.#Faction

factions: {
	"faction.party": {name: "text.faction.party", hostile_to: ["faction.wild", "faction.crawling"]}
	"faction.wild": {name: "text.faction.wild", hostile_to: ["faction.party", "faction.crawling"]}
	// A severed part that crawls off: hostile to everyone, itself included.
	"faction.crawling": {
		name: "text.faction.crawling"
		hostile_to: ["faction.party", "faction.wild", "faction.crawling"]
	}
}

text: {
	"text.faction.crawling": "the crawling"
	"text.faction.party":    "the party"
	"text.faction.wild":     "the wild"
}
