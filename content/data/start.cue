package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

start: content.#Start

// The prototype's room, with the main party at its center.
start: {
	map:           "map.test.hall"
	party:         "template.party.main"
	party_at_mark: "party_start"
}
