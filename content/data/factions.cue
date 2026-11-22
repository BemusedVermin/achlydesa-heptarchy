package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

factions: [string]: content.#Faction

factions: {
	"faction.party": {name: "text.faction.party", hostile_to: ["faction.deep"]}
	"faction.deep": {name: "text.faction.deep", hostile_to: ["faction.party"]}
}

text: {
	"text.faction.party": "the party"
	"text.faction.deep":  "the deep"
}
