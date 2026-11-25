package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

factions: [string]: content.#Faction

factions: {
	"faction.party": {name: "text.faction.party", hostile_to: ["faction.wild", "faction.crawling", "faction.archon"]}
	"faction.wild": {name: "text.faction.wild", hostile_to: ["faction.party", "faction.crawling"]}
	// Townspeople: hostile to no one by default.
	"faction.people": {name: "text.faction.people"}
	// The archon's creatures: hostile to outsiders (the party) and to no believer.
	"faction.archon": {name: "text.faction.archon", hostile_to: ["faction.party"]}
	"faction.believers": {name: "text.faction.believers"}
	// A severed part that crawls off: hostile to everyone, itself included.
	"faction.crawling": {
		name: "text.faction.crawling"
		hostile_to: ["faction.party", "faction.wild", "faction.crawling"]
	}
}

text: {
	"text.faction.archon":    "the archon's creatures"
	"text.faction.believers": "the believers"
	"text.faction.crawling":  "the crawling"
	"text.faction.party":     "the party"
	"text.faction.people":    "the people"
	"text.faction.wild":      "the wild"
}
