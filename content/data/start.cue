package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

start: content.#Start

// The playground: a hub with a gate to each of the test rooms.
start: {
	map:           "map.play.hub"
	party:         "template.party.main"
	party_at_mark: "party_start"
}
