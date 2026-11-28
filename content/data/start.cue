package data

import "github.com/bemusedvermin/achlydesa-heptarchy/internal/content"

start: content.#Start

// The Reaper's crawlspace, whose ways out lead into the generated deep.
start: {
	place: "settlement.reaper"
	mark:  "party_start"
	party: "template.party.main"
}
