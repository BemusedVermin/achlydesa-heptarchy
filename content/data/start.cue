package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

start: content.#Start

// The prototype's room, with the main party at its center.
start: {
	place: "place.test.hall"
	mark:  "party_start"
	party: "template.party.main"
}
