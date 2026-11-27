package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

claims: [string]:  content.#Claim
origins: [string]: content.#Origin
knowledge: content.#KnowledgeConfig
agents:    content.#AgentsConfig

claims: {
	"claim.material": {name: "text.claim.material", value: "material"}
	"claim.name": name: "text.claim.name"
	// A thing is at a cell, written "q,r,depth". It moves, so a newer sighting replaces it.
	"claim.at": name:   "text.claim.at"
	"claim.dead": name: "text.claim.dead"
	// What a landmark gives: its affordance IDs, joined by commas.
	"claim.gives": name: "text.claim.gives"
	// How decayed the ground is at a cell: the geography's decay level there.
	"claim.ground": name: "text.claim.ground"
}

origins: {
	"origin.sweep": name: "text.origin.sweep"
	"origin.rumor": name: "text.origin.rumor"
	"origin.seen": name:  "text.origin.seen"
	"origin.told": name:  "text.origin.told"
}

// First values, for tuning.
knowledge: {
	// A knower outside the party forgets the least sure, then the oldest, past this many facts.
	cap: 64
}

// First values, for tuning.
agents: {
	sight_radius: 8
	sight_depth:  6
	// What an agent records of a body it sees.
	at_claim:        "claim.at"
	dead_claim:      "claim.dead"
	seen_origin:     "origin.seen"
	seen_confidence: 1
	// What an agent records of a landmark it sees, and of the ground.
	gives_claim:  "claim.gives"
	ground_claim: "claim.ground"
	// A landmark believed this many cells away or fewer counts as near, for planning.
	near_radius: 1
}

text: {
	"text.claim.material": "material"
	"text.claim.name":     "name"
	"text.claim.at":       "whereabouts"
	"text.claim.dead":     "death"
	"text.claim.gives":    "what it gives"
	"text.claim.ground":   "how decayed the ground is"
	"text.origin.sweep":   "a sweep"
	"text.origin.rumor":   "a rumor"
	"text.origin.seen":    "sight"
	"text.origin.told":    "being told"
}
